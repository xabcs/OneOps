package system

import (
	"encoding/json"
	"time"

	modelcmdb "oneops/backend3/model/cmdb"
	modelk8s "oneops/backend3/model/k8s"
	modelsystem "oneops/backend3/model/system"
	"oneops/backend3/pkg/database"
	"oneops/backend3/pkg/logger"
	"oneops/backend3/pkg/utils"

	"go.uber.org/zap"
)

// InitService 初始化服务
type InitService struct{}

// NewInitService 创建初始化服务
func NewInitService() *InitService {
	return &InitService{}
}

// InitDatabase 初始化数据库（创建表和初始数据）
// 使用新的初始化协调器，保持API兼容
func (s *InitService) InitDatabase() error {
	initializer := NewInitializer()
	return initializer.Initialize()
}

// runMigrations 执行 SQL 迁移脚本
func (s *InitService) runMigrations() error {
	logger.Info("开始执行 SQL 迁移...")

	db := database.GetDB()

	// 添加 menu_ids 字段到 roles 表
	menuIDsSQL := `
		ALTER TABLE sys_roles
		ADD COLUMN IF NOT EXISTS menu_ids JSON NULL
		COMMENT '菜单ID列表(JSON数组)，如 [1,2,3]'
		AFTER description;
	`

	if err := db.Exec(menuIDsSQL).Error; err != nil {
		logger.Debug("添加 menu_ids 字段（可能已存在）", zap.Error(err))
	}

	// 添加 resource 字段到 menus 表（用于权限码自动关联菜单）
	menuResourceSQL := `
		ALTER TABLE sys_menus
		ADD COLUMN IF NOT EXISTS resource VARCHAR(30) DEFAULT ''
		COMMENT '对应的资源名称，用于权限码自动关联菜单'
		AFTER permission;
	`

	if err := db.Exec(menuResourceSQL).Error; err != nil {
		logger.Debug("添加 resource 字段（可能已存在）", zap.Error(err))
	}

	// 添加索引
	createResourceIndex := `
		CREATE INDEX IF NOT EXISTS idx_menus_resource ON sys_menus(resource);
	`

	if err := db.Exec(createResourceIndex).Error; err != nil {
		logger.Debug("创建 resource 索引（可能已存在）", zap.Error(err))
	}

	// 添加 disk_partitions 字段到 servers 表
	migrationSQL := `
		ALTER TABLE cmdb_servers
		ADD COLUMN IF NOT EXISTS disk_partitions JSON NULL
		COMMENT '磁盘分区信息 [{"mount":"/","usage":80.5},{"mount":"/var","usage":90.2}]'
		AFTER disk_usage;
	`

	if err := db.Exec(migrationSQL).Error; err != nil {
		logger.Debug("添加 disk_partitions 字段（可能已存在）", zap.Error(err))
	}

	// 创建 agent_metrics 表（监控指标存储）
	createMetricsTable := `
		CREATE TABLE IF NOT EXISTS mon_agent_metrics (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			server_id INT UNSIGNED NOT NULL COMMENT '主机ID',
			metric_type VARCHAR(50) NOT NULL COMMENT '指标类型: performance/system/hardware/service/process/network/security',
			metric_data LONGTEXT NOT NULL COMMENT 'JSON格式指标数据',
			report_time DATETIME NOT NULL COMMENT '采集时间',
			received_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '接收时间',
			INDEX idx_server_type_time (server_id, metric_type, received_at),
			INDEX idx_report_time (report_time)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Agent指标数据表';
	`
	if err := db.Exec(createMetricsTable).Error; err != nil {
		logger.Warn("创建 agent_metrics 表失败", zap.Error(err))
	}

	// 创建告警表
	createAlertsTable := `
		CREATE TABLE IF NOT EXISTS mon_agent_alerts (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			server_id INT UNSIGNED NOT NULL COMMENT '主机ID',
			hostname VARCHAR(100) NOT NULL COMMENT '主机名',
			ip VARCHAR(50) NOT NULL COMMENT 'IP地址',
			rule_id VARCHAR(50) NOT NULL COMMENT '规则ID',
			level VARCHAR(20) NOT NULL COMMENT '告警级别: critical/high/medium/low/info',
			message VARCHAR(500) NOT NULL COMMENT '告警内容',
			metric_value DECIMAL(10,2) COMMENT '当前值',
			threshold DECIMAL(10,2) COMMENT '阈值',
			first_seen DATETIME NOT NULL COMMENT '首次触发时间',
			last_seen DATETIME NOT NULL COMMENT '最近触发时间',
			acknowledged TINYINT(1) DEFAULT 0 COMMENT '是否已确认',
			acknowledged_by VARCHAR(100) COMMENT '确认人',
			acknowledged_at DATETIME COMMENT '确认时间',
			resolved_at DATETIME COMMENT '恢复时间',
			INDEX idx_server_level (server_id, level),
			INDEX idx_acknowledged (acknowledged, first_seen),
			INDEX idx_first_seen (first_seen)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='告警记录表';
	`
	if err := db.Exec(createAlertsTable).Error; err != nil {
		logger.Warn("创建 agent_alerts 表失败", zap.Error(err))
	}

	// 创建 agent_metrics_archive 表（监控指标归档存储）
	createMetricsArchiveTable := `
		CREATE TABLE IF NOT EXISTS mon_agent_metrics_archive (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			server_id INT UNSIGNED NOT NULL COMMENT '主机ID',
			metric_type VARCHAR(50) NOT NULL COMMENT '指标类型',
			metric_data LONGTEXT NOT NULL COMMENT 'JSON格式指标数据',
			received_at DATETIME COMMENT '接收时间',
			archived_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '归档时间',
			INDEX idx_server_type (server_id, metric_type),
			INDEX idx_archived_at (archived_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Agent指标归档表';
	`
	if err := db.Exec(createMetricsArchiveTable).Error; err != nil {
		logger.Warn("创建 agent_metrics_archive 表失败", zap.Error(err))
	}

	// 创建告警规则表
	createAlertRulesTable := `
		CREATE TABLE IF NOT EXISTS mon_agent_alert_rules (
			id VARCHAR(50) PRIMARY KEY COMMENT '规则ID',
			name VARCHAR(100) NOT NULL COMMENT '规则名称',
			level VARCHAR(20) NOT NULL COMMENT '告警级别',
			metric VARCHAR(50) NOT NULL COMMENT '监控指标',
			` + "`condition`" + ` VARCHAR(20) NOT NULL COMMENT '比较条件: gt/lt/gte/lte/eq',
			threshold DECIMAL(10,2) NOT NULL COMMENT '阈值',
			duration INT DEFAULT 0 COMMENT '持续时间(秒)',
			description VARCHAR(500) COMMENT '描述',
			enabled TINYINT(1) DEFAULT 1 COMMENT '是否启用',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='告警规则表';
	`
	if err := db.Exec(createAlertRulesTable).Error; err != nil {
		logger.Warn("创建 agent_alert_rules 表失败", zap.Error(err))
	}

	// 创建通知渠道表
	createNotificationChannelsTable := `
		CREATE TABLE IF NOT EXISTS mon_notification_channels (
			id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			channel_type VARCHAR(50) NOT NULL COMMENT '渠道类型: email/wechat/dingtalk',
			channel_name VARCHAR(100) NOT NULL COMMENT '渠道名称',
			config TEXT COMMENT '配置(JSON)',
			enabled TINYINT(1) DEFAULT 1 COMMENT '是否启用',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='通知渠道表';
	`
	if err := db.Exec(createNotificationChannelsTable).Error; err != nil {
		logger.Warn("创建 notification_channels 表失败", zap.Error(err))
	}

	// 创建巡检报告表
	createInspectionReportsTable := `
		CREATE TABLE IF NOT EXISTS mon_inspection_reports (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			report_type VARCHAR(50) NOT NULL COMMENT '报告类型',
			title VARCHAR(200) NOT NULL COMMENT '报告标题',
			server_ids TEXT COMMENT '服务器ID列表(JSON)',
			status VARCHAR(20) DEFAULT 'pending' COMMENT '状态: pending/running/completed/failed',
			created_by VARCHAR(100) COMMENT '创建人',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_status (status),
			INDEX idx_created_at (created_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='巡检报告表';
	`
	if err := db.Exec(createInspectionReportsTable).Error; err != nil {
		logger.Warn("创建 inspection_reports 表失败", zap.Error(err))
	}

	// 修复 group_bindings 表的 application_permission_id 字段
	fixGroupBindingsTable := `
		ALTER TABLE auth_group_bindings
		ADD COLUMN IF NOT EXISTS application_permission_id BIGINT UNSIGNED NULL COMMENT '关联的权限ID（可选）'
		AFTER application_role_id;
	`
	if err := db.Exec(fixGroupBindingsTable).Error; err != nil {
		logger.Debug("修复 group_bindings 表字段（可能已存在）", zap.Error(err))
	}

	// 如果字段已存在但不允许 NULL，则修改它
	fixGroupBindingsNull := `
		ALTER TABLE auth_group_bindings
		MODIFY COLUMN application_permission_id BIGINT UNSIGNED NULL COMMENT '关联的权限ID（可选）';
	`
	if err := db.Exec(fixGroupBindingsNull).Error; err != nil {
		logger.Debug("修改 group_bindings 表字段允许 NULL（可能已正确）", zap.Error(err))
	}

	logger.Info("SQL 迁移执行完成")
	return nil
}

// initDiagnosticData 初始化诊断相关数据
func (s *InitService) initDiagnosticData() error {
	logger.Info("开始初始化诊断功能数据...")

	// 1. 初始化诊断权限
	if err := s.initDiagnosticPermissions(); err != nil {
		return err
	}

	// 2. 初始化诊断配置
	if err := s.initDiagnosticConfig(); err != nil {
		return err
	}

	logger.Info("诊断功能数据初始化完成")
	return nil
}

// initDiagnosticPermissions 初始化诊断权限
func (s *InitService) initDiagnosticPermissions() error {
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
func (s *InitService) initDiagnosticConfig() error {
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

// initMenus 初始化菜单数据（动态路由模式）
func (s *InitService) initMenus() error {
	db := database.GetDB()

	menus := []modelsystem.Menu{
		// 一级菜单
		{ID: 1, Name: "首页", Icon: "mdi:monitor-dashboard", Path: "/home", Permission: "", Sort: 1, Status: 1, ParentID: 0},
		{ID: 2, Name: "系统管理", Icon: "carbon:cloud-service-management", Path: "/manage", Permission: "", Sort: 2, Status: 1, ParentID: 0},
		{ID: 3, Name: "用户管理", Icon: "ic:round-manage-accounts", Path: "/manage/user", Permission: "system.user.view", Sort: 1, Status: 1, ParentID: 2},
		{ID: 4, Name: "角色管理", Icon: "carbon:user-role", Path: "/manage/role", Permission: "system.role.view", Sort: 2, Status: 1, ParentID: 2},
		{ID: 5, Name: "菜单管理", Icon: "material-symbols:route", Path: "/manage/menu", Permission: "system.menu.view", Sort: 3, Status: 1, ParentID: 2},
		{ID: 13, Name: "关于", Icon: "fluent:book-information-24-regular", Path: "/about", Permission: "", Sort: 5, Status: 1, ParentID: 0},
		{ID: 14, Name: "用户中心", Icon: "mdi:user-circle-outline", Path: "/user-center", Permission: "", Sort: 6, Status: 1, ParentID: 0},
		{ID: 20, Name: "资产管理", Icon: "mdi:server-network", Path: "/cmdb", Permission: "", Sort: 3, Status: 1, ParentID: 0},
		{ID: 21, Name: "主机管理", Icon: "mdi:server", Path: "/cmdb/servers", Permission: "cmdb:server:query", Sort: 1, Status: 1, ParentID: 20},
		{ID: 22, Name: "业务管理", Icon: "mdi:sitemap", Path: "/cmdb/business", Permission: "cmdb:business:query", Sort: 2, Status: 1, ParentID: 20},
		{ID: 23, Name: "机房管理", Icon: "mdi:office-building-marker", Path: "/cmdb/rooms", Permission: "cmdb:room:query", Sort: 3, Status: 1, ParentID: 20},
		{ID: 24, Name: "标签管理", Icon: "mdi:tag-multiple", Path: "/cmdb/tags", Permission: "cmdb:tag:query", Sort: 4, Status: 1, ParentID: 20},
		{ID: 25, Name: "变更记录", Icon: "mdi:history", Path: "/cmdb/changes", Permission: "cmdb:change:query", Sort: 5, Status: 1, ParentID: 20},
		{ID: 26, Name: "SSH凭证", Icon: "mdi:key-variant", Path: "/cmdb/ssh-credentials", Permission: "cmdb:credential:query", Sort: 6, Status: 1, ParentID: 20},
	}

	for _, menu := range menus {
		if err := db.Create(&menu).Error; err != nil {
			return err
		}
	}

	return nil
}

// syncMenus 同步菜单数据（自动检测并添加新菜单）
func (s *InitService) syncMenus() error {
	logger.Info("开始同步菜单数据...")

	db := database.GetDB()

	// 定义动态路由菜单（用于 SoybeanAdmin 动态路由模式）
	// V4 - 匹配新的前端模块化目录结构
	menus := []modelsystem.Menu{
		// ========== 一级菜单 ==========
		{ID: 1, Name: "首页", Icon: "mdi:monitor-dashboard", Path: "/home", Permission: "", MenuType: "menu", Sort: 1, Status: 1, ParentID: 0},

		{ID: 2, Name: "资产管理", Icon: "mdi:server-network", Path: "/cmdb", Permission: "", MenuType: "directory", Sort: 2, Status: 1, ParentID: 0},
		{ID: 3, Name: "监控中心", Icon: "mdi:chart-line", Path: "/monitoring", Permission: "", MenuType: "directory", Sort: 3, Status: 1, ParentID: 0},
		{ID: 7, Name: "授权中心", Icon: "mdi:shield-account", Path: "/auth", Permission: "", MenuType: "directory", Sort: 4, Status: 1, ParentID: 0},
		{ID: 4, Name: "审计中心", Icon: "mdi:file-document", Path: "/audit", Permission: "", MenuType: "directory", Sort: 5, Status: 1, ParentID: 0},
		{ID: 5, Name: "K8s管理", Icon: "mdi:kubernetes", Path: "/k8s", Permission: "", MenuType: "directory", Sort: 6, Status: 1, ParentID: 0},
		// web终端作为独立的一级路由（无导航栏），使用 Sort=7
		{ID: 60, Name: "web终端", Icon: "mdi:console", Path: "/webterminal", Permission: "", MenuType: "menu", Sort: 7, Status: 1, ParentID: 0},
		{ID: 6, Name: "系统管理", Icon: "mdi:cog", Path: "/manage", Permission: "", MenuType: "directory", Sort: 8, Status: 1, ParentID: 0},

		// ========== CMDB 二级菜单 (ID: 20-39) ==========
		{ID: 20, Name: "主机管理", Icon: "mdi:server", Path: "/cmdb/servers", Permission: "cmdb:server:query", MenuType: "menu", Sort: 1, Status: 1, ParentID: 2},
		{ID: 21, Name: "业务管理", Icon: "mdi:sitemap", Path: "/cmdb/business", Permission: "cmdb:business:query", MenuType: "menu", Sort: 2, Status: 1, ParentID: 2},
		// 凭证管理目录
		{ID: 22, Name: "凭证管理", Icon: "mdi:key", Path: "/cmdb/credentials", Permission: "", MenuType: "directory", Sort: 3, Status: 1, ParentID: 2},
		{ID: 23, Name: "访问凭证", Icon: "mdi:key-variant", Path: "/cmdb/credentials/access", Permission: "cmdb:credentials:access", MenuType: "menu", Sort: 1, Status: 1, ParentID: 22},
		{ID: 24, Name: "SSH密钥", Icon: "mdi:ssh", Path: "/cmdb/credentials/ssh", Permission: "cmdb:credentials:ssh", MenuType: "menu", Sort: 2, Status: 1, ParentID: 22},
		// 访问策略
		{ID: 25, Name: "访问策略", Icon: "mdi:shield-lock", Path: "/cmdb/policies", Permission: "cmdb:policies:query", MenuType: "menu", Sort: 4, Status: 1, ParentID: 2},
		// 配置管理目录
		{ID: 26, Name: "配置管理", Icon: "mdi:cog", Path: "/cmdb/config", Permission: "", MenuType: "directory", Sort: 5, Status: 1, ParentID: 2},
		{ID: 27, Name: "业务配置", Icon: "mdi:sitemap", Path: "/cmdb/config/business", Permission: "cmdb:config:business", MenuType: "menu", Sort: 1, Status: 1, ParentID: 26},
		{ID: 28, Name: "机房管理", Icon: "mdi:server", Path: "/cmdb/config/rooms", Permission: "cmdb:rooms:query", MenuType: "menu", Sort: 2, Status: 1, ParentID: 26},
		{ID: 29, Name: "标签管理", Icon: "mdi:tag-multiple", Path: "/cmdb/config/tags", Permission: "cmdb:tags:query", MenuType: "menu", Sort: 3, Status: 1, ParentID: 26},
		{ID: 30, Name: "代理配置", Icon: "mdi:robot", Path: "/cmdb/config/agents", Permission: "cmdb:agents:query", MenuType: "menu", Sort: 4, Status: 1, ParentID: 26},
		// 资产总览
		{ID: 31, Name: "资产总览", Icon: "mdi:chart-pie", Path: "/cmdb/dashboard", Permission: "cmdb:dashboard:query", MenuType: "menu", Sort: 6, Status: 1, ParentID: 2},
		// 审计记录目录
		{ID: 32, Name: "审计记录", Icon: "mdi:history", Path: "/cmdb/audit", Permission: "", MenuType: "directory", Sort: 7, Status: 1, ParentID: 2},
		{ID: 33, Name: "变更记录", Icon: "mdi:file-document", Path: "/cmdb/audit/changes", Permission: "cmdb:audit:changes", MenuType: "menu", Sort: 1, Status: 1, ParentID: 32},
		// 命令审计目录
		{ID: 34, Name: "命令审计", Icon: "mdi:terminal", Path: "/cmdb/audit/command", Permission: "", MenuType: "directory", Sort: 2, Status: 1, ParentID: 32},
		{ID: 35, Name: "命令历史", Icon: "mdi:history", Path: "/cmdb/audit/command/history", Permission: "cmdb:audit:command:history", MenuType: "menu", Sort: 1, Status: 1, ParentID: 34},
		{ID: 36, Name: "在线会话", Icon: "mdi:laptop", Path: "/cmdb/audit/online", Permission: "cmdb:audit:online", MenuType: "menu", Sort: 3, Status: 1, ParentID: 32},
		{ID: 37, Name: "历史会话", Icon: "mdi:history", Path: "/cmdb/audit/sessions", Permission: "cmdb:audit:sessions", MenuType: "menu", Sort: 4, Status: 1, ParentID: 32},
		{ID: 38, Name: "命令记录", Icon: "mdi:code-tags", Path: "/cmdb/audit/commands", Permission: "cmdb:audit:commands", MenuType: "menu", Sort: 5, Status: 1, ParentID: 32},

		// ========== 监控中心二级菜单 (ID: 40-49) ==========
		{ID: 40, Name: "监控概览", Icon: "mdi:chart-line", Path: "/monitoring/overview", Permission: "monitoring:overview:query", MenuType: "menu", Sort: 1, Status: 1, ParentID: 3},
		{ID: 41, Name: "主机监控", Icon: "mdi:server-network", Path: "/monitoring/servers", Permission: "monitoring:servers:query", MenuType: "directory", Sort: 2, Status: 1, ParentID: 3},
		{ID: 42, Name: "告警管理", Icon: "mdi:alert-circle", Path: "/monitoring/alerts", Permission: "monitoring:alerts:query", MenuType: "menu", Sort: 3, Status: 1, ParentID: 3},
		{ID: 43, Name: "趋势分析", Icon: "mdi:chart-areaspline", Path: "/monitoring/trends", Permission: "monitoring:trends:query", MenuType: "menu", Sort: 4, Status: 1, ParentID: 3},
		{ID: 44, Name: "巡检报告", Icon: "mdi:file-document", Path: "/monitoring/reports", Permission: "monitoring:reports:query", MenuType: "menu", Sort: 5, Status: 1, ParentID: 3},
		{ID: 45, Name: "监控设置", Icon: "mdi:cog", Path: "/monitoring/settings", Permission: "monitoring:settings:query", MenuType: "menu", Sort: 6, Status: 1, ParentID: 3},

		// ========== 审计中心二级菜单 (ID: 50-59) ==========
		{ID: 50, Name: "登录审计", Icon: "mdi:login", Path: "/audit/login", Permission: "audit.login.view", MenuType: "menu", Sort: 1, Status: 1, ParentID: 4},
		{ID: 51, Name: "操作审计", Icon: "mdi:account-edit", Path: "/audit/operation", Permission: "audit.operation.view", MenuType: "menu", Sort: 2, Status: 1, ParentID: 4},
		{ID: 52, Name: "系统事件", Icon: "mdi:information", Path: "/audit/system", Permission: "audit.system.view", MenuType: "menu", Sort: 3, Status: 1, ParentID: 4},

		// ========== 授权中心二级菜单 (ID: 100-107) ==========
		{ID: 100, Name: "用户", Icon: "mdi:account", Path: "/auth/users", Permission: "auth:user:query", MenuType: "menu", Sort: 1, Status: 1, ParentID: 7},
		{ID: 101, Name: "用户组", Icon: "mdi:shield-account", Path: "/auth/roles", Permission: "auth:role:query", MenuType: "menu", Sort: 2, Status: 1, ParentID: 7},
		{ID: 102, Name: "应用", Icon: "mdi:application", Path: "/auth/applications", Permission: "auth:app:query", MenuType: "menu", Sort: 3, Status: 1, ParentID: 7},
		{ID: 103, Name: "权限映射", Icon: "mdi:link", Path: "/auth/rolebindings", Permission: "auth:binding:query", MenuType: "menu", Sort: 4, Status: 1, ParentID: 7},
		{ID: 104, Name: "用户授权", Icon: "mdi:account-key", Path: "/auth/userauthorization", Permission: "auth:authorization:query", MenuType: "menu", Sort: 5, Status: 1, ParentID: 7},
		{ID: 105, Name: "操作日志", Icon: "mdi:file-document", Path: "/auth/operationlogs", Permission: "auth:log:query", MenuType: "menu", Sort: 6, Status: 1, ParentID: 7},
		{ID: 106, Name: "用户身份映射", Icon: "mdi:account-switch", Path: "/auth/user-identities", Permission: "auth:identity:query", MenuType: "menu", Sort: 7, Status: 1, ParentID: 7},
		{ID: 107, Name: "用户有效权限", Icon: "mdi:shield-check", Path: "/auth/user-permissions", Permission: "auth:permission:query", MenuType: "menu", Sort: 8, Status: 1, ParentID: 7},

		// ========== K8s管理二级菜单 (ID: 80-99) ==========
		{ID: 80, Name: "集群管理", Icon: "mdi:server-network", Path: "/k8s/clusters", Permission: "k8s:cluster:query", MenuType: "menu", Sort: 1, Status: 1, ParentID: 5},
		{ID: 87, Name: "工作负载", Icon: "mdi:cube-outline", Path: "/k8s/workloads", Permission: "k8s:workload:query", MenuType: "menu", Sort: 2, Status: 1, ParentID: 5},
		{ID: 90, Name: "诊断中心", Icon: "mdi:stethoscope", Path: "/k8s/diagnostic", Permission: "k8s:diagnostic:execute", MenuType: "menu", Sort: 3, Status: 1, ParentID: 5},
		{ID: 88, Name: "网络", Icon: "mdi:network-outline", Path: "/k8s/network", Permission: "k8s:network:query", MenuType: "menu", Sort: 4, Status: 1, ParentID: 5},
		{ID: 89, Name: "配置管理", Icon: "mdi:cog", Path: "/k8s/config", Permission: "k8s:config:query", MenuType: "menu", Sort: 5, Status: 1, ParentID: 5},

		// ========== 系统管理二级菜单 (ID: 70-79) ==========
		{ID: 70, Name: "用户管理", Icon: "mdi:account-multiple", Path: "/manage/user", Permission: "system.user.view", MenuType: "menu", Sort: 1, Status: 1, ParentID: 6},
		{ID: 71, Name: "角色管理", Icon: "mdi:shield-account", Path: "/manage/role", Permission: "system.role.view", MenuType: "menu", Sort: 2, Status: 1, ParentID: 6},
		{ID: 72, Name: "菜单管理", Icon: "mdi:menu", Path: "/manage/menu", Permission: "system.menu.view", MenuType: "menu", Sort: 3, Status: 1, ParentID: 6},
	}

	addedCount := 0
	updatedCount := 0

	for _, menu := range menus {
		var existingMenu modelsystem.Menu
		err := db.Where("id = ?", menu.ID).First(&existingMenu).Error

		if err == nil {
			// 菜单已存在，更新数据（保持数据同步）
			// 注意：不更新 sort 字段，保留用户在菜单管理中修改的排序
			db.Model(&existingMenu).Updates(map[string]interface{}{
				"name":       menu.Name,
				"icon":       menu.Icon,
				"path":       menu.Path,
				"permission": menu.Permission,
				"resource":   menu.Resource,
				"parent_id":  menu.ParentID,
				"status":     menu.Status,
				"menu_type":  menu.MenuType,
			})
			updatedCount++
			logger.Debug("更新菜单",
				zap.String("name", menu.Name),
				zap.Uint("id", menu.ID))
		} else {
			// 菜单不存在，添加新菜单
			if err := db.Create(&menu).Error; err != nil {
				logger.Error("添加菜单失败",
					zap.String("name", menu.Name),
					zap.Any("error", err))
				return err
			}
			addedCount++
			logger.Info("添加新菜单",
				zap.String("name", menu.Name),
				zap.Uint("id", menu.ID),
				zap.String("path", menu.Path))
		}
	}
	// 更新菜单的 resource 字段（根据路径映射到权限码的 resource）
	logger.Info("开始更新菜单resource字段...")

	resourceMappings := []struct {
		path     string
		resource string
	}{
		{"/manage/user", "user"},
		{"/manage/role", "role"},
		{"/manage/menu", "menu"},
		{"/cmdb/servers", "server"},
		{"/cmdb/business", "business"},
		{"/cmdb/config/rooms", "rooms"},
		{"/cmdb/config/tags", "tags"},
		{"/cmdb/config/agents", "agents"},
		{"/cmdb/config/business", "config_business"},
		{"/k8s/clusters", "cluster"},
		{"/k8s/workloads", "workload"},
		{"/k8s/network", "network"},
		{"/k8s/config", "k8s_config"},
		{"/k8s/diagnostic", "diagnostic"},
		{"/monitoring/overview", "overview"},
		{"/monitoring/servers", "monitoring_servers"},
		{"/monitoring/alerts", "alerts"},
		{"/monitoring/trends", "trends"},
		{"/monitoring/reports", "reports"},
		{"/monitoring/settings", "settings"},
		{"/audit/login", "login_audit"},
		{"/audit/operation", "operation_audit"},
		{"/audit/system", "system_audit"},
	}

	for _, mapping := range resourceMappings {
		result := db.Model(&modelsystem.Menu{}).Where("path = ?", mapping.path).Update("resource", mapping.resource)
		if result.Error != nil {
			logger.Warn("更新菜单resource字段失败",
				zap.String("path", mapping.path),
				zap.Error(result.Error))
		} else if result.RowsAffected > 0 {
			logger.Info("更新菜单resource字段",
				zap.String("path", mapping.path),
				zap.String("resource", mapping.resource))
		}
	}

	logger.Info("菜单resource字段更新完成")

	// 删除废弃的 K8s 会话审计菜单（如果存在）
	deletedK8sAuditResult := db.Where("path = ?", "/k8s/audit/sessions").Delete(&modelsystem.Menu{})
	if deletedK8sAuditResult.Error != nil {
		logger.Error("删除废弃K8s会话审计菜单失败", zap.Error(deletedK8sAuditResult.Error))
		return deletedK8sAuditResult.Error
	}
	if deletedK8sAuditResult.RowsAffected > 0 {
		logger.Info("已删除废弃K8s会话审计菜单", zap.Int64("count", deletedK8sAuditResult.RowsAffected))
	}

	deletedResult := db.Where("path = ? OR name IN ?", "/cmdb/groups", []string{"主机分组", "cmdb_groups"}).Delete(&modelsystem.Menu{})
	if deletedResult.Error != nil {
		logger.Error("删除废弃主机分组菜单失败", zap.Error(deletedResult.Error))
		return deletedResult.Error
	}
	if deletedResult.RowsAffected > 0 {
		logger.Info("已删除废弃主机分组菜单", zap.Int64("count", deletedResult.RowsAffected))
	}

	// 删除废弃的 terminal 菜单（已改名为 webterminal）
	deletedTerminalResult := db.Where("path IN ?", []string{"/terminal", "/terminal/workbench"}).Delete(&modelsystem.Menu{})
	if deletedTerminalResult.Error != nil {
		logger.Error("删除废弃终端菜单失败", zap.Error(deletedTerminalResult.Error))
		return deletedTerminalResult.Error
	}
	if deletedTerminalResult.RowsAffected > 0 {
		logger.Info("已删除废弃终端菜单", zap.Int64("count", deletedTerminalResult.RowsAffected))
	}

	logger.Info("菜单同步完成",
		zap.Int("added", addedCount),
		zap.Int("updated", updatedCount),
		zap.Int("total", len(menus)))

	return nil
}

// SyncMenus 公开的菜单同步方法（用于外部调用）
func (s *InitService) SyncMenus() error {
	return s.syncMenus()
}

// syncBuiltinRoles 同步内置角色（权限码推导模式）
func (s *InitService) syncBuiltinRoles() error {
	logger.Info("开始同步内置角色...")

	db := database.GetDB()

	// 定义需要同步的内置角色
	builtinRoles := []struct {
		code        string
		name        string
		description string
	}{
		{"admin", "超级管理员", "拥有系统所有权限"},
		{"ops", "运维工程师", "负责主机和任务管理"},
		{"auditor", "审计员", "仅拥有查看权限"},
		{"viewer", "查看者", "仅拥有查看权限"},
		{"user", "普通用户", "系统普通用户，拥有基础权限"},
		{"test", "测试角色", "用于测试的角色，拥有部分权限"},
	}

	// 同步或创建每个内置角色
	for _, builtinRole := range builtinRoles {
		var existingRole modelsystem.Role
		err := db.Where("code = ?", builtinRole.code).First(&existingRole).Error

		if err == nil {
			// 角色已存在：只更新名称和描述
			db.Model(&existingRole).Updates(map[string]interface{}{
				"name":        builtinRole.name,
				"description": builtinRole.description,
			})
			logger.Info("更新内置角色信息",
				zap.String("name", builtinRole.name),
				zap.String("code", builtinRole.code))
		} else {
			// 角色不存在，创建新角色
			newRole := modelsystem.Role{
				Name:        builtinRole.name,
				Code:        builtinRole.code,
				Description: builtinRole.description,
				Status:      1,
			}
			if err := db.Create(&newRole).Error; err != nil {
				logger.Error("创建内置角色失败",
					zap.String("name", builtinRole.name),
					zap.String("code", builtinRole.code),
					zap.Any("error", err))
				return err
			}
			logger.Info("创建内置角色（权限通过权限码配置）",
				zap.String("name", builtinRole.name),
				zap.String("code", builtinRole.code),
				zap.Uint("id", newRole.ID))
		}
	}

	logger.Info("内置角色同步完成")

	return nil
}

// initUsers 初始化用户数据
func (s *InitService) initUsers() error {
	db := database.GetDB()

	// 加密密码
	hashedPassword, err := utils.HashPassword("123456")
	if err != nil {
		return err
	}

	adminRoleIDs := []uint{1}
	adminRoleIDsJSON, _ := json.Marshal(adminRoleIDs)

	user := modelsystem.User{
		Username: "admin",
		Password: hashedPassword,
		Nickname: "超级管理员",
		Email:    "admin@example.com",
		RoleIDs:  string(adminRoleIDsJSON),
		Status:   "active",
		HomePath: "/home",
	}

	return db.Create(&user).Error
}

// initAttributes 初始化属性定义数据
func (s *InitService) initAttributes() error {
	logger.Info("开始初始化属性定义数据...")

	db := database.GetDB()

	attributes := []modelsystem.AttributeDefinition{
		{
			Name:        "业务系统",
			Key:         "business_system",
			Category:    "system",
			Type:        "select",
			Options:     `[{"value":"ecommerce","label":"电商系统"},{"value":"crm","label":"CRM系统"},{"value":"erp","label":"ERP系统"},{"value":"monitor","label":"监控系统"}]`,
			SortOrder:   1,
			Status:      1,
			Description: "主机所属的业务系统",
		},
		{
			Name:        "机房",
			Key:         "room",
			Category:    "location",
			Type:        "select",
			Options:     `[{"value":"hz","label":"杭州机房"},{"value":"bj","label":"北京机房"},{"value":"sh","label":"上海机房"},{"value":"sz","label":"深圳机房"}]`,
			SortOrder:   2,
			Status:      1,
			Description: "主机所在的机房",
		},
		{
			Name:        "机柜",
			Key:         "cabinet",
			Category:    "location",
			Type:        "text",
			SortOrder:   3,
			Status:      1,
			Description: "主机所在的机柜",
		},
		{
			Name:         "环境",
			Key:          "env",
			Category:     "environment",
			Type:         "select",
			Options:      `[{"value":"prod","label":"生产"},{"value":"test","label":"测试"},{"value":"dev","label":"开发"}]`,
			DefaultValue: "test",
			Required:     false,
			SortOrder:    4,
			Status:       1,
			Description:  "主机运行环境",
		},
		{
			Name:        "标签",
			Key:         "tags",
			Category:    "system",
			Type:        "multiselect",
			Options:     `[{"value":"important","label":"重要"},{"value":"backup","label":"备份节点"},{"value":"monitor","label":"监控节点"},{"value":"web","label":"Web服务器"},{"value":"db","label":"数据库服务器"}]`,
			SortOrder:   5,
			Status:      1,
			Description: "主机的标签分类",
		},
		{
			Name:        "所属项目",
			Key:         "project",
			Category:    "system",
			Type:        "text",
			SortOrder:   6,
			Status:      1,
			Description: "主机所属的项目",
		},
		{
			Name:        "购买日期",
			Key:         "purchase_date",
			Category:    "hardware",
			Type:        "date",
			SortOrder:   7,
			Status:      1,
			Description: "主机购买日期",
		},
		{
			Name:        "过保日期",
			Key:         "warranty_date",
			Category:    "hardware",
			Type:        "date",
			SortOrder:   8,
			Status:      1,
			Description: "主机过保日期",
		},
		{
			Name:        "责任人",
			Key:         "owner",
			Category:    "system",
			Type:        "text",
			SortOrder:   9,
			Status:      1,
			Description: "主机责任人",
		},
		{
			Name:        "联系方式",
			Key:         "contact",
			Category:    "system",
			Type:        "text",
			SortOrder:   10,
			Status:      1,
			Description: "责任人联系方式",
		},
		{
			Name:        "备注",
			Key:         "remark",
			Category:    "custom",
			Type:        "text",
			SortOrder:   11,
			Status:      1,
			Description: "主机备注信息",
		},
	}

	for _, attr := range attributes {
		if err := db.Create(&attr).Error; err != nil {
			logger.Error("创建属性定义失败",
				zap.String("name", attr.Name),
				zap.String("key", attr.Key),
				zap.Any("error", err))
			return err
		}
		logger.Info("创建属性定义",
			zap.String("name", attr.Name),
			zap.String("key", attr.Key),
			zap.Uint("id", attr.ID))
	}

	logger.Info("属性定义初始化完成", zap.Int("total", len(attributes)))
	return nil
}

// syncAttributes 同步属性定义（增量更新）
func (s *InitService) syncAttributes() error {
	logger.Info("开始同步属性定义数据...")

	db := database.GetDB()

	// 定义预置属性（通过 key 唯一标识）
	builtinAttributes := []struct {
		name        string
		key         string
		category    string
		attrType    string
		options     string
		defaultVal  string
		required    bool
		sortOrder   int
		description string
	}{
		{"业务系统", "business_system", "system", "select", `[{"value":"ecommerce","label":"电商系统"},{"value":"crm","label":"CRM系统"},{"value":"erp","label":"ERP系统"},{"value":"monitor","label":"监控系统"}]`, "", false, 1, "主机所属的业务系统"},
		{"机房", "room", "location", "select", `[{"value":"hz","label":"杭州机房"},{"value":"bj","label":"北京机房"},{"value":"sh","label":"上海机房"},{"value":"sz","label":"深圳机房"}]`, "", false, 2, "主机所在的机房"},
		{"机柜", "cabinet", "location", "text", "", "", false, 3, "主机所在的机柜"},
		{"环境", "env", "environment", "select", `[{"value":"prod","label":"生产"},{"value":"test","label":"测试"},{"value":"dev","label":"开发"}]`, "test", false, 4, "主机运行环境"},
		{"标签", "tags", "system", "multiselect", `[{"value":"important","label":"重要"},{"value":"backup","label":"备份节点"},{"value":"monitor","label":"监控节点"},{"value":"web","label":"Web服务器"},{"value":"db","label":"数据库服务器"}]`, "", false, 5, "主机的标签分类"},
		{"所属项目", "project", "system", "text", "", "", false, 6, "主机所属的项目"},
		{"购买日期", "purchase_date", "hardware", "date", "", "", false, 7, "主机购买日期"},
		{"过保日期", "warranty_date", "hardware", "date", "", "", false, 8, "主机过保日期"},
		{"责任人", "owner", "system", "text", "", "", false, 9, "主机责任人"},
		{"联系方式", "contact", "system", "text", "", "", false, 10, "责任人联系方式"},
		{"备注", "remark", "custom", "text", "", "", false, 11, "主机备注信息"},
	}

	addedCount := 0
	updatedCount := 0

	for _, builtinAttr := range builtinAttributes {
		var existingAttr modelsystem.AttributeDefinition
		err := db.Where(&modelsystem.AttributeDefinition{Key: builtinAttr.key}).First(&existingAttr).Error

		if err == nil {
			// 属性已存在，更新数据（保持数据同步）
			db.Model(&existingAttr).Updates(map[string]interface{}{
				"name":          builtinAttr.name,
				"category":      builtinAttr.category,
				"type":          builtinAttr.attrType,
				"options":       builtinAttr.options,
				"default_value": builtinAttr.defaultVal,
				"required":      builtinAttr.required,
				"sort_order":    builtinAttr.sortOrder,
				"status":        1,
				"description":   builtinAttr.description,
			})
			updatedCount++
			logger.Debug("更新属性定义",
				zap.String("name", builtinAttr.name),
				zap.String("key", builtinAttr.key))
		} else {
			// 属性不存在，添加新属性
			newAttr := modelsystem.AttributeDefinition{
				Name:         builtinAttr.name,
				Key:          builtinAttr.key,
				Category:     builtinAttr.category,
				Type:         builtinAttr.attrType,
				Options:      builtinAttr.options,
				DefaultValue: builtinAttr.defaultVal,
				Required:     builtinAttr.required,
				SortOrder:    builtinAttr.sortOrder,
				Status:       1,
				Description:  builtinAttr.description,
			}
			if err := db.Create(&newAttr).Error; err != nil {
				logger.Error("添加属性定义失败",
					zap.String("name", builtinAttr.name),
					zap.String("key", builtinAttr.key),
					zap.Any("error", err))
				return err
			}
			addedCount++
			logger.Info("添加属性定义",
				zap.String("name", builtinAttr.name),
				zap.String("key", builtinAttr.key),
				zap.Uint("id", newAttr.ID))
		}
	}

	logger.Info("属性定义同步完成",
		zap.Int("added", addedCount),
		zap.Int("updated", updatedCount),
		zap.Int("total", len(builtinAttributes)))
	return nil
}

// initAgentVersions 初始化 Agent 版本数据
func (s *InitService) initAgentVersions() error {
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

// assignDefaultPermissions 为内置角色分配默认权限
func (s *InitService) assignDefaultPermissions() error {
	logger.Info("开始为内置角色分配默认权限...")

	db := database.GetDB()

	// 定义角色默认权限映射（使用权限代码）
	rolePermissions := map[string][]string{
		"admin": {
			"*.*.*", // 超级管理员拥有所有权限（通配符）
		},
		"ops": {
			// 系统管理
			"system.user.list", "system.user.create", "system.user.update", "system.user.delete",
			"system.role.list", "system.role.update",
			"system.menu.list",
			// CMDB
			"cmdb.server.list", "cmdb.server.view", "cmdb.server.create", "cmdb.server.update", "cmdb.server.delete", "cmdb.server.connect",
			"cmdb.business.list", "cmdb.business.create", "cmdb.business.update", "cmdb.business.delete",
			"cmdb.rooms.list", "cmdb.rooms.create", "cmdb.rooms.update", "cmdb.rooms.delete",
			"cmdb.tags.list", "cmdb.tags.create", "cmdb.tags.update", "cmdb.tags.delete",
			"cmdb.group.list", "cmdb.group.view", "cmdb.group.create", "cmdb.group.update", "cmdb.group.delete", "cmdb.group.assign",
			"cmdb.agents.list", "cmdb.agents.deploy", "cmdb.agents.restart", "cmdb.agents.uninstall",
			// 监控
			"monitor.data.view", "monitor.data.export",
			"monitor.alert.list", "monitor.alert.ack", "monitor.alert.handle",
			"monitor.task.list", "monitor.task.view", "monitor.task.create", "monitor.task.update", "monitor.task.delete", "monitor.task.execute",
			// K8s
			"k8s.cluster.list", "k8s.cluster.view", "k8s.cluster.create", "k8s.cluster.update", "k8s.cluster.delete", "k8s.cluster.connect",
			"k8s.resource.view", "k8s.resource.create", "k8s.resource.update", "k8s.resource.delete",
			"k8s.permission.list", "k8s.permission.assign", "k8s.permission.revoke",
			// 审计
			"audit.login_log.list", "audit.login_log.export",
			"audit.operation_log.list", "audit.operation_log.export",
			"audit.system_event.list",
			"audit.stats.view",
		},
		"auditor": {
			// 只读权限（查看和列表）
			"cmdb.server.list", "cmdb.server.view",
			"cmdb.business.list",
			"cmdb.rooms.list",
			"cmdb.tags.list",
			"cmdb.group.list", "cmdb.group.view",
			"cmdb.agents.list",
			"monitor.data.view",
			"monitor.alert.list",
			"monitor.task.list", "monitor.task.view",
			"k8s.cluster.list", "k8s.cluster.view",
			"k8s.resource.view",
			"k8s.permission.list",
			"audit.login_log.list",
			"audit.operation_log.list",
			"audit.system_event.list",
			"audit.stats.view",
		},
		"viewer": {
			// 最小只读权限
			"cmdb.server.list",
			"monitor.data.view",
		},
		"user": {
			// 普通用户基础权限
			"monitor.data.view",
		},
		"test": {
			// 测试角色权限（用于测试）
			"system.user.list", "system.user.create",
			"system.role.list",
			"cmdb.server.list", "cmdb.server.create",
			"monitor.data.view",
			"k8s.cluster.list",
		},
	}

	for roleCode, permissionCodes := range rolePermissions {
		// 获取角色
		var role modelsystem.Role
		if err := db.Where("code = ?", roleCode).First(&role).Error; err != nil {
			logger.Warn("角色不存在，跳过权限分配",
				zap.String("code", roleCode),
				zap.Error(err))
			continue
		}

		// 获取权限
		var permissions []modelsystem.Permission
		if err := db.Where("code IN ?", permissionCodes).Find(&permissions).Error; err != nil {
			logger.Warn("查询权限失败",
				zap.String("role", roleCode),
				zap.Error(err))
			continue
		}

		// 分配权限
		assignedCount := 0
		for _, perm := range permissions {
			var rolePerm modelsystem.RolePermission
			err := db.Where("role_id = ? AND permission_id = ?", role.ID, perm.ID).
				First(&rolePerm).Error

			if err != nil {
				// 创建新的角色权限关联
				rolePerm = modelsystem.RolePermission{
					RoleID:       role.ID,
					PermissionID: perm.ID,
				}
				if err := db.Create(&rolePerm).Error; err != nil {
					logger.Error("分配权限失败",
						zap.String("role", roleCode),
						zap.String("permission", perm.Code),
						zap.Error(err))
				} else {
					assignedCount++
					logger.Debug("分配权限",
						zap.String("role", roleCode),
						zap.String("permission", perm.Code))
				}
			}
		}

		logger.Info("角色权限分配完成",
			zap.String("role", roleCode),
			zap.Int("总权限数", len(permissions)),
			zap.Int("新分配", assignedCount))
	}

	return nil
}

// initAPIPermissions 初始化层级权限代码到Casbin
func (s *InitService) initAPIPermissions() error {
	logger.Info("开始初始化层级权限代码到Casbin...")

	// 获取 PermissionService 实例
	permService, err := GetPermissionService()
	if err != nil {
		logger.Warn("获取PermissionService失败，跳过权限初始化", zap.Error(err))
		return nil // 不阻塞启动
	}

	// 清除所有旧的API路径格式的策略
	logger.Info("清除旧的API路径格式策略...")
	if err := permService.ClearAllPolicies(); err != nil {
		logger.Error("清除旧策略失败", zap.Error(err))
		return err
	}

	// 同步所有角色的权限到Casbin（使用层级权限代码）
	logger.Info("同步所有角色权限到Casbin...")
	if err := permService.SyncAllRolesToCasbin(); err != nil {
		logger.Error("同步角色权限失败", zap.Error(err))
		return err
	}

	// 验证策略数量
	policies := permService.GetAllPolicies()
	logger.Info("层级权限代码初始化完成", zap.Int("总策略数", len(policies)))

	return nil
}
