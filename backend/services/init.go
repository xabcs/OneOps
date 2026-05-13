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

	// 定义动态路由菜单（用于 SoybeanAdmin 动态路由模式）
	// 只包含前端实际存在的页面
	menus := []models.Menu{
		// 一级菜单
		{ID: 1, Name: "首页", Icon: "mdi:monitor-dashboard", Path: "/home", Permission: "", Sort: 1, Status: 1, ParentID: 0},
		{ID: 2, Name: "系统管理", Icon: "carbon:cloud-service-management", Path: "/manage", Permission: "", Sort: 2, Status: 1, ParentID: 0},
		{ID: 3, Name: "用户管理", Icon: "ic:round-manage-accounts", Path: "/manage/user", Permission: "system:user:query", Sort: 1, Status: 1, ParentID: 2},
		{ID: 4, Name: "角色管理", Icon: "carbon:user-role", Path: "/manage/role", Permission: "system:role:query", Sort: 2, Status: 1, ParentID: 2},
		{ID: 5, Name: "菜单管理", Icon: "material-symbols:route", Path: "/manage/menu", Permission: "system:menu:query", Sort: 3, Status: 1, ParentID: 2},
		{ID: 13, Name: "关于", Icon: "fluent:book-information-24-regular", Path: "/about", Permission: "", Sort: 5, Status: 1, ParentID: 0},
		{ID: 14, Name: "用户中心", Icon: "mdi:user-circle-outline", Path: "/user-center", Permission: "", Sort: 6, Status: 1, ParentID: 0},
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

	// 定义5个内置角色的菜单权限（动态路由模式）
	// 菜单ID映射：1=首页, 2=系统管理, 3=用户管理, 4=角色管理, 5=菜单管理, 13=关于, 14=用户中心
	adminMenuIDs := []uint{1, 2, 3, 4, 5, 13, 14}           // 超级管理员：所有权限
	opsMenuIDs := []uint{1, 2, 3, 4, 5, 13, 14}              // 运维工程师：系统管理权限
	auditorMenuIDs := []uint{1, 13}                    // 审计员：仅基础权限
	userMenuIDs := []uint{1}                           // 普通用户：仅首页
	testMenuIDs := []uint{1, 13}                       // 测试角色：首页和关于

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
	// 菜单ID映射：1=首页, 2=系统管理, 3=用户管理, 4=角色管理, 5=菜单管理, 13=关于, 14=用户中心
	adminMenuIDs := []uint{1, 2, 3, 4, 5, 13, 14}           // 超级管理员：所有权限
	opsMenuIDs := []uint{1, 2, 3, 4, 5, 13, 14}              // 运维工程师：系统管理权限
	auditorMenuIDs := []uint{1, 13}                    // 审计员：仅基础权限
	userMenuIDs := []uint{1}                           // 普通用户：仅首页
	testMenuIDs := []uint{1, 13}                       // 测试角色：首页和关于

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
