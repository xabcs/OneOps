package system

import (
	"encoding/json"
	"fmt"

	modelaudit "oneops/backend3/model/audit"
	modelauth "oneops/backend3/model/authorization"
	modelcmdb "oneops/backend3/model/cmdb"
	modelk8s "oneops/backend3/model/k8s"
	modelsystem "oneops/backend3/model/system"
	modelticket "oneops/backend3/model/ticket"
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
	logger.Info("开始数据库初始化流程...")

	// 阶段1：数据库模式迁移
	if err := i.migrateSchema(); err != nil {
		return fmt.Errorf("模式迁移失败: %w", err)
	}

	// 阶段1.5：用户-角色关系迁移（role_ids JSON 列 → sys_user_roles 关联表）
	if err := i.migrateUserRoleRelations(); err != nil {
		return fmt.Errorf("用户角色关系迁移失败: %w", err)
	}

	// 阶段2：创建监控表
	if err := i.runMigrations(); err != nil {
		logger.Warn("监控表创建失败，继续执行", zap.Error(err))
	}

	// 阶段3：初始化系统数据（菜单、角色、用户、权限等）
	if err := i.initSystemData(); err != nil {
		return fmt.Errorf("系统数据初始化失败: %w", err)
	}

	logger.Info("数据库初始化流程完成")
	return nil
}

// migrateSchema 迁移数据库模式
func (i *Initializer) migrateSchema() error {
	logger.Info("阶段1：开始数据库模式迁移...")

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
		&modelsystem.PermissionRoute{},
		&modelsystem.UserGroup{},
		&modelsystem.UserGroupMember{},
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
		// 工单系统
		&modelticket.Workflow{},
		&modelticket.WorkflowNode{},
		&modelticket.TicketType{},
		&modelticket.Ticket{},
		&modelticket.TicketNodeRecord{},
		&modelticket.TicketFlowLog{},
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
		&modelk8s.K8sClusterRole{},
		&modelk8s.K8sClusterRolePermission{},
		&modelk8s.ClusterRoleBinding{},
		&modelk8s.ClusterGroupBinding{},
		&modelk8s.K8sNativeRoleBinding{},
		&modelk8s.K8sSession{},
		&modelk8s.K8sCommand{},
	); err != nil {
		logger.Warn("K8s表迁移失败", zap.Error(err))
	}

	logger.Info("数据库模式迁移完成")
	return nil
}

// runMigrations 创建无 GORM model 的监控表（原生 SQL 访问）
func (i *Initializer) runMigrations() error {
	logger.Info("开始创建监控表...")

	db := database.GetDB()

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

	logger.Info("监控表创建完成")
	return nil
}

// migrateUserRoleRelations 用户-角色关系迁移：
// 历史实现将角色 ID 存于 sys_users.role_ids JSON 列，现改为 sys_user_roles 关联表（many2many）。
// 启动时将 JSON 数据幂等回填至关联表，全部成功后删除旧列。旧列不存在（新库）则跳过。
func (i *Initializer) migrateUserRoleRelations() error {
	db := database.GetDB()

	if !db.Migrator().HasTable("sys_users") {
		return nil
	}
	// AutoMigrate 已按 many2many 标签建好关联表；此处仅防御性确认
	if !db.Migrator().HasTable("sys_user_roles") {
		if err := db.Exec(`CREATE TABLE IF NOT EXISTS sys_user_roles (
			user_id BIGINT UNSIGNED NOT NULL,
			role_id BIGINT UNSIGNED NOT NULL,
			PRIMARY KEY (user_id, role_id)
		)`).Error; err != nil {
			return fmt.Errorf("创建 sys_user_roles 失败: %w", err)
		}
	}

	if !db.Migrator().HasColumn("sys_users", "role_ids") {
		return nil // 新库或已迁移
	}

	type legacyRow struct {
		ID      uint   `gorm:"column:id"`
		RoleIDs string `gorm:"column:role_ids"`
	}
	var rows []legacyRow
	if err := db.Raw("SELECT id, role_ids FROM sys_users WHERE role_ids IS NOT NULL AND role_ids NOT IN ('', 'null', '[]')").Scan(&rows).Error; err != nil {
		return fmt.Errorf("读取旧 role_ids 失败: %w", err)
	}

	migrated := 0
	for _, row := range rows {
		var roleIDs []uint
		if err := json.Unmarshal([]byte(row.RoleIDs), &roleIDs); err != nil {
			logger.Warn("用户 role_ids 解析失败，跳过",
				zap.Uint("user_id", row.ID), zap.String("role_ids", row.RoleIDs), zap.Error(err))
			continue
		}
		for _, roleID := range roleIDs {
			if err := db.Exec("INSERT IGNORE INTO sys_user_roles (user_id, role_id) VALUES (?, ?)",
				row.ID, roleID).Error; err != nil {
				return fmt.Errorf("写入关联表失败 user_id=%d role_id=%d: %w", row.ID, roleID, err)
			}
			migrated++
		}
	}

	if err := db.Migrator().DropColumn("sys_users", "role_ids"); err != nil {
		return fmt.Errorf("删除旧列 sys_users.role_ids 失败: %w", err)
	}

	logger.Info("用户-角色关系迁移完成", zap.Int("rows", len(rows)), zap.Int("relations", migrated))
	return nil
}

// initSystemData 初始化系统数据（菜单、角色、用户、权限、诊断、Agent等）
func (i *Initializer) initSystemData() error {
	logger.Info("初始化系统数据...")

	// 同步菜单
	if err := i.syncMenus(); err != nil {
		logger.Warn("菜单同步失败", zap.Error(err))
	}

	// 同步内置角色
	if err := i.syncBuiltinRoles(); err != nil {
		logger.Warn("内置角色同步失败", zap.Error(err))
	}

	// 同步管理员用户
	if err := i.syncUsers(); err != nil {
		logger.Warn("管理员用户同步失败", zap.Error(err))
	}

	// 同步属性定义
	if err := i.syncAttributes(); err != nil {
		logger.Warn("属性定义同步失败", zap.Error(err))
	}

	// 同步工单示例数据（流程定义/工单类型）
	if err := i.syncTicketSeeds(); err != nil {
		logger.Warn("工单示例数据同步失败", zap.Error(err))
	}

	// 同步权限数据
	if err := i.syncPermissions(); err != nil {
		logger.Warn("权限数据同步失败", zap.Error(err))
	}

	// 同步权限树层级（Level 1/2 分组节点 + Level 3 parent_id 回填，依赖权限目录）
	if err := i.syncPermissionHierarchy(); err != nil {
		logger.Warn("权限层级同步失败", zap.Error(err))
	}

	// 同步权限路由映射（依赖权限目录，必须在 syncPermissions 之后）
	if err := i.syncPermissionRoutes(); err != nil {
		logger.Warn("权限路由映射同步失败", zap.Error(err))
	}

	// 菜单 resource 与权限目录一致性校验（防演进错位：resource 在权限码中无对应项则告警）
	i.validateMenuResourceConsistency()

	// 同步默认角色权限分配
	if err := i.syncDefaultPermissions(); err != nil {
		logger.Warn("默认权限分配失败", zap.Error(err))
	}

	// 同步诊断数据
	if err := i.syncDiagnosticData(); err != nil {
		logger.Warn("诊断数据同步失败", zap.Error(err))
	}

	// 同步 Agent 版本
	if err := i.syncAgentVersions(); err != nil {
		logger.Warn("Agent版本同步失败", zap.Error(err))
	}

	// 同步 Casbin 策略
	if err := i.syncAPIPermissions(); err != nil {
		logger.Warn("Casbin策略同步失败", zap.Error(err))
	}

	return nil
}

// validateMenuResourceConsistency 校验启用菜单的 resource 在权限目录中存在对应权限码
// 菜单可见性按"权限码第二段 == 菜单 resource"匹配，resource 错位会导致菜单对所有非管理员不可见
func (i *Initializer) validateMenuResourceConsistency() {
	db := database.GetDB()

	var menus []modelsystem.Menu
	if err := db.Where("status = 1 AND resource != ''").Find(&menus).Error; err != nil {
		logger.Warn("菜单一致性校验失败：查询菜单失败", zap.Error(err))
		return
	}

	for _, m := range menus {
		var count int64
		db.Model(&modelsystem.Permission{}).
			Where("resource = ? AND status = 1", m.Resource).
			Count(&count)
		if count == 0 {
			logger.Warn("菜单 resource 在权限目录中无对应权限码，该菜单对所有非管理员不可见",
				zap.Uint("menu_id", m.ID),
				zap.String("menu_name", m.Name),
				zap.String("menu_path", m.Path),
				zap.String("resource", m.Resource))
		}
	}
}
