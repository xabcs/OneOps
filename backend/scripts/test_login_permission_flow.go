package main

import (
	"encoding/json"
	"fmt"
	"log"
	"oneops/backend/config"
	"oneops/backend/models"
	"oneops/backend/services"
	"oneops/backend/utils"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("完整请求流程模拟")
	fmt.Println("========================================")

	cfg := config.GetConfig()
	services.InitDB(&cfg.Database)
	db := services.GetDB()
	fmt.Println("✓ 数据库连接完成\n")

	// 模拟用户登录并获取 token
	fmt.Println("=== 1. 模拟用户登录 ===")
	username := "test"
	password := "123456"

	var user models.User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		log.Fatalf("❌ 用户不存在: %v", err)
	}
	fmt.Printf("✓ 找到用户: ID=%d, Username=%s\n", user.ID, user.Username)

	// 验证密码
	if !utils.CheckPassword(password, user.Password) {
		log.Fatalf("❌ 密码错误")
	}
	fmt.Println("✓ 密码验证通过")

	// 生成 JWT token
	token, err := utils.GenerateToken(user.ID, user.Username, 24)
	if err != nil {
		log.Fatalf("❌ Token生成失败: %v", err)
	}
	fmt.Printf("✓ 生成Token: %s...\n", token[:20])
	fmt.Println()

	// 解析 Token 查看内容
	fmt.Println("=== 2. 解析 Token 内容 ===")
	claims, err := utils.ParseToken(token)
	if err != nil {
		log.Fatalf("❌ Token解析失败: %v", err)
	}
	fmt.Printf("UserID: %d\n", claims.UserID)
	fmt.Printf("Username: %s\n", claims.Username)
	fmt.Println()

	// 模拟权限检查
	fmt.Println("=== 3. 模拟权限检查 ===")
	permService, err := services.GetPermissionService()
	if err != nil {
		log.Fatalf("❌ 权限服务初始化失败: %v", err)
	}

	// 使用 Token 中的 UserID
	hasPermission, err := permService.HasPermission(claims.UserID, "system.user.list")
	if err != nil {
		log.Fatalf("❌ 权限检查出错: %v", err)
	}

	fmt.Printf("检查结果: HasPermission(%d, 'system.user.list') = %v\n", claims.UserID, hasPermission)
	fmt.Println()

	// 检查用户角色
	fmt.Println("=== 4. 检查用户角色 ===")
	var roleIDs []uint
	if err := json.Unmarshal([]byte(user.RoleIDs), &roleIDs); err != nil {
		log.Fatalf("❌ JSON解析失败: %v", err)
	}
	fmt.Printf("用户RoleIDs: %v\n", roleIDs)

	var roles []models.Role
	if err := db.Where("id IN ?", roleIDs).Find(&roles).Error; err != nil {
		log.Fatalf("❌ 角色查询失败: %v", err)
	}

	for _, r := range roles {
		fmt.Printf("  角色: ID=%d, Code=%s, Name=%s\n", r.ID, r.Code, r.Name)
	}
	fmt.Println()

	// 测试 HTTP 请求
	fmt.Println("=== 5. 测试 HTTP 请求 ===")
	testHTTPPermissionCheck(token)
	fmt.Println()

	// 结论
	fmt.Println("========================================")
	fmt.Println("诊断结论")
	fmt.Println("========================================")
	if hasPermission {
		fmt.Println("✅ 权限系统工作正常")
		fmt.Println("✓ HasPermission() 返回 true")
		fmt.Println()
		fmt.Println("如果前端仍然提示权限不足，问题在于：")
		fmt.Println("1. 前端发送的 Authorization header 格式错误")
		fmt.Println("2. 后端中间件链执行顺序有问题")
		fmt.Println("3. 某个中间件提前中断了请求")
		fmt.Println("4. gin.Context 中的 user_id 设置不正确")
	} else {
		fmt.Println("❌ 权限检查失败")
		fmt.Println("问题：用户角色或权限配置不正确")
	}
	fmt.Println("========================================")
}

func testHTTPPermissionCheck(token string) {
	// 这里可以添加实际的 HTTP 请求测试
	fmt.Println("提示：需要实际启动后端服务来测试 HTTP 请求")
	fmt.Println("可以使用 curl 命令测试：")
	fmt.Printf("curl -H \"Authorization: Bearer %s\" http://localhost:8080/api/system/users\n", token)
}
