package services

import (
	"encoding/json"
	"oneops/backend/logger"
	"oneops/backend/models"
	"oneops/backend/utils"

	"go.uber.org/zap"
)

// InitService 初始化服务
type InitService struct{}

// NewInitService 创建初始化服务
func NewInitService() *InitService {
	return &InitService{}
}

// InitDatabase 初始化数据库（创建表和初始数据）
func (s *InitService) InitDatabase() error {
	// 在底层 sql.DB 上关闭外键检查（确保与 AutoMigrate 使用同一连接池生效）
	sqlDB, _ := db.DB()
	sqlDB.Exec("SET FOREIGN_KEY_CHECKS=0")
	defer sqlDB.Exec("SET FOREIGN_KEY_CHECKS=1")

	// 清理历史遗留的外键约束（忽略错误，约束不存在时正常失败）
	db.Exec("ALTER TABLE cabinets DROP FOREIGN KEY IF EXISTS cabinets_ibfk_1")
	db.Exec("ALTER TABLE cabinets DROP FOREIGN KEY IF EXISTS fk_server_rooms_cabinets")

	migrateErr := db.AutoMigrate(
		// 无外键依赖的基础表
		&models.User{},
		&models.Role{},
		&models.Menu{},
		&models.LoginLog{},
		&models.OperationLog{},
		&models.SystemEventLog{},
		&models.BusinessUnit{},
		&models.SSHCredential{},
		&models.AttributeDefinition{},
		// 有外键依赖的表（按依赖顺序）
		&models.ServerRoom{},
		&models.Cabinet{},
		&models.Server{},
		&models.ServerTag{},
		&models.ServerTagRelation{},
		&models.ServerGroup{},
		&models.ServerGroupRelation{},
		&models.ServerCredential{},
		&models.ServerAttribute{},
		&models.CloudServer{},
		&models.AssetChange{},
		// 堡垒机相关表
		&models.AssetAccessPolicy{},
		&models.BastionSession{},
		&models.BastionCommand{},
		&models.BastionFileTransfer{},
		&models.BastionApproval{},
	)

	if migrateErr != nil {
		// AutoMigrate 失败只记录警告，不阻止数据初始化（菜单/角色同步必须执行）
		logger.Warn("AutoMigrate 部分失败，继续执行数据初始化", zap.Error(migrateErr))
	}

	// 执行 SQL 迁移脚本（添加磁盘分区字段等）
	if err := s.runMigrations(); err != nil {
		logger.Warn("SQL 迁移执行失败，继续执行数据初始化", zap.Error(err))
	}

	// 无论 AutoMigrate 是否完全成功，都执行数据初始化
	return s.initData()
}

// runMigrations 执行 SQL 迁移脚本
func (s *InitService) runMigrations() error {
	logger.Info("开始执行 SQL 迁移...")

	// 添加 menu_ids 字段到 roles 表
	menuIDsSQL := `
		ALTER TABLE roles
		ADD COLUMN IF NOT EXISTS menu_ids JSON NULL
		COMMENT '菜单ID列表(JSON数组)，如 [1,2,3]'
		AFTER description;
	`

	if err := db.Exec(menuIDsSQL).Error; err != nil {
		logger.Debug("添加 menu_ids 字段（可能已存在）", zap.Error(err))
	}

	// 添加 disk_partitions 字段到 servers 表
	migrationSQL := `
		ALTER TABLE servers
		ADD COLUMN IF NOT EXISTS disk_partitions JSON NULL
		COMMENT '磁盘分区信息 [{"mount":"/","usage":80.5},{"mount":"/var","usage":90.2}]'
		AFTER disk_usage;
	`

	if err := db.Exec(migrationSQL).Error; err != nil {
		logger.Debug("添加 disk_partitions 字段（可能已存在）", zap.Error(err))
	}

	// 创建 agent_metrics 表（监控指标存储）
	createMetricsTable := `
		CREATE TABLE IF NOT EXISTS agent_metrics (
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
		CREATE TABLE IF NOT EXISTS agent_alerts (
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

	logger.Info("SQL 迁移执行完成")
	return nil
}

// initData 初始化和同步数据（自动检测并添加新菜单）
func (s *InitService) initData() error {
	// 同步菜单数据（增量更新，并清理已废弃菜单）
	if err := s.syncMenus(); err != nil {
		return err
	}

	// 检查角色表是否为空
	var roleCount int64
	db.Model(&models.Role{}).Count(&roleCount)
	if roleCount == 0 {
		if err := s.initRoles(); err != nil {
			return err
		}
	} else {
		// 同步角色权限（确保角色包含所有新菜单）
		if err := s.syncRoleMenus(); err != nil {
			return err
		}
	}

	// 检查用户表是否为空
	var userCount int64
	db.Model(&models.User{}).Count(&userCount)
	if userCount == 0 {
		if err := s.initUsers(); err != nil {
			return err
		}
	}

	// 检查属性定义表是否为空
	var attrCount int64
	db.Model(&models.AttributeDefinition{}).Count(&attrCount)
	if attrCount == 0 {
		if err := s.initAttributes(); err != nil {
			return err
		}
	} else {
		// 同步属性定义（确保包含所有预置属性）
		if err := s.syncAttributes(); err != nil {
			return err
		}
	}

	return nil
}

// initMenus 初始化菜单数据（动态路由模式）
func (s *InitService) initMenus() error {
	menus := []models.Menu{
		// 一级菜单
		{ID: 1, Name: "首页", Icon: "mdi:monitor-dashboard", Path: "/home", Permission: "", Sort: 1, Status: 1, ParentID: 0},
		{ID: 2, Name: "系统管理", Icon: "carbon:cloud-service-management", Path: "/manage", Permission: "", Sort: 2, Status: 1, ParentID: 0},
		{ID: 3, Name: "用户管理", Icon: "ic:round-manage-accounts", Path: "/manage/user", Permission: "system:user:query", Sort: 1, Status: 1, ParentID: 2},
		{ID: 4, Name: "角色管理", Icon: "carbon:user-role", Path: "/manage/role", Permission: "system:role:query", Sort: 2, Status: 1, ParentID: 2},
		{ID: 5, Name: "菜单管理", Icon: "material-symbols:route", Path: "/manage/menu", Permission: "system:menu:query", Sort: 3, Status: 1, ParentID: 2},
		{ID: 13, Name: "关于", Icon: "fluent:book-information-24-regular", Path: "/about", Permission: "", Sort: 5, Status: 1, ParentID: 0},
		{ID: 14, Name: "用户中心", Icon: "mdi:user-circle-outline", Path: "/user-center", Permission: "", Sort: 6, Status: 1, ParentID: 0},
		{ID: 20, Name: "资产管理", Icon: "mdi:server-network", Path: "/cmdb", Permission: "", Sort: 3, Status: 1, ParentID: 0},
		{ID: 21, Name: "服务器管理", Icon: "mdi:server", Path: "/cmdb/servers", Permission: "cmdb:server:query", Sort: 1, Status: 1, ParentID: 20},
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

	// 清除旧菜单（ID 20-50），重新写入新层级结构（包含监控中心）
	if err := db.Where("id >= 20 AND id <= 50").Delete(&models.Menu{}).Error; err != nil {
		logger.Error("清除旧菜单失败", zap.Error(err))
		return err
	}

	// 定义动态路由菜单（用于 SoybeanAdmin 动态路由模式）
	menus := []models.Menu{
		// 一级菜单
		{ID: 1, Name: "首页", Icon: "mdi:monitor-dashboard", Path: "/home", Permission: "", MenuType: "menu", Sort: 1, Status: 1, ParentID: 0},
		{ID: 2, Name: "系统管理", Icon: "carbon:cloud-service-management", Path: "/manage", Permission: "", MenuType: "directory", Sort: 2, Status: 1, ParentID: 0},
		{ID: 3, Name: "用户管理", Icon: "ic:round-manage-accounts", Path: "/manage/user", Permission: "system:user:query", Sort: 1, Status: 1, ParentID: 2, MenuType: "menu"},
		{ID: 4, Name: "角色管理", Icon: "carbon:user-role", Path: "/manage/role", Permission: "system:role:query", Sort: 2, Status: 1, ParentID: 2, MenuType: "menu"},
		{ID: 5, Name: "菜单管理", Icon: "material-symbols:route", Path: "/manage/menu", Permission: "system:menu:query", Sort: 3, Status: 1, ParentID: 2, MenuType: "menu"},
		{ID: 13, Name: "关于", Icon: "fluent:book-information-24-regular", Path: "/about", Permission: "", MenuType: "menu", Sort: 5, Status: 1, ParentID: 0},
		{ID: 14, Name: "用户中心", Icon: "mdi:user-circle-outline", Path: "/user-center", Permission: "", MenuType: "menu", Sort: 6, Status: 1, ParentID: 0},
		// 资产管理一级目录
		{ID: 20, Name: "资产管理", Icon: "mdi:server-network", Path: "/cmdb", Permission: "", MenuType: "directory", Sort: 3, Status: 1, ParentID: 0},
		// 资产管理二级菜单
		{ID: 21, Name: "资产总览", Icon: "mdi:view-dashboard", Path: "/cmdb/dashboard", Permission: "cmdb:server:query", MenuType: "menu", Sort: 1, Status: 1, ParentID: 20},
		{ID: 22, Name: "主机资产", Icon: "mdi:server", Path: "/cmdb/servers", Permission: "cmdb:server:query", MenuType: "menu", Sort: 2, Status: 1, ParentID: 20},
		// 访问控制目录
		{ID: 23, Name: "访问控制", Icon: "mdi:shield-lock", Path: "/cmdb/access", Permission: "", MenuType: "directory", Sort: 3, Status: 1, ParentID: 20},
		{ID: 24, Name: "访问策略", Icon: "mdi:file-lock", Path: "/cmdb/access/policies", Permission: "cmdb:access-policy:query", MenuType: "menu", Sort: 1, Status: 1, ParentID: 23},
		{ID: 25, Name: "凭证库", Icon: "mdi:key", Path: "/cmdb/access/credentials", Permission: "cmdb:credential:query", MenuType: "menu", Sort: 2, Status: 1, ParentID: 23},
		// 会话审计目录
		{ID: 26, Name: "会话审计", Icon: "mdi:clipboard-text-clock", Path: "/cmdb/audit", Permission: "", MenuType: "directory", Sort: 4, Status: 1, ParentID: 20},
		{ID: 27, Name: "在线会话", Icon: "mdi:monitor", Path: "/cmdb/audit/online", Permission: "cmdb:session:query", MenuType: "menu", Sort: 1, Status: 1, ParentID: 26},
		{ID: 28, Name: "历史会话", Icon: "mdi:history", Path: "/cmdb/audit/sessions", Permission: "cmdb:session:query", MenuType: "menu", Sort: 2, Status: 1, ParentID: 26},
		{ID: 29, Name: "命令审计", Icon: "mdi:console", Path: "/cmdb/audit/commands", Permission: "cmdb:command:query", MenuType: "menu", Sort: 3, Status: 1, ParentID: 26},
		// 资产配置目录
		{ID: 30, Name: "资产配置", Icon: "mdi:cog", Path: "/cmdb/config", Permission: "", MenuType: "directory", Sort: 5, Status: 1, ParentID: 20},
		{ID: 31, Name: "业务系统", Icon: "mdi:domain", Path: "/cmdb/config/business", Permission: "cmdb:business:query", MenuType: "menu", Sort: 1, Status: 1, ParentID: 30},
		{ID: 32, Name: "机房机柜", Icon: "mdi:office-building", Path: "/cmdb/config/rooms", Permission: "cmdb:room:query", MenuType: "menu", Sort: 2, Status: 1, ParentID: 30},
		{ID: 33, Name: "标签管理", Icon: "mdi:tag", Path: "/cmdb/config/tags", Permission: "cmdb:tag:query", MenuType: "menu", Sort: 3, Status: 1, ParentID: 30},
		{ID: 35, Name: "Agent 管理", Icon: "mdi:robot", Path: "/cmdb/config/agents", Permission: "cmdb:agent:query", MenuType: "menu", Sort: 4, Status: 1, ParentID: 30},
		{ID: 36, Name: "属性管理", Icon: "mdi:format-list-bulleted", Path: "/system/attributes", Permission: "system:attribute:query", MenuType: "menu", Sort: 5, Status: 1, ParentID: 30},
		// 资产变更
		{ID: 34, Name: "资产变更", Icon: "mdi:clock-edit", Path: "/cmdb/changes", Permission: "cmdb:change:query", MenuType: "menu", Sort: 6, Status: 1, ParentID: 20},
		// 监控中心一级目录
		{ID: 40, Name: "监控中心", Icon: "mdi:chart-line", Path: "/monitoring", Permission: "", MenuType: "directory", Sort: 4, Status: 1, ParentID: 0},
		// 监控中心二级菜单
		{ID: 41, Name: "监控概览", Icon: "mdi:gauge", Path: "/monitoring/overview", Permission: "monitoring:overview:view", MenuType: "menu", Sort: 1, Status: 1, ParentID: 40},
		{ID: 42, Name: "主机监控", Icon: "mdi:server", Path: "/monitoring/servers", Permission: "monitoring:servers:view", MenuType: "menu", Sort: 2, Status: 1, ParentID: 40},
			{ID: 43, Name: "趋势分析", Icon: "mdi:chart-areaspline", Path: "/monitoring/trends", Permission: "monitoring:trends:view", MenuType: "menu", Sort: 3, Status: 1, ParentID: 40},
			{ID: 44, Name: "告警管理", Icon: "mdi:bell-alert", Path: "/monitoring/alerts", Permission: "monitoring:alerts:view", MenuType: "menu", Sort: 4, Status: 1, ParentID: 40},
			{ID: 45, Name: "监控设置", Icon: "mdi:cog", Path: "/monitoring/settings", Permission: "monitoring:settings:view", MenuType: "menu", Sort: 5, Status: 1, ParentID: 40},
			{ID: 46, Name: "巡检报告", Icon: "mdi:file-document", Path: "/monitoring/reports", Permission: "monitoring:reports:view", MenuType: "menu", Sort: 6, Status: 1, ParentID: 40},
			// 终端管理一级目录
			{ID: 47, Name: "终端管理", Icon: "mdi:console", Path: "/terminal", Permission: "", MenuType: "directory", Sort: 5, Status: 1, ParentID: 0},
			// 终端管理二级菜单
			{ID: 48, Name: "终端工作台", Icon: "mdi:monitor-dashboard", Path: "/terminal/workbench", Permission: "terminal:workbench:view", MenuType: "menu", Sort: 1, Status: 1, ParentID: 47},
		}

	addedCount := 0
	updatedCount := 0

	for _, menu := range menus {
		var existingMenu models.Menu
		err := db.Where("id = ?", menu.ID).First(&existingMenu).Error

		if err == nil {
			// 菜单已存在，更新数据（保持数据同步）
			db.Model(&existingMenu).Updates(map[string]interface{}{
				"name":       menu.Name,
				"icon":       menu.Icon,
				"path":       menu.Path,
				"permission": menu.Permission,
				"parent_id":  menu.ParentID,
				"sort":       menu.Sort,
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

	deletedResult := db.Where("path = ? OR name IN ?", "/cmdb/groups", []string{"主机分组", "cmdb_groups"}).Delete(&models.Menu{})
	if deletedResult.Error != nil {
		logger.Error("删除废弃主机分组菜单失败", zap.Error(deletedResult.Error))
		return deletedResult.Error
	}
	if deletedResult.RowsAffected > 0 {
		logger.Info("已删除废弃主机分组菜单", zap.Int64("count", deletedResult.RowsAffected))
	}

	logger.Info("菜单同步完成",
		zap.Int("added", addedCount),
		zap.Int("updated", updatedCount),
		zap.Int("total", len(menus)))

	return nil
}

// syncRoleMenus 同步角色菜单权限
func (s *InitService) syncRoleMenus() error {
	logger.Info("开始同步角色菜单权限...")

	// 定义5个内置角色的菜单权限（动态路由模式）
	// 菜单ID映射：1=首页, 2=系统管理, 20=资产管理, 40=监控中心
	adminMenuIDs := []uint{1, 2, 3, 4, 5, 13, 14, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 40, 41, 42, 43, 44, 45, 46} // 超级管理员：所有权限
	opsMenuIDs := []uint{1, 13, 14, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 40, 41, 42, 43, 44, 45, 46}               // 运维工程师：含监控权限
	auditorMenuIDs := []uint{1, 13, 20, 21, 22, 26, 27, 28, 29, 34, 40, 41, 42, 43, 44}                                                    // 审计员：含监控查看权限
	userMenuIDs := []uint{1}                                                                                                            // 普通用户：仅首页
	testMenuIDs := []uint{1, 13, 20, 21, 22, 26, 27, 28, 29, 34, 40, 41, 42, 43, 44, 45, 46}                                                       // 测试角色：含监控权限

	adminMenuIDsJSON, _ := json.Marshal(adminMenuIDs)
	opsMenuIDsJSON, _ := json.Marshal(opsMenuIDs)
	auditorMenuIDsJSON, _ := json.Marshal(auditorMenuIDs)
	userMenuIDsJSON, _ := json.Marshal(userMenuIDs)
	testMenuIDsJSON, _ := json.Marshal(testMenuIDs)

	// 定义需要同步的内置角色
	builtinRoles := []struct {
		code        string
		name        string
		description string
		menuIDs     []uint
		menuIDsJSON []byte
	}{
		{"admin", "超级管理员", "拥有系统所有权限", adminMenuIDs, adminMenuIDsJSON},
		{"ops", "运维工程师", "负责主机和任务管理", opsMenuIDs, opsMenuIDsJSON},
		{"auditor", "审计员", "仅拥有查看权限", auditorMenuIDs, auditorMenuIDsJSON},
		{"user", "普通用户", "系统普通用户，拥有基础权限", userMenuIDs, userMenuIDsJSON},
		{"test", "测试角色", "用于测试的角色，拥有部分权限", testMenuIDs, testMenuIDsJSON},
	}

	// 同步或创建每个内置角色
	for _, builtinRole := range builtinRoles {
		var existingRole models.Role
		err := db.Where("code = ?", builtinRole.code).First(&existingRole).Error

		if err == nil {
			// 角色存在，更新权限
			db.Model(&existingRole).Updates(map[string]interface{}{
				"name":        builtinRole.name,
				"description": builtinRole.description,
				"menu_ids":    string(builtinRole.menuIDsJSON),
			})
			logger.Info("更新内置角色权限",
				zap.String("name", builtinRole.name),
				zap.String("code", builtinRole.code),
				zap.Int("menus", len(builtinRole.menuIDs)))
		} else {
			// 角色不存在，创建新角色
			newRole := models.Role{
				Name:        builtinRole.name,
				Code:        builtinRole.code,
				Description: builtinRole.description,
				MenuIDs:     string(builtinRole.menuIDsJSON),
				Status:      1,
			}
			if err := db.Create(&newRole).Error; err != nil {
				logger.Error("创建内置角色失败",
					zap.String("name", builtinRole.name),
					zap.String("code", builtinRole.code),
					zap.Any("error", err))
				return err
			}
			logger.Info("创建内置角色",
				zap.String("name", builtinRole.name),
				zap.String("code", builtinRole.code),
				zap.Uint("id", newRole.ID))
		}
	}

	logger.Info("角色菜单权限同步完成")
	return nil
}

// initRoles 初始化角色数据
func (s *InitService) initRoles() error {
	// 定义5个内置角色的菜单权限（动态路由模式）
	// 菜单ID映射：1=首页, 2=系统管理, 20=资产管理, 40=监控中心
	adminMenuIDs := []uint{1, 2, 3, 4, 5, 13, 14, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 40, 41, 42, 43, 44, 45, 46} // 超级管理员：所有权限
	opsMenuIDs := []uint{1, 13, 14, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 40, 41, 42, 43, 44, 45, 46}               // 运维工程师：含监控权限
	auditorMenuIDs := []uint{1, 13, 20, 21, 22, 26, 27, 28, 29, 34, 40, 41, 42, 43, 44}                                                    // 审计员：含监控查看权限
	userMenuIDs := []uint{1}                                                                                                            // 普通用户：仅首页
	testMenuIDs := []uint{1, 13, 20, 21, 22, 26, 27, 28, 29, 34, 40, 41, 42, 43, 44, 45, 46}                                                       // 测试角色：含监控权限

	adminMenuIDsJSON, _ := json.Marshal(adminMenuIDs)
	opsMenuIDsJSON, _ := json.Marshal(opsMenuIDs)
	auditorMenuIDsJSON, _ := json.Marshal(auditorMenuIDs)
	userMenuIDsJSON, _ := json.Marshal(userMenuIDs)
	testMenuIDsJSON, _ := json.Marshal(testMenuIDs)

	// 5个系统内置角色
	roles := []models.Role{
		{Name: "超级管理员", Code: "admin", Description: "拥有系统所有权限", MenuIDs: string(adminMenuIDsJSON), Status: 1},
		{Name: "运维工程师", Code: "ops", Description: "负责主机和任务管理", MenuIDs: string(opsMenuIDsJSON), Status: 1},
		{Name: "审计员", Code: "auditor", Description: "仅拥有查看权限", MenuIDs: string(auditorMenuIDsJSON), Status: 1},
		{Name: "普通用户", Code: "user", Description: "系统普通用户，拥有基础权限", MenuIDs: string(userMenuIDsJSON), Status: 1},
		{Name: "测试角色", Code: "test", Description: "用于测试的角色，拥有部分权限", MenuIDs: string(testMenuIDsJSON), Status: 1},
	}

	for _, role := range roles {
		if err := db.Create(&role).Error; err != nil {
			logger.Error("创建角色失败",
				zap.String("name", role.Name),
				zap.String("code", role.Code),
				zap.Any("error", err))
			return err
		}
		logger.Info("创建内置角色",
			zap.String("name", role.Name),
			zap.String("code", role.Code),
			zap.Uint("id", role.ID))
	}

	return nil
}

// initUsers 初始化用户数据
func (s *InitService) initUsers() error {
	// 加密密码
	hashedPassword, err := utils.HashPassword("123456")
	if err != nil {
		return err
	}

	adminRoleIDs := []uint{1}
	adminRoleIDsJSON, _ := json.Marshal(adminRoleIDs)

	user := models.User{
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

	attributes := []models.AttributeDefinition{
		{
			Name:     "业务系统",
			Key:      "business_system",
			Category: "system",
			Type:     "select",
			Options:  `[{"value":"ecommerce","label":"电商系统"},{"value":"crm","label":"CRM系统"},{"value":"erp","label":"ERP系统"},{"value":"monitor","label":"监控系统"}]`,
			SortOrder: 1,
			Status:    1,
			Description: "主机所属的业务系统",
		},
		{
			Name:     "机房",
			Key:      "room",
			Category: "location",
			Type:     "select",
			Options:  `[{"value":"hz","label":"杭州机房"},{"value":"bj","label":"北京机房"},{"value":"sh","label":"上海机房"},{"value":"sz","label":"深圳机房"}]`,
			SortOrder: 2,
			Status:    1,
			Description: "主机所在的机房",
		},
		{
			Name:     "机柜",
			Key:      "cabinet",
			Category: "location",
			Type:     "text",
			SortOrder: 3,
			Status:    1,
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
			Name:     "标签",
			Key:      "tags",
			Category: "system",
			Type:     "multiselect",
			Options:  `[{"value":"important","label":"重要"},{"value":"backup","label":"备份节点"},{"value":"monitor","label":"监控节点"},{"value":"web","label":"Web服务器"},{"value":"db","label":"数据库服务器"}]`,
			SortOrder: 5,
			Status:    1,
			Description: "主机的标签分类",
		},
		{
			Name:     "所属项目",
			Key:      "project",
			Category: "system",
			Type:     "text",
			SortOrder: 6,
			Status:    1,
			Description: "主机所属的项目",
		},
		{
			Name:     "购买日期",
			Key:      "purchase_date",
			Category: "hardware",
			Type:     "date",
			SortOrder: 7,
			Status:    1,
			Description: "主机购买日期",
		},
		{
			Name:     "过保日期",
			Key:      "warranty_date",
			Category: "hardware",
			Type:     "date",
			SortOrder: 8,
			Status:    1,
			Description: "主机过保日期",
		},
		{
			Name:     "责任人",
			Key:      "owner",
			Category: "system",
			Type:     "text",
			SortOrder: 9,
			Status:    1,
			Description: "主机责任人",
		},
		{
			Name:     "联系方式",
			Key:      "contact",
			Category: "system",
			Type:     "text",
			SortOrder: 10,
			Status:    1,
			Description: "责任人联系方式",
		},
		{
			Name:     "备注",
			Key:      "remark",
			Category: "custom",
			Type:     "text",
			SortOrder: 11,
			Status:    1,
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
		var existingAttr models.AttributeDefinition
		// 使用 GORM 的方式查询，避免 SQL 保留字问题
		err := db.Where(&models.AttributeDefinition{Key: builtinAttr.key}).First(&existingAttr).Error

		if err == nil {
			// 属性已存在，更新数据（保持数据同步）
			db.Model(&existingAttr).Updates(map[string]interface{}{
				"name":         builtinAttr.name,
				"category":     builtinAttr.category,
				"type":         builtinAttr.attrType,
				"options":      builtinAttr.options,
				"default_value": builtinAttr.defaultVal,
				"required":     builtinAttr.required,
				"sort_order":   builtinAttr.sortOrder,
				"status":       1,
				"description":  builtinAttr.description,
			})
			updatedCount++
			logger.Debug("更新属性定义",
				zap.String("name", builtinAttr.name),
				zap.String("key", builtinAttr.key))
		} else {
			// 属性不存在，添加新属性
			newAttr := models.AttributeDefinition{
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
