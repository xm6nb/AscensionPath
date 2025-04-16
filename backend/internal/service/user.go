package service

import (
	"errors"
	"time"

	"AscensionPath/internal/middleware"
	"AscensionPath/internal/model"
	"AscensionPath/internal/utils"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	UserDTO
}

// 在service层添加角色校验
const (
	RoleUser  = "user"
	RoleAdmin = "admin"
	RoleVip   = "vip"
)

var RoleMap = map[string]int{
	RoleAdmin: 0,
	RoleVip:   500,
	RoleUser:  999,
}

// 转换前端需要的DTO对象
type UserDTO struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Status    int       `json:"status"`
	Role      string    `json:"role"`
	Score     float64   `json:"score"`
	LastLogin time.Time `json:"last_login"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func invertUserDTO(user model.User) UserDTO {
	return UserDTO{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Status:    user.Status,
		Role:      user.Role,
		Score:     user.Score,
		LastLogin: user.LastLogin,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func invertUserDTOList(users []model.User) []UserDTO {
	var userDTOs []UserDTO
	for _, user := range users {
		userDTOs = append(userDTOs, invertUserDTO(user))
	}
	return userDTOs
}

// 判断非法角色
func IsValidRole(role string) bool {
	switch role {
	case RoleUser, RoleAdmin, RoleVip:
		return true
	default:
		return false
	}
}

// Register 用户注册
func (s *UserService) Register(username, password, email string) (*model.User, error) {
	// 检查用户是否存在
	if exists, _ := model.UserExists(username, email); exists {
		return nil, utils.ErrUserAlreadyExists
	}

	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username: username,
		Password: string(hashedPassword),
		Email:    email,
		Status:   0,        // 默认状态为禁用
		Role:     RoleUser, // 设置默认角色
		Score:    0,        // 初始分数
	}

	if err := model.CreateUser(user); err != nil {
		middleware.SugarLogger.Errorw("创建用户失败",
			"username", username,
			"error", err.Error(),
		)
		return nil, err
	}

	middleware.SugarLogger.Infow("用户注册成功",
		"userID", user.ID,
		"username", username,
	)
	return user, nil
}

// Login 用户登录
func (s *UserService) Login(username, password string) (*model.User, error) {
	middleware.SugarLogger.Infow("登录尝试",
		"username", username,
	)

	user, err := model.GetUserByUsername(username)
	if err != nil {
		middleware.SugarLogger.Warnw("用户不存在",
			"username", username,
			"error", err.Error(),
		)
		return nil, utils.ErrInvalidCredentials
	}
	// 判断密码正确性
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		middleware.SugarLogger.Warnw("密码验证失败",
			"userID", user.ID,
			"username", username,
		)
		return nil, utils.ErrInvalidCredentials
	}

	// 检查用户状态
	if user.Status == 0 {
		middleware.SugarLogger.Warnw("用户已被禁用",
			"userID", user.ID,
			"username", username,
		)
		return nil, utils.ErrInvalidCredentials
	}

	middleware.SugarLogger.Infow("登录成功",
		"userID", user.ID,
		"username", username,
	)

	// 仅更新最后登录时间字段
	if err := model.UpdateUser(user.ID, map[string]interface{}{
		"last_login": time.Now(),
	}); err != nil {
		middleware.SugarLogger.Errorw("更新最后登录时间失败",
			"userID", user.ID,
			"error", err,
		)
	}
	return user, nil
}

// UpdateProfile 更新用户资料（添加权限验证）
func (s *UserService) UpdateProfile(id uint, email string, status int, role string, score float64, username string, password string) error {
	logFields := []interface{}{
		"operatorID", s.ID,
		"targetUserID", id,
		"operatorRole", s.Role,
		"newEmail", email,
		"newStatus", status,
		"newRole", role,
		"newScore", score,
	}

	middleware.SugarLogger.Infow("用户资料更新请求", logFields...)

	// 权限验证
	if s.Role != RoleAdmin {
		// 普通用户只能修改自己的资料
		if s.ID != id {
			middleware.SugarLogger.Warnw("越权修改尝试", logFields...)
			return errors.New("无权修改其他用户资料")
		}

		// 普通用户不能修改status、role和score字段
		if status != -1 && status != 1 {
			return errors.New("无效的状态值")
		}
	}

	// 构建更新字段
	updates := make(map[string]interface{})
	if email != "" {
		updates["email"] = email
	}
	if username != "" {
		updates["username"] = username
	}

	if s.Role == RoleAdmin {
		if status != -1 {
			updates["status"] = status
		}
		if role != "" && IsValidRole(role) {
			updates["role"] = role
		}
		if score != -1 {
			updates["score"] = score
		}
	}

	// 密码修改单独处理
	if password != "" {
		if s.Role != RoleAdmin && s.ID != id {
			return errors.New("无权修改其他用户密码")
		}
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		updates["password"] = string(hashedPassword)
	}

	// 如果没有可更新字段
	if len(updates) == 0 {
		middleware.SugarLogger.Warnw("无有效更新字段", logFields...)
		return errors.New("没有需要更新的字段")
	}

	if err := model.UpdateUser(id, updates); err != nil {
		middleware.SugarLogger.Errorw("用户资料更新失败",
			append(logFields, "error", err.Error())...,
		)
		return err
	}

	middleware.SugarLogger.Infow("用户资料更新成功",
		"targetUserID", id,
		"updatedFields", logFields,
	)
	return nil
}

// ChangePassword 修改密码（添加权限验证）
func (s *UserService) ChangePassword(targetUserID uint, oldPassword, newPassword string) error {
	logFields := []interface{}{
		"operatorID", s.ID,
		"targetUserID", targetUserID,
		"isAdmin", s.Role == "admin",
	}

	middleware.SugarLogger.Infow("密码修改请求", logFields...)
	// 权限验证：当前用户是管理员或修改自己的密码
	if s.Role != "admin" && s.ID != targetUserID {
		middleware.SugarLogger.Warnw("越权密码修改尝试", logFields...)
		return errors.New("无权修改其他用户密码")
	}

	// 当修改他人密码时（管理员操作），跳过旧密码验证
	if s.Role != "admin" {
		user, err := model.GetUserByID(targetUserID)
		if err != nil {
			return err
		}

		// 只有修改自己密码时需要验证旧密码
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
			return utils.ErrInvalidCredentials
		}
	}

	if err := model.UpdatePassword(targetUserID, newPassword); err != nil {
		middleware.SugarLogger.Errorw("密码更新失败",
			append(logFields, "error", err.Error())...,
		)
		return err
	}

	middleware.SugarLogger.Infow("密码更新成功", logFields...)

	return nil
}

// DeleteAccount 删除账户（添加权限验证）
func (s *UserService) DeleteAccount(targetUserID uint) error {
	logFields := []interface{}{
		"operatorID", s.ID,
		"targetUserID", targetUserID,
		"operatorRole", s.Role,
	}

	middleware.SugarLogger.Infow("账户删除请求", logFields...)

	// 权限验证：当前用户是管理员或删除自己账户
	if s.Role != "admin" && s.ID != targetUserID {
		return errors.New("无权删除其他用户账户")
	}

	// 管理员删除时需要额外验证（可选）
	if s.Role == "admin" && s.ID == targetUserID {
		return errors.New("管理员不能删除自己账户")
	}

	// 普通用户删除自己时需要密码验证（可选）
	if s.Role == "user" {
		user, err := model.GetUserByID(targetUserID)
		if err != nil {
			return err
		}
		if user.Status == 0 {
			return errors.New("账户已被禁用")
		}
	}

	if err := model.DeleteUser(targetUserID); err != nil {
		middleware.SugarLogger.Errorw("账户删除失败",
			append(logFields, "error", err.Error())...,
		)
		return err
	}

	middleware.SugarLogger.Infow("账户已删除",
		"targetUserID", targetUserID,
		"operatorID", s.ID,
	)
	return nil
}

// 在更新用户信息时校验角色
func (s *UserService) UpdateUserRole(id uint, newRole string) error {
	middleware.SugarLogger.Infow("角色变更请求",
		"operatorID", s.ID,
		"targetUserID", id,
		"newRole", newRole,
	)

	if !IsValidRole(newRole) {
		middleware.SugarLogger.Warnw("无效角色尝试",
			"inputRole", newRole,
		)
		return errors.New("无效的角色类型")
	}

	if s.Role != "admin" {
		middleware.SugarLogger.Warnw("非管理员角色修改尝试",
			"operatorRole", s.Role,
		)
		return errors.New("权限不足，无法修改用户角色")
	}

	if err := model.UpdateUserRole(id, newRole); err != nil {
		middleware.SugarLogger.Errorw("角色更新失败",
			"targetUserID", id,
			"error", err.Error(),
		)
		return err
	}

	middleware.SugarLogger.Infow("角色更新成功",
		"targetUserID", id,
		"newRole", newRole,
	)
	return nil
}

// GetAllUsers 获取用户列表（Service层）
func (s *UserService) GetAllUsers(page, pageSize int) ([]UserDTO, error) {
	// 添加访问控制逻辑
	if s.Role != "admin" {
		return nil, errors.New("无权访问")
	}
	// 转换为DTO对象
	users, err := model.GetAllUsers(page, pageSize)
	if err != nil {
		return nil, err
	}
	userDTOs := invertUserDTOList(users)
	return userDTOs, nil
}

// GetUserCount 获取所有用户的总数
func (s *UserService) GetUserCount() (int64, error) {
	// 添加访问控制逻辑
	if s.Role != "admin" {
		return 0, errors.New("无权访问")
	}
	return model.GetUserCount()
}

// GetUserByID 获取用户信息（Service层）
func (s *UserService) GetUserByID(id uint) (*UserDTO, error) {
	// 添加访问控制逻辑
	if s.ID == id || s.Role == "admin" {
		user, err := model.GetUserByID(id)
		if err != nil {
			return nil, err
		}
		userDTO := invertUserDTO(*user)
		return &userDTO, nil
	}
	return nil, errors.New("无权访问")
}

// AddUser 添加一个用户
func (s *UserService) AddUser(user *model.User) error {
	if s.Role != "admin" {
		return errors.New("只有管理员可以添加用户")
	}
	// 检查用户是否已存在
	if exists, _ := model.UserExists(user.Username, user.Email); exists {
		return utils.ErrUserAlreadyExists
	}
	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)
	if !IsValidRole(user.Role) {
		return errors.New("无效的角色类型")
	}
	return model.CreateUser(user)
}

// SearchUsers 搜索用户
func (s *UserService) SearchUsers(username, email string, status int, role string, page, pageSize int) ([]UserDTO, error) {
	// 添加访问控制逻辑
	if s.Role != "admin" {
		return nil, errors.New("只有管理员可以搜索用户")
	}
	users, err := model.SearchUsers(username, email, status, role, page, pageSize)
	if err != nil {
		return nil, err
	}
	UserDTOs := invertUserDTOList(users)
	return UserDTOs, nil
}

// SearchUsersCount 获取符合搜索条件的用户总数
func (s *UserService) SearchUsersCount(username, email string, status int, role string) (int64, error) {
	// 添加访问控制逻辑
	if s.Role != "admin" {
		return 0, errors.New("只有管理员可以搜索用户")
	}
	return model.SearchUsersCount(username, email, status, role)
}
