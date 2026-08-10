package models

import (
	"time"
)

// Role 角色模型
type Role struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Name        string `json:"name" gorm:"size:50;not null"`
	Code        string `json:"code" gorm:"uniqueIndex;size:50;not null"`
	Description string `json:"description" gorm:"size:200"`
	// MenuIds 字段已移除 - 权限通过 role_permissions 表管理
	Status    int       `json:"status" gorm:"default:1"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// TableName 指定表名
func (Role) TableName() string {
	return "sys_roles"
}
