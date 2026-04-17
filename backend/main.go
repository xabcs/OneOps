package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"oneops/backend/config"
	"oneops/backend/routes"
	"oneops/backend/services"
	"oneops/backend/utils"
)

func main() {
	// 加载配置
	cfg := config.GetConfig()

	// 设置 Gin 模式
	gin.SetMode(cfg.Server.Mode)

	// 初始化数据库
	if err := services.InitDB(&cfg.Database); err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	fmt.Println("数据库连接成功")

	// 设置 JWT 密钥
	utils.SetJWTSecret(cfg.JWT.Secret)

	// 初始化数据库表和数据
	initService := services.NewInitService()
	if err := initService.InitDatabase(); err != nil {
		log.Printf("初始化数据库数据失败: %v", err)
	} else {
		fmt.Println("数据库初始化完成")
	}

	// 创建 Gin 引擎
	r := gin.Default()

	// 注册路由
	routes.SetupRoutes(r)

	// 启动服务器
	addr := cfg.Server.GetServerAddr()
	fmt.Printf("服务器启动成功，监听端口 %s\n", cfg.Server.Port)
	fmt.Printf("API 地址: http://localhost:%s\n", cfg.Server.Port)

	if err := r.Run(addr); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
