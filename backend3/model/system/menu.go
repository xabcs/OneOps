package modelsystem

import "time"

// Menu 菜单模型
type Menu struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	Name       string    `json:"name" gorm:"size:50;not null"`
	Icon       string    `json:"icon" gorm:"size:50"`
	Path       string    `json:"path" gorm:"size:200"`
	Permission string    `json:"permission" gorm:"size:100"`
	Resource   string    `json:"resource" gorm:"size:30;index"`
	MenuType   string    `json:"menuType" gorm:"size:20;default:menu"`
	// OpenType 打开方式：internal=站内跳转（默认），new-tab=新标签页打开（配合 meta.href 下发）
	OpenType   string    `json:"openType" gorm:"size:20;default:internal"`
	ParentID   uint      `json:"parentId" gorm:"default:0"`
	Sort       int       `json:"sort" gorm:"default:0"`
	Status     int       `json:"status" gorm:"default:1"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
	Children   []*Menu   `json:"children,omitempty" gorm:"-"`
}

func (Menu) TableName() string { return "sys_menus" }
