package modelsystem

import "time"

// UserGroup 平台用户组（团队维度，用于批量授权）
// 与授权中心 auth_groups（堡垒机独立账号体系）无关，组员是平台登录用户 sys_users
// 组本身不持有全局功能权限（那属于 sys_roles），仅作为集群等资源的批量授权主体
type UserGroup struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Code        string    `json:"code" gorm:"size:50;not null;uniqueIndex"`
	Name        string    `json:"name" gorm:"size:50;not null"`
	Description string    `json:"description" gorm:"size:200"`
	Status      int       `json:"status" gorm:"default:1;comment:1=启用,0=停用"`
	CreatedAt   time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

func (UserGroup) TableName() string { return "sys_user_groups" }

// UserGroupMember 用户组成员关联
type UserGroupMember struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	GroupID   uint      `json:"groupId" gorm:"not null;uniqueIndex:uk_group_member"`
	UserID    uint      `json:"userId" gorm:"not null;uniqueIndex:uk_group_member;index"`
	User      *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
}

func (UserGroupMember) TableName() string { return "sys_user_group_members" }

// --- DTO ---

type CreateUserGroupRequest struct {
	Code        string `json:"code" binding:"required,max=50"`
	Name        string `json:"name" binding:"required,max=50"`
	Description string `json:"description" binding:"max=200"`
}

type UpdateUserGroupRequest struct {
	Name        string `json:"name" binding:"required,max=50"`
	Description string `json:"description" binding:"max=200"`
	Status      *int   `json:"status" binding:"omitempty,oneof=0 1"`
}
