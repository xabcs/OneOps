package main

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// CasbinRule Casbin规则模型
type CasbinRule struct {
	ID    uint   `gorm:"primaryKey"`
	PType string `gorm:"column:p_type"`
	V0    string `gorm:"column:v0"` // 角色
	V1    string `gorm:"column:v1"` // API路径
	V2    string `gorm:"column:v2"` // HTTP方法
	V3    string `gorm:"column:v3"` // API名称
	V4    string `gorm:"column:v4"` // 描述
	V5    string `gorm:"column:v5"` // 模块
}

func (CasbinRule) TableName() string {
	return "casbin_rule"
}

func main() {
	// 数据库连接配置
	dsn := "root:123456@tcp(60.191.116.75:38089)/msre?charset=utf8mb4&parseTime=True&loc=Local"

	fmt.Println("====================================")
	fmt.Println("更新Casbin规则的业务信息")
	fmt.Println("====================================")
	fmt.Println()

	// 连接数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ 数据库连接失败: %v", err)
	}
	fmt.Println("✅ 数据库连接成功")
	fmt.Println()

	// API业务信息映射
	apiMetadata := []struct {
		path        string
		method      string
		name        string
		description string
		module      string
	}{
		// 用户管理
		{"/api/system/users", "GET", "用户列表", "获取用户列表", "user"},
		{"/api/system/users/:id", "GET", "用户详情", "获取单个用户详情", "user"},
		{"/api/system/users", "POST", "创建用户", "创建新用户", "user"},
		{"/api/system/users/:id", "PUT", "更新用户", "更新用户信息", "user"},
		{"/api/system/users/:id", "DELETE", "删除用户", "删除用户", "user"},
		{"/api/system/users/:id/password", "PUT", "重置密码", "重置用户密码", "user"},
		{"/api/system/users/:id/roles", "GET", "用户角色", "获取用户角色列表", "user"},

		// 角色管理
		{"/api/system/roles", "GET", "角色列表", "获取角色列表", "role"},
		{"/api/system/roles/:id", "GET", "角色详情", "获取单个角色详情", "role"},
		{"/api/system/roles", "POST", "创建角色", "创建新角色", "role"},
		{"/api/system/roles/:id", "PUT", "更新角色", "更新角色信息", "role"},
		{"/api/system/roles/:id", "DELETE", "删除角色", "删除角色", "role"},
		{"/api/system/roles/:id/permissions", "GET", "角色权限", "获取角色权限列表", "role"},
		{"/api/system/roles/:id/permissions", "POST", "分配权限", "为角色分配权限", "role"},

		// 菜单管理
		{"/api/system/menus", "GET", "菜单列表", "获取菜单列表", "menu"},
		{"/api/system/menus/tree", "GET", "菜单树", "获取菜单树结构", "menu"},
		{"/api/system/menus/:id", "GET", "菜单详情", "获取菜单详情", "menu"},
		{"/api/system/menus", "POST", "创建菜单", "创建新菜单", "menu"},
		{"/api/system/menus/:id", "PUT", "更新菜单", "更新菜单信息", "menu"},
		{"/api/system/menus/:id", "DELETE", "删除菜单", "删除菜单", "menu"},

		// 权限管理
		{"/api/system/permissions", "GET", "权限列表", "获取权限列表", "permission"},
		{"/api/system/permissions/tree", "GET", "权限树", "获取权限树结构", "permission"},
		{"/api/system/permissions/:id", "GET", "权限详情", "获取权限详情", "permission"},
		{"/api/system/permissions", "POST", "创建权限", "创建新权限", "permission"},
		{"/api/system/permissions/:id", "PUT", "更新权限", "更新权限信息", "permission"},
		{"/api/system/permissions/:id", "DELETE", "删除权限", "删除权限", "permission"},
	}

	// 更新业务信息
	fmt.Println("开始更新API业务信息...")
	fmt.Println("----------------------------------------")
	updatedCount := 0

	for _, meta := range apiMetadata {
		result := db.Model(&CasbinRule{}).
			Where("p_type = ? AND v1 = ? AND v2 = ?", "p", meta.path, meta.method).
			Updates(map[string]interface{}{
				"v3": meta.name,
				"v4": meta.description,
				"v5": meta.module,
			})

		if result.Error != nil {
			log.Printf("❌ 更新失败 [%s %s]: %v", meta.method, meta.path, result.Error)
			continue
		}

		if result.RowsAffected > 0 {
			fmt.Printf("✅ %-6s %-40s → %s\n", meta.method, meta.path, meta.name)
			updatedCount++
		}
	}

	fmt.Println()
	fmt.Printf("总计更新 %d 条记录\n", updatedCount)
	fmt.Println()

	// 验证更新结果
	fmt.Println("验证更新结果:")
	fmt.Println("----------------------------------------")
	var rules []CasbinRule
	db.Where("p_type = ?", "p").Order("v5, v1, v2").Limit(20).Find(&rules)

	for _, rule := range rules {
		fmt.Printf("%-10s %-40s %-15s %-20s\n",
			rule.V2, rule.V1, rule.V3, rule.V5)
	}

	fmt.Println()
	fmt.Println("====================================")
	fmt.Println("✅ 更新完成！")
	fmt.Println("====================================")
	fmt.Println()
	fmt.Println("下一步操作:")
	fmt.Println("1. 刷新前端页面")
	fmt.Println("2. 点击\"同步常用API\"按钮")
	fmt.Println("3. 选择角色查看权限列表")
}
