package models

import (
	"time"
)

// AuthUser 授权中心用户模型（独立于系统登录用户）
type AuthUser struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Username    string    `json:"username" gorm:"size:50;not null;uniqueIndex"`
	Password    string    `json:"-" gorm:"size:255;comment:'初始密码（加密存储）'"` // 不返回给前端
	Nickname    string    `json:"nickname" gorm:"size:50"`
	Email       string    `json:"email" gorm:"size:100"`
	Phone       string    `json:"phone" gorm:"size:20"`
	Description string    `json:"description" gorm:"size:200"`
	Status      int       `json:"status" gorm:"default:1;comment:1=启用,0=禁用"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
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
	ID          uint      `json:"id" gorm:"primaryKey"`
	UserID      uint      `json:"userId" gorm:"not null;index;comment:授权中心用户ID"`
	UserIDField AuthUser  `json:"user,omitempty" gorm:"foreignKey:UserID"`
	GroupID     uint      `json:"groupId" gorm:"not null;index;comment:授权中心用户组ID"`
	GroupIDField AuthGroup `json:"group,omitempty" gorm:"foreignKey:GroupID"`
	GrantedBy   string    `json:"grantedBy" gorm:"size:50"`
	GrantedAt   time.Time `json:"grantedAt"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
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
	Username string `json:"username"`  // Basic Auth用户名
	Password string `json:"password"` // Basic Auth密码
	Token    string `json:"token"`     // API Token
}

// Application 外部应用注册模型
type Application struct {
	ID           uint                   `json:"id" gorm:"primaryKey"`
	Name         string                 `json:"name" gorm:"size:50;not null;uniqueIndex"`
	Code         string                 `json:"code" gorm:"size:50;not null;uniqueIndex"`
	Type         string                 `json:"type" gorm:"size:20;not null"`
	BaseURL      string                 `json:"baseUrl" gorm:"size:255;not null"`
	Endpoints    string                 `json:"endpoints" gorm:"type:json"`     // JSON: ApplicationEndpoints
	AuthConfig   string                 `json:"authConfig" gorm:"type:json"`    // JSON: ApplicationAuthConfig
	SyncInterval int                    `json:"syncInterval" gorm:"default:300"`
	LastSyncTime *time.Time             `json:"lastSyncTime"`
	Description  string                 `json:"description" gorm:"size:200"`
	Status       int                    `json:"status" gorm:"default:1"`
	CreatedAt    time.Time              `json:"createdAt"`
	UpdatedAt    time.Time              `json:"updatedAt"`
}

func (Application) TableName() string {
	return "applications"
}

// ApplicationRole 外部应用角色模型（同步过来的）
type ApplicationRole struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	AppID       uint      `json:"appId" gorm:"not null;index"`
	AppIDField  Application `json:"-" gorm:"foreignKey:AppID"`
	RoleCode    string    `json:"roleCode" gorm:"size:100;not null"`
	RoleName    string    `json:"roleName" gorm:"size:100;not null"`
	RoleType    string    `json:"roleType" gorm:"size:50;not null"`
	Description string    `json:"description" gorm:"size:200"`
	SyncTime    time.Time `json:"syncTime"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (ApplicationRole) TableName() string {
	return "application_roles"
}

// ApplicationUser 外部应用用户模型（同步过来的）
type ApplicationUser struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	AppID       uint      `json:"appId" gorm:"not null;index"`
	AppIDField  Application `json:"-" gorm:"foreignKey:AppID"`
	Username    string    `json:"username" gorm:"size:100;not null"`
	DisplayName string    `json:"displayName" gorm:"size:100"`
	Email       string    `json:"email" gorm:"size:100"`
	Status      string    `json:"status" gorm:"size:20;default:active"`
	SyncTime    time.Time `json:"syncTime"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (ApplicationUser) TableName() string {
	return "application_users"
}

// GroupBinding 用户组权限绑定模型（授权中心用户组 → 外部系统角色）
type GroupBinding struct {
	ID                   uint             `json:"id" gorm:"primaryKey"`
	GroupID              uint             `json:"groupId" gorm:"not null;index;comment:授权中心用户组ID"`
	GroupIDField         AuthGroup        `json:"group,omitempty" gorm:"foreignKey:GroupID"`
	AppID                uint             `json:"appId" gorm:"not null;index"`
	AppIDField           Application      `json:"-" gorm:"foreignKey:AppID"`
	ApplicationRoleID     uint             `json:"applicationRoleId" gorm:"not null;index"`
	ApplicationRole      ApplicationRole  `json:"applicationRole,omitempty" gorm:"foreignKey:ApplicationRoleID"`
	ApplicationPermissionID *uint            `json:"applicationPermissionId,omitempty" gorm:"comment:关联的权限ID（可选）"`
	CreatedAt            time.Time        `json:"createdAt"`
	UpdatedAt            time.Time        `json:"updatedAt"`
}

func (GroupBinding) TableName() string {
	return "group_bindings"
}

// ApplicationOperationLog 外部应用操作日志
type ApplicationOperationLog struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	AppID       uint      `json:"appId" gorm:"not null;index"`
	AppIDField  Application `json:"-" gorm:"foreignKey:AppID"`
	Operation   string    `json:"operation" gorm:"size:50;not null"`
	Target      string    `json:"target" gorm:"size:100"`
	RequestData string    `json:"requestData" gorm:"type:text"`
	ResponseData string   `json:"responseData" gorm:"type:text"`
	Status      string    `json:"status" gorm:"size:20"`
	ErrorMsg    string    `json:"errorMsg" gorm:"type:text"`
	Operator    string    `json:"operator" gorm:"size:50"`
	CreatedAt   time.Time `json:"createdAt"`
}

func (ApplicationOperationLog) TableName() string {
	return "application_operation_logs"
}
