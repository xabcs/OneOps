package main

import (
	"fmt"
	"log"
	"oneops/backend/config"
	"oneops/backend/services"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("检查 Casbin 策略状态")
	fmt.Println("========================================")

	cfg := config.GetConfig()
	services.InitDB(&cfg.Database)
	fmt.Println("✓ 数据库连接完成\n")

	// 获取权限服务
	permService, err := services.GetPermissionService()
	if err != nil {
		log.Fatalf("❌ 权限服务初始化失败: %v", err)
	}
	fmt.Println("✓ 权限服务初始化完成\n")

	// 检查 Casbin 规则
	fmt.Println("=== Casbin 规则检查 ===")
	policies := permService.GetAllPolicies()
	fmt.Printf("总策略数: %d\n\n", len(policies))

	// 查找 aaa 角色的 system.user.list 规则
	found := false
	for i, p := range policies {
		if len(p) >= 2 && p[0] == "aaa" && p[1] == "system.user.list" {
			fmt.Printf("✓ 找到目标规则 [索引%d]: %s, %s, %s\n", i, p[0], p[1], p[2])
			found = true
		}
	}

	if !found {
		fmt.Println("❌ 没有找到 aaa, system.user.list 规则")
		fmt.Println("\n所有 aaa 角色的规则：")
		count := 0
		for _, p := range policies {
			if len(p) >= 2 && p[0] == "aaa" {
				count++
				if count <= 10 { // 只显示前10条
					fmt.Printf("  [%d] %s, %s, %s\n", count, p[0], p[1], p[2])
				}
			}
		}
		fmt.Printf("  ... 共 %d 条规则\n", count)
	}
	fmt.Println()

	// 重新同步 Casbin 策略
	fmt.Println("=== 重新同步 Casbin 策略 ===")
	err = permService.InitializeCasbinPolicies()
	if err != nil {
		log.Fatalf("❌ 策略同步失败: %v", err)
	}
	fmt.Println("✓ 策略同步完成")
	fmt.Println()

	// 再次检查
	fmt.Println("=== 再次检查 Casbin 规则 ===")
	policies = permService.GetAllPolicies()
	fmt.Printf("总策略数: %d\n\n", len(policies))

	found = false
	for i, p := range policies {
		if len(p) >= 2 && p[0] == "aaa" && p[1] == "system.user.list" {
			fmt.Printf("✓ 找到目标规则 [索引%d]: %s, %s, %s\n", i, p[0], p[1], p[2])
			found = true
		}
	}

	if !found {
		fmt.Println("❌ 仍然没有找到规则")
	} else {
		fmt.Println("✓ 规则已存在")
	}
	fmt.Println()

	// 测试权限检查
	fmt.Println("=== 测试权限检查 ===")
	hasPermission, err := permService.HasPermission(2, "system.user.list")
	if err != nil {
		log.Fatalf("❌ 权限检查出错: %v", err)
	}
	fmt.Printf("HasPermission(2, 'system.user.list') = %v\n", hasPermission)

	fmt.Println("\n========================================")
	fmt.Println("结论")
	fmt.Println("========================================")
	if hasPermission {
		fmt.Println("✅ 权限系统正常工作")
		fmt.Println("✅ 问题解决了！")
	} else {
		fmt.Println("❌ 权限检查仍然失败")
		fmt.Println("需要进一步调查")
	}
	fmt.Println("========================================")
}
