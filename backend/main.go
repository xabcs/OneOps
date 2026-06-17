package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"oneops/backend/config"
	"oneops/backend/container"
	"oneops/backend/dto"
	"oneops/backend/handlers"
	"oneops/backend/logger"
	"oneops/backend/routes"
	"oneops/backend/services"
	"oneops/backend/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
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

	// 初始化参数验证器
	if err := dto.InitValidator(); err != nil {
		log.Fatalf("验证器初始化失败: %v", err)
	}

	// 设置 Gin 模式
	gin.SetMode(cfg.Server.Mode)

	// 初始化数据库
	if err := services.InitDB(&cfg.Database); err != nil {
		logger.Fatal("数据库连接失败", zap.Error(err))
	}
	logger.Info("数据库连接成功")

	// 初始化加密模块（敏感数据加密）
	// 优先使用配置文件中的密钥，如果没有则使用环境变量
	encryptionKey := cfg.Encryption.Key
	if encryptionKey != "" {
		// 从配置文件读取
		if err := utils.InitEncryptionWithKey(encryptionKey); err != nil {
			logger.Warn("加密模块初始化失败，SSH凭证将以明文存储", zap.Error(err))
		} else {
			logger.Info("加密模块初始化成功（使用配置文件密钥）")
		}
	} else {
		// 尝试从环境变量读取
		if err := utils.InitEncryption(); err != nil {
			logger.Warn("加密模块初始化失败，SSH凭证将以明文存储", zap.Error(err))
		} else {
			logger.Info("加密模块初始化成功（使用环境变量）")
		}
	}

	// 初始化服务容器（依赖数据库连接）
	logger.Info("开始初始化服务容器")
	container.Initialize(services.GetDB())
	logger.Info("服务容器初始化成功")

	// 设置 JWT 密钥
	utils.SetJWTSecret(cfg.JWT.Secret)

	// 初始化 Redis（如果启用）
	if err := services.InitRedis(&cfg.Redis); err != nil {
		logger.Warn("Redis 初始化失败，缓存功能将不可用", zap.Error(err))
	}

	// 初始化数据库表和数据
	initService := services.NewInitService()
	if err := initService.InitDatabase(); err != nil {
		logger.Warn("初始化数据库数据失败", zap.Error(err))
	} else {
		logger.Info("数据库初始化完成")
	}

	// 清理上次运行遗留的孤儿活跃会话（重启时必须执行，避免在线会话列表显示失效数据）
	services.CleanupOrphanedSessions()

	// 初始化 SessionManager（必须在服务启动前初始化）
	handlers.InitSessionManager()

	// 启动 Agent 指标采集调度器（每 5 分钟 HTTP 拉取 agent_status=running 主机，每 1 分钟检测心跳超时）
	go services.StartAgentMetricsScheduler()

	// 创建 Gin 引擎
	r := gin.Default()

	// 注册日志中间件
	r.Use(logger.GinLogger())
	r.Use(logger.GinRecovery())

	// 注册路由
	routes.SetupRoutes(r)

	// 启动服务器
	addr := cfg.Server.GetServerAddr()

	// 创建 HTTP 服务器并配置超时
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  120 * time.Second, // 空闲连接超时
	}

	logger.Info("服务器启动成功",
		zap.String("port", cfg.Server.Port),
		zap.String("address", fmt.Sprintf("http://localhost:%s", cfg.Server.Port)),
		zap.Int("read_timeout", cfg.Server.ReadTimeout),
		zap.Int("write_timeout", cfg.Server.WriteTimeout),
	)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal("服务器启动失败", zap.Error(err))
	}

	// 关闭 Redis 连接
	defer services.CloseRedis()
}
