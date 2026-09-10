package dto

import (
	"oneops/backend3/pkg/logger"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

// InitValidator 初始化验证器
func InitValidator() error {
	validate = validator.New()

	// 注册自定义验证器
	if err := RegisterCustomValidators(validate); err != nil {
		return err
	}

	// 注：项目请求校验走 gin 的 binding 标签，不在此处 SetTagName。
	// 此前误设 SetTagName("true") 会把结构体标签名改成不存在的 "true"，
	// 使本实例的校验静默通过，已删除

	logger.Info("验证器初始化成功")
	return nil
}

// GetValidator 获取验证器实例
func GetValidator() *validator.Validate {
	if validate == nil {
		// 如果未初始化，使用默认配置
		validate = validator.New()
		RegisterCustomValidators(validate)
	}
	return validate
}
