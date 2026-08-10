package modelauth

import (
	"time"
)

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
