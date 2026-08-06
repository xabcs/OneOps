package services

import (
	"encoding/json"
	"fmt"
	"oneops/backend/models"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Initializer 初始化协调器
type Initializer struct {
	dataLoader   *DataLoader
	initService  *InitService
}

// NewInitializer 创建初始化协调器
func NewInitializer() *Initializer {
	return &Initializer{
		dataLoader:  NewDataLoader("services/data"),
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
	if err := i.initPermissionsFromJSON(); err != nil {
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

// initPermissionsFromJSON 从JSON文件初始化权限数据
func (i *Initializer) initPermissionsFromJSON() error {
	zap.L().Info("初始化权限数据...")

	// 从JSON文件加载权限数据
	permissions, err := i.dataLoader.LoadPermissions()
	if err != nil {
		return fmt.Errorf("加载权限数据失败: %w", err)
	}

	addedCount := 0
	updatedCount := 0

	for _, perm := range permissions {
		var existingPerm models.Permission
		err := db.Where("code = ?", perm.Code).First(&existingPerm).Error

		if err == gorm.ErrRecordNotFound {
			// 新权限，插入
			if err := db.Create(&perm).Error; err != nil {
				zap.L().Warn("创建权限失败",
					zap.String("code", perm.Code),
					zap.Error(err))
				continue
			}
			addedCount++
		} else if err == nil {
			// 权限已存在，更新
			updates := map[string]interface{}{
				"name":        perm.Name,
				"description": perm.Description,
				"module":      perm.Module,
				"resource":    perm.Resource,
				"action":      perm.Action,
				"level":       perm.Level,
				"sort_order":  perm.SortOrder,
				"status":      perm.Status,
			}
			if err := db.Model(&existingPerm).Updates(updates).Error; err != nil {
				zap.L().Warn("更新权限失败",
					zap.String("code", perm.Code),
					zap.Error(err))
				continue
			}
			updatedCount++
		}
	}

	zap.L().Info("权限数据初始化完成",
		zap.Int("added", addedCount),
		zap.Int("updated", updatedCount),
		zap.Int("total", len(permissions)))

	return nil
}

// DataLoader 数据加载器
type DataLoader struct {
	dataDir string
}

// NewDataLoader 创建数据加载器
func NewDataLoader(dataDir string) *DataLoader {
	return &DataLoader{
		dataDir: dataDir,
	}
}

// PermissionData 权限数据结构
type PermissionData struct {
	Version     string               `json:"version"`
	Description string               `json:"description"`
	Modules     []PermissionModule   `json:"modules"`
}

// PermissionModule 权限模块结构
type PermissionModule struct {
	Code        string               `json:"code"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Level       int                  `json:"level"`
	Status      int                  `json:"status"`
	Resources   []PermissionResource `json:"resources"`
}

// PermissionResource 权限资源结构
type PermissionResource struct {
	Code        string               `json:"code"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Level       int                  `json:"level"`
	Status      int                  `json:"status"`
	Actions     []PermissionAction   `json:"actions"`
}

// PermissionAction 权限操作结构
type PermissionAction struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Level       int    `json:"level"`
	Status      int    `json:"status"`
}

// LoadPermissions 加载权限数据
func (d *DataLoader) LoadPermissions() ([]models.Permission, error) {
	filePath := filepath.Join(d.dataDir, "permissions.json")

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取权限配置文件失败: %w", err)
	}

	var permData PermissionData
	if err := json.Unmarshal(data, &permData); err != nil {
		return nil, fmt.Errorf("解析权限配置文件失败: %w", err)
	}

	// 转换为Permission模型列表
	permissions := make([]models.Permission, 0)
	sortOrder := 1

	for _, module := range permData.Modules {
		// 添加模块级权限
		permissions = append(permissions, models.Permission{
			Code:        module.Code,
			Name:        module.Name,
			Description: module.Description,
			Module:      module.Code,
			Resource:    "",
			Action:      "",
			Level:       module.Level,
			SortOrder:   sortOrder,
			Status:      module.Status,
		})
		sortOrder++

		for _, resource := range module.Resources {
			// 添加资源级权限
			resourceCode := fmt.Sprintf("%s.%s", module.Code, resource.Code)
			permissions = append(permissions, models.Permission{
				Code:        resourceCode,
				Name:        resource.Name,
				Description: resource.Description,
				Module:      module.Code,
				Resource:    resource.Code,
				Action:      "",
				Level:       resource.Level,
				SortOrder:   sortOrder,
				Status:      resource.Status,
			})
			sortOrder++

			for _, action := range resource.Actions {
				// 添加操作级权限
				actionCode := fmt.Sprintf("%s.%s.%s", module.Code, resource.Code, action.Code)
				permissions = append(permissions, models.Permission{
					Code:        actionCode,
					Name:        action.Name,
					Description: action.Description,
					Module:      module.Code,
					Resource:    resource.Code,
					Action:      action.Code,
					Level:       action.Level,
					SortOrder:   sortOrder,
					Status:      action.Status,
				})
				sortOrder++
			}
		}
	}

	zap.L().Info("成功加载权限数据",
		zap.Int("total", len(permissions)),
		zap.Int("modules", len(permData.Modules)))

	return permissions, nil
}
