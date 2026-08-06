package main

import (
	"fmt"
	"log"
	"oneops/backend/config"
	"oneops/backend/services"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("清理废弃的API权限管理菜单")
	fmt.Println("========================================")

	// 获取配置
	cfg := config.GetConfig()

	// 初始化数据库连接
	err := services.InitDB(&cfg.Database)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	db := services.GetDB()
	fmt.Println("✓ 数据库连接完成")

	// 删除菜单项
	fmt.Println("\n1. 删除菜单项...")
	result := db.Exec("DELETE FROM menus WHERE id = 73 AND name = 'API权限管理'")
	if result.Error != nil {
		log.Printf("删除菜单失败: %v", result.Error)
	} else {
		fmt.Printf("✓ 删除了 %d 条菜单记录\n", result.RowsAffected)
	}

	// 删除相关的角色菜单关联
	fmt.Println("\n2. 删除角色菜单关联...")
	result = db.Exec("DELETE FROM role_menus WHERE menu_id = 73")
	if result.Error != nil {
		log.Printf("删除角色菜单关联失败: %v", result.Error)
	} else {
		fmt.Printf("✓ 删除了 %d 条角色菜单关联\n", result.RowsAffected)
	}

	// 确认删除
	fmt.Println("\n3. 验证清理结果...")
	var count int64
	db.Raw("SELECT COUNT(*) FROM menus WHERE name = 'API权限管理'").Scan(&count)
	if count == 0 {
		fmt.Println("✓ 验证成功：API权限管理菜单已完全删除")
	} else {
		fmt.Printf("⚠ 警告：仍有 %d 条相关菜单记录\n", count)
	}

	fmt.Println("\n========================================")
	fmt.Println("清理完成！")
	fmt.Println("========================================")
	fmt.Println("\n请重启后端服务以应用更改")
}
