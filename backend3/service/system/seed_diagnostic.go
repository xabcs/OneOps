package system

import (
	"time"

	modelcmdb "oneops/backend3/model/cmdb"
	modelk8s "oneops/backend3/model/k8s"
	"oneops/backend3/pkg/database"
	"oneops/backend3/pkg/logger"

	"go.uber.org/zap"
)

// initDiagnosticData 初始化诊断相关数据
func (i *Initializer) initDiagnosticData() error {
	logger.Info("开始初始化诊断功能数据...")

	// 1. 初始化诊断权限
	if err := i.initDiagnosticPermissions(); err != nil {
		return err
	}

	// 2. 初始化诊断配置
	if err := i.initDiagnosticConfig(); err != nil {
		return err
	}

	logger.Info("诊断功能数据初始化完成")
	return nil
}

// initDiagnosticPermissions 初始化诊断权限
func (i *Initializer) initDiagnosticPermissions() error {
	logger.Info("开始初始化诊断权限...")

	db := database.GetDB()

	permissions := []modelk8s.DiagnosticPermission{
		{
			Name:        "执行K8s诊断",
			Code:        "k8s:diagnostic:execute",
			Category:    "K8s诊断",
			Description: "对K8s集群中的Java应用执行Arthas诊断",
			RiskLevel:   "high",
			Status:      1,
		},
		{
			Name:        "查看诊断结果",
			Code:        "k8s:diagnostic:view",
			Category:    "K8s诊断",
			Description: "查看K8s诊断结果和历史记录",
			RiskLevel:   "medium",
			Status:      1,
		},
	}

	for _, permission := range permissions {
		var existingPerm modelk8s.DiagnosticPermission
		err := db.Where("code = ?", permission.Code).First(&existingPerm).Error

		if err == nil {
			// 权限已存在，更新数据
			db.Model(&existingPerm).Updates(map[string]interface{}{
				"name":        permission.Name,
				"description": permission.Description,
				"risk_level":  permission.RiskLevel,
				"status":      permission.Status,
			})
			logger.Debug("更新诊断权限", zap.String("name", permission.Name))
		} else {
			// 权限不存在，创建新权限
			if err := db.Create(&permission).Error; err != nil {
				logger.Error("创建诊断权限失败",
					zap.String("name", permission.Name),
					zap.String("code", permission.Code),
					zap.Error(err))
				return err
			}
			logger.Info("创建诊断权限",
				zap.String("name", permission.Name),
				zap.Uint("id", permission.ID))
		}
	}

	return nil
}

// initDiagnosticConfig 初始化诊断配置
func (i *Initializer) initDiagnosticConfig() error {
	logger.Info("开始初始化诊断配置...")

	db := database.GetDB()

	configs := []modelk8s.DiagnosticConfig{
		{
			ClusterID:   "*",
			Namespace:   "*",
			ConfigKey:   "default_timeout",
			ConfigValue: "60",
			Description: "默认诊断超时时间(秒)",
		},
		{
			ClusterID:   "*",
			Namespace:   "*",
			ConfigKey:   "max_output_size",
			ConfigValue: "10485760", // 10MB
			Description: "最大输出大小(字节)",
		},
		{
			ClusterID:   "*",
			Namespace:   "*",
			ConfigKey:   "enable_auto_cleanup",
			ConfigValue: "true",
			Description: "是否自动清理临时文件",
		},
		{
			ClusterID:   "*",
			Namespace:   "*",
			ConfigKey:   "history_retention_days",
			ConfigValue: "30",
			Description: "历史记录保留天数",
		},
	}

	for _, config := range configs {
		var existingConfig modelk8s.DiagnosticConfig
		err := db.Where("cluster_id = ? AND namespace = ? AND config_key = ?",
			config.ClusterID, config.Namespace, config.ConfigKey).First(&existingConfig).Error

		if err == nil {
			// 配置已存在，更新数据
			db.Model(&existingConfig).Updates(map[string]interface{}{
				"config_value": config.ConfigValue,
				"description":  config.Description,
			})
			logger.Debug("更新诊断配置", zap.String("key", config.ConfigKey))
		} else {
			// 配置不存在，创建新配置
			if err := db.Create(&config).Error; err != nil {
				logger.Error("创建诊断配置失败",
					zap.String("key", config.ConfigKey),
					zap.Error(err))
				return err
			}
			logger.Info("创建诊断配置",
				zap.String("key", config.ConfigKey))
		}
	}

	return nil
}

// initAgentVersions 初始化 Agent 版本数据
func (i *Initializer) initAgentVersions() error {
	logger.Info("开始初始化 Agent 版本数据...")

	db := database.GetDB()

	defaultVersion := modelcmdb.AgentVersion{
		Version:       "1.0.0",
		ReleaseNotes:  "Agent 初始版本，支持基础监控指标采集",
		Changelog:     "- 支持基础监控指标采集\n- 支持心跳上报\n- 支持 /metrics 端点",
		ReleasedAt:    time.Now(),
		IsLatest:      true,
		IsDeprecated:  false,
		Features:      `{"extended_metrics":false,"custom_configs":false}`,
		DownloadCount: 0,
		DeployCount:   0,
	}

	if err := db.Create(&defaultVersion).Error; err != nil {
		logger.Error("创建默认 Agent 版本失败", zap.Error(err))
		return err
	}

	logger.Info("创建默认 Agent 版本",
		zap.String("version", defaultVersion.Version),
		zap.Uint("id", defaultVersion.ID))

	return nil
}
