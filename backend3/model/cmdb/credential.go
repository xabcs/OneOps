package modelcmdb

import "time"

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
	ID             uint             `json:"id" gorm:"primaryKey"`
	Name           string           `json:"name" gorm:"size:100;not null"`
	Description    string           `json:"description" gorm:"type:text"`
	Username       string           `json:"username" gorm:"size:50;not null"`
	AuthType       SSHPublicKeyAuth `json:"authType" gorm:"size:20;default:'password'"`
	Password       string           `json:"password,omitempty" gorm:"size:255"`    // 加密存储
	PrivateKey     string           `json:"privateKey,omitempty" gorm:"type:text"` // 加密存储
	Passphrase     string           `json:"passphrase,omitempty" gorm:"size:255"`  // 加密存储
	Port           int              `json:"port" gorm:"default:22"`
	CredentialType CredentialType   `json:"credentialType" gorm:"type:varchar(10);not null;default:'user'"` // user | system
	SortOrder      int              `json:"sortOrder" gorm:"default:0"`
	Status         int              `json:"status" gorm:"default:1"`
	CreatedAt      time.Time        `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt      time.Time        `json:"updatedAt" gorm:"autoUpdateTime"`

	// 关联
	Servers []Server `json:"servers,omitempty" gorm:"many2many:cmdb_server_credentials;joinForeignKey:CredentialID;joinReferences:ServerID"`
}

// TableName 指定表名
func (SSHCredential) TableName() string {
	return "cmdb_ssh_credentials"
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
	return "cmdb_cloud_servers"
}

// ServerCredential 服务器-凭证多对多关联表
type ServerCredential struct {
	ServerID     uint `json:"serverId" gorm:"primaryKey"`
	CredentialID uint `json:"credentialId" gorm:"primaryKey"`
}

// TableName 指定表名
func (ServerCredential) TableName() string {
	return "cmdb_server_credentials"
}
