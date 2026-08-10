package models

import (
	"time"
)

// Permission 权限定义模型
type Permission struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Code        string    `json:"code" gorm:"type:varchar(100);uniqueIndex;not null"`
	Name        string    `json:"name" gorm:"type:varchar(50);not null"`
	Description string    `json:"description" gorm:"type:varchar(200)"`
	Module      string    `json:"module" gorm:"type:varchar(30);not null;index"`
	Resource    string    `json:"resource" gorm:"type:varchar(30);not null;index"`
	Action      string    `json:"action" gorm:"type:varchar(20);not null"`
	RouteMethod string    `json:"routeMethod" gorm:"type:varchar(10)"`                // HTTP方法: GET/POST/PUT/DELETE
	RoutePath   string    `json:"routePath" gorm:"type:varchar(255)"`                 // 路由路径
	Level       int       `json:"level" gorm:"type:tinyint;not null;default:3;index"` // 权限级别：1-模块级，2-页面级，3-按钮级，4-API级
	ParentID    *uint     `json:"parentId,omitempty" gorm:"index"`
	SortOrder   int       `json:"sortOrder" gorm:"default:0"`
	Status      int       `json:"status" gorm:"type:tinyint;not null;default:1;index"`
	CreatedAt   time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updatedAt" gorm:"autoUpdateTime"`

	// 关联关系
	Parent   *Permission  `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Children []Permission `json:"children,omitempty" gorm:"foreignKey:ParentID"`
}

// TableName 指定表名
func (Permission) TableName() string {
	return "sys_permissions"
}

// RolePermission 角色-权限关联模型
type RolePermission struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	RoleID       uint      `json:"roleId" gorm:"not null;uniqueIndex:uk_role_permission;index"`
	PermissionID uint      `json:"permissionId" gorm:"not null;uniqueIndex:uk_role_permission;index"`
	CreatedAt    time.Time `json:"createdAt" gorm:"autoCreateTime"`

	// 关联关系
	Role       Role       `json:"role,omitempty" gorm:"foreignKey:RoleID"`
	Permission Permission `json:"permission,omitempty" gorm:"foreignKey:PermissionID"`
}

// TableName 指定表名
func (RolePermission) TableName() string {
	return "sys_role_permissions"
}

// UserPermission 用户-权限关联模型（可选，用于用户级权限覆盖）
type UserPermission struct {
	ID           uint       `json:"id" gorm:"primaryKey"`
	UserID       uint       `json:"userId" gorm:"not null;uniqueIndex:uk_user_permission;index"`
	PermissionID uint       `json:"permissionId" gorm:"not null;uniqueIndex:uk_user_permission;index"`
	GrantedBy    *uint      `json:"grantedBy,omitempty" gorm:"index"`  // 授权人ID
	ExpireTime   *time.Time `json:"expireTime,omitempty" gorm:"index"` // 权限过期时间
	CreatedAt    time.Time  `json:"createdAt" gorm:"autoCreateTime"`

	// 关联关系
	User       User       `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Permission Permission `json:"permission,omitempty" gorm:"foreignKey:PermissionID"`
	Granter    *User      `json:"granter,omitempty" gorm:"foreignKey:GrantedBy"`
}

// TableName 指定表名
func (UserPermission) TableName() string {
	return "sys_user_permissions"
}

// PermissionLog 权限操作日志模型（审计用途）
type PermissionLog struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	UserID         uint      `json:"userId" gorm:"not null;index"`
	PermissionCode string    `json:"permissionCode" gorm:"type:varchar(100);not null;index"`
	ResourceType   string    `json:"resourceType,omitempty" gorm:"type:varchar(50)"`
	ResourceID     *uint     `json:"resourceId,omitempty"`
	Action         string    `json:"action" gorm:"type:varchar(50);not null"`
	Result         string    `json:"result" gorm:"type:varchar(20);not null;index"` // allowed, denied
	IPAddress      string    `json:"ipAddress,omitempty" gorm:"type:varchar(50)"`
	UserAgent      string    `json:"userAgent,omitempty" gorm:"type:varchar(500)"`
	CreatedAt      time.Time `json:"createdAt" gorm:"autoCreateTime;index"`

	// 关联关系
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName 指定表名
func (PermissionLog) TableName() string {
	return "sys_permission_logs"
}

// IsExpired 检查用户权限是否过期
func (up *UserPermission) IsExpired() bool {
	if up.ExpireTime == nil {
		return false
	}
	return time.Now().After(*up.ExpireTime)
}

// GetTreePermissions 获取权限树结构（用于权限管理界面）
type GetTreePermissions struct {
	Permissions []Permission `json:"permissions"`
}

// CreatePermissionRequest 创建权限请求
type CreatePermissionRequest struct {
	Code        string `json:"code" binding:"required,max=100"`
	Name        string `json:"name" binding:"required,max=50"`
	Description string `json:"description" binding:"max=200"`
	Module      string `json:"module" binding:"required,max=30"`
	Resource    string `json:"resource" binding:"required,max=30"`
	Action      string `json:"action" binding:"required,max=20"`
	Level       int    `json:"level" binding:"required,oneof=1 2 3 4"` // 1-模块级，2-页面级，3-按钮级，4-API级
	ParentID    *uint  `json:"parentId,omitempty"`
	SortOrder   int    `json:"sortOrder"`
}

// UpdatePermissionRequest 更新权限请求
type UpdatePermissionRequest struct {
	Name        string `json:"name" binding:"required,max=50"`
	Description string `json:"description" binding:"max=200"`
	Level       int    `json:"level" binding:"required,oneof=1 2 3 4"` // 1-模块级，2-页面级，3-按钮级，4-API级
	SortOrder   int    `json:"sortOrder"`
	Status      int    `json:"status" binding:"oneof=0 1"`
}

// AssignRolePermissionsRequest 分配角色权限请求
type AssignRolePermissionsRequest struct {
	PermissionIDs []uint `json:"permissionIds" binding:"required,min=1"`
}

// CheckPermissionRequest 检查权限请求（内部使用）
type CheckPermissionRequest struct {
	UserID       uint   `json:"userId" binding:"required"`
	Permission   string `json:"permission" binding:"required"`
	ResourceType string `json:"resourceType,omitempty"`
	ResourceID   *uint  `json:"resourceId,omitempty"`
}

// CheckPermissionResponse 检查权限响应
type CheckPermissionResponse struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason,omitempty"`
}

// GetUserPermissionsResponse 获取用户权限响应
type GetUserPermissionsResponse struct {
	UserID      uint     `json:"userId"`
	Username    string   `json:"username"`
	Permissions []string `json:"permissions"`
	Roles       []string `json:"roles"`
}

// PermissionTreeNode 权限树节点（用于前端展示）
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
	Checked     bool                 `json:"checked,omitempty"` // 用于权限分配界面
}

// ToTreeNode 转换为树节点
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
