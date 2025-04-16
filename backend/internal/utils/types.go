package utils

import "errors"

const (
	CodeSuccess       = 200 // 成功
	CodeBadRequest    = 400 // 请求参数错误
	CodeUnauthorized  = 401 // 未授权
	CodeInternalError = 500 // 服务器内部错误
)

var (
	ErrInvalidCredentials = errors.New("无效的凭证")
	ErrUserAlreadyExists  = errors.New("用户已存在")
	ErrJsonMarshal        = errors.New("JSON解析失败")
)

// Message 基础响应结构体
type Message[T any] struct {
	Code    int    `json:"code"`            // 状态码
	Message string `json:"message"`         // 消息内容
	Data    T      `json:"data"`            // 业务数据
	Token   string `json:"token,omitempty"` // 认证令牌（可选）
}

// UploadImageRequest 上传文件结构体
type UploadImageRequest struct {
	Filename       string `json:"filename" binding:"required"`
	Base64FileData string `json:"base64FileData" binding:"required"`
}
