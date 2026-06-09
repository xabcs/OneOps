package errors

import (
	"fmt"
	"net/http"
)

// AppError 应用错误类型
type AppError struct {
	Code       int    `json:"code"`
	Message    string `json:"message"`
	Err        error  `json:"-"` // 不序列化到 JSON
	StatusCode int    `json:"-"` // HTTP 状态码
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// New 创建新的应用错误
func New(code int, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// Wrap 包装错误
func Wrap(err error, code int, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// WithStatus 设置 HTTP 状态码
func (e *AppError) WithStatus(statusCode int) *AppError {
	e.StatusCode = statusCode
	return e
}

// GetStatusCode 获取 HTTP 状态码
func (e *AppError) GetStatusCode() int {
	if e.StatusCode != 0 {
		return e.StatusCode
	}
	// 根据 Code 映射到默认 HTTP 状态码
	switch {
	case e.Code >= 42000 && e.Code < 43000: // 服务器错误
		return http.StatusBadRequest
	case e.Code == ErrUnauthorized:
		return http.StatusUnauthorized
	case e.Code == ErrForbidden:
		return http.StatusForbidden
	case e.Code == ErrNotFound || e.Code >= 43000: // 菜单、角色等未找到
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}

// 业务错误码定义
const (
	// 通用错误 (50000-50099)
	ErrInternal = 50000

	// 参数验证错误 (40000-40099)
	ErrInvalidParams = 40000
	ErrMissingParam   = 40001
	ErrBadRequest     = 40002

	// 认证授权错误 (40100-40199)
	ErrUnauthorized = 40100
	ErrForbidden    = 40300
	ErrNotFound     = 40400

	// 用户相关错误 (41000-41099)
	ErrUserNotFound    = 41000
	ErrInvalidPassword = 41001
	ErrTokenExpired    = 41002
	ErrTokenInvalid    = 41003

	// 服务器相关错误 (42000-42099)
	ErrServerNotFound    = 42000
	ErrDuplicateHostname = 42001
	ErrDuplicateIP      = 42002
	ErrInvalidCredential = 42003

	// 菜单相关错误 (43000-43099)
	ErrMenuNotFound    = 43000
	ErrMenuHasChildren = 43001
	ErrMenuInUse       = 43002

	// 角色相关错误 (44000-44099)
	ErrRoleNotFound = 44000
	ErrRoleInUse    = 44001
	ErrRoleHasUsers = 44002

	// 凭证相关错误 (45000-45099)
	ErrCredentialNotFound = 45000
	ErrCredentialInUse   = 45001

	// 审计相关错误 (46000-46099)
	ErrAuditNotFound = 46000
)

// 预定义错误实例
var (
	ErrBadRequestRequest     = New(ErrInvalidParams, "请求参数错误")
	ErrUnauthorizedRequest   = New(ErrUnauthorized, "未授权访问")
	ErrForbiddenRequest      = New(ErrForbidden, "禁止访问")
	ErrNotFoundRequest       = New(ErrNotFound, "资源不存在")
	ErrInternalServer        = New(ErrInternal, "服务器内部错误")
	ErrInvalidParamsRequest  = New(ErrInvalidParams, "请求参数格式错误")
)

// IsAppError 判断是否是 AppError
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// GetCode 获取错误码
func GetCode(err error) int {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code
	}
	return ErrInternal
}
