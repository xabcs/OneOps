package utils

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message"`
}

// SuccessResponse 成功响应
func SuccessResponse(data interface{}, message string) Response {
	if message == "" {
		message = "success"
	}
	return Response{
		Code:    200,
		Success: true,
		Data:    data,
		Message: message,
	}
}

// ErrorResponse 错误响应
func ErrorResponse(code int, message string) Response {
	return Response{
		Code:    code,
		Success: false,
		Message: message,
	}
}

// SuccessWithData 成功响应（带数据）
func SuccessWithData(data interface{}) Response {
	return SuccessResponse(data, "success")
}

// SuccessWithMessage 成功响应（带消息）
func SuccessWithMessage(message string) Response {
	return SuccessResponse(nil, message)
}

// ErrorUnauthorized 401 未授权
func ErrorUnauthorized(message string) Response {
	if message == "" {
		message = "未授权"
	}
	return ErrorResponse(401, message)
}

// ErrorBadRequest 400 错误请求
func ErrorBadRequest(message string) Response {
	if message == "" {
		message = "请求参数错误"
	}
	return ErrorResponse(400, message)
}

// ErrorInternal 500 内部错误
func ErrorInternal(message string) Response {
	if message == "" {
		message = "服务器内部错误"
	}
	return ErrorResponse(500, message)
}
