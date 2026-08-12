package authorization

import (
	"time"

	modelauth "oneops/backend3/model/authorization"

	"gorm.io/gorm"
)

// AuthorizationRepository 授权中心数据访问层
type AuthorizationRepository struct {
	db *gorm.DB
}

// NewAuthorizationRepository 创建授权中心仓库
func NewAuthorizationRepository(db *gorm.DB) *AuthorizationRepository {
	return &AuthorizationRepository{db: db}
}

// ========== Application ==========

// CreateApplication 创建应用
func (r *AuthorizationRepository) CreateApplication(app *modelauth.Application) error {
	return r.db.Create(app).Error
}

// FindApplications 分页查询应用列表
func (r *AuthorizationRepository) FindApplications(page, pageSize int, name string) ([]*modelauth.Application, int64, error) {
	var apps []*modelauth.Application
	var total int64

	query := r.db.Model(&modelauth.Application{})

	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&apps).Error; err != nil {
		return nil, 0, err
	}

	return apps, total, nil
}

// FindAllApplications 获取所有应用（不分页）
func (r *AuthorizationRepository) FindAllApplications() ([]*modelauth.Application, error) {
	var apps []*modelauth.Application
	err := r.db.Where("status = ?", 1).Select("id, name, code, type").Find(&apps).Error
	return apps, err
}

// FindApplicationByID 根据ID获取应用
func (r *AuthorizationRepository) FindApplicationByID(id uint) (*modelauth.Application, error) {
	var app modelauth.Application
	if err := r.db.First(&app, id).Error; err != nil {
		return nil, err
	}
	return &app, nil
}

// SaveApplication 保存应用
func (r *AuthorizationRepository) SaveApplication(app *modelauth.Application) error {
	return r.db.Save(app).Error
}

// DeleteApplication 删除应用
func (r *AuthorizationRepository) DeleteApplication(id uint) error {
	return r.db.Delete(&modelauth.Application{}, id).Error
}

// ========== ApplicationRole ==========

// DeleteRolesByAppID 删除应用的所有角色
func (r *AuthorizationRepository) DeleteRolesByAppID(appID uint) error {
	return r.db.Where("app_id = ?", appID).Delete(&modelauth.ApplicationRole{}).Error
}

// CreateRole 创建应用角色
func (r *AuthorizationRepository) CreateRole(role *modelauth.ApplicationRole) error {
	return r.db.Create(role).Error
}

// FindRolesByAppID 获取应用角色列表
func (r *AuthorizationRepository) FindRolesByAppID(appID uint) ([]*modelauth.ApplicationRole, error) {
	var roles []*modelauth.ApplicationRole
	err := r.db.Where("app_id = ?", appID).Find(&roles).Error
	return roles, err
}

// FindRoleByAppIDAndID 根据应用ID和角色ID获取角色
func (r *AuthorizationRepository) FindRoleByAppIDAndID(appID, id uint) (*modelauth.ApplicationRole, error) {
	var role modelauth.ApplicationRole
	if err := r.db.Where("id = ? AND app_id = ?", id, appID).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

// FindRolesByAppIDOrdered 获取应用角色列表（按类型和名称排序，用于矩阵视图）
func (r *AuthorizationRepository) FindRolesByAppIDOrdered(appID uint) ([]modelauth.ApplicationRole, error) {
	var roles []modelauth.ApplicationRole
	if err := r.db.Where("app_id = ?", appID).Order("role_type, role_name").Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

// ========== ApplicationUser ==========

// DeleteUsersByAppID 删除应用的所有用户，返回受影响行数
func (r *AuthorizationRepository) DeleteUsersByAppID(appID uint) (int64, error) {
	result := r.db.Where("app_id = ?", appID).Delete(&modelauth.ApplicationUser{})
	return result.RowsAffected, result.Error
}

// CreateUser 创建应用用户
func (r *AuthorizationRepository) CreateUser(user *modelauth.ApplicationUser) error {
	return r.db.Create(user).Error
}

// FindUsersByAppID 获取应用用户列表
func (r *AuthorizationRepository) FindUsersByAppID(appID uint) ([]*modelauth.ApplicationUser, error) {
	var users []*modelauth.ApplicationUser
	err := r.db.Where("app_id = ?", appID).Find(&users).Error
	return users, err
}

// ========== ApplicationGroup ==========

// DeleteGroupsByAppID 删除应用的所有用户组
func (r *AuthorizationRepository) DeleteGroupsByAppID(appID uint) error {
	return r.db.Where("app_id = ?", appID).Delete(&modelauth.ApplicationGroup{}).Error
}

// CreateGroup 创建应用用户组
func (r *AuthorizationRepository) CreateGroup(group *modelauth.ApplicationGroup) error {
	return r.db.Create(group).Error
}

// FindGroupsByAppID 获取应用用户组列表
func (r *AuthorizationRepository) FindGroupsByAppID(appID uint) ([]modelauth.ApplicationGroup, error) {
	var groups []modelauth.ApplicationGroup
	if err := r.db.Where("app_id = ?", appID).Find(&groups).Error; err != nil {
		return nil, err
	}
	return groups, nil
}

// ========== ApplicationAuthorizationRule ==========

// DeleteAuthRulesByAppID 删除应用的所有授权规则，返回受影响行数
func (r *AuthorizationRepository) DeleteAuthRulesByAppID(appID uint) (int64, error) {
	result := r.db.Where("app_id = ?", appID).Delete(&modelauth.ApplicationAuthorizationRule{})
	return result.RowsAffected, result.Error
}

// CreateAuthRule 创建应用授权规则
func (r *AuthorizationRepository) CreateAuthRule(rule *modelauth.ApplicationAuthorizationRule) error {
	return r.db.Create(rule).Error
}

// FindAuthRulesByAppID 获取应用授权规则列表（不排序）
func (r *AuthorizationRepository) FindAuthRulesByAppID(appID uint) ([]modelauth.ApplicationAuthorizationRule, error) {
	var rules []modelauth.ApplicationAuthorizationRule
	if err := r.db.Where("app_id = ?", appID).Find(&rules).Error; err != nil {
		return nil, err
	}
	return rules, nil
}

// FindAuthRulesByAppIDOrdered 获取应用授权规则列表（按规则名称排序，用于矩阵视图）
func (r *AuthorizationRepository) FindAuthRulesByAppIDOrdered(appID uint) ([]modelauth.ApplicationAuthorizationRule, error) {
	var rules []modelauth.ApplicationAuthorizationRule
	if err := r.db.Where("app_id = ?", appID).Order("rule_name").Find(&rules).Error; err != nil {
		return nil, err
	}
	return rules, nil
}

// FindAuthRuleByID 根据ID获取授权规则
func (r *AuthorizationRepository) FindAuthRuleByID(id uint) (*modelauth.ApplicationAuthorizationRule, error) {
	var rule modelauth.ApplicationAuthorizationRule
	if err := r.db.First(&rule, id).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}

// ========== AuthUser ==========

// CreateAuthUser 创建授权中心用户
func (r *AuthorizationRepository) CreateAuthUser(user *modelauth.AuthUser) error {
	return r.db.Create(user).Error
}

// FindAuthUsers 分页查询授权中心用户列表
func (r *AuthorizationRepository) FindAuthUsers(page, pageSize int, username string) ([]*modelauth.AuthUser, int64, error) {
	var users []*modelauth.AuthUser
	var total int64

	query := r.db.Model(&modelauth.AuthUser{})

	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Preload("Groups").Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// FindAuthUserByID 根据ID获取授权中心用户
func (r *AuthorizationRepository) FindAuthUserByID(id uint) (*modelauth.AuthUser, error) {
	var user modelauth.AuthUser
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// SaveAuthUser 保存授权中心用户
func (r *AuthorizationRepository) SaveAuthUser(user *modelauth.AuthUser) error {
	return r.db.Save(user).Error
}

// DeleteAuthUser 删除授权中心用户
func (r *AuthorizationRepository) DeleteAuthUser(id uint) error {
	return r.db.Delete(&modelauth.AuthUser{}, id).Error
}

// FindAllActiveAuthUsers 获取所有启用的授权中心用户
func (r *AuthorizationRepository) FindAllActiveAuthUsers() ([]*modelauth.AuthUser, error) {
	var users []*modelauth.AuthUser
	err := r.db.Where("status = ?", 1).Find(&users).Error
	return users, err
}

// ========== AuthGroup ==========

// CreateAuthGroup 创建授权中心用户组
func (r *AuthorizationRepository) CreateAuthGroup(group *modelauth.AuthGroup) error {
	return r.db.Create(group).Error
}

// FindAuthGroups 分页查询授权中心用户组列表
func (r *AuthorizationRepository) FindAuthGroups(page, pageSize int, name string) ([]*modelauth.AuthGroup, int64, error) {
	var groups []*modelauth.AuthGroup
	var total int64

	query := r.db.Model(&modelauth.AuthGroup{})

	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&groups).Error; err != nil {
		return nil, 0, err
	}

	return groups, total, nil
}

// FindAuthGroupByID 根据ID获取授权中心用户组
func (r *AuthorizationRepository) FindAuthGroupByID(id uint) (*modelauth.AuthGroup, error) {
	var group modelauth.AuthGroup
	if err := r.db.First(&group, id).Error; err != nil {
		return nil, err
	}
	return &group, nil
}

// SaveAuthGroup 保存授权中心用户组
func (r *AuthorizationRepository) SaveAuthGroup(group *modelauth.AuthGroup) error {
	return r.db.Save(group).Error
}

// DeleteAuthGroup 删除授权中心用户组
func (r *AuthorizationRepository) DeleteAuthGroup(id uint) error {
	return r.db.Delete(&modelauth.AuthGroup{}, id).Error
}

// FindAllActiveAuthGroups 获取所有启用的授权中心用户组
func (r *AuthorizationRepository) FindAllActiveAuthGroups() ([]*modelauth.AuthGroup, error) {
	var groups []*modelauth.AuthGroup
	err := r.db.Where("status = ?", 1).Find(&groups).Error
	return groups, err
}

// ========== AuthUserGroup ==========

// CountUserGroup 统计用户组成员数量（检查是否已分配）
func (r *AuthorizationRepository) CountUserGroup(userID, groupID uint) (int64, error) {
	var count int64
	err := r.db.Model(&modelauth.AuthUserGroup{}).
		Where("user_id = ? AND group_id = ?", userID, groupID).
		Count(&count).Error
	return count, err
}

// CreateUserGroup 创建用户组成员
func (r *AuthorizationRepository) CreateUserGroup(ug *modelauth.AuthUserGroup) error {
	return r.db.Create(ug).Error
}

// FindUserGroupsByUserID 获取用户的用户组列表（预加载用户组信息）
func (r *AuthorizationRepository) FindUserGroupsByUserID(userID uint) ([]*modelauth.AuthUserGroup, error) {
	var userGroups []*modelauth.AuthUserGroup
	err := r.db.Where("user_id = ?", userID).Preload("GroupIDField").Find(&userGroups).Error
	return userGroups, err
}

// FindUserGroupsByGroupID 获取用户组的成员列表（预加载用户信息）
func (r *AuthorizationRepository) FindUserGroupsByGroupID(groupID uint) ([]*modelauth.AuthUserGroup, error) {
	var userGroups []*modelauth.AuthUserGroup
	err := r.db.Where("group_id = ?", groupID).Preload("UserIDField").Find(&userGroups).Error
	return userGroups, err
}

// DeleteUserGroup 删除用户组成员
func (r *AuthorizationRepository) DeleteUserGroup(userID, groupID uint) error {
	return r.db.Where("user_id = ? AND group_id = ?", userID, groupID).
		Delete(&modelauth.AuthUserGroup{}).Error
}

// ========== GroupBinding ==========

// CreateGroupBinding 创建用户组绑定
func (r *AuthorizationRepository) CreateGroupBinding(binding *modelauth.GroupBinding) error {
	return r.db.Create(binding).Error
}

// FindGroupBindingsByGroupID 获取用户组的所有绑定（预加载角色和应用）
func (r *AuthorizationRepository) FindGroupBindingsByGroupID(groupID uint) ([]*modelauth.GroupBinding, error) {
	var bindings []*modelauth.GroupBinding
	err := r.db.Where("group_id = ?", groupID).
		Preload("ApplicationRole").
		Preload("AppIDField").
		Find(&bindings).Error
	return bindings, err
}

// DeleteGroupBinding 删除用户组绑定
func (r *AuthorizationRepository) DeleteGroupBinding(id uint) error {
	return r.db.Delete(&modelauth.GroupBinding{}, id).Error
}

// ========== GroupBindingExecution ==========

// CreateGroupBindingExecution 创建角色绑定执行记录
func (r *AuthorizationRepository) CreateGroupBindingExecution(exec *modelauth.GroupBindingExecution) error {
	return r.db.Create(exec).Error
}

// FindGroupBindingExecutionsByBindingID 获取权限绑定执行记录
func (r *AuthorizationRepository) FindGroupBindingExecutionsByBindingID(bindingID uint) ([]modelauth.GroupBindingExecution, error) {
	var executions []modelauth.GroupBindingExecution
	err := r.db.Where("group_binding_id = ?", bindingID).
		Preload("AuthUserField").
		Order("created_at DESC").
		Find(&executions).Error
	if err != nil {
		return nil, err
	}
	return executions, nil
}

// ========== OperationLog ==========

// CreateOperationLog 创建操作日志
func (r *AuthorizationRepository) CreateOperationLog(log *modelauth.ApplicationOperationLog) error {
	return r.db.Create(log).Error
}

// FindOperationLogs 分页查询操作日志
func (r *AuthorizationRepository) FindOperationLogs(appID uint, page, pageSize int) ([]*modelauth.ApplicationOperationLog, int64, error) {
	var logs []*modelauth.ApplicationOperationLog
	var total int64

	query := r.db.Model(&modelauth.ApplicationOperationLog{}).Where("app_id = ?", appID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// ========== UserIdentityMapping ==========

// FirstOrCreateUserIdentityMapping 查询或创建用户身份映射
func (r *AuthorizationRepository) FirstOrCreateUserIdentityMapping(mapping *modelauth.UserIdentityMapping, authUserID, appID uint) error {
	return r.db.Where("auth_user_id = ? AND app_id = ?", authUserID, appID).
		FirstOrCreate(mapping).Error
}

// FindUserIdentityMappings 分页查询用户身份映射列表
func (r *AuthorizationRepository) FindUserIdentityMappings(page, pageSize int, username string, appID uint, status string) ([]*modelauth.UserIdentityMapping, int64, error) {
	var mappings []*modelauth.UserIdentityMapping
	var total int64

	query := r.db.Model(&modelauth.UserIdentityMapping{}).Preload("AuthUserField").Preload("AppIDField")

	if username != "" {
		query = query.Joins("JOIN auth_users ON auth_users.id = user_identity_mappings.auth_user_id").
			Where("auth_users.username LIKE ?", "%"+username+"%")
	}

	if appID > 0 {
		query = query.Where("app_id = ?", appID)
	}

	if status != "" {
		query = query.Where("mapping_status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&mappings).Error; err != nil {
		return nil, 0, err
	}

	return mappings, total, nil
}

// DeleteUserIdentityMapping 删除用户身份映射
func (r *AuthorizationRepository) DeleteUserIdentityMapping(id uint) error {
	return r.db.Delete(&modelauth.UserIdentityMapping{}, id).Error
}

// FindUserIdentityMappingByUserAndApp 根据用户ID和应用ID获取身份映射
func (r *AuthorizationRepository) FindUserIdentityMappingByUserAndApp(authUserID, appID uint) (*modelauth.UserIdentityMapping, error) {
	var mapping modelauth.UserIdentityMapping
	if err := r.db.Where("auth_user_id = ? AND app_id = ?", authUserID, appID).First(&mapping).Error; err != nil {
		return nil, err
	}
	return &mapping, nil
}

// ========== 矩阵视图查询结果类型 ==========

// MatrixUserResult 角色-用户矩阵用户查询结果
type MatrixUserResult struct {
	ID               uint   `json:"id"`
	Username         string `json:"username"`
	Nickname         string `json:"nickname"`
	ExternalUsername string `json:"external_username"`
}

// RuleMatrixUserResult 用户-规则矩阵用户查询结果（含外部用户ID）
type RuleMatrixUserResult struct {
	ID               uint   `json:"id"`
	Username         string `json:"username"`
	Nickname         string `json:"nickname"`
	ExternalUsername string `json:"external_username"`
	ExternalUserID   string `json:"external_user_id"`
}

// MatrixPermissionResult 角色-用户矩阵权限查询结果
type MatrixPermissionResult struct {
	UserID           uint
	RoleID           uint
	Username         string
	RoleCode         string
	Status           string
	GroupName        string
	AssignedAt       time.Time
	ExternalUsername string
}

// FindMatrixUsers 获取角色-用户矩阵的用户列表
func (r *AuthorizationRepository) FindMatrixUsers(appID uint) ([]MatrixUserResult, error) {
	userQuery := `
		SELECT DISTINCT
			auth_users.id,
			auth_users.username,
			auth_users.nickname,
			auth_user_identity_mappings.external_username
		FROM auth_users
		INNER JOIN auth_user_groups ON auth_user_groups.user_id = auth_users.id
		INNER JOIN auth_groups ON auth_groups.id = auth_user_groups.group_id
		INNER JOIN auth_group_bindings ON auth_group_bindings.group_id = auth_groups.id
		LEFT JOIN auth_user_identity_mappings ON auth_user_identity_mappings.auth_user_id = auth_users.id
			AND auth_user_identity_mappings.app_id = ?
		WHERE auth_group_bindings.app_id = ?
			AND auth_users.status = 1
			AND auth_groups.status = 1
		ORDER BY auth_users.username
	`
	var users []MatrixUserResult
	if err := r.db.Raw(userQuery, appID, appID).Scan(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// FindMatrixPermissions 获取角色-用户矩阵的权限数据
func (r *AuthorizationRepository) FindMatrixPermissions(appID uint) ([]MatrixPermissionResult, error) {
	permQuery := `
		SELECT
			auth_users.id as user_id,
			auth_application_roles.id as role_id,
			auth_users.username,
			auth_application_roles.role_code,
			CASE
				WHEN auth_user_identity_mappings.mapping_status = 'active' THEN 'active'
				ELSE 'inactive'
			END as status,
			auth_groups.name as group_name,
			MIN(auth_user_groups.created_at) as assigned_at,
			auth_user_identity_mappings.external_username
		FROM auth_users
		INNER JOIN auth_user_groups ON auth_user_groups.user_id = auth_users.id
		INNER JOIN auth_groups ON auth_groups.id = auth_user_groups.group_id
		INNER JOIN auth_group_bindings ON auth_group_bindings.group_id = auth_groups.id
		INNER JOIN auth_application_roles ON auth_application_roles.id = auth_group_bindings.application_role_id
		LEFT JOIN auth_user_identity_mappings ON auth_user_identity_mappings.auth_user_id = auth_users.id
			AND auth_user_identity_mappings.app_id = ?
		WHERE auth_group_bindings.app_id = ?
			AND auth_users.status = 1
			AND auth_groups.status = 1
		GROUP BY auth_users.id, auth_application_roles.id, auth_users.username, auth_application_roles.role_code,
			auth_user_identity_mappings.mapping_status, auth_groups.name, auth_user_identity_mappings.external_username
	`
	var permissions []MatrixPermissionResult
	if err := r.db.Raw(permQuery, appID, appID).Scan(&permissions).Error; err != nil {
		return nil, err
	}
	return permissions, nil
}

// FindRuleMatrixUsers 获取用户-规则矩阵的用户列表（含外部用户ID）
func (r *AuthorizationRepository) FindRuleMatrixUsers(appID uint) ([]RuleMatrixUserResult, error) {
	userQuery := `
		SELECT DISTINCT
			auth_users.id,
			auth_users.username,
			auth_users.nickname,
			auth_user_identity_mappings.external_username,
			auth_user_identity_mappings.external_user_id
		FROM auth_users
		INNER JOIN auth_user_groups ON auth_user_groups.user_id = auth_users.id
		INNER JOIN auth_groups ON auth_groups.id = auth_user_groups.group_id
		INNER JOIN auth_group_bindings ON auth_group_bindings.group_id = auth_groups.id
		LEFT JOIN auth_user_identity_mappings ON auth_user_identity_mappings.auth_user_id = auth_users.id
			AND auth_user_identity_mappings.app_id = ?
		WHERE auth_group_bindings.app_id = ?
			AND auth_users.status = 1
			AND auth_groups.status = 1
		ORDER BY auth_users.username
	`
	var users []RuleMatrixUserResult
	if err := r.db.Raw(userQuery, appID, appID).Scan(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// FindUserEffectivePermissions 获取用户有效权限列表
func (r *AuthorizationRepository) FindUserEffectivePermissions(page, pageSize int, username string, appID uint) ([]map[string]interface{}, int64, error) {
	var results []map[string]interface{}
	var total int64

	query := r.db.Table("auth_users").
		Select(`
			DISTINCT
			auth_users.username,
			auth_users.nickname,
			auth_applications.id as app_id,
			auth_applications.name as app_name,
			auth_application_roles.role_code,
			auth_application_roles.role_name,
			auth_application_roles.role_type,
			CASE
				WHEN auth_user_identity_mappings.mapping_status = 'active' THEN 'active'
				ELSE 'inactive'
			END as status,
			auth_user_identity_mappings.external_username,
			auth_groups.id as group_id,
			auth_groups.name as group_name,
			auth_groups.code as group_code,
			MIN(auth_user_groups.created_at) as assigned_at
		`).
		Joins("JOIN auth_user_groups ON auth_user_groups.user_id = auth_users.id").
		Joins("JOIN auth_groups ON auth_groups.id = auth_user_groups.group_id").
		Joins("JOIN auth_group_bindings ON auth_group_bindings.group_id = auth_groups.id").
		Joins("JOIN auth_applications ON auth_applications.id = auth_group_bindings.app_id").
		Joins("JOIN auth_application_roles ON auth_application_roles.id = auth_group_bindings.application_role_id").
		Joins("LEFT JOIN auth_user_identity_mappings ON auth_user_identity_mappings.auth_user_id = auth_users.id AND auth_user_identity_mappings.app_id = auth_applications.id").
		Where("auth_users.status = ?", 1).
		Where("auth_groups.status = ?", 1).
		Group("auth_users.id, auth_applications.id, auth_application_roles.id")

	if username != "" {
		query = query.Where("auth_users.username LIKE ?", "%"+username+"%")
	}

	if appID > 0 {
		query = query.Where("auth_applications.id = ?", appID)
	}

	countQuery := r.db.Table("auth_users").
		Select("COUNT(DISTINCT CONCAT(auth_users.id, '-', auth_applications.id, '-', auth_application_roles.id))").
		Joins("JOIN auth_user_groups ON auth_user_groups.user_id = auth_users.id").
		Joins("JOIN auth_groups ON auth_groups.id = auth_user_groups.group_id").
		Joins("JOIN auth_group_bindings ON auth_group_bindings.group_id = auth_groups.id").
		Joins("JOIN auth_applications ON auth_applications.id = auth_group_bindings.app_id").
		Joins("JOIN auth_application_roles ON auth_application_roles.id = auth_group_bindings.application_role_id").
		Where("auth_users.status = ?", 1).
		Where("auth_groups.status = ?", 1)

	if username != "" {
		countQuery = countQuery.Where("auth_users.username LIKE ?", "%"+username+"%")
	}

	if appID > 0 {
		countQuery = countQuery.Where("auth_applications.id = ?", appID)
	}

	if err := countQuery.Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("auth_users.username, auth_applications.name").Find(&results).Error; err != nil {
		return nil, 0, err
	}

	return results, total, nil
}
