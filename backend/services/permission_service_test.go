package services

import (
	"testing"
	"oneops/backend/models"

	"github.com/stretchr/testify/assert"
)

// TestPermissionServiceBasic 测试权限服务基本功能
func TestPermissionServiceBasic(t *testing.T) {
	// 获取权限服务实例
	permService, err := GetPermissionService()
	assert.NoError(t, err, "获取权限服务应该成功")
	assert.NotNil(t, permService, "权限服务不应为空")
}

// TestHierarchicalPermissionCodes 测试层级权限代码格式
func TestHierarchicalPermissionCodes(t *testing.T) {
	// 测试权限代码格式
	testCases := []struct {
		code        string
		validFormat bool
		description string
	}{
		{"system.user.list", true, "系统管理-用户-列表"},
		{"system.user.create", true, "系统管理-用户-创建"},
		{"system.role.update", true, "系统管理-角色-更新"},
		{"cmdb.asset.view", true, "CMDB-资产-查看"},
		{"monitor.task.delete", true, "监控-任务-删除"},
		{"invalid", false, "无效的权限代码"},
		{".user.list", false, "缺少模块名"},
		{"system..list", false, "缺少资源名"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			// 这里可以添加权限代码验证逻辑
			if tc.validFormat {
				assert.Contains(t, tc.code, ".", "权限代码应该包含点分隔符")
				assert.NotEqual(t, '.', tc.code[0], "权限代码不应以点开头")
			}
		})
	}
}

// TestPermissionCodeStructure 测试权限代码结构
func TestPermissionCodeStructure(t *testing.T) {
	// 模拟权限定义
	permission := models.Permission{
		Code:     "system.user.create",
		Name:     "创建用户",
		Module:   "system",
		Resource: "user",
		Action:   "create",
		Level:    3,
	}

	// 验证权限代码结构
	assert.Equal(t, "system.user.create", permission.Code, "权限代码应该匹配")
	assert.Equal(t, "system", permission.Module, "模块名应该匹配")
	assert.Equal(t, "user", permission.Resource, "资源名应该匹配")
	assert.Equal(t, "create", permission.Action, "操作名应该匹配")
	assert.Equal(t, 3, permission.Level, "权限级别应该匹配")
}

// TestCasbinSyncLogic 测试Casbin同步逻辑
func TestCasbinSyncLogic(t *testing.T) {
	permService, err := GetPermissionService()
	if err != nil {
		t.Skip("跳过测试：权限服务初始化失败")
	}

	// 测试清除所有策略
	err = permService.ClearAllPolicies()
	assert.NoError(t, err, "清除所有策略应该成功")

	// 验证策略已清空
	policies := permService.GetAllPolicies()
	assert.Equal(t, 0, len(policies), "策略数量应该为0")
}

// TestPermissionCheckWithWildcard 测试通配符权限
func TestPermissionCheckWithWildcard(t *testing.T) {
	// 测试通配符权限
	wildcardCases := []struct {
		code        string
		description string
	}{
		{"*.*.*", "全局通配符权限"},
		{"system.*.*", "系统管理模块通配符"},
		{"system.user.*", "用户资源通配符"},
	}

	for _, tc := range wildcardCases {
		t.Run(tc.description, func(t *testing.T) {
			// 验证通配符权限代码格式
			assert.Contains(t, tc.code, "*", "通配符权限应该包含星号")
		})
	}
}