package system

import (
	"oneops/backend3/pkg/logger"

	"go.uber.org/zap"
)

// InitializeAPIManagement 初始化API管理系统（基于Casbin方案）
func InitializeAPIManagement() error {
	logger.Info("开始初始化API管理系统...")

	// 1. 初始化 Casbin API 管理器
	manager, err := NewCasbinAPIManager()
	if err != nil {
		logger.Error("创建Casbin API管理器失败", zap.Error(err))
		return err
	}

	// 2. 同步常用API到casbin_rule
	if err := manager.SyncCommonAPIs(); err != nil {
		logger.Error("同步常用API失败", zap.Error(err))
		return err
	}

	logger.Info("API管理系统初始化完成")
	return nil
}
