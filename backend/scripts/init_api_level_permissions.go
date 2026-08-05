package main

import (
	"encoding/json"
	"fmt"
	"log"
	"oneops/backend/config"
	"oneops/backend/models"
	"oneops/backend/services"

	"gorm.io/gorm"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("Level 4 API级权限系统初始化（真正版本）")
	fmt.Println("========================================")

	// 获取配置
	cfg := config.GetConfig()

	// 初始化数据库连接
	fmt.Println("\n1. 初始化数据库连接...")
	err := services.InitDB(&cfg.Database)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	db := services.GetDB()
	fmt.Println("✓ 数据库连接完成")

	// 初始化权限服务
	fmt.Println("\n2. 初始化权限服务...")
	permService, err := services.NewPermissionService()
	if err != nil {
		log.Fatalf("权限服务初始化失败: %v", err)
	}
	fmt.Println("✓ 权限服务初始化完成")

	// 清除旧的 Casbin 策略
	fmt.Println("\n3. 清除旧的 Casbin 策略...")
	if err := permService.ClearAllPolicies(); err != nil {
		log.Printf("警告: 清除旧策略失败: %v", err)
	} else {
		fmt.Println("✓ 旧策略清除完成")
	}

	// 为不同角色设置默认的 API 权限
	fmt.Println("\n4. 为角色设置默认API权限...")

	// 为 admin 角色设置所有权限
	fmt.Println("\n  4.1 为 admin 角色设置所有API权限...")
	err = setupAdminPermissions(permService)
	if err != nil {
		log.Printf("警告: admin权限设置失败: %v", err)
	} else {
		fmt.Println("✓ admin权限设置完成")
	}

	// 为 ops 角色设置系统管理权限
	fmt.Println("\n  4.2 为 ops 角色设置系统管理API权限...")
	err = setupOpsPermissions(permService)
	if err != nil {
		log.Printf("警告: ops权限设置失败: %v", err)
	} else {
		fmt.Println("✓ ops权限设置完成")
	}

	// 为 viewer 角色设置只读权限
	fmt.Println("\n  4.3 为 viewer 角色设置只读API权限...")
	err = setupViewerPermissions(permService)
	if err != nil {
		log.Printf("警告: viewer权限设置失败: %v", err)
	} else {
		fmt.Println("✓ viewer权限设置完成")
	}

	// 显示同步结果
	fmt.Println("\n5. 验证同步结果...")
	policies := permService.GetAllPolicies()
	if policies == nil {
		fmt.Println("警告: 获取策略失败")
	} else {
		fmt.Printf("✓ 当前共有 %d 条 Casbin 策略\n", len(policies))

		// 按角色统计策略
		roleStats := make(map[string]int)
		for _, policy := range policies {
			if len(policy) >= 1 {
				roleStats[policy[0]]++
			}
		}

		fmt.Println("\n各角色权限统计:")
		for role, count := range roleStats {
			fmt.Printf("  - %s: %d 条API权限\n", role, count)
		}

		// 显示部分策略示例
		fmt.Println("\n策略示例（前10条）:")
		for i, policy := range policies {
			if i >= 10 {
				break
			}
			fmt.Printf("  - %v\n", policy)
		}
	}

	// 测试权限检查
	fmt.Println("\n6. 测试权限检查...")
	testPermissionCheck(permService, db)

	fmt.Println("\n========================================")
	fmt.Println("Level 4 API级权限系统初始化完成！")
	fmt.Println("========================================")
	fmt.Println("\n说明:")
	fmt.Println("  - admin 角色: 拥有所有API权限")
	fmt.Println("  - ops 角色: 拥有系统管理API权限（用户、角色、菜单、权限）")
	fmt.Println("  - viewer 角色: 拥有只读API权限")
	fmt.Println("\n后续操作:")
	fmt.Println("  - 可通过管理界面为角色分配更多API权限")
	fmt.Println("  - 可通过权限管理界面添加新的API端点定义")
}

// setupAdminPermissions 为admin角色设置所有API权限
func setupAdminPermissions(permService *services.PermissionService) error {
	// admin角色拥有所有系统API的权限
	allAPIs := services.GetSystemAPIEndpoints()

	for _, api := range allAPIs {
		for _, method := range api.Methods {
			if err := permService.AssignAPIPermission("admin", api.Path, method); err != nil {
				return err
			}
		}
	}

	return nil
}

// setupOpsPermissions 为ops角色设置系统管理API权限
func setupOpsPermissions(permService *services.PermissionService) error {
	// ops角色可以管理系统用户、角色、菜单、权限
	opsPermissions := []struct {
		Path   string
		Method string
	}{
		// 用户管理
		{"/api/system/users", "GET"},
		{"/api/system/users/:id", "GET"},
		{"/api/system/users/:id/roles", "GET"},
		{"/api/system/users", "POST"},
		{"/api/system/users/:id", "PUT"},
		{"/api/system/users/:id", "DELETE"},
		{"/api/system/users/:id/password", "PUT"},

		// 角色管理
		{"/api/system/roles", "GET"},
		{"/api/system/roles/:id", "GET"},
		{"/api/system/roles/:id/permissions", "GET"},
		{"/api/system/roles", "POST"},
		{"/api/system/roles/:id", "PUT"},
		{"/api/system/roles/:id", "DELETE"},
		{"/api/system/roles/:id/permissions", "POST"},

		// 菜单管理
		{"/api/system/menus", "GET"},
		{"/api/system/menus/tree", "GET"},
		{"/api/system/menus/:id", "GET"},
		{"/api/system/menus", "POST"},
		{"/api/system/menus/:id", "PUT"},
		{"/api/system/menus/:id", "DELETE"},

		// 权限管理
		{"/api/system/permissions", "GET"},
		{"/api/system/permissions/tree", "GET"},
		{"/api/system/permissions/:id", "GET"},
		{"/api/system/permissions", "POST"},
		{"/api/system/permissions/:id", "PUT"},
		{"/api/system/permissions/:id", "DELETE"},
	}

	for _, perm := range opsPermissions {
		if err := permService.AssignAPIPermission("ops", perm.Path, perm.Method); err != nil {
			return err
		}
	}

	return nil
}

// setupViewerPermissions 为viewer角色设置只读权限
func setupViewerPermissions(permService *services.PermissionService) error {
	// viewer角色只有只读权限
	viewerPermissions := []struct {
		Path   string
		Method string
	}{
		// 用户查看
		{"/api/system/users", "GET"},
		{"/api/system/users/:id", "GET"},

		// 角色查看
		{"/api/system/roles", "GET"},
		{"/api/system/roles/:id", "GET"},

		// 菜单查看
		{"/api/system/menus", "GET"},
		{"/api/system/menus/tree", "GET"},

		// 权限查看
		{"/api/system/permissions", "GET"},
		{"/api/system/permissions/tree", "GET"},
	}

	for _, perm := range viewerPermissions {
		if err := permService.AssignAPIPermission("viewer", perm.Path, perm.Method); err != nil {
			return err
		}
	}

	return nil
}

// testPermissionCheck 测试权限检查
func testPermissionCheck(permService *services.PermissionService, db *gorm.DB) {
	// 获取测试用户
	var testUsers []struct {
		ID       uint
		Username string
		RoleIDs  string
	}

	db.Table("users").Select("id, username, role_ids").Find(&testUsers)

	for _, user := range testUsers {
		fmt.Printf("\n测试用户: %s (ID: %d)\n", user.Username, user.ID)

		// 解析角色ID
		var roleIDs []uint
		if err := json.Unmarshal([]byte(user.RoleIDs), &roleIDs); err == nil {
			for _, roleID := range roleIDs {
				var role models.Role
				if err := db.First(&role, roleID).Error; err == nil {
					fmt.Printf("  角色: %s\n", role.Code)

					// 测试几个API权限
					testCases := []struct {
						apiPath     string
						httpMethod  string
						description string
					}{
						{"/api/system/users", "GET", "查看用户列表"},
						{"/api/system/users", "POST", "创建用户"},
						{"/api/system/roles", "GET", "查看角色列表"},
					}

					for _, tc := range testCases {
						hasPermission, err := permService.HasAPIPermission(user.ID, tc.apiPath, tc.httpMethod)
						if err != nil {
							fmt.Printf("    ✗ %s: 错误 - %v\n", tc.description, err)
						} else if hasPermission {
							fmt.Printf("    ✓ %s: 有权限\n", tc.description)
						} else {
							fmt.Printf("    ✗ %s: 无权限\n", tc.description)
						}
					}
				}
			}
		}
	}
}