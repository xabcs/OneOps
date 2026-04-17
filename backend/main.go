package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"oneops/backend/config"
	"oneops/backend/logger"
	"oneops/backend/routes"
	"oneops/backend/services"
	"oneops/backend/utils"
)

func main() {
	// 加载配置（从配置文件）
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("配置加载失败: %v", err)
	}
	fmt.Printf("配置加载成功，运行环境: %s\n", cfg.App.Environment)

	// 初始化日志系统
	if err := logger.InitLogger(cfg.Log); err != nil {
		log.Fatalf("日志初始化失败: %v", err)
	}
	defer logger.Sync()
	logger.Info("OneOps后端启动",
		zap.String("version", cfg.App.Version),
		zap.String("environment", cfg.App.Environment),
		zap.String("mode", cfg.Server.Mode),
	)

	// 设置 Gin 模式
	gin.SetMode(cfg.Server.Mode)

	// 初始化数据库
	if err := services.InitDB(&cfg.Database); err != nil {
		logger.Fatal("数据库连接失败", zap.Error(err))
	}
	logger.Info("数据库连接成功")

	// 设置 JWT 密钥
	utils.SetJWTSecret(cfg.JWT.Secret)

	// 初始化数据库表和数据
	initService := services.NewInitService()
	if err := initService.InitDatabase(); err != nil {
		logger.Warn("初始化数据库数据失败", zap.Error(err))
	} else {
		logger.Info("数据库初始化完成")
	}

	// 创建 Gin 引擎
	r := gin.Default()

	// 注册日志中间件
	r.Use(logger.GinLogger())
	r.Use(logger.GinRecovery())

	// 注册路由
	routes.SetupRoutes(r)

	// 启动服务器
	addr := cfg.Server.GetServerAddr()
	logger.Info("服务器启动成功",
		zap.String("port", cfg.Server.Port),
		zap.String("address", fmt.Sprintf("http://localhost:%s", cfg.Server.Port)),
	)

	if err := r.Run(addr); err != nil {
		logger.Fatal("服务器启动失败", zap.Error(err))
	}
}
