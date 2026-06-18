package models

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
	return "business_units"
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
	return "server_rooms"
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
	return "cabinets"
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
	ID                uint       `json:"id" gorm:"primaryKey"`
	Hostname          string     `json:"hostname" gorm:"size:100;not null;uniqueIndex"` // 主机名
	IP                string     `json:"ip" gorm:"size:50;not null;index"`              // 外网IP
	InnerIP           string     `json:"innerIp" gorm:"size:50;index"`                  // 内网IP
	CPU               int        `json:"cpu" gorm:"default:0"`                          // CPU核心数
	Memory            int        `json:"memory" gorm:"default:0"`                       // 内存(GB)
	Disk              int        `json:"disk" gorm:"default:0"`                         // 磁盘(GB)
	OS                string     `json:"os" gorm:"size:50"`                             // 操作系统
	OSVersion         string     `json:"osVersion" gorm:"size:50"`                      // 系统版本
	Arch              string     `json:"arch" gorm:"size:20;default:'x86_64'"`          // 系统架构
	Env               string     `json:"env" gorm:"type:varchar(20);default:'test';index"`       // 环境
	Status            string     `json:"status" gorm:"type:varchar(20);default:'unknown';index"` // 状态
	SSHPort           int        `json:"sshPort" gorm:"default:22"`                     // SSH端口
	CredentialID      uint       `json:"credentialId" gorm:"index"`                     // SSH凭证ID（兼容旧字段）
	SSHCredentialID   uint       `json:"sshCredentialId" gorm:"index"`                  // SSH凭证ID
	CabinetID         uint       `json:"cabinetId" gorm:"index"`                        // 所在机柜ID
	UPosition         int        `json:"uPosition"`                                     // 机柜位置(U)
	SN                string     `json:"sn" gorm:"size:100"`                            // 序列号
	Manufacturer     string     `json:"manufacturer" gorm:"size:100"`                  // 厂商
	Model             string     `json:"model" gorm:"size:100"`                         // 型号
	PurchaseDate      *time.Time `json:"purchaseDate"`                                  // 购买日期
	ExpireWarranty    *time.Time `json:"expireWarranty"`                                // 保修到期
	AssetNumber       string     `json:"assetNumber" gorm:"size:100"`                   // 资产编号
	InstanceID        string     `json:"instanceId" gorm:"size:100"`                    // 云主机实例ID
	InstanceType      string     `json:"instanceType" gorm:"size:50"`                   // 云主机类型
	Region            string     `json:"region" gorm:"size:50"`                         // 区域
	Zone              string     `json:"zone" gorm:"size:50"`                           // 可用区
	Provider          string     `json:"provider" gorm:"size:50;index"`                 // 服务商
	ServerType        string     `json:"serverType" gorm:"size:20;default:'vm'"`        // 类型
	Remarks           string     `json:"remarks" gorm:"type:text"`                      // 备注
	LastCheckTime     *time.Time `json:"lastCheckTime"`                                 // 最后连通性检查时间
	LastConnectTime   *time.Time `json:"lastConnectTime"`                               // 最后连接时间
	ConnectivityStatus string     `json:"connectivityStatus" gorm:"type:enum('online','offline','unknown');default:'unknown'"` // 连通性状态
	BusinessID         uint       `json:"businessId" gorm:"index"`
	Business           *BusinessUnit `json:"business,omitempty" gorm:"foreignKey:BusinessID;constraint:OnDelete:SET NULL"`
	CPUUsage           float64    `json:"cpuUsage" gorm:"default:0"`
	MemoryUsage        float64    `json:"memoryUsage" gorm:"default:0"`
	DiskUsage          float64    `json:"diskUsage" gorm:"default:0"`
	Load1              float64    `json:"load1" gorm:"default:0"`
	Load5              float64    `json:"load5" gorm:"default:0"`
	Load15             float64    `json:"load15" gorm:"default:0"`
	DiskPartitions     DiskPartitions `json:"diskPartitions,omitempty" gorm:"type:json"`
	MetricsUpdatedAt   *time.Time `json:"metricsUpdatedAt"`
	AgentStatus        string     `json:"agentStatus" gorm:"type:varchar(20);default:'uninstalled';index"` // uninstalled | running | offline
	AgentPort          int        `json:"agentPort" gorm:"default:9100"`
	AgentVersion       string     `json:"agentVersion" gorm:"size:50"`
	LastHeartbeatAt    *time.Time `json:"lastHeartbeatAt"`
	SystemCredentialID uint       `json:"systemCredentialId" gorm:"index"` // 系统运维凭证（Agent部署/采集专用）
	// 冗余字段：优化列表查询性能（避免关联查询）
	GroupNames        string     `json:"groupNames" gorm:"type:varchar(500);default:'[]'"`         // 分组名称JSON数组
	CredentialNames   string     `json:"credentialNames" gorm:"type:varchar(500);default:'[]'"`   // 凭证名称JSON数组
	CreatedAt          time.Time  `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt          time.Time  `json:"updatedAt" gorm:"autoUpdateTime"`

	// 关联
	SSHCredential    *SSHCredential  `json:"sshCredential,omitempty" gorm:"foreignKey:SSHCredentialID;constraint:OnDelete:SET NULL"`
	SystemCredential *SSHCredential  `json:"systemCredential,omitempty" gorm:"foreignKey:SystemCredentialID;constraint:OnDelete:SET NULL"`
	Cabinet          *Cabinet        `json:"cabinet,omitempty" gorm:"foreignKey:CabinetID;constraint:OnDelete:SET NULL"`
	Tags          []ServerTag     `json:"tags,omitempty" gorm:"many2many:server_tag_relations;constraint:OnDelete:CASCADE"`
	Groups        []ServerGroup   `json:"groups,omitempty" gorm:"many2many:server_group_relations;constraint:OnDelete:CASCADE"`
	CloudInfo     *CloudServer    `json:"cloudInfo,omitempty" gorm:"foreignKey:ServerID;constraint:OnDelete:SET NULL"`
	Credentials   []SSHCredential `json:"credentials,omitempty" gorm:"many2many:server_credentials;joinForeignKey:ServerID;joinReferences:CredentialID"`
	Attributes    []ServerAttribute `json:"attributes,omitempty" gorm:"foreignKey:ServerID;constraint:OnDelete:CASCADE"`
	GroupIDs      []uint          `json:"groupIds,omitempty" gorm:"-"`
	CredentialIDs []uint          `json:"credentialIds,omitempty" gorm:"-"`
}

// TableName 指定表名
func (Server) TableName() string {
	return "servers"
}

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
	Servers []Server `json:"servers,omitempty" gorm:"many2many:server_tag_relations;joinForeignKey:TagID;joinReferences:ServerID"`
}

// TableName 指定表名
func (ServerTag) TableName() string {
	return "server_tags"
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
	return "asset_changes"
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
	return "server_tag_relations"
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
	Parent   *ServerGroup  `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Children []ServerGroup `json:"children,omitempty" gorm:"foreignKey:ParentID"`
	Servers  []Server      `json:"servers,omitempty" gorm:"many2many:server_group_relations;joinForeignKey:GroupID;joinReferences:ServerID"`
}

// TableName 指定表名
func (ServerGroup) TableName() string {
	return "server_groups"
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
	return "server_group_relations"
}

// SSHPublicKeyAuth 公钥认证类型
type SSHPublicKeyAuth string

const (
	SSHAuthPassword SSHPublicKeyAuth = "password" // 密码认证
	SSHAuthKey      SSHPublicKeyAuth = "key"      // 密钥认证
)

// CredentialType 凭证用途类型
type CredentialType string

const (
	CredentialTypeUser   CredentialType = "user"   // 用于用户堡垒连接
	CredentialTypeSystem CredentialType = "system" // 用于系统自动化运维（Agent 部署/采集）
)

// SSHCredential SSH认证凭证
type SSHCredential struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	Name           string         `json:"name" gorm:"size:100;not null"`
	Description    string         `json:"description" gorm:"type:text"`
	Username       string         `json:"username" gorm:"size:50;not null"`
	AuthType       SSHPublicKeyAuth `json:"authType" gorm:"size:20;default:'password'"`
	Password       string         `json:"password,omitempty" gorm:"size:255"`    // 加密存储
	PrivateKey     string         `json:"privateKey,omitempty" gorm:"type:text"` // 加密存储
	Passphrase     string         `json:"passphrase,omitempty" gorm:"size:255"`  // 加密存储
	Port           int            `json:"port" gorm:"default:22"`
	CredentialType CredentialType `json:"credentialType" gorm:"type:varchar(10);not null;default:'user'"` // user | system
	SortOrder      int            `json:"sortOrder" gorm:"default:0"`
	Status         int            `json:"status" gorm:"default:1"`
	CreatedAt      time.Time      `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt      time.Time      `json:"updatedAt" gorm:"autoUpdateTime"`

	// 关联
	Servers []Server `json:"servers,omitempty" gorm:"many2many:server_credentials;joinForeignKey:CredentialID;joinReferences:ServerID"`
}

// TableName 指定表名
func (SSHCredential) TableName() string {
	return "ssh_credentials"
}

// CloudServer 云主机信息
type CloudServer struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	ServerID       uint      `json:"serverId" gorm:"not null;uniqueIndex"`
	Provider       string    `json:"provider" gorm:"size:50;not null;index"` // aliyun, tencent, aws, huawei
	InstanceID     string    `json:"instanceId" gorm:"size:100"`             // 云主机实例ID
	InstanceName   string    `json:"instanceName" gorm:"size:100"`           // 实例名称
	InstanceType   string    `json:"instanceType" gorm:"size:50"`            // 实例规格
	Region         string    `json:"region" gorm:"size:50;index"`            // 地域
	Zone           string    `json:"zone" gorm:"size:50"`                    // 可用区
	VpcID          string    `json:"vpcId" gorm:"size:100"`                  // VPC ID
	SubnetID       string    `json:"subnetId" gorm:"size:100"`               // 子网ID
	PublicIP       string    `json:"publicIp" gorm:"size:50"`                // 公网IP
	PrivateIP      string    `json:"privateIp" gorm:"size:50"`               // 内网IP
	SecurityGroups string    `json:"securityGroups" gorm:"type:text"`        // 安全组JSON
	ChargeType     string    `json:"chargeType" gorm:"size:20"`              // postpay/prepay
	CreatedAt      time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updatedAt" gorm:"autoUpdateTime"`

	// 关联
	Server *Server `json:"server,omitempty" gorm:"foreignKey:ServerID"`
}

// TableName 指定表名
func (CloudServer) TableName() string {
	return "cloud_servers"
}

// ServerCredential 服务器-凭证多对多关联表
type ServerCredential struct {
	ServerID     uint `json:"serverId" gorm:"primaryKey"`
	CredentialID uint `json:"credentialId" gorm:"primaryKey"`
}

// TableName 指定表名
func (ServerCredential) TableName() string {
	return "server_credentials"
}

// ========================================
// Agent 版本管理相关模型
// ========================================

// AgentVersion Agent 版本模型
type AgentVersion struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Version   string    `json:"version" gorm:"size:50;not null;uniqueIndex"` // 版本号
	ReleaseNotes string  `json:"releaseNotes" gorm:"type:text"`                // 发布说明
	Changelog  string    `json:"changelog" gorm:"type:text"`                  // 更新日志
	ReleasedAt time.Time `json:"releasedAt"`                                  // 发布时间

	// 二进制文件信息
	AMD64BinaryPath string `json:"amd64BinaryPath" gorm:"size:255"`           // AMD64 二进制路径
	AMD64BinaryHash string `json:"amd64BinaryHash" gorm:"size:64"`           // AMD64 文件哈希
	AMD64BinarySize int64  `json:"amd64BinarySize"`                          // AMD64 文件大小
	ARM64BinaryPath string `json:"arm64BinaryPath" gorm:"size:255"`           // ARM64 二进制路径
	ARM64BinaryHash string `json:"arm64BinaryHash" gorm:"size:64"`           // ARM64 文件哈希
	ARM64BinarySize int64  `json:"arm64BinarySize"`                          // ARM64 文件大小

	// 版本状态
	IsLatest     bool   `json:"isLatest" gorm:"default:0"`                   // 是否最新版本
	IsDeprecated bool   `json:"isDeprecated" gorm:"default:0"`               // 是否弃用

	// 功能支持
	Features             string `json:"features" gorm:"type:json"`             // 功能列表 JSON
	MinCompatibleVersion string `json:"minCompatibleVersion" gorm:"size:50"`  // 最小兼容版本
	MaxCompatibleVersion string `json:"maxCompatibleVersion" gorm:"size:50"`  // 最大兼容版本

	// 统计信息
	DownloadCount int     `json:"downloadCount" gorm:"default:0"`             // 下载次数
	DeployCount   int     `json:"deployCount" gorm:"default:0"`               // 部署次数

	CreatedAt     time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

// TableName 指定表名
func (AgentVersion) TableName() string {
	return "agent_versions"
}

// AgentUpgradeTask Agent 升级任务模型
type AgentUpgradeTask struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	TaskName    string    `json:"taskName" gorm:"size:100"`

	// 升级目标信息
	TargetVersion   string `json:"targetVersion" gorm:"size:50;not null"`   // 目标版本
	TargetServerIDs string `json:"targetServerIds" gorm:"type:json"`         // 目标主机ID列表

	// 任务状态
	Status      string `json:"status" gorm:"type:varchar(20);default:'pending';index"` // pending|running|completed|failed|cancelled
	CurrentStep int    `json:"currentStep" gorm:"default:0"`
	TotalSteps  int    `json:"totalSteps" gorm:"default:0"`

	// 进度统计
	TotalCount   int `json:"totalCount" gorm:"default:0"`
	SuccessCount int `json:"successCount" gorm:"default:0"`
	FailedCount  int `json:"failedCount" gorm:"default:0"`
	SkippedCount int `json:"skippedCount" gorm:"default:0"`

	// 时间记录
	StartedAt   *time.Time `json:"startedAt"`
	CompletedAt *time.Time `json:"completedAt"`

	// 详细日志
	ErrorMessage string `json:"errorMessage" gorm:"type:text"`
	OperationLog string `json:"operationLog" gorm:"type:json"`

	CreatedBy string `json:"createdBy" gorm:"size:50"`
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

// TableName 指定表名
func (AgentUpgradeTask) TableName() string {
	return "agent_upgrade_tasks"
}

// AgentVersionFeature 版本功能支持结构
type AgentVersionFeature struct {
	ExtendedMetrics bool `json:"extendedMetrics"` // 支持扩展指标
	CustomConfigs   bool `json:"customConfigs"`  // 支持自定义配置
}
