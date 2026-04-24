package main

import (
	"fmt"
	"oneops/backend/config"
	"oneops/backend/services"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		return
	}

	// 初始化数据库连接
	if err := services.InitDB(&cfg.Database); err != nil {
		fmt.Printf("连接数据库失败: %v\n", err)
		return
	}

	db := services.GetDB()

	// 执行清理SQL
	sql := `
		UPDATE operation_logs
		SET module = CASE
		    WHEN path LIKE '/api/monitoring%' THEN '监控中心'
		    WHEN path LIKE '/api/audit%' THEN '审计管理'
		    WHEN path LIKE '/api/system/menus%' THEN '菜单管理'
		    WHEN path LIKE '/api/system/roles%' THEN '角色管理'
		    WHEN path LIKE '/api/system/users%' THEN '用户管理'
		    WHEN path LIKE '/api/user%' THEN '用户管理'
		    WHEN path LIKE '/api/login%' THEN '认证管理'
		    WHEN path LIKE '/api/logout%' THEN '认证管理'
		    WHEN path LIKE '/api/tasks%' THEN '任务管理'
		    WHEN path LIKE '/api/servers%' THEN '服务器管理'
		    WHEN path LIKE '/api/containers%' THEN '容器管理'
		    WHEN path LIKE '/api/certificates%' THEN '证书管理'
		    WHEN path LIKE '/api/system%' THEN '系统管理'
		    ELSE '其他'
		END
		WHERE module = '未知模块' OR module = ''
	`

	result := db.Exec(sql)
	if result.Error != nil {
		fmt.Printf("执行清理SQL失败: %v\n", result.Error)
		return
	}

	fmt.Printf("✓ 成功更新 %d 条审计日志记录\n", result.RowsAffected)

	// 查看修复结果
	type ModuleCount struct {
		Module string
		Count  int64
	}

	var results []ModuleCount
	db.Table("operation_logs").
		Select("module, COUNT(*) as count").
		Group("module").
		Order("count DESC").
		Scan(&results)

	fmt.Println("\n模块统计:")
	for _, r := range results {
		fmt.Printf("  %s: %d 条\n", r.Module, r.Count)
	}
}
