package modelcmdb

import (
	"time"

	modelsystem "oneops/backend3/model/system"
)

// ServerAttribute 主机属性值
type ServerAttribute struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	ServerID       uint      `json:"serverId" gorm:"not null;index"`
	AttributeID    uint      `json:"attributeId" gorm:"not null;index"`
	AttributeKey   string    `json:"attributeKey" gorm:"size:50;not null;index"`
	AttributeValue string    `json:"attributeValue" gorm:"type:text"`
	ValueType      string    `json:"valueType" gorm:"size:20;default:'string'"`
	Category       string    `json:"category" gorm:"size:50"`
	CreatedAt      time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updatedAt" gorm:"autoUpdateTime"`

	// 关联
	Server     *Server                          `json:"server,omitempty" gorm:"foreignKey:ServerID"`
	Definition *modelsystem.AttributeDefinition `json:"definition,omitempty" gorm:"foreignKey:AttributeID"`
}

// TableName 指定表名
func (ServerAttribute) TableName() string {
	return "cmdb_server_attributes"
}
