package services

import (
	"fmt"
	"oneops/backend/models"
	"os"

	"go.uber.org/zap"
)

// Initializer 初始化协调器
type Initializer struct {
	initService *InitService
}

// NewInitializer 创建初始化协调器
func NewInitializer() *Initializer {
	return &Initializer{
		initService: NewInitService(),
	}
}

// Initialize 执行初始化
func (i *Initializer) Initialize() error {
	zap.L().Info("开始数据库初始化流程...")

	// 阶段1：数据库模式迁移
	if err := i.migrateSchema(); err != nil {
		return fmt.Errorf("模式迁移失败: %w", err)
	}

	// 阶段2：执行SQL迁移脚本
	if err := i.initService.runMigrations(); err != nil {
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

	// 关闭外键检查
	sqlDB, _ := db.DB()
	sqlDB.Exec("SET FOREIGN_KEY_CHECKS=0")
	defer sqlDB.Exec("SET FOREIGN_KEY_CHECKS=1")

	// 清理历史遗留的外键约束
	db.Exec("ALTER TABLE cabinets DROP FOREIGN KEY cabinets_ibfk_1")
	db.Exec("ALTER TABLE cabinets DROP FOREIGN KEY fk_server_rooms_cabinets")

	// 基础表（无外键依赖）
	if err := db.AutoMigrate(
		&models.User{},
		&models.Role{},
		&models.Menu{},
		&models.Permission{},
		&models.RolePermission{},
		&models.UserPermission{},
		&models.PermissionLog{},
		&models.APIResource{},
		&models.LoginLog{},
		&models.OperationLog{},
		&models.SystemEventLog{},
		&models.BusinessUnit{},
		&models.SSHCredential{},
		&models.AttributeDefinition{},
		&models.AgentVersion{},
		&models.AgentUpgradeTask{},
		&models.DiagnosticHistory{},
		&models.DiagnosticConfig{},
		&models.DiagnosticPermission{},
		&models.Application{},
		&models.ApplicationRole{},
		&models.ApplicationUser{},
		&models.ApplicationGroup{},
		&models.ApplicationAuthorizationRule{},
		&models.GroupBinding{},
		&models.AuthUser{},
		&models.AuthGroup{},
		&models.AuthUserGroup{},
		&models.AuthGroupPermissionMapping{},
		&models.ApplicationOperationLog{},
		&models.UserIdentityMapping{},
		&models.GroupBindingExecution{},
		&models.PermissionAssignmentStatus{},
	); err != nil {
		return fmt.Errorf("基础表迁移失败: %w", err)
	}

	// 依赖表（按依赖顺序）
	if err := db.AutoMigrate(
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
		&models.AssetAccessPolicy{},
		&models.BastionSession{},
		&models.BastionCommand{},
		&models.BastionFileTransfer{},
		&models.BastionApproval{},
	); err != nil {
		return fmt.Errorf("依赖表迁移失败: %w", err)
	}

	// K8s相关表
	if err := db.AutoMigrate(
		&models.K8sCluster{},
		&models.ClusterRoleBinding{},
		&models.K8sSession{},
		&models.K8sCommand{},
	); err != nil {
		zap.L().Warn("K8s表迁移失败", zap.Error(err))
	}

	zap.L().Info("数据库模式迁移完成")
	return nil
}

// initSeedData 初始化基础数据
func (i *Initializer) initSeedData() error {
	zap.L().Info("阶段3：初始化基础数据...")

	// 同步菜单
	if err := i.initService.syncMenus(); err != nil {
		zap.L().Warn("菜单同步失败", zap.Error(err))
	}

	// 同步内置角色
	if err := i.initService.syncBuiltinRoles(); err != nil {
		zap.L().Warn("内置角色同步失败", zap.Error(err))
	}

	// 初始化管理员用户
	if err := i.initService.initUsers(); err != nil {
		zap.L().Warn("管理员用户初始化失败", zap.Error(err))
	}

	// 同步属性定义
	if err := i.initService.syncAttributes(); err != nil {
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
	if err := i.initService.assignDefaultPermissions(); err != nil {
		zap.L().Warn("默认权限分配失败", zap.Error(err))
	}

	// 初始化诊断数据
	if err := i.initService.initDiagnosticData(); err != nil {
		zap.L().Warn("诊断数据初始化失败", zap.Error(err))
	}

	// 初始化Agent版本
	if err := i.initService.initAgentVersions(); err != nil {
		zap.L().Warn("Agent版本初始化失败", zap.Error(err))
	}

	// 初始化API权限
	if err := i.initService.initAPIPermissions(); err != nil {
		zap.L().Warn("API权限初始化失败", zap.Error(err))
	}

	return nil
}

// initPermissionsFromSQL 从 SQL 文件初始化权限数据
func (i *Initializer) initPermissionsFromSQL() error {
	zap.L().Info("从 SQL 文件初始化权限数据...")

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
	if err := db.Model(&models.Permission{}).Count(&count).Error; err != nil {
		zap.L().Warn("统计权限数量失败", zap.Error(err))
	} else {
		zap.L().Info("权限数据初始化完成", zap.Int64("total_permissions", count))
	}

	return nil
}
