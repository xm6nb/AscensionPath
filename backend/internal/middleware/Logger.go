package middleware

import (
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var SugarLogger *zap.SugaredLogger

func init() {
	// 配置日志轮转
	lumberjackLogger := &lumberjack.Logger{
		Filename:   "./log/app.log",
		MaxSize:    500,
		MaxBackups: 10,
		MaxAge:     90,
		Compress:   true,
	}

	// 确保文件写入使用UTF-8编码
	fileWriter := &writerWithUTF8{writer: lumberjackLogger}

	// 文件输出
	// 文件输出的EncoderConfig也做同样修改
	fileEncoder := zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		TimeKey:        "T",
		LevelKey:       "L",
		NameKey:        "N",
		CallerKey:      "C",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "M",
		StacktraceKey:  "S",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     customTimeEncoder, // 使用相同的时间格式
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	})

	// 控制台输出
	// 修改控制台输出的EncoderConfig
	consoleEncoder := zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     customTimeEncoder, // 使用自定义时间格式
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	})

	// 创建多输出核心
	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zapcore.InfoLevel),
		zapcore.NewCore(fileEncoder, fileWriter, zapcore.InfoLevel),
	)

	logger := zap.New(core, zap.AddCaller())
	SugarLogger = logger.Sugar()
}

// 添加自定义时间格式化函数
func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

// 修改writerWithUTF8定义
type writerWithUTF8 struct {
	writer *lumberjack.Logger
}

func (w *writerWithUTF8) Write(p []byte) (n int, err error) {
	// 移除BOM头写入逻辑
	return w.writer.Write(p)
}

func (w *writerWithUTF8) Sync() error {
	return nil
}

// 修改 GinLogger 以输出格式化字符串
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		c.Next()

		cost := time.Since(start)
		// 修改错误处理逻辑
		errors := ""
		if c.Writer.Status() >= 400 { // 当状态码为4xx或5xx时记录错误
			if len(c.Errors) > 0 {
				errors = c.Errors.String()
			} else {
				// 如果没有c.Errors，尝试从响应体中获取错误信息
				if c.Writer.Status() >= 500 {
					errors = "Internal Server Error"
				} else {
					errors = http.StatusText(c.Writer.Status())
				}
			}
		}
		SugarLogger.Infof("[%s] %s | %d | %s | ip=%s | ua=%s | errors=%s | cost=%s",
			c.Request.Method,
			path,
			c.Writer.Status(),
			query,
			c.ClientIP(),
			c.Request.UserAgent(),
			errors,
			cost,
		)
	}
}

// GinRecovery recover掉项目可能出现的panic
func GinRecovery(stack bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 检查是否断开的连接
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				if brokenPipe {
					SugarLogger.Error(c.Request.URL.Path,
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
					c.Error(err.(error))
					c.Abort()
					return
				}

				if stack {
					SugarLogger.Error("[Recovery from panic]",
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
						zap.String("stack", string(debug.Stack())),
					)
				} else {
					SugarLogger.Error("[Recovery from panic]",
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
				}
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}
