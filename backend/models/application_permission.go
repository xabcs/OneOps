package models

import (
	"time"
)

// AuthUser 授权中心用户模型（独立于系统登录用户）
type AuthUser struct {
	ID          uint        `json:"id" gorm:"primaryKey"`
	Username    string      `json:"username" gorm:"size:50;not null;uniqueIndex"`
	Password    string      `json:"-" gorm:"size:255;comment:'初始密码（加密存储）'"` // 不返回给前端
	Nickname    string      `json:"nickname" gorm:"size:50"`
	Email       string      `json:"email" gorm:"size:100"`
	Phone       string      `json:"phone" gorm:"size:20"`
	Description string      `json:"description" gorm:"size:200"`
	Status      int         `json:"status" gorm:"default:1;comment:1=启用,0=禁用"`
	Groups      []AuthGroup `json:"groups" gorm:"many2many:auth_user_groups;joinForeignKey:UserID;joinReferences:GroupID;association_foreignkey:ID;references:ID"`
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt"`
}

func (AuthUser) TableName() string {
	return "auth_users"
}

// AuthGroup 授权中心用户组模型（独立于系统角色）
type AuthGroup struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"size:50;not null;uniqueIndex"`
	Code        string    `json:"code" gorm:"size:50;not null;uniqueIndex"`
	Description string    `json:"description" gorm:"size:200"`
	Status      int       `json:"status" gorm:"default:1;comment:1=启用,0=禁用"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (AuthGroup) TableName() string {
	return "auth_groups"
}

// AuthUserGroup 授权中心用户组成员关联模型
type AuthUserGroup struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	UserID       uint      `json:"userId" gorm:"not null;uniqueIndex:idx_user_group;comment:授权中心用户ID"`
	UserIDField  AuthUser  `json:"user,omitempty" gorm:"foreignKey:UserID"`
	GroupID      uint      `json:"groupId" gorm:"not null;uniqueIndex:idx_user_group;comment:授权中心用户组ID"`
	GroupIDField AuthGroup `json:"group,omitempty" gorm:"foreignKey:GroupID"`
	GrantedBy    string    `json:"grantedBy" gorm:"size:50"`
	GrantedAt    time.Time `json:"grantedAt"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// ApplicationEndpoints API端点配置
type ApplicationEndpoints struct {
	// 用户管理相关
	CreateUser string `json:"createUser"` // POST {baseUrl}/api/users
	GetUsers   string `json:"getUsers"`   // GET {baseUrl}/api/users
	GetUser    string `json:"getUser"`    // GET {baseUrl}/api/users/{username}

	// 角色管理相关
	GetRoles     string `json:"getRoles"`     // GET {baseUrl}/api/roles
	AssignRole   string `json:"assignRole"`   // POST {baseUrl}/api/users/{username}/roles
	RevokeRole   string `json:"revokeRole"`   // DELETE {baseUrl}/api/users/{username}/roles/{roleCode}
	GetUserRoles string `json:"getUserRoles"` // GET {baseUrl}/api/users/{username}/roles
}

// ApplicationAuthConfig 认证配置
type ApplicationAuthConfig struct {
	Type     string `json:"type"`     // basic, token, oauth2
	Username string `json:"username"` // Basic Auth用户名
	Password string `json:"password"` // Basic Auth密码
	Token    string `json:"token"`    // API Token
}

// Application 应用注册模型
type Application struct {
	ID           uint       `json:"id" gorm:"primaryKey"`
	Name         string     `json:"name" gorm:"size:50;not null;uniqueIndex"`
	Code         string     `json:"code" gorm:"size:50;not null;uniqueIndex"`
	Type         string     `json:"type" gorm:"size:20;not null"`
	BaseURL      string     `json:"baseUrl" gorm:"size:255;not null"`
	Endpoints    string     `json:"endpoints" gorm:"type:json"`  // JSON: ApplicationEndpoints
	AuthConfig   string     `json:"authConfig" gorm:"type:json"` // JSON: ApplicationAuthConfig
	SyncInterval int        `json:"syncInterval" gorm:"default:300"`
	LastSyncTime *time.Time `json:"lastSyncTime"`
	Description  string     `json:"description" gorm:"size:200"`
	Status       int        `json:"status" gorm:"default:1"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

func (Application) TableName() string {
	return "auth_applications"
}

// ApplicationRole 应用角色模型（同步过来的）
type ApplicationRole struct {
	ID          uint        `json:"id" gorm:"primaryKey"`
	AppID       uint        `json:"appId" gorm:"not null;index"`
	AppIDField  Application `json:"-" gorm:"foreignKey:AppID"`
	RoleCode    string      `json:"roleCode" gorm:"size:100;not null"`
	RoleName    string      `json:"roleName" gorm:"size:100;not null"`
	RoleType    string      `json:"roleType" gorm:"size:50;not null"`
	Description string      `json:"description" gorm:"size:200"`
	SyncTime    time.Time   `json:"syncTime"`
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt"`
}

func (ApplicationRole) TableName() string {
	return "auth_application_roles"
}

// ApplicationUser 应用用户模型（同步过来的）
type ApplicationUser struct {
	ID          uint        `json:"id" gorm:"primaryKey"`
	AppID       uint        `json:"appId" gorm:"not null;index"`
	AppIDField  Application `json:"-" gorm:"foreignKey:AppID"`
	Username    string      `json:"username" gorm:"size:100;not null"`
	DisplayName string      `json:"displayName" gorm:"size:100"`
	Email       string      `json:"email" gorm:"size:100"`
	Status      string      `json:"status" gorm:"size:20;default:active"`
	SyncTime    time.Time   `json:"syncTime"`
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt"`
}

func (ApplicationUser) TableName() string {
	return "auth_application_users"
}

// ApplicationGroup 应用用户组模型（同步过来的）
type ApplicationGroup struct {
	ID          uint        `json:"id" gorm:"primaryKey"`
	AppID       uint        `json:"appId" gorm:"not null;index"`
	AppIDField  Application `json:"-" gorm:"foreignKey:AppID"`
	GroupCode   string      `json:"groupCode" gorm:"size:100;not null"`
	GroupName   string      `json:"groupName" gorm:"size:100;not null"`
	Description string      `json:"description" gorm:"size:200"`
	SyncTime    time.Time   `json:"syncTime"`
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt"`
}

func (ApplicationGroup) TableName() string {
	return "auth_application_groups"
}

// GroupBinding 用户组权限绑定模型（授权中心用户组 → 外部系统角色）
type GroupBinding struct {
	ID                      uint            `json:"id" gorm:"primaryKey"`
	GroupID                 uint            `json:"groupId" gorm:"not null;index;comment:授权中心用户组ID"`
	GroupIDField            AuthGroup       `json:"group,omitempty" gorm:"foreignKey:GroupID"`
	AppID                   uint            `json:"appId" gorm:"not null;index"`
	AppIDField              Application     `json:"-" gorm:"foreignKey:AppID"`
	ApplicationRoleID       uint            `json:"applicationRoleId" gorm:"not null;index"`
	ApplicationRole         ApplicationRole `json:"applicationRole,omitempty" gorm:"foreignKey:ApplicationRoleID"`
	ApplicationPermissionID *uint           `json:"applicationPermissionId,omitempty" gorm:"comment:关联的权限ID（可选）"`
	CreatedAt               time.Time       `json:"createdAt"`
	UpdatedAt               time.Time       `json:"updatedAt"`
}

func (GroupBinding) TableName() string {
	return "auth_group_bindings"
}

// ApplicationOperationLog 应用操作日志
type ApplicationOperationLog struct {
	ID           uint        `json:"id" gorm:"primaryKey"`
	AppID        uint        `json:"appId" gorm:"not null;index"`
	AppIDField   Application `json:"-" gorm:"foreignKey:AppID"`
	Operation    string      `json:"operation" gorm:"size:50;not null"`
	Target       string      `json:"target" gorm:"size:100"`
	RequestData  string      `json:"requestData" gorm:"type:text"`
	ResponseData string      `json:"responseData" gorm:"type:text"`
	Status       string      `json:"status" gorm:"size:20"`
	ErrorMsg     string      `json:"errorMsg" gorm:"type:text"`
	Operator     string      `json:"operator" gorm:"size:50"`
	CreatedAt    time.Time   `json:"createdAt"`
}

func (ApplicationOperationLog) TableName() string {
	return "auth_operation_logs"
}

// UserIdentityMapping 用户身份映射表（授权中心用户 ↔ 外部应用用户）
type UserIdentityMapping struct {
	ID               uint        `json:"id" gorm:"primaryKey"`
	AuthUserID       uint        `json:"authUserId" gorm:"not null;index;comment:授权中心用户ID"`
	AuthUserField    AuthUser    `json:"authUser,omitempty" gorm:"foreignKey:AuthUserID"`
	AppID            uint        `json:"appId" gorm:"not null;index;comment:应用ID"`
	AppIDField       Application `json:"-" gorm:"foreignKey:AppID"`
	ExternalUsername string      `json:"externalUsername" gorm:"size:100;not null;index;comment:外部应用用户名"`
	ExternalUserID   string      `json:"externalUserId" gorm:"size:100;comment:外部系统用户ID（如果有的话）"`
	MappingType      string      `json:"mappingType" gorm:"size:20;default:auto;comment:auto=自动创建,manual=手动创建"`
	MappingStatus    string      `json:"mappingStatus" gorm:"size:20;default:active;comment:active=激活,inactive=禁用,deleted=已删除"`
	LastSyncAt       *time.Time  `json:"lastSyncAt" gorm:"comment:最后同步时间"`
	CreatedAt        time.Time   `json:"createdAt"`
	UpdatedAt        time.Time   `json:"updatedAt"`
}

func (UserIdentityMapping) TableName() string {
	return "auth_user_identity_mappings"
}

// GroupBindingExecution 角色绑定执行记录（记录权限分配的详细状态）
type GroupBindingExecution struct {
	ID               uint         `json:"id" gorm:"primaryKey"`
	GroupBindingID   uint         `json:"groupBindingId" gorm:"not null;index;comment:GroupBinding的ID"`
	GroupBinding     GroupBinding `json:"-" gorm:"foreignKey:GroupBindingID"`
	AuthUserID       uint         `json:"authUserId" gorm:"not null;index;comment:处理的授权中心用户ID"`
	AuthUserField    AuthUser     `json:"authUser,omitempty" gorm:"foreignKey:AuthUserID"`
	ExternalUsername string       `json:"externalUsername" gorm:"size:100;comment:外部用户名"`
	ActionType       string       `json:"actionType" gorm:"size:50;comment:操作类型:created=创建用户,granted=分配权限,removed=移除权限,failed=失败"`
	Status           string       `json:"status" gorm:"size:20;comment:状态:success=成功,failed=失败,pending=待处理"`
	Message          string       `json:"message" gorm:"type:text;comment:详细信息或错误信息"`
	Operator         string       `json:"operator" gorm:"size:50;comment:操作人"`
	CreatedAt        time.Time    `json:"createdAt"`
}

func (GroupBindingExecution) TableName() string {
	return "auth_group_binding_executions"
}

// PermissionAssignmentStatus 权限分配状态汇总（记录权限分配的整体状态）
type PermissionAssignmentStatus struct {
	ID                uint       `json:"id" gorm:"primaryKey"`
	GroupBindingID    uint       `json:"groupBindingId" gorm:"not null;uniqueIndex;comment:GroupBinding的ID"`
	TotalMembers      int        `json:"totalMembers" gorm:"default:0;comment:总成员数"`
	ProcessedMembers  int        `json:"processedMembers" gorm:"default:0;comment:已处理成员数"`
	PendingMembers    int        `json:"pendingMembers" gorm:"default:0;comment:待处理成员数"`
	CreatedIdentities int        `json:"createdIdentities" gorm:"default:0;comment:创建的外部账号数"`
	FailedMembers     int        `json:"failedMembers" gorm:"default:0;comment:失败成员数"`
	Status            string     `json:"status" gorm:"size:20;default:pending;comment:状态:pending=待处理,success=成功,partial=部分成功,failed=失败"`
	LastProcessedAt   *time.Time `json:"lastProcessedAt" gorm:"comment:最后处理时间"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
}

func (PermissionAssignmentStatus) TableName() string {
	return "auth_permission_assignment_statuses"
}
