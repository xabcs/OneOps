package models

import (
	"time"

	"gorm.io/gorm"
)

// APIResource API资源模型（用于可视化管理）
type APIResource struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"type:varchar(100);not null;comment:API名称"`
	Path        string         `json:"path" gorm:"type:varchar(200);not null;uniqueIndex:idx_path_method;comment:API路径"`
	Method      string         `json:"method" gorm:"type:varchar(10);not null;uniqueIndex:idx_path_method;comment:HTTP方法"`
	Description string         `json:"description" gorm:"type:varchar(200);comment:API描述"`
	Category    string         `json:"category" gorm:"type:varchar(50);comment:API分类"`
	Module      string         `json:"module" gorm:"type:varchar(50);index;comment:所属模块"`
	ParentID    uint           `json:"parentId" gorm:"default:0;comment:父级ID（支持分组）"`
	Sort        int            `json:"sort" gorm:"default:0;comment:排序"`
	Status      int            `json:"status" gorm:"default:1;comment:状态:1启用,0禁用"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	Children    []*APIResource `json:"children,omitempty" gorm:"-"`
}

// TableName 指定表名
func (APIResource) TableName() string {
	return "sys_api_resources"
}
