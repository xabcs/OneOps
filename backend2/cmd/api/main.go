package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"oneops/backend2/internal/authorization"
	"oneops/backend2/internal/cmdb"
	"oneops/backend2/internal/system"
	"oneops/backend2/pkg/config"
	"oneops/backend2/pkg/database"
	"oneops/backend2/pkg/dto"
	"oneops/backend2/pkg/initializer"
	"oneops/backend2/pkg/logger"
	"oneops/backend2/pkg/utils"

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
	if err := database.InitDB(&cfg.Database); err != nil {
		logger.Fatal("数据库连接失败", zap.Error(err))
	}
	logger.Info("数据库连接成功")

	// 注入 system 包的数据库连接
	system.SetDB(database.GetDB())

	// 初始化加密模块
	encryptionKey := cfg.Encryption.Key
	if encryptionKey != "" {
		if err := utils.InitEncryptionWithKey(encryptionKey); err != nil {
			logger.Warn("加密模块初始化失败，SSH凭证将以明文存储", zap.Error(err))
		} else {
			logger.Info("加密模块初始化成功（使用配置文件密钥）")
		}
	} else {
		if err := utils.InitEncryption(); err != nil {
			logger.Warn("加密模块初始化失败，SSH凭证将以明文存储", zap.Error(err))
		} else {
			logger.Info("加密模块初始化成功（使用环境变量）")
		}
	}

	// 设置 JWT 密钥
	utils.SetJWTSecret(cfg.JWT.Secret)

	// 初始化 Redis（如果启用）
	if err := database.InitRedis(&cfg.Redis); err != nil {
		logger.Warn("Redis 初始化失败，缓存功能将不可用", zap.Error(err))
	}

	// 注入菜单同步回调（避免 system → initializer 循环依赖）
	system.SetMenuSyncFunc(func() error {
		return initializer.NewInitService().SyncMenus()
	})

	// 初始化数据库表和数据
	initService := initializer.NewInitService()
	if err := initService.InitDatabase(); err != nil {
		logger.Warn("初始化数据库数据失败", zap.Error(err))
	} else {
		logger.Info("数据库初始化完成")
	}

	// 清理上次运行遗留的孤儿活跃会话
	cmdb.CleanupOrphanedSessions()

	// 初始化 SessionManager
	cmdb.InitSessionManager()

	// 启动 Agent 指标采集调度器
	go cmdb.StartAgentMetricsScheduler()

	// 创建 Gin 引擎
	r := gin.Default()

	// 注册日志中间件
	r.Use(logger.GinLogger())
	r.Use(logger.GinRecovery())

	// 注册路由
	SetupRoutes(r)

	// 启动服务器
	addr := cfg.Server.GetServerAddr()

	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  120 * time.Second,
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

	defer database.CloseRedis()
}
