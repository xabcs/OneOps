package models

import (
	"time"
)

// AttributeDefinition 属性定义
type AttributeDefinition struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Name         string    `json:"name" gorm:"size:100;not null"`              // 属性名称
	Key          string    `json:"key" gorm:"size:50;not null;uniqueIndex;column:key"`     // 属性键
	Category     string    `json:"category" gorm:"size:50;not null;index"`      // 分类
	Type         string    `json:"type" gorm:"size:20;not null;default:'text'"` // 类型
	Options      string    `json:"options" gorm:"type:text"`                    // 选项JSON
	Required     bool      `json:"required" gorm:"default:false"`               // 是否必填
	DefaultValue string    `json:"defaultValue" gorm:"size:255"`               // 默认值
	SortOrder    int       `json:"sortOrder" gorm:"default:0"`                  // 排序
	Status       int       `json:"status" gorm:"default:1"`                     // 状态
	Description  string    `json:"description" gorm:"type:text"`                 // 说明
	CreatedAt    time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updatedAt" gorm:"autoUpdateTime"`

	// 关联
	ServerAttributes []ServerAttribute `json:"serverAttributes,omitempty" gorm:"foreignKey:AttributeID"`
}

// AttributeOption 属性选项（用于 select/multiselect 类型）
type AttributeOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// AttributeOptions 属性选项列表
type AttributeOptions []AttributeOption

// TableName 指定表名
func (AttributeDefinition) TableName() string {
	return "sys_attribute_definitions"
}

// ServerAttribute 主机属性值
type ServerAttribute struct {
	ID              uint                `json:"id" gorm:"primaryKey"`
	ServerID        uint                `json:"serverId" gorm:"not null;index"`
	AttributeID     uint                `json:"attributeId" gorm:"not null;index"`
	AttributeKey    string              `json:"attributeKey" gorm:"size:50;not null;index"`
	AttributeValue string              `json:"attributeValue" gorm:"type:text"`
	ValueType       string              `json:"valueType" gorm:"size:20;default:'string'"`
	Category        string              `json:"category" gorm:"size:50"`
	CreatedAt       time.Time           `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt       time.Time           `json:"updatedAt" gorm:"autoUpdateTime"`

	// 关联
	Server     *Server             `json:"server,omitempty" gorm:"foreignKey:ServerID"`
	Definition *AttributeDefinition `json:"definition,omitempty" gorm:"foreignKey:AttributeID"`
}

// TableName 指定表名
func (ServerAttribute) TableName() string {
	return "cmdb_server_attributes"
}

// AttributeCategory 属性分类常量
const (
	AttrCategorySystem      = "system"      // 系统分类
	AttrCategoryLocation    = "location"    // 地理位置
	AttrCategoryEnvironment = "environment" // 环境信息
	AttrCategoryHardware    = "hardware"    // 硬件配置
	AttrCategoryCustom      = "custom"      // 自定义
)

// AttributeType 属性类型常量
const (
	AttrTypeText        = "text"        // 单行文本
	AttrTypeSelect      = "select"      // 下拉单选
	AttrTypeMultiselect = "multiselect" // 下拉多选
	AttrTypeNumber      = "number"      // 数字
	AttrTypeDate        = "date"        // 日期
	AttrTypeBoolean     = "boolean"     // 布尔值
)
