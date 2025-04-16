package utils

// SuccessResult 成功响应构造器
func SuccessResult[T any](data T, token ...string) Message[T] {
	result := Message[T]{
		Code:    200,
		Message: "success",
		Data:    data,
	}
	if len(token) > 0 {
		result.Token = token[0]
	}
	return result
}

// FailResult 失败响应构造器
func FailResult(code int, message string) Message[any] {
	return Message[any]{
		Code:    code,
		Message: message,
		Data:    nil,
	}
}

// GetDataFromResult 从结果中获取数据
func (message Message[T]) GetDataFromResult() T {
	return message.Data
}
