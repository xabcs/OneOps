package services

import (
	"encoding/json"
	"oneops/backend/models"
	"oneops/backend/utils"
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

// initData 初始化数据（仅在表为空时）
func (s *InitService) initData() error {
	// 检查菜单表是否为空
	var menuCount int64
	db.Model(&models.Menu{}).Count(&menuCount)
	if menuCount == 0 {
		if err := s.initMenus(); err != nil {
			return err
		}
	}

	// 检查角色表是否为空
	var roleCount int64
	db.Model(&models.Role{}).Count(&roleCount)
	if roleCount == 0 {
		if err := s.initRoles(); err != nil {
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
		{ID: 5, Name: "系统管理", Icon: "Setting", Path: "/system", Permission: "menu:system", Sort: 5, Status: 1, ParentID: 0},
		{ID: 51, Name: "菜单管理", Icon: "Menu", Path: "/system/menus", Permission: "menu:system:menus", Sort: 1, Status: 1, ParentID: 5},
		{ID: 52, Name: "角色管理", Icon: "UserFilled", Path: "/system/roles", Permission: "menu:system:roles", Sort: 2, Status: 1, ParentID: 5},
		{ID: 53, Name: "用户管理", Icon: "User", Path: "/system/users", Permission: "menu:system:users", Sort: 3, Status: 1, ParentID: 5},
		{ID: 6, Name: "操作审计", Icon: "Document", Path: "/audit", Permission: "menu:audit", Sort: 6, Status: 1, ParentID: 0},
		{ID: 61, Name: "事件查询", Icon: "List", Path: "/audit/behavior", Permission: "menu:audit:behavior", Sort: 1, Status: 1, ParentID: 6},
	}

	for _, menu := range menus {
		if err := db.Create(&menu).Error; err != nil {
			return err
		}
	}

	return nil
}

// initRoles 初始化角色数据（从 Mock 数据迁移）
func (s *InitService) initRoles() error {
	adminMenuIDs := []uint{1, 2, 21, 22, 3, 31, 32, 4, 41, 42, 5, 51, 52, 53, 6, 61}
	opsMenuIDs := []uint{1, 2, 21, 22, 3, 31, 32, 4, 41, 42}
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
		Username:  "admin",
		Password:  hashedPassword,
		Nickname:  "超级管理员",
		Email:     "admin@example.com",
		RoleIDs:   string(adminRoleIDsJSON),
		Status:    "active",
		HomePath:  "/",
	}

	return db.Create(&user).Error
}
