package main

import (
	"fmt"
	"log"
	"oneops/backend/config"
	"oneops/backend/services"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("初始化API权限")
	fmt.Println("========================================")

	cfg := config.GetConfig()
	err := services.InitDB(&cfg.Database)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	permService, err := services.NewPermissionService()
	if err != nil {
		log.Fatalf("权限服务初始化失败: %v", err)
	}

	// 为角色设置默认API权限
	setupAdminPermissions(permService)
	setupOpsPermissions(permService)
	setupViewerPermissions(permService)

	policies := permService.GetAllPolicies()
	fmt.Printf("✓ 当前共有 %d 条 Casbin 策略\n", len(policies))

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
}

func setupAdminPermissions(permService *services.PermissionService) error {
	allAPIs := services.GetSystemAPIEndpoints()
	for _, api := range allAPIs {
		for _, method := range api.Methods {
			permService.AssignAPIPermission("admin", api.Path, method)
		}
	}
	return nil
}

func setupOpsPermissions(permService *services.PermissionService) error {
	opsPermissions := []struct {
		Path   string
		Method string
	}{
		{"/api/system/users", "GET"}, {"/api/system/users/:id", "GET"},
		{"/api/system/users/:id/roles", "GET"}, {"/api/system/users", "POST"},
		{"/api/system/users/:id", "PUT"}, {"/api/system/users/:id", "DELETE"},
		{"/api/system/users/:id/password", "PUT"},
		{"/api/system/roles", "GET"}, {"/api/system/roles/:id", "GET"},
		{"/api/system/roles/:id/permissions", "GET"}, {"/api/system/roles", "POST"},
		{"/api/system/roles/:id", "PUT"}, {"/api/system/roles/:id", "DELETE"},
		{"/api/system/roles/:id/permissions", "POST"},
		{"/api/system/menus", "GET"}, {"/api/system/menus/tree", "GET"},
		{"/api/system/menus/:id", "GET"}, {"/api/system/menus", "POST"},
		{"/api/system/menus/:id", "PUT"}, {"/api/system/menus/:id", "DELETE"},
		{"/api/system/permissions", "GET"}, {"/api/system/permissions/tree", "GET"},
		{"/api/system/permissions/:id", "GET"}, {"/api/system/permissions", "POST"},
		{"/api/system/permissions/:id", "PUT"}, {"/api/system/permissions/:id", "DELETE"},
	}
	for _, perm := range opsPermissions {
		permService.AssignAPIPermission("ops", perm.Path, perm.Method)
	}
	return nil
}

func setupViewerPermissions(permService *services.PermissionService) error {
	viewerPermissions := []struct {
		Path   string
		Method string
	}{
		{"/api/system/users", "GET"}, {"/api/system/users/:id", "GET"},
		{"/api/system/roles", "GET"}, {"/api/system/roles/:id", "GET"},
		{"/api/system/menus", "GET"}, {"/api/system/menus/tree", "GET"},
		{"/api/system/permissions", "GET"}, {"/api/system/permissions/tree", "GET"},
	}
	for _, perm := range viewerPermissions {
		permService.AssignAPIPermission("viewer", perm.Path, perm.Method)
	}
	return nil
}
