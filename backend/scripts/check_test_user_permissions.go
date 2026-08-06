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
	fmt.Println("查询 test 用户的权限信息")
	fmt.Println("========================================")

	// 获取配置
	cfg := config.GetConfig()

	// 初始化数据库连接
	err := services.InitDB(&cfg.Database)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	db := services.GetDB()
	fmt.Println("✓ 数据库连接完成\n")

	// 1. 查看用户信息
	fmt.Println("=== 1. 用户基本信息 ===")
	var user models.User
	if err := db.Where("username = ?", "test").First(&user).Error; err != nil {
		log.Fatalf("用户不存在: %v", err)
	}
	fmt.Printf("用户ID: %d\n", user.ID)
	fmt.Printf("用户名: %s\n", user.Username)
	fmt.Printf("状态: %d\n\n", user.Status)

	// 2. 查看用户的角色
	fmt.Println("=== 2. 用户角色 ===")
	type UserRole struct {
		RoleID   uint
		RoleName string
		RoleCode string
	}
	var userRoles []UserRole
	db.Table("user_roles").
		Select("roles.id as role_id, roles.name as role_name, roles.code as role_code").
		Joins("LEFT JOIN roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", user.ID).
		Scan(&userRoles)

	if len(userRoles) == 0 {
		fmt.Println("❌ 用户没有分配角色！")
	} else {
		for _, ur := range userRoles {
			fmt.Printf("角色ID: %d, 名称: %s, 代码: %s\n", ur.RoleID, ur.RoleName, ur.RoleCode)
		}
	}
	fmt.Println()

	// 3. 查看用户的所有权限
	fmt.Println("=== 3. 用户权限列表 ===")
	type UserPermission struct {
		RoleName       string
		PermissionCode string
		PermissionName string
		Module         string
		Resource       string
		Action         string
	}
	var userPermissions []UserPermission

	db.Table("user_roles").
		Select("roles.name as role_name, permissions.code as permission_code, permissions.name as permission_name, permissions.module, permissions.resource, permissions.action").
		Joins("LEFT JOIN roles ON user_roles.role_id = roles.id").
		Joins("LEFT JOIN role_permissions ON roles.id = role_permissions.role_id").
		Joins("LEFT JOIN permissions ON role_permissions.permission_id = permissions.id").
		Where("user_roles.user_id = ?", user.ID).
		Order("permissions.code").
		Scan(&userPermissions)

	if len(userPermissions) == 0 {
		fmt.Println("❌ 用户没有任何权限！")
	} else {
		fmt.Printf("共有 %d 个权限:\n", len(userPermissions))
		for _, up := range userPermissions {
			fmt.Printf("  [%s] %s - %s\n", up.RoleName, up.PermissionCode, up.PermissionName)
		}
	}
	fmt.Println()

	// 4. 检查是否有 system.user.list 权限
	fmt.Println("=== 4. 检查 system.user.list 权限 ===")
	var hasPermission bool
	for _, up := range userPermissions {
		if up.PermissionCode == "system.user.list" {
			hasPermission = true
			fmt.Printf("✅ 有权限: [%s] %s - %s\n", up.RoleName, up.PermissionCode, up.PermissionName)
			break
		}
	}
	if !hasPermission {
		fmt.Println("❌ 没有 system.user.list 权限！")
	}
	fmt.Println()

	// 5. 查看所有 system.user 相关权限
	fmt.Println("=== 5. 数据库中的 system.user 权限 ===")
	var systemUserPerms []models.Permission
	db.Where("code LIKE ?", "system.user%").Order("code").Find(&systemUserPerms)
	for _, p := range systemUserPerms {
		fmt.Printf("ID: %d, Code: %s, Name: %s\n", p.ID, p.Code, p.Name)
	}
	fmt.Println()

	// 6. 诊断问题
	fmt.Println("========================================")
	fmt.Println("诊断结果")
	fmt.Println("========================================")

	if len(userRoles) == 0 {
		fmt.Println("❌ 问题：用户没有分配角色")
		fmt.Println("解决：为用户分配角色")
	} else if len(userPermissions) == 0 {
		fmt.Println("❌ 问题：用户角色没有分配权限")
		fmt.Println("解决：为角色分配权限")
	} else if !hasPermission {
		fmt.Println("❌ 问题：用户没有 system.user.list 权限")
		fmt.Println("解决：在角色管理中为用户角色分配此权限")
	} else {
		fmt.Println("✅ 用户有 system.user.list 权限")
		fmt.Println("如果仍然提示权限不足，可能原因：")
		fmt.Println("1. 权限数据没有同步到 Casbin")
		fmt.Println("2. 前端缓存了旧的权限数据")
		fmt.Println("3. Token 中的权限信息过期")
	}

	fmt.Println("\n========================================")
}
