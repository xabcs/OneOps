package models

import (
	"time"
)

// Menu 菜单模型
type Menu struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	Name       string    `json:"name" gorm:"size:50;not null"`
	Icon       string    `json:"icon" gorm:"size:50"`
	Path       string    `json:"path" gorm:"size:200"`
	Permission string    `json:"permission" gorm:"size:100"`
	ParentID   uint      `json:"parentId" gorm:"default:0"`
	Sort       int       `json:"sort" gorm:"default:0"`
	Status     int       `json:"status" gorm:"default:1"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
	Children   []*Menu   `json:"children,omitempty" gorm:"-"` // 不存储到数据库，仅用于返回
}

// TableName 指定表名
func (Menu) TableName() string {
	return "menus"
}
