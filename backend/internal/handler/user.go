package handler

import (
	"AscensionPath/internal/model"
	"AscensionPath/internal/service"
	"AscensionPath/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 新增注册handler函数
func registerUser(c *gin.Context) {
	var reqMessage utils.Message[struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
	}]

	// 绑定并验证请求参数
	if err := c.ShouldBindJSON(&reqMessage); err != nil {
		c.JSON(http.StatusBadRequest, utils.FailResult(utils.CodeBadRequest, "无效的请求参数: "+err.Error()))
		return
	}

	// 从Message中提取实际请求数据
	req := reqMessage.Data

	// 调用service层注册方法
	userService := &service.UserService{}
	user, err := userService.Register(req.Username, req.Password, req.Email)
	if err != nil {
		statusCode := utils.CodeInternalError
		if err == utils.ErrUserAlreadyExists {
			statusCode = http.StatusConflict
		}

		c.JSON(statusCode, utils.FailResult(statusCode, err.Error()))
		return
	}

	// 返回注册成功的用户信息(不包含敏感信息)
	responseData := gin.H{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
		"role":     user.Role,
	}
	c.JSON(http.StatusCreated, utils.SuccessResult(responseData))
}

// loginUser 用户登录
func loginUser(c *gin.Context) {
	var reqMessage utils.Message[struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}]
	// 绑定并验证请求参数
	if err := c.ShouldBindJSON(&reqMessage); err != nil {
		c.JSON(http.StatusBadRequest, utils.FailResult(utils.CodeBadRequest, "无效的请求参数: "+err.Error()))
		return
	}

	// 从Message中提取实际请求数据
	req := reqMessage.Data

	// 调用service层登录方法
	userService := &service.UserService{}
	user, err := userService.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, utils.FailResult(utils.CodeUnauthorized, "登录失败: "+err.Error()))
		return
	}
	// 生成token或其他认证信息
	token, err := generateToken(user.ID, user.Username, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.FailResult(utils.CodeInternalError, "生成token失败: "+err.Error()))
		return
	}

	// 设置cookie，有效期18小时
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		"auth_token", // cookie名称
		token,        // token值
		18*60*60,     // 过期时间(秒)
		"/",          // 路径
		"",           // 域名
		false,        // 仅HTTPS
		true,         // 仅HTTP访问
	)

	// 返回登录成功的用户信息和token
	responseData := gin.H{
		"id":       user.ID,
		"username": user.Username,
		"role":     user.Role,
		"token":    token,
		"email":    user.Email,
		"status":   user.Status,
	}
	c.JSON(http.StatusOK, utils.SuccessResult(responseData))
}

// updateProfile 更新用户资料
func updateProfile(c *gin.Context) {
	var reqMessage utils.Message[struct {
		ID       uint    `json:"id" binding:"required"`
		Email    string  `json:"email"`
		Status   int     `json:"status"`
		Role     string  `json:"role"`
		Username string  `json:"username"`
		Password string  `json:"password"`
		Score    float64 `json:"score"` // 新增的分数字段，用于更新用户的积分
	}]

	if err := c.ShouldBindJSON(&reqMessage); err != nil {
		c.JSON(http.StatusBadRequest, utils.FailResult(utils.CodeBadRequest, "无效的请求参数: "+err.Error()))
		return
	}

	req := reqMessage.Data

	userService, err := utils.GetDataFromContext(c, "UserInfo", &service.UserService{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.FailResult(utils.CodeInternalError, err.Error()))
		return
	}

	// 用户不能禁用自己
	if userService.ID == req.ID && req.Status == 0 {
		c.JSON(http.StatusInternalServerError, utils.FailResult(utils.CodeInternalError, "用户不能禁用自己的账号"))
		return
	}

	if err := userService.UpdateProfile(req.ID, req.Email, req.Status, req.Role, req.Score, req.Username, req.Password); err != nil {
		c.JSON(http.StatusForbidden, utils.FailResult(http.StatusForbidden, err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResult[interface{}](nil))
}

// changePassword 修改密码
func changePassword(c *gin.Context) {
	var reqMessage utils.Message[struct {
		UserID      uint   `json:"user_id" binding:"required"`
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password" binding:"required"`
	}]

	if err := c.ShouldBindJSON(&reqMessage); err != nil {
		c.JSON(http.StatusBadRequest, utils.FailResult(utils.CodeBadRequest, "无效的请求参数: "+err.Error()))
		return
	}

	req := reqMessage.Data
	userService, err := utils.GetDataFromContext(c, "UserInfo", &service.UserService{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.FailResult(utils.CodeInternalError, err.Error()))
		return
	}

	if err := userService.ChangePassword(req.UserID, req.OldPassword, req.NewPassword); err != nil {
		statusCode := http.StatusForbidden
		if err == utils.ErrInvalidCredentials {
			statusCode = http.StatusUnauthorized
		}
		c.JSON(statusCode, utils.FailResult(statusCode, err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResult[interface{}](nil))
}

// addUser 添加用户
func addUser(c *gin.Context) {
	var reqMessage utils.Message[struct {
		Username string  `json:"username" binding:"required"`
		Password string  `json:"password" binding:"required"`
		Email    string  `json:"email" binding:"required,email"`
		Role     string  `json:"role" binding:"required"`
		Status   int     `json:"status"`
		Score    float64 `json:"score"`
	}]
	if err := c.ShouldBindJSON(&reqMessage); err != nil {
		c.JSON(http.StatusBadRequest, utils.FailResult(utils.CodeBadRequest, "无效的请求参数: "+err.Error()))
		return
	}
	req := reqMessage.Data
	userService, err := utils.GetDataFromContext(c, "UserInfo", &service.UserService{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.FailResult(utils.CodeInternalError, err.Error()))
		return
	}
	user := model.User{
		Username: req.Username,
		Password: req.Password,
		Email:    req.Email,
		Role:     req.Role,
		Status:   req.Status,
		Score:    req.Score,
	}
	if err := userService.AddUser(&user); err != nil {
		c.JSON(http.StatusForbidden, utils.FailResult(http.StatusForbidden, err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessResult[interface{}](nil))
}

// deleteUser 删除用户
func deleteUser(c *gin.Context) {
	var reqMessage utils.Message[struct {
		UserID uint `json:"id" binding:"required"`
	}]

	if err := c.ShouldBindJSON(&reqMessage); err != nil {
		c.JSON(http.StatusBadRequest, utils.FailResult(utils.CodeBadRequest, "无效的请求参数: "+err.Error()))
		return
	}

	req := reqMessage.Data
	userService, err := utils.GetDataFromContext(c, "UserInfo", &service.UserService{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.FailResult(utils.CodeInternalError, err.Error()))
		return
	}

	if err := userService.DeleteAccount(req.UserID); err != nil {
		c.JSON(http.StatusForbidden, utils.FailResult(http.StatusForbidden, err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResult[interface{}](nil))
}

// updateUserRole 更新用户角色
func updateUserRole(c *gin.Context) {
	var reqMessage utils.Message[struct {
		UserID uint   `json:"user_id" binding:"required"`
		Role   string `json:"role" binding:"required"`
	}]

	if err := c.ShouldBindJSON(&reqMessage); err != nil {
		c.JSON(http.StatusBadRequest, utils.FailResult(utils.CodeBadRequest, "无效的请求参数: "+err.Error()))
		return
	}

	req := reqMessage.Data
	userService, err := utils.GetDataFromContext(c, "UserInfo", &service.UserService{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.FailResult(utils.CodeInternalError, err.Error()))
		return
	}

	if err := userService.UpdateUserRole(req.UserID, req.Role); err != nil {
		c.JSON(http.StatusForbidden, utils.FailResult(http.StatusForbidden, err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResult[interface{}](nil))
}

// getAllUsers 获取用户列表
func getAllUsers(c *gin.Context) {
	page, err := utils.StringToInt(c.DefaultQuery("page", "1"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.FailResult(utils.CodeBadRequest, "无效的页码: "+err.Error()))
		return
	}
	pageSize, err := utils.StringToInt(c.DefaultQuery("pageSize", "10"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.FailResult(utils.CodeBadRequest, "无效的每页数量: "+err.Error()))
		return
	}
	userService, err := utils.GetDataFromContext(c, "UserInfo", &service.UserService{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.FailResult(utils.CodeInternalError, err.Error()))
		return
	}

	users, err := userService.GetAllUsers(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.FailResult(utils.CodeInternalError, err.Error()))
		return
	}

	usercount, err := userService.GetUserCount()
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.FailResult(utils.CodeInternalError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResult(struct {
		Users []service.UserDTO `json:"users"`
		Count int64             `json:"count"`
	}{Users: users, Count: usercount}))
}

// getUserByID 获取单个用户信息
func getUserByID(c *gin.Context) {
	userID, err := utils.StringToInt(c.Query("id")) // 获取URL中的用户ID参数
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.FailResult(utils.CodeBadRequest, "无效的用户ID: "+err.Error()))
		return
	}

	userService, err := utils.GetDataFromContext(c, "UserInfo", &service.UserService{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.FailResult(utils.CodeInternalError, err.Error()))
		return
	}

	user, err := userService.GetUserByID(uint(userID))
	if err != nil {
		c.JSON(http.StatusNotFound, utils.FailResult(http.StatusNotFound, "用户不存在"))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessResult(user))
}

// searchUsers 搜索用户
func searchUsers(c *gin.Context) {
	var reqMessage utils.Message[struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Role     string `json:"role"`
		Status   int    `json:"status"`
		Page     int    `json:"page"`
		PageSize int    `json:"pageSize"`
	}]
	if err := c.ShouldBindJSON(&reqMessage); err != nil {
		c.JSON(http.StatusBadRequest, utils.FailResult(utils.CodeBadRequest, "无效的请求参数: "+err.Error()))
		return
	}
	req := reqMessage.Data
	userService, err := utils.GetDataFromContext(c, "UserInfo", &service.UserService{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.FailResult(utils.CodeInternalError, err.Error()))
		return
	}
	users, err := userService.SearchUsers(req.Username, req.Email, req.Status, req.Role, req.Page, req.PageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.FailResult(utils.CodeInternalError, err.Error()))
		return
	}
	// 获取所有符合条件的用户数量
	count, err := userService.SearchUsersCount(req.Username, req.Email, req.Status, req.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.FailResult(utils.CodeInternalError, err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessResult(struct {
		Users []service.UserDTO `json:"users"`
		Count int64             `json:"count"` // 添加用户数量的返回
	}{Users: users, Count: count}))
}
