package dto

import (
	"oneops/backend2/pkg/logger"

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

	// 自定义错误信息翻译
	validate.SetTagName("true")

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
