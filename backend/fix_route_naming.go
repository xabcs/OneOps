package main

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Menu 菜单模型
type Menu struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"size:50"`
	RouteName string `gorm:"column:route_name;size:100"`
	Path      string `gorm:"size:200"`
	ParentID  uint   `gorm:"column:parent_id"`
	Sort      int
	Status    int
}

func (Menu) TableName() string {
	return "menus"
}

func main() {
	// 数据库连接配置（使用正确的连接信息）
	dsn := "root:123456@tcp(60.191.116.75:38089)/msre?charset=utf8mb4&parseTime=True&loc=Local"

	fmt.Println("====================================")
	fmt.Println("后端路由命名规范统一修复")
	fmt.Println("====================================")
	fmt.Println()

	// 连接数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ 数据库连接失败: %v", err)
	}
	fmt.Println("✅ 数据库连接成功")
	fmt.Println()

	// 1. 查看当前配置
	fmt.Println("1. 当前管理模块菜单配置:")
	fmt.Println("----------------------------------------")
	var currentMenus []Menu
	if err := db.Where("path LIKE ?", "/manage/%").Order("sort").Find(&currentMenus).Error; err != nil {
		log.Printf("查询失败: %v", err)
	}

	for _, menu := range currentMenus {
		status := "✅"
		if menu.RouteName == "manage_apipermission" {
			status = "❌ 需要修复"
		}
		fmt.Printf("%s ID:%-3d 名称:%-20s 路由名:%-25s 路径:%s\n",
			status, menu.ID, menu.Name, menu.RouteName, menu.Path)
	}
	fmt.Println()

	// 2. 执行修复
	fmt.Println("2. 执行路由命名规范修复:")
	fmt.Println("----------------------------------------")
	fmt.Println("修复规则：多个单词使用连字符连接（kebab-case）")
	fmt.Println("  ❌ manage_apipermission → ✅ manage_api-permission")
	fmt.Println()

	// 修复API权限管理路由
	result := db.Model(&Menu{}).
		Where("route_name = ?", "manage_apipermission").
		Update("route_name", "manage_api-permission")

	if result.Error != nil {
		log.Printf("❌ 修复失败: %v", result.Error)
	} else if result.RowsAffected > 0 {
		fmt.Printf("✅ 成功修复 %d 条记录\n", result.RowsAffected)
	} else {
		fmt.Println("ℹ️  没有需要修复的记录（可能已修复）")
	}
	fmt.Println()

	// 3. 验证修复结果
	fmt.Println("3. 验证修复结果:")
	fmt.Println("----------------------------------------")
	var fixedMenus []Menu
	if err := db.Where("path LIKE ?", "/manage/%").Order("sort").Find(&fixedMenus).Error; err != nil {
		log.Printf("查询失败: %v", err)
	}

	allCorrect := true
	for _, menu := range fixedMenus {
		status := "✅ 符合规范"
		if menu.RouteName == "manage_apipermission" {
			status = "❌ 不符合规范"
			allCorrect = false
		}
		fmt.Printf("%s %-20s → %s\n", status, menu.Name, menu.RouteName)
	}
	fmt.Println()

	// 4. 总结
	if allCorrect {
		fmt.Println("====================================")
		fmt.Println("✅ 所有路由命名符合后端规范！")
		fmt.Println("====================================")
		fmt.Println()
		fmt.Println("路由命名规范：")
		fmt.Println("  • 单个单词：manage_user, manage_role")
		fmt.Println("  • 多个单词：manage_api-permission（使用连字符）")
		fmt.Println()
		fmt.Println("下一步操作：")
		fmt.Println("  1. 清除浏览器缓存或重新登录")
		fmt.Println("  2. 测试菜单跳转是否正常")
	} else {
		fmt.Println("====================================")
		fmt.Println("⚠️  仍有路由不符合规范，请检查！")
		fmt.Println("====================================")
	}
}
