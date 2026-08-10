package modelcmdb

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// BusinessUnit 业务系统模型
type BusinessUnit struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"size:100;not null"`            // 业务名称
	Code      string    `json:"code" gorm:"size:50;not null;uniqueIndex"` // 业务代码
	ParentID  uint      `json:"parentId" gorm:"default:0;index"`          // 父业务ID
	Level     int       `json:"level" gorm:"default:1"`                   // 层级
	Owner     string    `json:"owner" gorm:"size:100"`                    // 负责人
	Phone     string    `json:"phone" gorm:"size:20"`                     // 联系电话
	Email     string    `json:"email" gorm:"size:100"`                    // 联系邮箱
	SortOrder int       `json:"sortOrder" gorm:"default:0"`               // 排序
	Status    int       `json:"status" gorm:"default:1"`                  // 状态：1启用 0禁用
	Remarks   string    `json:"remarks" gorm:"type:text"`                 // 备注
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"autoUpdateTime"`

	// 关联
	Children []BusinessUnit `json:"children,omitempty" gorm:"-"` // 子业务（不映射到数据库）
}

// TableName 指定表名
func (BusinessUnit) TableName() string {
	return "cmdb_business_units"
}

// ServerRoom 机房模型
type ServerRoom struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"size:100;not null"`            // 机房名称
	Code      string    `json:"code" gorm:"size:50;not null;uniqueIndex"` // 机房代码
	Location  string    `json:"location" gorm:"size:200"`                 // 机房位置
	Address   string    `json:"address" gorm:"size:500"`                  // 详细地址
	Provider  string    `json:"provider" gorm:"size:100"`                 // 服务商
	Contact   string    `json:"contact" gorm:"size:100"`                  // 联系人
	Phone     string    `json:"phone" gorm:"size:20"`                     // 联系电话
	Status    int       `json:"status" gorm:"default:1"`                  // 状态：1启用 0禁用
	Remarks   string    `json:"remarks" gorm:"type:text"`                 // 备注
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"autoUpdateTime"`

	// 关联
	Cabinets []Cabinet `json:"cabinets,omitempty" gorm:"foreignKey:RoomID"`
}

// TableName 指定表名
func (ServerRoom) TableName() string {
	return "cmdb_server_rooms"
}

// Cabinet 机柜模型
type Cabinet struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	Name          string    `json:"name" gorm:"size:100;not null"`            // 机柜名称
	Code          string    `json:"code" gorm:"size:50;not null;uniqueIndex"` // 机柜代码
	RoomID        uint      `json:"roomId" gorm:"index"`                      // 机房ID
	Position      string    `json:"position" gorm:"size:50"`                  // 位置
	Capacity      int       `json:"capacity" gorm:"default:42"`               // U数
	UsedU         int       `json:"usedU" gorm:"default:0"`                   // 已用U数
	PowerUsage    float64   `json:"powerUsage" gorm:"default:0.00"`           // 已用电力(KW)
	PowerCapacity float64   `json:"powerCapacity" gorm:"default:0.00"`        // 总电力(KW)
	Status        int       `json:"status" gorm:"default:1"`                  // 状态：1启用 0禁用
	Remarks       string    `json:"remarks" gorm:"type:text"`                 // 备注
	CreatedAt     time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updatedAt" gorm:"autoUpdateTime"`

	// 关联
	Room    *ServerRoom `json:"room,omitempty" gorm:"foreignKey:RoomID"`
	Servers []Server    `json:"servers,omitempty" gorm:"foreignKey:CabinetID"`
}

// TableName 指定表名
func (Cabinet) TableName() string {
	return "cmdb_cabinets"
}

// DiskPartition 磁盘分区信息
type DiskPartition struct {
	Mount string  `json:"mount"` // 挂载点，如 /var, /home
	Usage float64 `json:"usage"` // 使用率百分比
}

// Scan 实现 sql.Scanner 接口，用于从数据库读取 JSON 数据
func (dp *DiskPartition) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal DiskPartition value: %v", value)
	}
	return json.Unmarshal(bytes, dp)
}

// Value 实现 driver.Valuer 接口，用于将数据写入数据库
func (dp DiskPartition) Value() (driver.Value, error) {
	return json.Marshal(dp)
}

// DiskPartitions 磁盘分区切片类型，用于实现 JSON 序列化
type DiskPartitions []DiskPartition

// Scan 实现 sql.Scanner 接口
func (dp *DiskPartitions) Scan(value interface{}) error {
	if value == nil {
		*dp = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		// 尝试字符串类型
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("failed to unmarshal DiskPartitions value: %v", value)
		}
		bytes = []byte(str)
	}
	return json.Unmarshal(bytes, dp)
}

// Value 实现 driver.Valuer 接口
func (dp DiskPartitions) Value() (driver.Value, error) {
	if len(dp) == 0 {
		return nil, nil
	}
	return json.Marshal(dp)
}

// Server 服务器模型
type Server struct {
	ID                 uint           `json:"id" gorm:"primaryKey"`
	Hostname           string         `json:"hostname" gorm:"size:100;not null;uniqueIndex"`                                       // 主机名
	IP                 string         `json:"ip" gorm:"size:50;not null;index"`                                                    // 外网IP
	InnerIP            string         `json:"innerIp" gorm:"size:50;index"`                                                        // 内网IP
	CPU                int            `json:"cpu" gorm:"default:0"`                                                                // CPU核心数
	Memory             int            `json:"memory" gorm:"default:0"`                                                             // 内存(GB)
	Disk               int            `json:"disk" gorm:"default:0"`                                                               // 磁盘(GB)
	OS                 string         `json:"os" gorm:"size:50"`                                                                   // 操作系统
	OSVersion          string         `json:"osVersion" gorm:"size:50"`                                                            // 系统版本
	Arch               string         `json:"arch" gorm:"size:20;default:'x86_64'"`                                                // 系统架构
	Env                string         `json:"env" gorm:"type:varchar(20);default:'test';index"`                                    // 环境
	Status             string         `json:"status" gorm:"type:varchar(20);default:'unknown';index"`                              // 状态
	SSHPort            int            `json:"sshPort" gorm:"default:22"`                                                           // SSH端口
	CredentialID       uint           `json:"credentialId" gorm:"index"`                                                           // SSH凭证ID（兼容旧字段）
	SSHCredentialID    uint           `json:"sshCredentialId" gorm:"index"`                                                        // SSH凭证ID
	CabinetID          uint           `json:"cabinetId" gorm:"index"`                                                              // 所在机柜ID
	UPosition          int            `json:"uPosition"`                                                                           // 机柜位置(U)
	SN                 string         `json:"sn" gorm:"size:100"`                                                                  // 序列号
	Manufacturer       string         `json:"manufacturer" gorm:"size:100"`                                                        // 厂商
	Model              string         `json:"model" gorm:"size:100"`                                                               // 型号
	PurchaseDate       *time.Time     `json:"purchaseDate"`                                                                        // 购买日期
	ExpireWarranty     *time.Time     `json:"expireWarranty"`                                                                      // 保修到期
	AssetNumber        string         `json:"assetNumber" gorm:"size:100"`                                                         // 资产编号
	InstanceID         string         `json:"instanceId" gorm:"size:100"`                                                          // 云主机实例ID
	InstanceType       string         `json:"instanceType" gorm:"size:50"`                                                         // 云主机类型
	Region             string         `json:"region" gorm:"size:50"`                                                               // 区域
	Zone               string         `json:"zone" gorm:"size:50"`                                                                 // 可用区
	Provider           string         `json:"provider" gorm:"size:50;index"`                                                       // 服务商
	ServerType         string         `json:"serverType" gorm:"size:20;default:'vm'"`                                              // 类型
	Remarks            string         `json:"remarks" gorm:"type:text"`                                                            // 备注
	LastCheckTime      *time.Time     `json:"lastCheckTime"`                                                                       // 最后连通性检查时间
	LastConnectTime    *time.Time     `json:"lastConnectTime"`                                                                     // 最后连接时间
	ConnectivityStatus string         `json:"connectivityStatus" gorm:"type:enum('online','offline','unknown');default:'unknown'"` // 连通性状态
	BusinessID         uint           `json:"businessId" gorm:"index"`
	Business           *BusinessUnit  `json:"business,omitempty" gorm:"foreignKey:BusinessID;constraint:OnDelete:SET NULL"`
	CPUUsage           float64        `json:"cpuUsage" gorm:"default:0"`
	MemoryUsage        float64        `json:"memoryUsage" gorm:"default:0"`
	DiskUsage          float64        `json:"diskUsage" gorm:"default:0"`
	Load1              float64        `json:"load1" gorm:"default:0"`
	Load5              float64        `json:"load5" gorm:"default:0"`
	Load15             float64        `json:"load15" gorm:"default:0"`
	DiskPartitions     DiskPartitions `json:"diskPartitions,omitempty" gorm:"type:json"`
	MetricsUpdatedAt   *time.Time     `json:"metricsUpdatedAt"`
	AgentStatus        string         `json:"agentStatus" gorm:"type:varchar(20);default:'uninstalled';index"` // uninstalled | running | offline
	AgentPort          int            `json:"agentPort" gorm:"default:9100"`
	AgentVersion       string         `json:"agentVersion" gorm:"size:50"`
	LastHeartbeatAt    *time.Time     `json:"lastHeartbeatAt"`
	SystemCredentialID uint           `json:"systemCredentialId" gorm:"index"` // 系统运维凭证（Agent部署/采集专用）
	// 冗余字段：优化列表查询性能（避免关联查询）
	GroupNames      string    `json:"groupNames" gorm:"type:varchar(500);default:'[]'"`      // 分组名称JSON数组
	CredentialNames string    `json:"credentialNames" gorm:"type:varchar(500);default:'[]'"` // 凭证名称JSON数组
	CreatedAt       time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt       time.Time `json:"updatedAt" gorm:"autoUpdateTime"`

	// 关联
	SSHCredential    *SSHCredential    `json:"sshCredential,omitempty" gorm:"foreignKey:SSHCredentialID;constraint:OnDelete:SET NULL"`
	SystemCredential *SSHCredential    `json:"systemCredential,omitempty" gorm:"foreignKey:SystemCredentialID;constraint:OnDelete:SET NULL"`
	Cabinet          *Cabinet          `json:"cabinet,omitempty" gorm:"foreignKey:CabinetID;constraint:OnDelete:SET NULL"`
	Tags             []ServerTag       `json:"tags,omitempty" gorm:"many2many:cmdb_server_tag_relations;constraint:OnDelete:CASCADE"`
	Groups           []ServerGroup     `json:"groups,omitempty" gorm:"many2many:cmdb_server_group_relations;constraint:OnDelete:CASCADE"`
	CloudInfo        *CloudServer      `json:"cloudInfo,omitempty" gorm:"foreignKey:ServerID;constraint:OnDelete:SET NULL"`
	Credentials      []SSHCredential   `json:"credentials,omitempty" gorm:"many2many:cmdb_server_credentials;joinForeignKey:ServerID;joinReferences:CredentialID"`
	Attributes       []ServerAttribute `json:"attributes,omitempty" gorm:"foreignKey:ServerID;constraint:OnDelete:CASCADE"`
	GroupIDs         []uint            `json:"groupIds,omitempty" gorm:"-"`
	CredentialIDs    []uint            `json:"credentialIds,omitempty" gorm:"-"`
}

// TableName 指定表名
func (Server) TableName() string {
	return "cmdb_servers"
}
