package utils

// 业务错误码定义
const (
	ErrCodeDuplicateHostname = 40001 // 主机名重复
	ErrCodeDuplicateIP       = 40002 // IP地址重复
	ErrCodeServerNotFound    = 40003 // 服务器不存在
	ErrCodeInvalidCredential = 40004 // 无效的SSH凭证
)

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

// ErrorDuplicateHostname 主机名重复错误
func ErrorDuplicateHostname() Response {
	return ErrorResponse(ErrCodeDuplicateHostname, "主机名已存在，请使用其他主机名")
}

// ErrorDuplicateIP IP地址重复错误
func ErrorDuplicateIP() Response {
	return ErrorResponse(ErrCodeDuplicateIP, "IP地址已存在，请使用其他IP地址")
}

// ErrorServerNotFound 服务器不存在错误
func ErrorServerNotFound() Response {
	return ErrorResponse(ErrCodeServerNotFound, "服务器不存在")
}

// ErrorInvalidCredential 无效的SSH凭证错误
func ErrorInvalidCredential() Response {
	return ErrorResponse(ErrCodeInvalidCredential, "无效的SSH凭证")
}

// ErrorForbidden 403 禁止访问
func ErrorForbidden(message string) Response {
	if message == "" {
		message = "禁止访问"
	}
	return ErrorResponse(403, message)
}
