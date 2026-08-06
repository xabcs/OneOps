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
	fmt.Println("权限系统运行时诊断")
	fmt.Println("========================================")

	cfg := config.GetConfig()
	services.InitDB(&cfg.Database)
	db := services.GetDB()
	fmt.Println("✓ 数据库连接完成\n")

	// 1. 检查权限服务初始化
	fmt.Println("=== 1. 权限服务初始化 ===")
	permService, err := services.GetPermissionService()
	if err != nil {
		log.Fatalf("❌ 权限服务初始化失败: %v", err)
	}
	fmt.Println("✓ 权限服务初始化成功")
	fmt.Println()

	// 2. 测试 HasPermission 方法
	fmt.Println("=== 2. 测试 HasPermission 方法 ===")
	userID := uint(2)
	permissionCode := "system.user.list"

	hasPermission, err := permService.HasPermission(userID, permissionCode)
	if err != nil {
		log.Fatalf("❌ 权限检查出错: %v", err)
	}

	fmt.Printf("用户ID: %d\n", userID)
	fmt.Printf("权限代码: %s\n", permissionCode)
	fmt.Printf("权限检查结果: %v\n", hasPermission)
	fmt.Println()

	// 3. 检查用户信息
	fmt.Println("=== 3. 用户详细信息 ===")
	var user models.User
	if err := db.First(&user, userID).Error; err != nil {
		log.Fatalf("❌ 用户不存在: %v", err)
	}

	fmt.Printf("用户ID: %d\n", user.ID)
	fmt.Printf("用户名: %s\n", user.Username)
	fmt.Printf("状态: %s\n", user.Status)
	fmt.Printf("RoleIDs: %s\n", user.RoleIDs)
	fmt.Println()

	// 4. 解析角色ID
	fmt.Println("=== 4. 角色ID解析 ===")
	var roleIDs []uint
	if err := json.Unmarshal([]byte(user.RoleIDs), &roleIDs); err != nil {
		log.Fatalf("❌ JSON解析失败: %v", err)
	}
	fmt.Printf("解析结果: %v\n", roleIDs)
	fmt.Println()

	// 5. 查询角色信息
	fmt.Println("=== 5. 角色信息 ===")
	var roles []models.Role
	if err := db.Where("id IN ?", roleIDs).Find(&roles).Error; err != nil {
		log.Fatalf("❌ 角色查询失败: %v", err)
	}

	for _, r := range roles {
		fmt.Printf("角色: ID=%d, Code=%s, Name=%s, Status=%d\n", r.ID, r.Code, r.Name, r.Status)
	}
	fmt.Println()

	// 6. 检查 Casbin 规则
	fmt.Println("=== 6. 相关 Casbin 规则 ===")
	policies := permService.GetAllPolicies()
	fmt.Printf("总规则数: %d\n\n", len(policies))

	for _, r := range roles {
		fmt.Printf("角色 %s 的规则:\n", r.Code)
		roleRules := 0
		for _, p := range policies {
			if len(p) >= 2 && p[0] == r.Code {
				roleRules++
				if p[1] == permissionCode {
					fmt.Printf("  ✓ [目标规则] %s, %s, %s\n", p[0], p[1], p[2])
				}
			}
		}
		fmt.Printf("  共 %d 条规则\n\n", roleRules)
	}

	// 7. 最终结论
	fmt.Println("========================================")
	fmt.Println("诊断结论")
	fmt.Println("========================================")
	if hasPermission {
		fmt.Println("✅ 后端权限系统工作正常")
		fmt.Println("✅ HasPermission(2, 'system.user.list') 返回 true")
		fmt.Println()
		fmt.Println("❌ 如果前端仍然提示权限不足，可能原因：")
		fmt.Println("   1. 用户的 JWT token 已过期或无效")
		fmt.Println("   2. token 中的 user_id 不是 2")
		fmt.Println("   3. 前端没有正确发送 Authorization header")
		fmt.Println("   4. 后端服务没有重启，使用了旧代码")
	} else {
		fmt.Println("❌ 权限系统有问题")
		fmt.Println("需要检查权限数据和配置")
	}
	fmt.Println("========================================")

	// 8. 测试建议
	fmt.Println("\n建议操作：")
	fmt.Println("1. 重启后端服务: go run main.go run")
	fmt.Println("2. 前端退出登录后重新登录（获取新 token）")
	fmt.Println("3. 在浏览器控制台检查 localStorage 中的 token")
}
