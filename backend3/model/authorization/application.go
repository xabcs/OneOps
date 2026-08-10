package modelauth

import (
	"time"
)

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
