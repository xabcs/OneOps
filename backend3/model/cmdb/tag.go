package modelcmdb

import "time"

// ServerTag 服务器标签模型
type ServerTag struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"size:50;not null;uniqueIndex"` // 标签名称
	Color       string    `json:"color" gorm:"size:20;default:'#409EFF'"`   // 标签颜色
	Description string    `json:"description" gorm:"size:200"`              // 标签描述
	SortOrder   int       `json:"sortOrder" gorm:"default:0"`               // 排序
	Status      int       `json:"status" gorm:"default:1"`                  // 状态：1启用 0禁用
	CreatedAt   time.Time `json:"createdAt" gorm:"autoCreateTime"`

	// 关联
	Servers []Server `json:"servers,omitempty" gorm:"many2many:cmdb_server_tag_relations;joinForeignKey:TagID;joinReferences:ServerID"`
}

// TableName 指定表名
func (ServerTag) TableName() string {
	return "cmdb_server_tags"
}

// AssetChange 资产变更记录模型
type AssetChange struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	AssetType   string    `json:"assetType" gorm:"size:50;not null"`          // 资产类型
	AssetID     uint      `json:"assetId" gorm:"not null;index"`              // 资产ID
	AssetName   string    `json:"assetName" gorm:"size:100"`                  // 资产名称
	FieldName   string    `json:"fieldName" gorm:"size:50;not null"`          // 变更字段
	OldValue    string    `json:"oldValue" gorm:"type:text"`                  // 旧值
	NewValue    string    `json:"newValue" gorm:"type:text"`                  // 新值
	ChangeType  string    `json:"changeType" gorm:"size:20;default:'update'"` // 变更类型
	Operator    string    `json:"operator" gorm:"size:100"`                   // 操作人
	OperatorID  uint      `json:"operatorId" gorm:"index"`                    // 操作人ID
	OperateTime time.Time `json:"operateTime" gorm:"autoCreateTime"`          // 操作时间
	Remarks     string    `json:"remarks" gorm:"type:text"`                   // 备注
}

// TableName 指定表名
func (AssetChange) TableName() string {
	return "cmdb_asset_changes"
}

// ServerTagRelation 服务器标签关联模型
type ServerTagRelation struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	ServerID  uint      `json:"serverId" gorm:"not null;index"` // 服务器ID
	TagID     uint      `json:"tagId" gorm:"not null;index"`    // 标签ID
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`

	// 关联
	Server *Server    `json:"server,omitempty" gorm:"foreignKey:ServerID"`
	Tag    *ServerTag `json:"tag,omitempty" gorm:"foreignKey:TagID"`
}

// TableName 指定表名
func (ServerTagRelation) TableName() string {
	return "cmdb_server_tag_relations"
}

// ServerGroup 主机分组模型
type ServerGroup struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"size:100;not null"`            // 分组名称
	Code        string    `json:"code" gorm:"size:50;not null;uniqueIndex"` // 分组代码
	ParentID    uint      `json:"parentId" gorm:"default:0;index"`          // 父分组ID
	Level       int       `json:"level" gorm:"default:1"`                   // 层级
	Description string    `json:"description" gorm:"type:text"`             // 分组描述
	Color       string    `json:"color" gorm:"size:20;default:'#409EFF'"`   // 分组颜色
	Icon        string    `json:"icon" gorm:"size:50;default:'mdi:folder'"` // 分组图标
	SortOrder   int       `json:"sortOrder" gorm:"default:0"`               // 排序
	Status      int       `json:"status" gorm:"default:1"`                  // 状态：1启用 0禁用
	CreatedAt   time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updatedAt" gorm:"autoUpdateTime"`

	// 关联
	Parent      *ServerGroup  `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Children    []ServerGroup `json:"children,omitempty" gorm:"foreignKey:ParentID"`
	Servers     []Server      `json:"servers,omitempty" gorm:"many2many:cmdb_server_group_relations;joinForeignKey:GroupID;joinReferences:ServerID"`
	ServerCount int           `json:"serverCount" gorm:"-"` // 直接关联的主机数量（不递归）
}

// TableName 指定表名
func (ServerGroup) TableName() string {
	return "cmdb_server_groups"
}

// ServerGroupRelation 服务器分组关联模型
type ServerGroupRelation struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	ServerID  uint      `json:"serverId" gorm:"not null;index"` // 服务器ID
	GroupID   uint      `json:"groupId" gorm:"not null;index"`  // 分组ID
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`

	// 关联
	Server *Server      `json:"server,omitempty" gorm:"foreignKey:ServerID"`
	Group  *ServerGroup `json:"group,omitempty" gorm:"foreignKey:GroupID"`
}

// TableName 指定表名
func (ServerGroupRelation) TableName() string {
	return "cmdb_server_group_relations"
}
