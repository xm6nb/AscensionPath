package handler

import (
	"AscensionPath/config"
	"AscensionPath/internal/service"
	"AscensionPath/internal/utils"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

// 定义JWT claims结构
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.StandardClaims
}

// 生成JWT token
func generateToken(userID uint, username, role string) (string, error) {
	// 设置token过期时间(18小时)
	expirationTime := time.Now().Add(18 * time.Hour)

	// 创建claims
	claims := &Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "AscensionPath",
		},
	}

	// 创建token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名token(请替换为您的实际密钥)
	secretKey := config.JwtSecretKey // 应该从配置中获取
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// 认证中间件
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 尝试从cookie或header中获取token
		var token string
		if cookieToken, err := c.Cookie("auth_token"); err == nil {
			token = cookieToken
		} else {
			authHeader := c.GetHeader("Authorization")
			token = authHeader
		}

		// 2. 验证token是否有效
		claims, err := validateToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, utils.FailResult(utils.CodeUnauthorized, "无效的token"))
			return
		}
		// 3. 将用户信息存入上下文
		userService := &service.UserService{
			// 临时用户信息，存储必要的查询条件
			UserDTO: service.UserDTO{
				ID:   claims.UserID,
				Role: claims.Role,
			},
		}
		userInfo, err := userService.GetUserByID(claims.UserID)
		if err != nil || userInfo.Status == 0 { // 检查用户状态
			c.AbortWithStatusJSON(http.StatusUnauthorized, utils.FailResult(utils.CodeUnauthorized, "用户不存在或已被禁用"))
			return
		}
		userService.UserDTO = *userInfo // 更新用户信息
		c.Set("UserInfo", userService)
		c.Next()
	}
}

// 认证是否为管理员
func isAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		userInfo, err := utils.GetDataFromContext(c, "UserInfo", &service.UserService{})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, utils.FailResult(utils.CodeUnauthorized, "未授权"))
			return
		}
		if userInfo.UserDTO.Role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, utils.FailResult(utils.CodeUnauthorized, "权限不足"))
			return
		}
		c.Next()
	}
}

// 验证token有效性
func validateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return config.JwtSecretKey, nil // 使用与生成token相同的密钥
	})

	if err != nil || !token.Valid {
		return nil, err
	}

	return claims, nil
}
