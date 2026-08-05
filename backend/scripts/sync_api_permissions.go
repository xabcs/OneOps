package main

import (
	"fmt"
	"log"
	"oneops/backend/config"
	"oneops/backend/models"
	"oneops/backend/services"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("Level 4 API级权限系统初始化")
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

	// 自动迁移
	fmt.Println("\n2. 执行数据库迁移...")
	err = db.AutoMigrate(
		&models.User{},
		&models.Role{},
		&models.Permission{},
		&models.RolePermission{},
		&models.Menu{},
	)
	if err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}
	fmt.Println("✓ 数据库迁移完成")

	// 初始化权限服务
	fmt.Println("\n3. 初始化权限服务...")
	permService, err := services.NewPermissionService()
	if err != nil {
		log.Fatalf("权限服务初始化失败: %v", err)
	}
	fmt.Println("✓ 权限服务初始化完成")

	// 清除旧的 Casbin 策略
	fmt.Println("\n4. 清除旧的 Casbin 策略...")
	if err := permService.ClearAllPolicies(); err != nil {
		log.Printf("警告: 清除旧策略失败: %v", err)
	} else {
		fmt.Println("✓ 旧策略清除完成")
	}

	// 同步 API 级权限到 Casbin
	fmt.Println("\n5. 同步 API 级权限到 Casbin...")
	if err := permService.SyncAPIPermissionsToCasbin(); err != nil {
		log.Fatalf("API权限同步失败: %v", err)
	}
	fmt.Println("✓ API权限同步完成")

	// 显示同步结果
	fmt.Println("\n6. 验证同步结果...")
	policies := permService.GetAllPolicies()
	if policies == nil {
		fmt.Println("警告: 获取策略失败")
	} else {
		fmt.Printf("✓ 当前共有 %d 条 Casbin 策略\n", len(policies))

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
	fmt.Println("\n7. 测试权限检查...")
	testPermissionCheck(permService)

	fmt.Println("\n========================================")
	fmt.Println("Level 4 API级权限系统初始化完成！")
	fmt.Println("========================================")
}

// testPermissionCheck 测试权限检查
func testPermissionCheck(permService *services.PermissionService) {
	// 获取测试用户（假设 user_id=2 的用户是 ops 角色）
	testUserID := uint(2)

	// 测试几个API权限
	testCases := []struct {
		apiPath     string
		httpMethod  string
		description string
	}{
		{"/api/system/users", "GET", "查看用户列表"},
		{"/api/system/users", "POST", "创建用户"},
		{"/api/system/roles", "GET", "查看角色列表"},
		{"/api/system/menus", "GET", "查看菜单列表"},
	}

	for _, tc := range testCases {
		hasPermission, err := permService.HasAPIPermission(testUserID, tc.apiPath, tc.httpMethod)
		if err != nil {
			fmt.Printf("  ✗ %s (%s %s): 错误 - %v\n", tc.description, tc.httpMethod, tc.apiPath, err)
		} else if hasPermission {
			fmt.Printf("  ✓ %s (%s %s): 有权限\n", tc.description, tc.httpMethod, tc.apiPath)
		} else {
			fmt.Printf("  ✗ %s (%s %s): 无权限\n", tc.description, tc.httpMethod, tc.apiPath)
		}
	}
}