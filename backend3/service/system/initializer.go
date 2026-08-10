package system

import (
	"fmt"
	"os"

	modelaudit "oneops/backend3/model/audit"
	modelauth "oneops/backend3/model/authorization"
	modelcmdb "oneops/backend3/model/cmdb"
	modelk8s "oneops/backend3/model/k8s"
	modelsystem "oneops/backend3/model/system"
	"oneops/backend3/pkg/database"
	"oneops/backend3/pkg/logger"

	"go.uber.org/zap"
)

// Initializer 初始化协调器
type Initializer struct{}

// NewInitializer 创建初始化协调器
func NewInitializer() *Initializer {
	return &Initializer{}
}

// Initialize 执行初始化
func (i *Initializer) Initialize() error {
	zap.L().Info("开始数据库初始化流程...")

	// 阶段1：数据库模式迁移
	if err := i.migrateSchema(); err != nil {
		return fmt.Errorf("模式迁移失败: %w", err)
	}

	// 阶段2：执行SQL迁移脚本
	if err := i.runMigrations(); err != nil {
		zap.L().Warn("SQL迁移执行失败，继续执行", zap.Error(err))
	}

	// 阶段3：基础数据初始化
	if err := i.initSeedData(); err != nil {
		return fmt.Errorf("基础数据初始化失败: %w", err)
	}

	// 阶段4：模块数据初始化
	if err := i.initModuleData(); err != nil {
		return fmt.Errorf("模块数据初始化失败: %w", err)
	}

	zap.L().Info("数据库初始化流程完成")
	return nil
}

// migrateSchema 迁移数据库模式
func (i *Initializer) migrateSchema() error {
	zap.L().Info("阶段1：开始数据库模式迁移...")

	db := database.GetDB()

	// 关闭外键检查
	sqlDB, _ := db.DB()
	sqlDB.Exec("SET FOREIGN_KEY_CHECKS=0")
	defer sqlDB.Exec("SET FOREIGN_KEY_CHECKS=1")

	// 清理历史遗留的外键约束
	db.Exec("ALTER TABLE cabinets DROP FOREIGN KEY cabinets_ibfk_1")
	db.Exec("ALTER TABLE cabinets DROP FOREIGN KEY fk_server_rooms_cabinets")

	// 基础表（无外键依赖）
	if err := db.AutoMigrate(
		&modelsystem.User{},
		&modelsystem.Role{},
		&modelsystem.Menu{},
		&modelsystem.Permission{},
		&modelsystem.RolePermission{},
		&modelsystem.UserPermission{},
		&modelsystem.PermissionLog{},
		&modelsystem.APIResource{},
		&modelaudit.LoginLog{},
		&modelaudit.OperationLog{},
		&modelaudit.SystemEventLog{},
		&modelcmdb.BusinessUnit{},
		&modelcmdb.SSHCredential{},
		&modelsystem.AttributeDefinition{},
		&modelcmdb.AgentVersion{},
		&modelcmdb.AgentUpgradeTask{},
		&modelk8s.DiagnosticHistory{},
		&modelk8s.DiagnosticConfig{},
		&modelk8s.DiagnosticPermission{},
		&modelauth.Application{},
		&modelauth.ApplicationRole{},
		&modelauth.ApplicationUser{},
		&modelauth.ApplicationGroup{},
		&modelauth.ApplicationAuthorizationRule{},
		&modelauth.GroupBinding{},
		&modelauth.AuthUser{},
		&modelauth.AuthGroup{},
		&modelauth.AuthUserGroup{},
		&modelauth.AuthGroupPermissionMapping{},
		&modelauth.ApplicationOperationLog{},
		&modelauth.UserIdentityMapping{},
		&modelauth.GroupBindingExecution{},
		&modelauth.PermissionAssignmentStatus{},
	); err != nil {
		return fmt.Errorf("基础表迁移失败: %w", err)
	}

	// 依赖表（按依赖顺序）
	if err := db.AutoMigrate(
		&modelcmdb.ServerRoom{},
		&modelcmdb.Cabinet{},
		&modelcmdb.Server{},
		&modelcmdb.ServerTag{},
		&modelcmdb.ServerTagRelation{},
		&modelcmdb.AssetChange{},
		&modelcmdb.ServerGroup{},
		&modelcmdb.ServerGroupRelation{},
		&modelcmdb.ServerCredential{},
		&modelcmdb.ServerAttribute{},
		&modelcmdb.CloudServer{},
		&modelcmdb.AssetAccessPolicy{},
		&modelcmdb.BastionSession{},
		&modelcmdb.BastionCommand{},
		&modelcmdb.BastionFileTransfer{},
		&modelcmdb.BastionApproval{},
	); err != nil {
		return fmt.Errorf("依赖表迁移失败: %w", err)
	}

	// K8s相关表
	if err := db.AutoMigrate(
		&modelk8s.K8sCluster{},
		&modelk8s.ClusterRoleBinding{},
		&modelk8s.K8sSession{},
		&modelk8s.K8sCommand{},
	); err != nil {
		zap.L().Warn("K8s表迁移失败", zap.Error(err))
	}

	zap.L().Info("数据库模式迁移完成")
	return nil
}

// runMigrations 执行 SQL 迁移脚本
func (i *Initializer) runMigrations() error {
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

// initSeedData 初始化基础数据
func (i *Initializer) initSeedData() error {
	zap.L().Info("阶段3：初始化基础数据...")

	// 同步菜单
	if err := i.syncMenus(); err != nil {
		zap.L().Warn("菜单同步失败", zap.Error(err))
	}

	// 同步内置角色
	if err := i.syncBuiltinRoles(); err != nil {
		zap.L().Warn("内置角色同步失败", zap.Error(err))
	}

	// 初始化管理员用户
	if err := i.initUsers(); err != nil {
		zap.L().Warn("管理员用户初始化失败", zap.Error(err))
	}

	// 同步属性定义
	if err := i.syncAttributes(); err != nil {
		zap.L().Warn("属性定义同步失败", zap.Error(err))
	}

	return nil
}

// initModuleData 初始化模块数据
func (i *Initializer) initModuleData() error {
	zap.L().Info("阶段4：初始化模块数据...")

	// 初始化权限数据
	if err := i.initPermissionsFromSQL(); err != nil {
		zap.L().Warn("权限数据初始化失败", zap.Error(err))
	}

	// 分配默认权限
	if err := i.assignDefaultPermissions(); err != nil {
		zap.L().Warn("默认权限分配失败", zap.Error(err))
	}

	// 初始化诊断数据
	if err := i.initDiagnosticData(); err != nil {
		zap.L().Warn("诊断数据初始化失败", zap.Error(err))
	}

	// 初始化Agent版本
	if err := i.initAgentVersions(); err != nil {
		zap.L().Warn("Agent版本初始化失败", zap.Error(err))
	}

	// 初始化API权限
	if err := i.initAPIPermissions(); err != nil {
		zap.L().Warn("API权限初始化失败", zap.Error(err))
	}

	return nil
}

// initPermissionsFromSQL 从 SQL 文件初始化权限数据
func (i *Initializer) initPermissionsFromSQL() error {
	zap.L().Info("从 SQL 文件初始化权限数据...")

	db := database.GetDB()

	// SQL 文件路径
	sqlFile := "migrations/v2_permissions.sql"

	// 读取 SQL 文件内容
	content, err := os.ReadFile(sqlFile)
	if err != nil {
		return fmt.Errorf("读取 SQL 文件失败: %w", err)
	}

	// 执行 SQL
	if err := db.Exec(string(content)).Error; err != nil {
		return fmt.Errorf("执行权限初始化 SQL 失败: %w", err)
	}

	// 统计权限数量
	var count int64
	if err := db.Model(&modelsystem.Permission{}).Count(&count).Error; err != nil {
		zap.L().Warn("统计权限数量失败", zap.Error(err))
	} else {
		zap.L().Info("权限数据初始化完成", zap.Int64("total_permissions", count))
	}

	return nil
}
