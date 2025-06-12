package handler

import (
	"AscensionPath/internal/middleware"
	"AscensionPath/internal/model"
	"AscensionPath/internal/service"
	"AscensionPath/internal/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// 判断当前docker是否可用
func IsDockerAvailable() gin.HandlerFunc {
	return func(c *gin.Context) {
		available, err := service.IsDockerAvailable()
		if !available {
			c.AbortWithStatusJSON(http.StatusInternalServerError, utils.FailResult(utils.CodeInternalError, "docker服务不可用:"+err.Error()))
			return
		}
		c.Next()
	}
}

// 获取所有漏洞镜像信息
func GetVulImages(c *gin.Context) {
	vul := service.VulService{}
	vulImagesList, err := vul.GetVulImages()
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.FailResult(utils.CodeBadRequest, err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessResult(vulImagesList))
}

type ImageLoadConfig struct {
	LoadPath     string `json:"load_path"`     // 镜像加载路径
	DockerHealth bool   `json:"docker_health"` // Docker 健康状态
}

// 获取镜像加载配置
func GetImageLoadConfig(c *gin.Context) {
	var err error
	vulService := service.VulService{}
	config := ImageLoadConfig{}
	config.LoadPath = vulService.GetVulStoragePath()
	config.DockerHealth, err = service.IsDockerAvailable()
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.FailResult(utils.CodeInternalError, err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessResult(config))
}

// 上传JSON文件到镜像仓库
func UploadImageFile(c *gin.Context) {
	var req utils.Message[utils.UploadImageRequest]

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.FailResult(utils.CodeBadRequest, "请求参数错误"))
		return
	}

	userService, err := utils.GetDataFromContext(c, "UserInfo", &service.UserService{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.FailResult(utils.CodeInternalError, err.Error()))
		return
	}

	middleware.SugarLogger.Infof("用户: %s 请求上传文件: %s", userService.Username, req.Data.Filename)

	vul := service.VulService{}
	if err := vul.SaveUploadedJsonFile(req.Data.Filename, req.Data.Base64FileData); err != nil {
		c.JSON(http.StatusInternalServerError, utils.FailResult(utils.CodeBadRequest, err.Error()))
		middleware.SugarLogger.Errorf("用户: %s 上传文件 %s 失败: %s", userService.Username, req.Data.Filename, err.Error())
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResult(""))
}

// 获取所有漏洞环境信息
func GetVulEnv(c *gin.Context) {
	vul := service.VulService{}
	vulEnvList, err := vul.GetVulEnv()
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.FailResult(utils.CodeBadRequest, err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessResult(vulEnvList))
}

// 上传docker compose 压缩包
func UploadVulZip(c *gin.Context) {
	var req utils.Message[utils.UploadImageRequest]

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.FailResult(utils.CodeBadRequest, "请求参数错误"))
		return
	}

	userService, err := utils.GetDataFromContext(c, "UserInfo", &service.UserService{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.FailResult(utils.CodeInternalError, err.Error()))
		return
	}

	middleware.SugarLogger.Infof("用户: %s 请求上传文件: %s", userService.Username, req.Data.Filename)
	vul := service.VulService{}
	if err := vul.UploadVulZip(req.Data.Base64FileData); err != nil {
		c.JSON(http.StatusInternalServerError, utils.FailResult(utils.CodeBadRequest, err.Error()))
		middleware.SugarLogger.Errorf("用户: %s 上传文件 %s 失败: %s", userService.Username, req.Data.Filename, err.Error())
		return
	}
	middleware.SugarLogger.Infof("用户: %s 上传文件 %s 成功", userService.Username, req.Data.Filename)
	c.JSON(http.StatusOK, utils.SuccessResult(""))
}

// 创建漏洞环境
func CreateVulEnv(c *gin.Context) {
	// 先升级WebSocket连接
	conn, err := utils.UpgradeToWebSocket(c.Writer, c.Request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.FailResult(utils.CodeInternalError, "WebSocket创建失败"))
		return
	}
	defer conn.Close()

	var req utils.Message[service.VulEnv]

	// 从WebSocket连接读取消息
	if err = conn.ReadJSON(&req); err != nil { // 替换原ReadMessage和ShouldBindJSON
		utils.SendError(conn, utils.CodeInternalError, err.Error())
		return
	}

	userService, err := utils.GetDataFromContext(c, "UserInfo", &service.UserService{})
	if err != nil {
		utils.SendError(conn, utils.CodeInternalError, err.Error())
		return
	}
	middleware.SugarLogger.Infof("用户: %s 请求创建环境: %s", userService.Username, req.Data.EnvName)
	vul := service.VulService{}
	if err := vul.CreateVulEnv(&req.Data, conn); err != nil {
		utils.SendError(conn, utils.CodeInternalError, err.Error())
		middleware.SugarLogger.Errorf("用户: %s 创建环境 %s 失败: %s", userService.Username, req.Data.EnvName, err.Error())
		return
	}
	utils.SendWSMessage(conn, utils.CodeSuccess, "漏洞环境创建成功", nil)
	middleware.SugarLogger.Infof("用户: %s 创建环境 %s 成功", userService.Username, req.Data.EnvName)
}

// 删除创建的漏洞环境
func DeleteVulEnv(c *gin.Context) {
	var req utils.Message[struct {
		VulEnvID      uint `json:"vul_env_id"`
		IsDeleteImage bool `json:"is_delete_image"`
	}]

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.FailResult(utils.CodeBadRequest, "请求参数错误"))
		return
	}

	userService, err := utils.GetDataFromContext(c, "UserInfo", &service.UserService{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.FailResult(utils.CodeInternalError, err.Error()))
		return
	}

	middleware.SugarLogger.Infof("用户: %s 请求删除环境: %d", userService.Username, req.Data.VulEnvID)
	vul := service.VulService{}
	if err := vul.DeleteVulEnv(req.Data.VulEnvID, req.Data.IsDeleteImage); err != nil {
		c.JSON(http.StatusInternalServerError, utils.FailResult(utils.CodeInternalError, err.Error()))
		middleware.SugarLogger.Errorf("用户: %s 删除环境 %d 失败: %s", userService.Username, req.Data.VulEnvID, err.Error())
		return
	}
	middleware.SugarLogger.Infof("用户: %s 删除环境 %d 成功", userService.Username, req.Data.VulEnvID)
	c.JSON(http.StatusOK, utils.SuccessResult(""))
}

// 获取所有对当前用户身份可见的已经创建的漏洞环境以及开启的场景
func GetCreatedVulEnv(c *gin.Context) {
	userService, err := utils.GetDataFromContext(c, "UserInfo", &service.UserService{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.FailResult(utils.CodeInternalError, err.Error()))
		return
	}

	result := service.VulInstanceList{}
	vul := service.VulService{}
	// 获取所有创建的漏洞环境
	vulEnvList, err := vul.GetVulEnvList(service.RoleMap[userService.Role])
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.FailResult(utils.CodeBadRequest, err.Error()))
		return
	}

	// 获取所有漏洞实例
	vulInstanceList, err := vul.GetVulInstanceByUserID(userService.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.FailResult(utils.CodeBadRequest, err.Error()))
		return
	}

	// 构建返回结果
	for _, vulEnv := range vulEnvList {
		instance := service.VulInstanceService{}
		instance.VulEnvID = vulEnv.ID
		instance.UserID = userService.ID
		instance.Status = 0
		instance.VulEnv = vulEnv
		for _, vulInstance := range vulInstanceList {
			if vulEnv.ID == vulInstance.VulEnvID {
				// instance.ContainerID = vulInstance.ContainerID
				// instance.StackName = vulInstance.StackName // 没有必要返回
				instance.ID = vulInstance.ID
				instance.Status = vulInstance.Status
				instance.Ports = vulInstance.Ports
				instance.StartTime = vulInstance.StartTime
				instance.ExpireTime = vulInstance.ExpireTime
				instance.EndTime = vulInstance.EndTime
			}
		}
		result = append(result, instance)
	}

	// 构建返回结果
	c.JSON(http.StatusOK, utils.SuccessResult(result))
}

// 创建场景实例
func CreateVulInstance(c *gin.Context) {
	var req utils.Message[struct {
		EnvName  string `json:"env_name"`
		VulEnvID uint   `json:"vul_env_id"`
	}]

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.FailResult(utils.CodeBadRequest, "请求参数错误"))
		return
	}

	userService, err := utils.GetDataFromContext(c, "UserInfo", &service.UserService{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.FailResult(utils.CodeInternalError, err.Error()))
		return
	}
	middleware.SugarLogger.Infof("用户: %s 请求创建场景: %s", userService.Username, req.Data.EnvName)
	vul := service.VulService{}
	instance, err := vul.CreateVulInstance(userService.ID, req.Data.VulEnvID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.FailResult(utils.CodeInternalError, err.Error()))
		middleware.SugarLogger.Errorf("用户: %s 创建场景 %s 失败: %s", userService.Username, req.Data.EnvName, err.Error())
		return
	}
	c.JSON(http.StatusOK, utils.SuccessResult(instance))
}

// 停止并移除实例
func RemoveInstance(c *gin.Context) {
	var req utils.Message[struct {
		UserID   uint `json:"user_id"`
		VulEnvID uint `json:"vul_env_id"`
	}]

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.FailResult(utils.CodeBadRequest, "请求参数错误"))
		return
	}

	userService, err := utils.GetDataFromContext(c, "UserInfo", &service.UserService{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.FailResult(utils.CodeInternalError, err.Error()))
		return
	}

	if userService.Role != "admin" && req.Data.UserID != userService.ID {
		c.JSON(http.StatusInternalServerError, utils.FailResult(utils.CodeInternalError, "无权删除该实例"))
		return
	}

	vul := service.VulService{}
	if err := vul.DeleteVulInstance(req.Data.UserID, req.Data.VulEnvID); err != nil {
		c.JSON(http.StatusInternalServerError, utils.FailResult(utils.CodeInternalError, err.Error()))
		middleware.SugarLogger.Errorf("用户: %s 停止实例失败: %s", userService.Username, err.Error())
		return
	}

	middleware.SugarLogger.Infof("用户: %s 成功停止实例", userService.Username)
	c.JSON(http.StatusOK, utils.SuccessResult("实例已停止并移除"))
}

// 获取所有漏洞场景实例
func GetAllInstance(c *gin.Context) {
	userService, err := utils.GetDataFromContext(c, "UserInfo", &service.UserService{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.FailResult(utils.CodeInternalError, err.Error()))
		return
	}
	if userService.Role != "admin" {
		c.JSON(http.StatusInternalServerError, utils.FailResult(utils.CodeInternalError, "未授权获取漏洞实例"))
		return
	}
	vul := service.VulService{}
	vulInstanceList, err := vul.GetAllVulInstances()
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.FailResult(utils.CodeBadRequest, err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessResult(vulInstanceList))
}

// 延长实例过期时间
func ExtendExpireTime(c *gin.Context) {
	// 从URL参数获取实例ID
	idStr := c.Query("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, utils.FailResult(utils.CodeBadRequest, "缺少实例ID参数"))
		return
	}

	// 转换ID为uint类型
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.FailResult(utils.CodeBadRequest, "无效的实例ID"))
		return
	}

	// 获取用户信息
	userService, err := utils.GetDataFromContext(c, "UserInfo", &service.UserService{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.FailResult(utils.CodeInternalError, err.Error()))
		return
	}

	// 检查实例是否存在且属于该用户
	instance, err := model.GetVulInstanceByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, utils.FailResult(utils.CodeInternalError, "实例不存在"))
		return
	}

	// 非管理员只能操作自己的实例
	if userService.Role != "admin" && instance.UserID != userService.ID {
		c.JSON(http.StatusForbidden, utils.FailResult(utils.CodeInternalError, "无权操作该实例"))
		return
	}

	// 调用服务层方法延长过期时间
	v := service.VulService{}
	if err := v.ExtendExpireTime(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, utils.FailResult(utils.CodeInternalError, err.Error()))
		middleware.SugarLogger.Errorf("用户: %s 延长实例ID: %d 过期时间失败: %s", userService.Username, id, err.Error())
		return
	}

	middleware.SugarLogger.Infof("用户: %s 成功延长实例ID: %d 过期时间", userService.Username, id)
	c.JSON(http.StatusOK, utils.SuccessResult("实例过期时间已延长"))
}
