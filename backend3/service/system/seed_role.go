package system

import (
	"encoding/json"

	modelsystem "oneops/backend3/model/system"
	"oneops/backend3/pkg/database"
	"oneops/backend3/pkg/logger"
	"oneops/backend3/pkg/utils"

	"go.uber.org/zap"
)

// syncBuiltinRoles 同步内置角色（权限码推导模式）
func (i *Initializer) syncBuiltinRoles() error {
	logger.Info("开始同步内置角色...")

	db := database.GetDB()

	// 定义需要同步的内置角色
	builtinRoles := []struct {
		code        string
		name        string
		description string
	}{
		{"admin", "超级管理员", "拥有系统所有权限"},
		{"ops", "运维工程师", "负责主机和任务管理"},
		{"auditor", "审计员", "仅拥有查看权限"},
		{"viewer", "查看者", "仅拥有查看权限"},
		{"user", "普通用户", "系统普通用户，拥有基础权限"},
		{"test", "测试角色", "用于测试的角色，拥有部分权限"},
	}

	// 同步或创建每个内置角色
	for _, builtinRole := range builtinRoles {
		var existingRole modelsystem.Role
		err := db.Where("code = ?", builtinRole.code).First(&existingRole).Error

		if err == nil {
			// 角色已存在：只更新名称和描述
			db.Model(&existingRole).Updates(map[string]interface{}{
				"name":        builtinRole.name,
				"description": builtinRole.description,
			})
			logger.Info("更新内置角色信息",
				zap.String("name", builtinRole.name),
				zap.String("code", builtinRole.code))
		} else {
			// 角色不存在，创建新角色
			newRole := modelsystem.Role{
				Name:        builtinRole.name,
				Code:        builtinRole.code,
				Description: builtinRole.description,
				Status:      1,
			}
			if err := db.Create(&newRole).Error; err != nil {
				logger.Error("创建内置角色失败",
					zap.String("name", builtinRole.name),
					zap.String("code", builtinRole.code),
					zap.Any("error", err))
				return err
			}
			logger.Info("创建内置角色（权限通过权限码配置）",
				zap.String("name", builtinRole.name),
				zap.String("code", builtinRole.code),
				zap.Uint("id", newRole.ID))
		}
	}

	logger.Info("内置角色同步完成")

	return nil
}

// initUsers 初始化用户数据
func (i *Initializer) initUsers() error {
	db := database.GetDB()

	// 加密密码
	hashedPassword, err := utils.HashPassword("123456")
	if err != nil {
		return err
	}

	adminRoleIDs := []uint{1}
	adminRoleIDsJSON, _ := json.Marshal(adminRoleIDs)

	user := modelsystem.User{
		Username: "admin",
		Password: hashedPassword,
		Nickname: "超级管理员",
		Email:    "admin@example.com",
		RoleIDs:  string(adminRoleIDsJSON),
		Status:   "active",
		HomePath: "/home",
	}

	return db.Create(&user).Error
}
