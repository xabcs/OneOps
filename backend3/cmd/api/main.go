// @title           OneOps API
// @version         1.0
// @description     OneOps 一体化运维平台后端 API
// @description     包含系统管理、CMDB、K8s、审计、监控、应用授权等模块
// @host            localhost:8082
// @BasePath        /api
// @schemes         http https
// @securityDefinitions.apikey BearerAuth
// @in   header
// @name Authorization
// @description 输入 Bearer {token}，token 由 /api/login 获取

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"oneops/backend3/config"
	_ "oneops/backend3/docs" // 注册 swag 生成的 OpenAPI 文档
	"oneops/backend3/pkg/database"
	"oneops/backend3/pkg/dto"
	"oneops/backend3/pkg/logger"
	"oneops/backend3/pkg/utils"
	"oneops/backend3/routes"
	cmdbsvc "oneops/backend3/service/cmdb"
	syssvc "oneops/backend3/service/system"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// 1. 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("配置加载失败: %v", err)
	}
	fmt.Printf("配置加载成功，运行环境: %s\n", cfg.App.Environment)

	// 2. 初始化日志系统
	logCfg := logger.LogConfig{
		Level:      cfg.Log.Level,
		Filename:   cfg.Log.Filename,
		MaxSize:    cfg.Log.MaxSize,
		MaxBackups: cfg.Log.MaxBackups,
		MaxAge:     cfg.Log.MaxAge,
		Compress:   cfg.Log.Compress,
	}
	if err := logger.InitLogger(logCfg); err != nil {
		log.Fatalf("日志初始化失败: %v", err)
	}
	defer logger.Sync()
	logger.Info("OneOps后端启动",
		zap.String("version", cfg.App.Version),
		zap.String("environment", cfg.App.Environment),
		zap.String("mode", cfg.Server.Mode),
	)

	// 3. 初始化参数验证器
	if err := dto.InitValidator(); err != nil {
		log.Fatalf("验证器初始化失败: %v", err)
	}

	// 4. 设置 Gin 模式
	gin.SetMode(cfg.Server.Mode)

	// 5. 初始化数据库
	if err := database.InitDB(&cfg.Database); err != nil {
		logger.Fatal("数据库连接失败", zap.Error(err))
	}
	logger.Info("数据库连接成功")

	// 6. 初始化加密模块（敏感数据加密）
	// 安全约束：初始化失败直接终止启动——若降级继续运行，SSH 凭证等敏感数据将以明文落库
	encryptionKey := cfg.Encryption.Key
	if encryptionKey != "" {
		if err := utils.InitEncryptionWithKey(encryptionKey); err != nil {
			logger.Fatal("加密模块初始化失败（拒绝以明文降级运行，请检查 encryption.key 配置）", zap.Error(err))
		}
		logger.Info("加密模块初始化成功（使用配置文件密钥）")
	} else {
		if err := utils.InitEncryption(); err != nil {
			logger.Fatal("加密模块初始化失败（拒绝以明文降级运行，请配置 encryption.key 或 ENCRYPTION_KEY 环境变量）", zap.Error(err))
		}
		logger.Info("加密模块初始化成功（使用环境变量）")
	}

	// 7. 设置 JWT 密钥
	utils.SetJWTSecret(cfg.JWT.Secret)

	// 8. 初始化 Redis（如果启用）
	if err := database.InitRedis(&cfg.Redis); err != nil {
		logger.Warn("Redis 初始化失败，缓存功能将不可用", zap.Error(err))
	}

	// 9. 初始化数据库表和数据
	initializer := syssvc.NewInitializer()
	if err := initializer.Initialize(); err != nil {
		logger.Warn("初始化数据库数据失败", zap.Error(err))
	} else {
		logger.Info("数据库初始化完成")
	}

	// 10. 清理上次运行遗留的孤儿活跃会话
	cmdbsvc.CleanupOrphanedSessions()

	// 11. 启动 Agent 指标采集调度器
	go cmdbsvc.StartAgentMetricsScheduler()

	// 12. 创建 Gin 引擎（用 gin.New 而非 gin.Default：访问日志统一由 RequestLogger 输出、
	// Recovery 统一由下方 GinRecovery 处理，避免重复注册导致一条请求打多份日志/多层恢复）
	r := gin.New()

	// 注册 panic 恢复中间件（全局仅此一处，routes 中不再重复挂 gin.Recovery）
	r.Use(logger.GinRecovery())

	// 13. 注册路由
	routes.SetupRoutes(r, cfg)

	// 13.1 路由权限对账：列出缺少映射的受保护路由（会被中间件拒绝）与指向已删路由的死映射
	if err := syssvc.AuditRoutePermissions(r); err != nil {
		logger.Warn("路由权限对账失败", zap.Error(err))
	} else {
		logger.Info("路由权限对账完成")
	}

	// 注册 Swagger UI（仅在非生产环境开放）
	if cfg.App.Environment != "production" {
		routes.SetupSwagger(r)
		logger.Info("Swagger UI 已启用",
			zap.String("url", fmt.Sprintf("http://localhost:%s/swagger/index.html", cfg.Server.Port)),
		)
	}

	// 14. 启动服务器
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

	// 15. 优雅停机：监听 SIGINT/SIGTERM，收到信号后给在途请求最多 10s 缓冲再退出
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// 独立 goroutine 监听服务退出错误（SafeGo 防 panic 打崩进程）
	errCh := make(chan error, 1)
	utils.SafeGo("http-server", func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	})

	select {
	case err := <-errCh:
		logger.Fatal("服务器启动失败", zap.Error(err))
	case <-ctx.Done():
		logger.Info("收到退出信号，开始优雅停机")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("服务器优雅停机失败（超时或出错）", zap.Error(err))
	} else {
		logger.Info("服务器已优雅停机")
	}

	// 关闭 Redis 连接
	database.CloseRedis()
}
