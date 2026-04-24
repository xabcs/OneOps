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
	// 自动创建表
	err := db.AutoMigrate(
		&models.User{},
		&models.Role{},
		&models.Menu{},
		&models.LoginLog{},
		&models.OperationLog{},
		&models.SystemEventLog{},
	)
	if err != nil {
		return err
	}

	// 检查是否需要初始化数据
	return s.initData()
}

// initData 初始化和同步数据（自动检测并添加新菜单）
func (s *InitService) initData() error {
	// 同步菜单数据（增量更新，不删除现有数据）
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

	return nil
}

// initMenus 初始化菜单数据（从 Mock 数据迁移）
func (s *InitService) initMenus() error {
	menus := []models.Menu{
		{ID: 1, Name: "仪表盘概览", Icon: "House", Path: "/", Permission: "menu:home", Sort: 1, Status: 1, ParentID: 0},
		{ID: 2, Name: "资产管理", Icon: "Monitor", Path: "/servers", Permission: "menu:servers", Sort: 2, Status: 1, ParentID: 0},
		{ID: 21, Name: "主机管理", Icon: "Monitor", Path: "/servers", Permission: "menu:servers", Sort: 1, Status: 1, ParentID: 2},
		{ID: 22, Name: "添加资产", Icon: "Plus", Path: "/servers/add", Permission: "menu:servers", Sort: 2, Status: 1, ParentID: 2},
		{ID: 3, Name: "自动化任务", Icon: "Timer", Path: "/tasks", Permission: "menu:tasks", Sort: 3, Status: 1, ParentID: 0},
		{ID: 31, Name: "任务列表", Icon: "Timer", Path: "/tasks", Permission: "menu:tasks", Sort: 1, Status: 1, ParentID: 3},
		{ID: 32, Name: "新建任务", Icon: "Plus", Path: "/tasks/create", Permission: "menu:tasks", Sort: 2, Status: 1, ParentID: 3},
		{ID: 4, Name: "监控中心", Icon: "DataLine", Path: "/monitoring", Permission: "menu:monitoring", Sort: 4, Status: 1, ParentID: 0},
		{ID: 41, Name: "系统监控", Icon: "DataLine", Path: "/monitoring", Permission: "menu:monitoring", Sort: 1, Status: 1, ParentID: 4},
		{ID: 42, Name: "告警管理", Icon: "Bell", Path: "/monitoring/alerts", Permission: "menu:monitoring", Sort: 2, Status: 1, ParentID: 4},
		{ID: 43, Name: "证书监控", Icon: "Key", Path: "/monitoring/certificate", Permission: "menu:monitoring:certificate", Sort: 3, Status: 1, ParentID: 4},
		{ID: 5, Name: "系统管理", Icon: "Setting", Path: "/system", Permission: "menu:system", Sort: 5, Status: 1, ParentID: 0},
		{ID: 51, Name: "菜单管理", Icon: "Menu", Path: "/system/menus", Permission: "menu:system:menus", Sort: 1, Status: 1, ParentID: 5},
		{ID: 52, Name: "角色管理", Icon: "UserFilled", Path: "/system/roles", Permission: "menu:system:roles", Sort: 2, Status: 1, ParentID: 5},
		{ID: 53, Name: "用户管理", Icon: "User", Path: "/system/users", Permission: "menu:system:users", Sort: 3, Status: 1, ParentID: 5},
		{ID: 6, Name: "操作审计", Icon: "Document", Path: "/audit", Permission: "menu:audit", Sort: 6, Status: 1, ParentID: 0},
		{ID: 61, Name: "行为日志", Icon: "List", Path: "/audit/behavior", Permission: "menu:audit:behavior", Sort: 1, Status: 1, ParentID: 6},
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

	// 定义所有菜单
	menus := []models.Menu{
		{ID: 1, Name: "仪表盘概览", Icon: "House", Path: "/", Permission: "menu:home", Sort: 1, Status: 1, ParentID: 0},
		{ID: 2, Name: "资产管理", Icon: "Monitor", Path: "/servers", Permission: "menu:servers", Sort: 2, Status: 1, ParentID: 0},
		{ID: 21, Name: "主机管理", Icon: "Monitor", Path: "/servers", Permission: "menu:servers", Sort: 1, Status: 1, ParentID: 2},
		{ID: 22, Name: "添加资产", Icon: "Plus", Path: "/servers/add", Permission: "menu:servers", Sort: 2, Status: 1, ParentID: 2},
		{ID: 3, Name: "自动化任务", Icon: "Timer", Path: "/tasks", Permission: "menu:tasks", Sort: 3, Status: 1, ParentID: 0},
		{ID: 31, Name: "任务列表", Icon: "Timer", Path: "/tasks", Permission: "menu:tasks", Sort: 1, Status: 1, ParentID: 3},
		{ID: 32, Name: "新建任务", Icon: "Plus", Path: "/tasks/create", Permission: "menu:tasks", Sort: 2, Status: 1, ParentID: 3},
		{ID: 4, Name: "监控中心", Icon: "DataLine", Path: "/monitoring", Permission: "menu:monitoring", Sort: 4, Status: 1, ParentID: 0},
		{ID: 41, Name: "系统监控", Icon: "DataLine", Path: "/monitoring", Permission: "menu:monitoring", Sort: 1, Status: 1, ParentID: 4},
		{ID: 42, Name: "告警管理", Icon: "Bell", Path: "/monitoring/alerts", Permission: "menu:monitoring", Sort: 2, Status: 1, ParentID: 4},
		{ID: 43, Name: "证书监控", Icon: "Key", Path: "/monitoring/certificate", Permission: "menu:monitoring:certificate", Sort: 3, Status: 1, ParentID: 4},
		{ID: 5, Name: "系统管理", Icon: "Setting", Path: "/system", Permission: "menu:system", Sort: 5, Status: 1, ParentID: 0},
		{ID: 51, Name: "菜单管理", Icon: "Menu", Path: "/system/menus", Permission: "menu:system:menus", Sort: 1, Status: 1, ParentID: 5},
		{ID: 52, Name: "角色管理", Icon: "UserFilled", Path: "/system/roles", Permission: "menu:system:roles", Sort: 2, Status: 1, ParentID: 5},
		{ID: 53, Name: "用户管理", Icon: "User", Path: "/system/users", Permission: "menu:system:users", Sort: 3, Status: 1, ParentID: 5},
		{ID: 6, Name: "操作审计", Icon: "Document", Path: "/audit", Permission: "menu:audit", Sort: 6, Status: 1, ParentID: 0},
		{ID: 61, Name: "行为日志", Icon: "List", Path: "/audit/behavior", Permission: "menu:audit:behavior", Sort: 1, Status: 1, ParentID: 6},
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

	logger.Info("菜单同步完成",
		zap.Int("added", addedCount),
		zap.Int("updated", updatedCount),
		zap.Int("total", len(menus)))

	return nil
}

// syncRoleMenus 同步角色菜单权限
func (s *InitService) syncRoleMenus() error {
	logger.Info("开始同步角色菜单权限...")

	adminMenuIDs := []uint{1, 2, 21, 22, 3, 31, 32, 4, 41, 42, 43, 5, 51, 52, 53, 6, 61}
	opsMenuIDs := []uint{1, 2, 21, 22, 3, 31, 32, 4, 41, 42, 43}

	adminMenuIDsJSON, _ := json.Marshal(adminMenuIDs)
	opsMenuIDsJSON, _ := json.Marshal(opsMenuIDs)

	// 更新超级管理员角色
	result := db.Model(&models.Role{}).Where("code = ?", "admin").Update("menu_ids", string(adminMenuIDsJSON))
	if result.Error != nil {
		logger.Error("更新超级管理员角色权限失败", zap.Any("error", result.Error))
		return result.Error
	}
	logger.Info("更新超级管理员角色权限", zap.Int("menus", len(adminMenuIDs)))

	// 更新或创建运维工程师角色
	var opsRole models.Role
	err := db.Where("code = ?", "ops").First(&opsRole).Error
	if err == nil {
		// 角色存在，更新权限
		db.Model(&opsRole).Update("menu_ids", string(opsMenuIDsJSON))
		logger.Info("更新运维工程师角色权限", zap.Int("menus", len(opsMenuIDs)))
	} else {
		// 角色不存在，创建新角色
		opsRole = models.Role{
			Name:    "运维工程师",
			Code:    "ops",
			Description: "负责主机和任务管理",
			MenuIDs: string(opsMenuIDsJSON),
			Status:  1,
		}
		if err := db.Create(&opsRole).Error; err != nil {
			logger.Error("创建运维工程师角色失败", zap.Any("error", err))
			return err
		}
		logger.Info("创建运维工程师角色", zap.Uint("id", opsRole.ID))
	}

	logger.Info("角色菜单权限同步完成")
	return nil
}

// initRoles 初始化角色数据（从 Mock 数据迁移）
func (s *InitService) initRoles() error {
	adminMenuIDs := []uint{1, 2, 21, 22, 3, 31, 32, 4, 41, 42, 43, 5, 51, 52, 53, 6, 61}
	opsMenuIDs := []uint{1, 2, 21, 22, 3, 31, 32, 4, 41, 42, 43}
	auditorMenuIDs := []uint{1, 4, 41}

	adminMenuIDsJSON, _ := json.Marshal(adminMenuIDs)
	opsMenuIDsJSON, _ := json.Marshal(opsMenuIDs)
	auditorMenuIDsJSON, _ := json.Marshal(auditorMenuIDs)

	roles := []models.Role{
		{Name: "超级管理员", Code: "admin", Description: "拥有系统所有权限", MenuIDs: string(adminMenuIDsJSON), Status: 1},
		{Name: "运维工程师", Code: "ops", Description: "负责主机和任务管理", MenuIDs: string(opsMenuIDsJSON), Status: 1},
		{Name: "审计员", Code: "auditor", Description: "仅拥有查看权限", MenuIDs: string(auditorMenuIDsJSON), Status: 1},
	}

	for _, role := range roles {
		if err := db.Create(&role).Error; err != nil {
			return err
		}
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
		HomePath: "/",
	}

	return db.Create(&user).Error
}
