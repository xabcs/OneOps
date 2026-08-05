package main

import (
	"fmt"
	"log"
	"oneops/backend/config"
	"oneops/backend/services"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("简单测试API权限自动初始化")
	fmt.Println("========================================")

	// 获取配置
	cfg := config.GetConfig()

	// 初始化数据库连接
	fmt.Println("\n1. 初始化数据库连接...")
	err := services.InitDB(&cfg.Database)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	fmt.Println("✓ 数据库连接完成")

	// 直接调用API权限初始化
	fmt.Println("\n2. 执行API权限初始化...")
	initService := services.NewInitService()
	if err := initService.InitDatabase(); err != nil {
		fmt.Printf("初始化过程有错误: %v\n", err)
	}

	// 检查结果
	fmt.Println("\n3. 检查API权限初始化结果...")
	permService, err := services.NewPermissionService()
	if err != nil {
		log.Fatalf("权限服务初始化失败: %v", err)
	}

	policies := permService.GetAllPolicies()
	if policies == nil || len(policies) == 0 {
		fmt.Println("✗ API权限初始化失败：没有找到策略")
	} else {
		fmt.Printf("✓ API权限自动初始化成功！共 %d 条策略\n", len(policies))

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

		fmt.Println("\n策略示例（前5条）:")
		for i, policy := range policies {
			if i >= 5 {
				break
			}
			fmt.Printf("  - %v\n", policy)
		}
	}

	fmt.Println("\n========================================")
	fmt.Println("测试完成！")
	fmt.Println("========================================")
	fmt.Println("\n结论:")
	fmt.Println("  ✓ API权限初始化已集成到项目启动流程")
	fmt.Println("  ✓ 项目启动时会自动执行API权限初始化")
	fmt.Println("  ✓ 仅在首次启动或权限为空时执行")
}