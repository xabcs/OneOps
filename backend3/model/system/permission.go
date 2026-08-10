package modelsystem

import "time"

// Permission 权限定义模型
type Permission struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Code        string    `json:"code" gorm:"type:varchar(100);uniqueIndex;not null"`
	Name        string    `json:"name" gorm:"type:varchar(50);not null"`
	Description string    `json:"description" gorm:"type:varchar(200)"`
	Module      string    `json:"module" gorm:"type:varchar(30);not null;index"`
	Resource    string    `json:"resource" gorm:"type:varchar(30);not null;index"`
	Action      string    `json:"action" gorm:"type:varchar(20);not null"`
	RouteMethod string    `json:"routeMethod" gorm:"type:varchar(10)"`
	RoutePath   string    `json:"routePath" gorm:"type:varchar(255)"`
	Level       int       `json:"level" gorm:"type:tinyint;not null;default:3;index"`
	ParentID    *uint     `json:"parentId,omitempty" gorm:"index"`
	SortOrder   int       `json:"sortOrder" gorm:"default:0"`
	Status      int       `json:"status" gorm:"type:tinyint;not null;default:1;index"`
	CreatedAt   time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updatedAt" gorm:"autoUpdateTime"`

	Parent   *Permission  `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Children []Permission `json:"children,omitempty" gorm:"foreignKey:ParentID"`
}

func (Permission) TableName() string { return "sys_permissions" }

// RolePermission 角色-权限关联模型
type RolePermission struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	RoleID       uint      `json:"roleId" gorm:"not null;uniqueIndex:uk_role_permission;index"`
	PermissionID uint      `json:"permissionId" gorm:"not null;uniqueIndex:uk_role_permission;index"`
	CreatedAt    time.Time `json:"createdAt" gorm:"autoCreateTime"`

	Role       Role       `json:"role,omitempty" gorm:"foreignKey:RoleID"`
	Permission Permission `json:"permission,omitempty" gorm:"foreignKey:PermissionID"`
}

func (RolePermission) TableName() string { return "sys_role_permissions" }

// UserPermission 用户-权限关联模型
type UserPermission struct {
	ID           uint       `json:"id" gorm:"primaryKey"`
	UserID       uint       `json:"userId" gorm:"not null;uniqueIndex:uk_user_permission;index"`
	PermissionID uint       `json:"permissionId" gorm:"not null;uniqueIndex:uk_user_permission;index"`
	GrantedBy    *uint      `json:"grantedBy,omitempty" gorm:"index"`
	ExpireTime   *time.Time `json:"expireTime,omitempty" gorm:"index"`
	CreatedAt    time.Time  `json:"createdAt" gorm:"autoCreateTime"`

	User       User       `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Permission Permission `json:"permission,omitempty" gorm:"foreignKey:PermissionID"`
	Granter    *User      `json:"granter,omitempty" gorm:"foreignKey:GrantedBy"`
}

func (UserPermission) TableName() string { return "sys_user_permissions" }

// PermissionLog 权限操作日志模型
type PermissionLog struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	UserID         uint      `json:"userId" gorm:"not null;index"`
	PermissionCode string    `json:"permissionCode" gorm:"type:varchar(100);not null;index"`
	ResourceType   string    `json:"resourceType,omitempty" gorm:"type:varchar(50)"`
	ResourceID     *uint     `json:"resourceId,omitempty"`
	Action         string    `json:"action" gorm:"type:varchar(50);not null"`
	Result         string    `json:"result" gorm:"type:varchar(20);not null;index"`
	IPAddress      string    `json:"ipAddress,omitempty" gorm:"type:varchar(50)"`
	UserAgent      string    `json:"userAgent,omitempty" gorm:"type:varchar(500)"`
	CreatedAt      time.Time `json:"createdAt" gorm:"autoCreateTime;index"`

	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (PermissionLog) TableName() string { return "sys_permission_logs" }

func (up *UserPermission) IsExpired() bool {
	if up.ExpireTime == nil {
		return false
	}
	return time.Now().After(*up.ExpireTime)
}

// --- DTO ---

type GetTreePermissions struct {
	Permissions []Permission `json:"permissions"`
}

type CreatePermissionRequest struct {
	Code        string `json:"code" binding:"required,max=100"`
	Name        string `json:"name" binding:"required,max=50"`
	Description string `json:"description" binding:"max=200"`
	Module      string `json:"module" binding:"required,max=30"`
	Resource    string `json:"resource" binding:"required,max=30"`
	Action      string `json:"action" binding:"required,max=20"`
	Level       int    `json:"level" binding:"required,oneof=1 2 3 4"`
	ParentID    *uint  `json:"parentId,omitempty"`
	SortOrder   int    `json:"sortOrder"`
}

type UpdatePermissionRequest struct {
	Name        string `json:"name" binding:"required,max=50"`
	Description string `json:"description" binding:"max=200"`
	Level       int    `json:"level" binding:"required,oneof=1 2 3 4"`
	SortOrder   int    `json:"sortOrder"`
	Status      int    `json:"status" binding:"oneof=0 1"`
}

type AssignRolePermissionsRequest struct {
	PermissionIDs []uint `json:"permissionIds" binding:"required,min=1"`
}

type CheckPermissionRequest struct {
	UserID       uint   `json:"userId" binding:"required"`
	Permission   string `json:"permission" binding:"required"`
	ResourceType string `json:"resourceType,omitempty"`
	ResourceID   *uint  `json:"resourceId,omitempty"`
}

type CheckPermissionResponse struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason,omitempty"`
}

type GetUserPermissionsResponse struct {
	UserID      uint     `json:"userId"`
	Username    string   `json:"username"`
	Permissions []string `json:"permissions"`
	Roles       []string `json:"roles"`
}

type PermissionTreeNode struct {
	ID          uint                 `json:"id"`
	Code        string               `json:"code"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Level       int                  `json:"level"`
	Module      string               `json:"module"`
	Resource    string               `json:"resource"`
	Action      string               `json:"action"`
	ParentID    *uint                `json:"parentId,omitempty"`
	SortOrder   int                  `json:"sortOrder"`
	Status      int                  `json:"status"`
	Children    []PermissionTreeNode `json:"children,omitempty"`
	Checked     bool                 `json:"checked,omitempty"`
}

func (p *Permission) ToTreeNode(checked bool) PermissionTreeNode {
	node := PermissionTreeNode{
		ID:          p.ID,
		Code:        p.Code,
		Name:        p.Name,
		Description: p.Description,
		Level:       p.Level,
		Module:      p.Module,
		Resource:    p.Resource,
		Action:      p.Action,
		ParentID:    p.ParentID,
		SortOrder:   p.SortOrder,
		Status:      p.Status,
		Checked:     checked,
	}
	if len(p.Children) > 0 {
		node.Children = make([]PermissionTreeNode, len(p.Children))
		for i, child := range p.Children {
			node.Children[i] = child.ToTreeNode(checked)
		}
	}
	return node
}
