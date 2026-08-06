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
	fmt.Println("检查 test 用户角色关联")
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

	// 1. 查看 test 用户信息
	fmt.Println("=== test 用户基本信息 ===")
	var user models.User
	if err := db.Where("username = ?", "test").First(&user).Error; err != nil {
		log.Fatalf("用户不存在: %v", err)
	}
	fmt.Printf("用户ID: %d\n", user.ID)
	fmt.Printf("用户名: %s\n", user.Username)
	fmt.Printf("role_ids: %s\n", user.RoleIDs)
	fmt.Printf("role_ids 字段类型: %T\n", user.RoleIDs)
	fmt.Println()

	// 2. 解析 role_ids JSON 数组
	fmt.Println("=== 解析 role_ids ===")
	var roleIDs []uint
	if user.RoleIDs != "" && user.RoleIDs != "[]" {
		// 使用标准 JSON 解析
		if err := json.Unmarshal([]byte(user.RoleIDs), &roleIDs); err != nil {
			fmt.Printf("JSON解析失败: %v, 原始字符串: %s\n", err, user.RoleIDs)
		} else {
			fmt.Printf("✓ JSON解析成功\n")
		}
	}
	fmt.Printf("解析后的 role_ids: %v (长度: %d)\n", roleIDs, len(roleIDs))
	fmt.Println()

	// 3. 查询对应的角色
	fmt.Println("=== 查询角色信息 ===")
	if len(roleIDs) > 0 {
		var roles []models.Role
		if err := db.Where("id IN ?", roleIDs).Find(&roles).Error; err != nil {
			fmt.Printf("角色查询失败: %v\n", err)
		} else {
			fmt.Printf("找到 %d 个角色:\n", len(roles))
			for _, r := range roles {
				fmt.Printf("  ID: %d, Code: %s, Name: %s, Status: %d\n", r.ID, r.Code, r.Name, r.Status)
			}
		}
	} else {
		fmt.Println("❌ 没有角色ID，无法查询角色")
	}
	fmt.Println()

	// 4. 检查 Casbin 规则
	fmt.Println("=== 检查 Casbin 规则 ===")
	permService, err := services.GetPermissionService()
	if err != nil {
		fmt.Printf("权限服务初始化失败: %v\n", err)
	} else {
		policies := permService.GetAllPolicies()
		fmt.Printf("共有 %d 条 Casbin 规则:\n", len(policies))
		aaaRules := 0
		for _, p := range policies {
			if len(p) >= 2 && p[0] == "aaa" {
				aaaRules++
				fmt.Printf("  [%d] v0=%s, v1=%s, v2=%s\n", aaaRules, p[0], p[1], p[2])
			}
		}
		if aaaRules == 0 {
			fmt.Println("❌ 没有 aaa 角色的 Casbin 规则")
		}
	}
	fmt.Println()

	fmt.Println("========================================")
	fmt.Println("诊断结论")
	fmt.Println("========================================")
	if len(roleIDs) == 0 {
		fmt.Println("❌ 用户的 role_ids 字段为空，需要为用户分配角色")
	} else if user.RoleIDs == "" || user.RoleIDs == "[]" {
		fmt.Println("❌ 用户的 role_ids 字段为空字符串或空数组")
	} else {
		fmt.Println("✅ 用户有角色ID，需要检查角色是否正确关联到 Casbin 规则")
	}
	fmt.Println("========================================")
}
