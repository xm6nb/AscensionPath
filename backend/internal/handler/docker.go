package handler

import (
	"context"
	"net/http"

	"AscensionPath/internal/middleware"
	"AscensionPath/internal/service"
	"AscensionPath/internal/utils"

	"github.com/gin-gonic/gin"
)

// PullImage 拉取Docker镜像
func PullImage(c *gin.Context) {
	// 先升级WebSocket连接
	conn, err := utils.UpgradeToWebSocket(c.Writer, c.Request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.FailResult(utils.CodeInternalError, "WebSocket创建失败"))
		return
	}
	defer conn.Close()

	// 再获取请求参数
	imageName := c.Query("image")
	if imageName == "" {
		utils.SendError(conn, utils.CodeBadRequest, "镜像名称不能为空") // 改用WebSocket发送错误
		return
	}

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go service.MonitorStopOperation(conn, cancel)

	// 调用service层的PullImage函数
	err = service.PullImage(ctx, imageName, conn)
	if err != nil {
		utils.SendError(conn, utils.CodeInternalError, err.Error())
		return
	}

	// 发送最终完成消息
	utils.SendWSMessage(conn, utils.CodeSuccess, "镜像拉取完成", nil)
	middleware.SugarLogger.Infof("用户: %s 拉取镜像 %s 成功", "admin", imageName)
}
