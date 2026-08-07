package models

import (
	"time"
)

// AuthGroupPermissionMapping 授权中心用户组权限映射表
// 核心映射关系：授权中心用户组 ←→ 外部应用权限对象
type AuthGroupPermissionMapping struct {
	ID    uint   `json:"id" gorm:"primaryKey"`

	// ========== 授权中心侧 ==========
	AuthGroupID uint `json:"authGroupId" gorm:"not null;index:idx_auth_app;comment:授权中心用户组ID"`
	AuthGroupIDField AuthGroup `json:"authGroup,omitempty" gorm:"foreignKey:AuthGroupID"`

	// ========== 外部应用侧 ==========
	AppID      uint       `json:"appId" gorm:"not null;index:idx_auth_app;index:idx_app;comment:外部应用ID"`
	AppIDField Application `json:"-" gorm:"foreignKey:AppID"`

	// ========== 映射对象（支持多种外部系统） ==========
	// JumpServer: role(角色), rule(授权规则), asset(资产), node(节点)
	// Jenkins: role(角色)
	// GitLab: role(角色), project(项目), group(GitLab组)
	MappingType   string `json:"mappingType" gorm:"size:20;not null;comment:权限对象类型"`
	ExternalID    string `json:"externalId" gorm:"size:100;comment:外部系统中的对象ID"`
	ExternalName  string `json:"externalName" gorm:"size:200;comment:外部系统中的对象名称（用于显示）"`

	// ========== 权限详情（灵活支持不同系统） ==========
	// JumpServer: {"actions":["connect","upload"],"assets":["server-1","server-2"],"protocols":["ssh"]}
	// Jenkins: {"jobs":["job-1","job-2"],"permissions":["build","read"]}
	// GitLab: {"access_level":30,"permissions":["push","issue"]}
	PermissionDetail string `json:"permissionDetail" gorm:"type:json;comment:权限详情(JSON格式)"`

	// ========== 状态与管理 ==========
	IsEnabled  bool       `json:"isEnabled" gorm:"default:true;comment:是否启用"`
	Priority   int        `json:"priority" gorm:"default:0;comment:优先级"`
	GrantedBy  string     `json:"grantedBy" gorm:"size:50;comment:授权人"`
	GrantedAt  time.Time  `json:"grantedAt;comment:授权时间"`
	ExpireTime *time.Time `json:"expireTime;comment:过期时间"`

	// ========== 审计字段 ==========
	LastSyncedAt   *time.Time `json:"lastSyncedAt;comment:最后同步时间"`
	SyncStatus     string     `json:"syncStatus" gorm:"size:20;default:pending;comment:同步状态:pending,success,failed"`
	SyncErrorMessage string    `json:"syncErrorMessage" gorm:"type:text;comment:同步错误信息"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (AuthGroupPermissionMapping) TableName() string {
	return "auth_group_permission_mappings"
}

// AuthGroupEffectivePermission 用户组有效权限视图
// 用于查询和展示用户组的实际权限
type AuthGroupEffectivePermission struct {
	ID              uint   `json:"id"`
	AuthGroupID     uint   `json:"authGroupId"`
	AuthGroupName   string `json:"authGroupName"`
	AppID           uint   `json:"appId"`
	AppName         string `json:"appName"`
	AppType         string `json:"appType"`
	MappingType     string `json:"mappingType"`
	ExternalID      string `json:"externalId"`
	ExternalName    string `json:"externalName"`
	PermissionDetail string `json:"permissionDetail"`
	IsEnabled       bool   `json:"isEnabled"`
	ExpireTime      *time.Time `json:"expireTime"`
}

// PermissionMappingResult 权限映射操作结果
type PermissionMappingResult struct {
	Success      bool     `json:"success"`
	Message      string   `json:"message"`
	MappingID    uint     `json:"mappingId,omitempty"`
	ExternalID   string   `json:"externalId,omitempty"`
	Warnings     []string `json:"warnings,omitempty"`
}

// PermissionSyncStatus 权限同步状态
type PermissionSyncStatus struct {
	MappingID          uint    `json:"mappingId"`
	AuthGroupID        uint    `json:"authGroupId"`
	AppID              uint    `json:"appId"`
	Status             string  `json:"status"` // pending, success, failed
	LastSyncedAt       *time.Time `json:"lastSyncedAt"`
	ErrorMessage       string  `json:"errorMessage,omitempty"`
	NextRetryAt        *time.Time `json:"nextRetryAt"`
	RetryCount         int     `json:"retryCount"`
}
