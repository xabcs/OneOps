package services

import (
	"fmt"
	"testing"
)

// TestCasbinAPIManager 测试Casbin API管理器
func TestCasbinAPIManager(t *testing.T) {
	// 初始化服务
	manager, err := NewCasbinAPIManager()
	if err != nil {
		t.Fatalf("创建Casbin API管理器失败: %v", err)
	}

	// 测试添加API权限
	t.Run("AssignPermission", func(t *testing.T) {
		err := manager.AssignPermissionToRole(
			"test_role",
			"/api/test",
			"GET",
			"测试API",
			"测试API接口",
			"test",
		)
		if err != nil {
			t.Errorf("添加权限失败: %v", err)
		}
		fmt.Println("✓ 权限添加成功")
	})

	// 测试获取角色权限
	t.Run("GetRolePermissions", func(t *testing.T) {
		permissions, err := manager.GetRolePermissions("test_role")
		if err != nil {
			t.Errorf("获取角色权限失败: %v", err)
		}
		fmt.Printf("✓ 获取成功，权限数量: %d\n", len(permissions))
		for _, perm := range permissions {
			fmt.Printf("  - %s %s (%s)\n", perm.Method, perm.Path, perm.Name)
		}
	})

	// 测试权限检查
	t.Run("CheckPermission", func(t *testing.T) {
		allowed, err := manager.CheckPermission("test_role", "/api/test", "GET")
		if err != nil {
			t.Errorf("权限检查失败: %v", err)
		}
		fmt.Printf("✓ 权限检查结果: %v\n", allowed)
		if !allowed {
			t.Error("权限检查失败：应该有权限")
		}
	})

	// 测试批量分配权限
	t.Run("BatchAssignPermissions", func(t *testing.T) {
		apis := []APIResource{
			{
				Path:        "/api/users",
				Method:      "GET",
				Name:        "用户列表",
				Description: "获取用户列表",
				Module:      "system",
			},
			{
				Path:        "/api/roles",
				Method:      "GET",
				Name:        "角色列表",
				Description: "获取角色列表",
				Module:      "system",
			},
		}

		err := manager.BatchAssignPermissions("test_role", apis)
		if err != nil {
			t.Errorf("批量分配权限失败: %v", err)
		}
		fmt.Println("✓ 批量分配成功")
	})

	// 测试撤销权限
	t.Run("RevokePermission", func(t *testing.T) {
		err := manager.RevokePermissionFromRole("test_role", "/api/test", "GET")
		if err != nil {
			t.Errorf("撤销权限失败: %v", err)
		}
		fmt.Println("✓ 权限撤销成功")
	})

	// 测试获取所有API资源
	t.Run("GetUniqueAPIResources", func(t *testing.T) {
		resources, err := manager.GetUniqueAPIResources()
		if err != nil {
			t.Errorf("获取API资源失败: %v", err)
		}
		fmt.Printf("✓ 获取成功，API数量: %d\n", len(resources))
		for _, res := range resources {
			fmt.Printf("  - %s %s (%s)\n", res.Method, res.Path, res.Name)
		}
	})
}

// TestPermissionService 测试权限服务
func TestPermissionService(t *testing.T) {
	// 获取权限服务
	permService, err := GetPermissionService()
	if err != nil {
		t.Fatalf("获取权限服务失败: %v", err)
	}

	// 测试API权限检查
	t.Run("HasAPIPermission", func(t *testing.T) {
		// 假设用户ID为1，检查是否有权限访问 /api/system/users GET
		allowed, err := permService.HasAPIPermission(1, "/api/system/users", "GET")
		if err != nil {
			t.Errorf("权限检查失败: %v", err)
		}
		fmt.Printf("✓ 权限检查结果: %v\n", allowed)
	})
}

// ExampleUsage 示例用法
func ExampleUsage() {
	// 1. 创建Casbin API管理器
	manager, err := NewCasbinAPIManager()
	if err != nil {
		fmt.Printf("初始化失败: %v\n", err)
		return
	}

	// 2. 为角色分配权限
	err = manager.AssignPermissionToRole(
		"operator",               // 角色代码
		"/api/system/users",      // API路径
		"GET",                    // HTTP方法
		"用户列表",                // API名称
		"获取用户列表，支持分页", // API描述
		"system",                 // 模块
	)
	if err != nil {
		fmt.Printf("分配权限失败: %v\n", err)
		return
	}

	// 3. 检查权限
	allowed, err := manager.CheckPermission("operator", "/api/system/users", "GET")
	if err != nil {
		fmt.Printf("权限检查失败: %v\n", err)
		return
	}

	fmt.Printf("用户是否有权限: %v\n", allowed)

	// 4. 批量分配权限
	apis := []APIResource{
		{
			Path:        "/api/system/users",
			Method:      "POST",
			Name:        "创建用户",
			Description: "创建新用户",
			Module:      "system",
		},
		{
			Path:        "/api/system/roles",
			Method:      "GET",
			Name:        "角色列表",
			Description: "获取角色列表",
			Module:      "system",
		},
	}

	err = manager.BatchAssignPermissions("operator", apis)
	if err != nil {
		fmt.Printf("批量分配失败: %v\n", err)
		return
	}

	fmt.Println("批量分配成功")
}