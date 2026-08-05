package main

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Menu 菜单模型
type Menu struct {
	ID         uint   `gorm:"primaryKey"`
	Name       string `gorm:"size:50"`
	RouteName  string `gorm:"column:route_name;size:100"`
	Path       string `gorm:"size:200"`
	ParentID   uint   `gorm:"column:parent_id"`
	Sort       int
	Status     int
	Resource   string `gorm:"size:30"`
}

func (Menu) TableName() string {
	return "menus"
}

func main() {
	// 数据库连接配置
	dsn := "root:YourPassword@tcp(60.191.116.75:38089)/ops?charset=utf8mb4&parseTime=True&loc=Local"
	// 请修改为实际的数据库密码

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	fmt.Println("=== API权限管理菜单修复 ===")
	fmt.Println()

	// 1. 查看当前配置
	fmt.Println("1. 查看当前的API权限菜单配置:")
	fmt.Println("----------------------------------------")
	var menus []Menu
	if err := db.Where("name LIKE ? OR route_name LIKE ?", "%API权限%", "%apipermission%").Find(&menus).Error; err != nil {
		log.Printf("查询失败: %v", err)
	}

	for _, menu := range menus {
		fmt.Printf("ID: %d, 名称: %s, 路由名: %s, 路径: %s\n",
			menu.ID, menu.Name, menu.RouteName, menu.Path)
	}
	fmt.Println()

	// 2. 修复路由名称
	fmt.Println("2. 修复路由名称...")
	fmt.Println("----------------------------------------")
	result := db.Model(&Menu{}).
		Where("route_name = ?", "manage_apipermission").
		Update("route_name", "manage_api-permission")

	if result.Error != nil {
		log.Printf("修复失败: %v", result.Error)
	} else {
		fmt.Printf("✓ 成功修复 %d 条记录\n", result.RowsAffected)
	}
	fmt.Println()

	// 3. 验证修复结果
	fmt.Println("3. 验证修复结果:")
	fmt.Println("----------------------------------------")
	var fixedMenus []Menu
	if err := db.Where("route_name = ?", "manage_api-permission").Find(&fixedMenus).Error; err != nil {
		log.Printf("查询失败: %v", err)
	}

	for _, menu := range fixedMenus {
		fmt.Printf("✓ ID: %d, 名称: %s, 路由名: %s, 路径: %s\n",
			menu.ID, menu.Name, menu.RouteName, menu.Path)
	}
	fmt.Println()

	fmt.Println("=== 修复完成 ===")
	fmt.Println()
	fmt.Println("下一步操作:")
	fmt.Println("1. 清除浏览器缓存或重新登录")
	fmt.Println("2. 菜单应该可以正常跳转到API权限管理页面")
}
