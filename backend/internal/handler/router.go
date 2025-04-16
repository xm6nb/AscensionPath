package handler

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	// 统一使用CORS中间件替代全局OPTIONS处理
	// 修改后的CORS配置
	r.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			// allowedOrigins := []string{
			// 	"http://localhost:3006",
			// 	"http://127.0.0.1:3006",
			// 	"http://localhost:8000",
			// 	"http://127.0.0.1:8000",
			// }
			// for _, o := range allowedOrigins {
			// 	if origin == o {
			// 		return true
			// 	}
			// }
			// return false
			return true
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "Upgrade", "Connection"}, // 新增WebSocket头
		ExposeHeaders:    []string{"Content-Length", "Sec-WebSocket-Accept"},                 // 暴露WebSocket头
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 健康检查路由
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// API v1 版本路由组
	v1 := r.Group("/api/v1")
	{
		// 用户相关路由
		userGroup := v1.Group("/users")
		{
			userGroup.POST("/register", registerUser)
			userGroup.POST("/login", loginUser)

			// 认证路由组使用明确路径
			authGroup := userGroup.Group("")
			authGroup.Use(authMiddleware())
			{
				authGroup.POST("/profile", updateProfile)
				authGroup.POST("/updatePassword", changePassword)
				authGroup.GET("/getUserInfo", getUserByID)
				adminGroup := authGroup.Group("")
				adminGroup.Use(isAdmin())
				{
					adminGroup.POST("/deleteUser", deleteUser)
					adminGroup.GET("/getAllUsers", getAllUsers)
					adminGroup.POST("/addUser", addUser)
					adminGroup.POST("/searchUsers", searchUsers)
				}
			}
		}

		// 漏洞相关路由
		vulGroup := v1.Group("/vul")
		vulGroup.Use(authMiddleware()).Use(IsDockerAvailable())
		{
			vulGroup.POST("/createVulInstance", CreateVulInstance)
			vulGroup.POST("/removeInstance", RemoveInstance)
			vulGroup.GET("/extendExpireTime", ExtendExpireTime)
			vulGroup.GET("/getCreatedVulEnv", GetCreatedVulEnv) // 获取所有创建的漏洞环境以及开启的场景

			adminGroup := vulGroup.Group("")
			adminGroup.Use(isAdmin())
			{
				adminGroup.GET("/getAllInstance", GetAllInstance)
				adminGroup.GET("/getVulImages", GetVulImages)
				adminGroup.GET("/getImageLoadConfig", GetImageLoadConfig)
				adminGroup.POST("/uploadImageFile", UploadImageFile)
				adminGroup.GET("/pullImage", PullImage)
				adminGroup.GET("/getVulEnv", GetVulEnv) // 获取镜像和compose信息
				adminGroup.POST("/uploadVulZip", UploadVulZip)
				adminGroup.GET("/createVulEnv", CreateVulEnv)
				adminGroup.POST("/deleteVulEnv", DeleteVulEnv)
			}
		}
	}
}
