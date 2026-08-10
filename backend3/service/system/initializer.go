package system

import (
	"fmt"

	modelaudit "oneops/backend3/model/audit"
	modelauth "oneops/backend3/model/authorization"
	modelcmdb "oneops/backend3/model/cmdb"
	modelk8s "oneops/backend3/model/k8s"
	modelsystem "oneops/backend3/model/system"
	"oneops/backend3/pkg/database"

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
