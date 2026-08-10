package system

import (
	modelsystem "oneops/backend3/model/system"
	"oneops/backend3/pkg/database"
	"oneops/backend3/pkg/logger"

	"go.uber.org/zap"
)

// assignDefaultPermissions 为内置角色分配默认权限
func (i *Initializer) assignDefaultPermissions() error {
	logger.Info("开始为内置角色分配默认权限...")

	db := database.GetDB()

	// 定义角色默认权限映射（使用权限代码）
	rolePermissions := map[string][]string{
		"admin": {
			"*.*.*", // 超级管理员拥有所有权限（通配符）
		},
		"ops": {
			// 系统管理
			"system.user.list", "system.user.create", "system.user.update", "system.user.delete",
			"system.role.list", "system.role.update",
			"system.menu.list",
			// CMDB
			"cmdb.server.list", "cmdb.server.view", "cmdb.server.create", "cmdb.server.update", "cmdb.server.delete", "cmdb.server.connect",
			"cmdb.business.list", "cmdb.business.create", "cmdb.business.update", "cmdb.business.delete",
			"cmdb.rooms.list", "cmdb.rooms.create", "cmdb.rooms.update", "cmdb.rooms.delete",
			"cmdb.tags.list", "cmdb.tags.create", "cmdb.tags.update", "cmdb.tags.delete",
			"cmdb.group.list", "cmdb.group.view", "cmdb.group.create", "cmdb.group.update", "cmdb.group.delete", "cmdb.group.assign",
			"cmdb.agents.list", "cmdb.agents.deploy", "cmdb.agents.restart", "cmdb.agents.uninstall",
			// 监控
			"monitor.data.view", "monitor.data.export",
			"monitor.alert.list", "monitor.alert.ack", "monitor.alert.handle",
			"monitor.task.list", "monitor.task.view", "monitor.task.create", "monitor.task.update", "monitor.task.delete", "monitor.task.execute",
			// K8s
			"k8s.cluster.list", "k8s.cluster.view", "k8s.cluster.create", "k8s.cluster.update", "k8s.cluster.delete", "k8s.cluster.connect",
			"k8s.resource.view", "k8s.resource.create", "k8s.resource.update", "k8s.resource.delete",
			"k8s.permission.list", "k8s.permission.assign", "k8s.permission.revoke",
			// 审计
			"audit.login_log.list", "audit.login_log.export",
			"audit.operation_log.list", "audit.operation_log.export",
			"audit.system_event.list",
			"audit.stats.view",
		},
		"auditor": {
			// 只读权限（查看和列表）
			"cmdb.server.list", "cmdb.server.view",
			"cmdb.business.list",
			"cmdb.rooms.list",
			"cmdb.tags.list",
			"cmdb.group.list", "cmdb.group.view",
			"cmdb.agents.list",
			"monitor.data.view",
			"monitor.alert.list",
			"monitor.task.list", "monitor.task.view",
			"k8s.cluster.list", "k8s.cluster.view",
			"k8s.resource.view",
			"k8s.permission.list",
			"audit.login_log.list",
			"audit.operation_log.list",
			"audit.system_event.list",
			"audit.stats.view",
		},
		"viewer": {
			// 最小只读权限
			"cmdb.server.list",
			"monitor.data.view",
		},
		"user": {
			// 普通用户基础权限
			"monitor.data.view",
		},
		"test": {
			// 测试角色权限（用于测试）
			"system.user.list", "system.user.create",
			"system.role.list",
			"cmdb.server.list", "cmdb.server.create",
			"monitor.data.view",
			"k8s.cluster.list",
		},
	}

	for roleCode, permissionCodes := range rolePermissions {
		// 获取角色
		var role modelsystem.Role
		if err := db.Where("code = ?", roleCode).First(&role).Error; err != nil {
			logger.Warn("角色不存在，跳过权限分配",
				zap.String("code", roleCode),
				zap.Error(err))
			continue
		}

		// 获取权限
		var permissions []modelsystem.Permission
		if err := db.Where("code IN ?", permissionCodes).Find(&permissions).Error; err != nil {
			logger.Warn("查询权限失败",
				zap.String("role", roleCode),
				zap.Error(err))
			continue
		}

		// 分配权限
		assignedCount := 0
		for _, perm := range permissions {
			var rolePerm modelsystem.RolePermission
			err := db.Where("role_id = ? AND permission_id = ?", role.ID, perm.ID).
				First(&rolePerm).Error

			if err != nil {
				// 创建新的角色权限关联
				rolePerm = modelsystem.RolePermission{
					RoleID:       role.ID,
					PermissionID: perm.ID,
				}
				if err := db.Create(&rolePerm).Error; err != nil {
					logger.Error("分配权限失败",
						zap.String("role", roleCode),
						zap.String("permission", perm.Code),
						zap.Error(err))
				} else {
					assignedCount++
					logger.Debug("分配权限",
						zap.String("role", roleCode),
						zap.String("permission", perm.Code))
				}
			}
		}

		logger.Info("角色权限分配完成",
			zap.String("role", roleCode),
			zap.Int("总权限数", len(permissions)),
			zap.Int("新分配", assignedCount))
	}

	return nil
}

// initAPIPermissions 初始化层级权限代码到Casbin
func (i *Initializer) initAPIPermissions() error {
	logger.Info("开始初始化层级权限代码到Casbin...")

	// 获取 PermissionService 实例
	permService, err := GetPermissionService()
	if err != nil {
		logger.Warn("获取PermissionService失败，跳过权限初始化", zap.Error(err))
		return nil // 不阻塞启动
	}

	// 清除所有旧的API路径格式的策略
	logger.Info("清除旧的API路径格式策略...")
	if err := permService.ClearAllPolicies(); err != nil {
		logger.Error("清除旧策略失败", zap.Error(err))
		return err
	}

	// 同步所有角色的权限到Casbin（使用层级权限代码）
	logger.Info("同步所有角色权限到Casbin...")
	if err := permService.SyncAllRolesToCasbin(); err != nil {
		logger.Error("同步角色权限失败", zap.Error(err))
		return err
	}

	// 验证策略数量
	policies := permService.GetAllPolicies()
	logger.Info("层级权限代码初始化完成", zap.Int("总策略数", len(policies)))

	return nil
}
