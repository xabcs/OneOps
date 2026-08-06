package main

import (
	"encoding/json"
	"fmt"
	"log"
	"oneops/backend/config"
	"oneops/backend/models"
	"oneops/backend/services"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("测试权限检查执行流程")
	fmt.Println("========================================")

	cfg := config.GetConfig()
	services.InitDB(&cfg.Database)
	db := services.GetDB()
	fmt.Println("✓ 数据库连接完成\n")

	// 获取权限服务
	permService, err := services.GetPermissionService()
	if err != nil {
		log.Fatalf("权限服务初始化失败: %v", err)
	}
	fmt.Println("✓ 权限服务初始化完成\n")

	// 测试用户ID和权限代码
	userID := uint(2)
	permissionCode := "system.user.list"

	fmt.Printf("=== 测试参数 ===\n")
	fmt.Printf("用户ID: %d\n", userID)
	fmt.Printf("权限代码: %s\n", permissionCode)
	fmt.Println()

	// 步骤1：获取用户信息
	fmt.Println("=== 步骤1：获取用户信息 ===")
	var user models.User
	if err := db.First(&user, userID).Error; err != nil {
		log.Fatalf("用户不存在: %v", err)
	}
	fmt.Printf("用户ID: %d, 用户名: %s, RoleIDs: %s\n", user.ID, user.Username, user.RoleIDs)

	// 超级管理员检查
	if user.Username == "admin" {
		fmt.Println("✓ 是超级管理员用户，应该通过权限检查")
		return
	}
	fmt.Println("× 不是 admin 用户，继续检查")
	fmt.Println()

	// 步骤2：获取用户角色（模拟 permission_service.go:191-204）
	fmt.Println("=== 步骤2：获取用户角色 ===")
	var roleIDs []uint
	if err := json.Unmarshal([]byte(user.RoleIDs), &roleIDs); err != nil {
		log.Fatalf("JSON解析失败: %v", err)
	}
	fmt.Printf("解析 role_ids: %v\n", roleIDs)

	var roles []models.Role
	if err := db.Where("id IN ?", roleIDs).Find(&roles).Error; err != nil {
		log.Fatalf("角色查询失败: %v", err)
	}

	if len(roles) == 0 {
		fmt.Println("❌ 没有找到角色，这就是问题所在！")
		return
	}
	fmt.Printf("找到 %d 个角色:\n", len(roles))
	for _, r := range roles {
		fmt.Printf("  - ID: %d, Code: %s, Name: %s, Status: %d\n", r.ID, r.Code, r.Name, r.Status)
	}
	fmt.Println()

	// 步骤3：超级管理员角色检查
	fmt.Println("=== 步骤3：超级管理员角色检查 ===")
	isSuperAdmin := false
	for _, r := range roles {
		if r.Code == "admin" || r.Code == "super_admin" {
			isSuperAdmin = true
			fmt.Printf("✓ 找到超级管理员角色: %s\n", r.Code)
			break
		}
	}
	if !isSuperAdmin {
		fmt.Println("× 不是超级管理员角色，继续检查")
	}
	fmt.Println()

	// 步骤4：Casbin 权限检查（模拟 permission_service.go:113）
	fmt.Println("=== 步骤4：Casbin 权限检查 ===")
	policies := permService.GetAllPolicies()
	fmt.Printf("Casbin 中共有 %d 条规则\n", len(policies))

	// 检查每个角色
	hasPermission := false
	for _, role := range roles {
		fmt.Printf("\n检查角色 %s (Code: %s):\n", role.Name, role.Code)

		// 找到该角色的相关规则
		roleRules := 0
		targetRuleFound := false
		for _, p := range policies {
			if len(p) >= 2 && p[0] == role.Code {
				roleRules++
				if p[1] == permissionCode {
					targetRuleFound = true
					fmt.Printf("  ✓ 找到目标规则: [%s, %s, %s]\n", p[0], p[1], p[2])
				}
			}
		}

		fmt.Printf("  该角色共有 %d 条规则\n", roleRules)
		if !targetRuleFound {
			fmt.Printf("  ❌ 没有找到 '%s' 的规则\n", permissionCode)
		} else {
			fmt.Printf("  ✅ 应该通过权限检查\n")
			hasPermission = true
		}
	}
	fmt.Println()

	// 结论
	fmt.Println("========================================")
	fmt.Println("诊断结论")
	fmt.Println("========================================")
	if hasPermission {
		fmt.Println("✅ 数据层面完全正确，用户应该有权限")
		fmt.Println("❌ 问题可能在于：")
		fmt.Println("   1. 后端服务没有重新加载权限数据")
		fmt.Println("   2. 中间件没有正确调用 HasPermission 方法")
		fmt.Println("   3. 用户上下文中的 user_id 不正确")
	} else {
		fmt.Println("❌ Casbin 中没有对应的权限规则")
	}
	fmt.Println("========================================")
}
