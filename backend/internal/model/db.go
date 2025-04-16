package model

import (
	"AscensionPath/internal/middleware"
	"context"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() {
	// 初始化数据库连接
	dsn := "gorm.db"

	// 创建自定义GORM日志器
	gormLogger := NewGormLogger()

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})

	if err != nil {
		panic("failed to connect database: " + err.Error())
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		panic("get underlying db error: " + err.Error())
	}

	sqlDB.SetMaxIdleConns(10)           // 最大空闲连接数
	sqlDB.SetMaxOpenConns(100)          // 最大打开连接数
	sqlDB.SetConnMaxLifetime(time.Hour) // 连接最大存活时间

	// 自动迁移表结构
	err = db.AutoMigrate(&User{}, &VulEnv{}, &VulInstance{})
	if err != nil {
		panic("auto migrate error: " + err.Error())
	}

	DB = db

	// 初始化管理员账号
	count, err := GetUserCount()
	if err != nil {
		panic("初始化数据库时获取用户数量出现错误: " + err.Error())
	}

	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	if err != nil {
		panic("初始化时密码加密失败: " + err.Error())
	}

	if count == 0 {
		user := User{
			Username: "admin",
			Password: string(hashedPassword),
			Status:   1,
			Role:     "admin",
		}
		CreateUser(&user)
	}
}

// 添加关闭数据库连接的函数
func CloseDB() error {
	if DB == nil {
		return nil
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	// 设置连接最大空闲时间为0（立即关闭）
	sqlDB.SetConnMaxLifetime(0)
	return sqlDB.Close()
}

// GormLogger 自定义GORM日志器
type GormLogger struct {
	ZapLogger *zap.SugaredLogger
}

func NewGormLogger() *GormLogger {
	return &GormLogger{
		ZapLogger: middleware.SugarLogger,
	}
}

// 实现gorm/logger.Interface
func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	return l // 使用全局日志级别
}

func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	l.ZapLogger.Infow(msg, "gorm", data)
}

func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.ZapLogger.Warnw(msg, "gorm", data)
}

func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	l.ZapLogger.Errorw(msg, "gorm", data)
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()

	l.ZapLogger.Debugw("SQL Query",
		"sql", sql,
		"rows", rows,
		"time", elapsed,
		"error", err,
	)
}
