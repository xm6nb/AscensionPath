package main

import (
	"AscensionPath/internal/handler"
	"AscensionPath/internal/middleware"
	"AscensionPath/internal/model"
	"embed"
	"flag"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	port := flag.String("port", "8080", "服务器监听的端口号")
	flag.Parse()

	// 1. 初始化配置
	// config.Load()
	gin.SetMode(gin.ReleaseMode)

	// 2. 初始化数据库
	model.InitDB()
	defer model.CloseDB()

	// 3. 创建Gin实例
	r := gin.Default()

	// 4. 注册中间件
	r.Use(middleware.GinLogger(), middleware.GinRecovery(true))

	// 5. 注册路由
	handler.RegisterRoutes(r)

	// 6. 设置嵌入的静态文件服务
	setupEmbeddedStaticFiles(r)

	println("服务器已启动，监听端口：" + *port)
	// 7. 启动服务
	err := r.Run(":" + *port)
	if err != nil {
		panic("无法启动服务器: " + err.Error())
	}
}

//go:embed dist/*
var staticFS embed.FS

func setupEmbeddedStaticFiles(r *gin.Engine) {
	// 从embed.FS中获取子文件系统
	subFS, err := fs.Sub(staticFS, "dist")
	if err != nil {
		panic("无法加载嵌入的静态文件: " + err.Error())
	}

	// 修改静态文件路由前缀，避免与API路由冲突
	r.StaticFS("/static", http.FS(subFS))

	// 修改重定向逻辑，只处理未匹配API路由的请求
	r.NoRoute(func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/static/")
	})
}
