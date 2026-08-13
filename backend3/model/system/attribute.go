package modelsystem

import "time"

// AttributeDefinition 属性定义
// 注意: ServerAttributes 关联已移除（原 gorm:"foreignKey:AttributeID" 会导致 system→cmdb 循环依赖）
// 如需查询某属性关联的服务器属性值，请在 repository 层手动 Join
type AttributeDefinition struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Name         string    `json:"name" gorm:"size:100;not null"`
	Key          string    `json:"key" gorm:"size:50;not null;uniqueIndex;column:key"`
	Category     string    `json:"category" gorm:"size:50;not null;index"`
	Type         string    `json:"type" gorm:"size:20;not null;default:'text'"`
	Options      string    `json:"options" gorm:"type:text"`
	Required     bool      `json:"required" gorm:"default:false"`
	DefaultValue string    `json:"defaultValue" gorm:"size:255"`
	SortOrder    int       `json:"sortOrder" gorm:"default:0"`
	Status       int       `json:"status" gorm:"default:1"`
	Description  string    `json:"description" gorm:"type:text"`
	CreatedAt    time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

func (AttributeDefinition) TableName() string { return "cmdb_attribute_definitions" }

// AttributeOption 属性选项
type AttributeOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type AttributeOptions []AttributeOption

const (
	AttrCategorySystem      = "system"
	AttrCategoryLocation    = "location"
	AttrCategoryEnvironment = "environment"
	AttrCategoryHardware    = "hardware"
	AttrCategoryCustom      = "custom"
)

const (
	AttrTypeText    = "text"
	AttrTypeSelect  = "select"
	AttrTypeNumber  = "number"
	AttrTypeBoolean = "boolean"
)
