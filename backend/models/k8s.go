package models

import (
	"time"
)

// K8sCluster Kubernetes集群配置
type K8sCluster struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:100;not null;uniqueIndex" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	Endpoint    string    `gorm:"size:512;not null" json:"endpoint"`
	Kubeconfig  string    `gorm:"type:longtext;not null" json:"-"` // 加密存储，不在JSON中返回
	ClusterType string    `gorm:"type:enum('standard','managed','edge');default:'standard'" json:"clusterType"`
	Region      string    `gorm:"size:100" json:"region"`
	Version     string    `gorm:"size:50" json:"version"`
	NodeCount   int       `gorm:"default:0" json:"nodeCount"`
	Status      int       `gorm:"default:1" json:"status"` // 1=正常, 0=禁用
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`

	// 关联数据（不存储在数据库）
	RoleBindings []ClusterRoleBinding `gorm:"-" json:"roleBindings,omitempty"`
}

// TableName 指定表名
func (K8sCluster) TableName() string {
	return "k8s_clusters"
}

// ClusterRoleBinding 集群角色绑定
type ClusterRoleBinding struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null" json:"userId"`
	User      *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	ClusterID uint      `gorm:"not null" json:"clusterId"`
	Cluster   *K8sCluster `gorm:"foreignKey:ClusterID" json:"cluster,omitempty"`
	RoleID    uint      `gorm:"not null" json:"roleId"`
	Role      *Role     `gorm:"foreignKey:RoleID" json:"role,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

// TableName 指定表名
func (ClusterRoleBinding) TableName() string {
	return "k8s_cluster_role_bindings"
}

// K8sSession K8s终端会话
type K8sSession struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	UserID        uint           `gorm:"not null" json:"userId"`
	User          *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	ClusterID     uint           `gorm:"not null" json:"clusterId"`
	Cluster       *K8sCluster    `gorm:"foreignKey:ClusterID" json:"cluster,omitempty"`
	PodName       string         `gorm:"size:255;not null" json:"podName"`
	ContainerName string         `gorm:"size:255" json:"containerName"` // NULL表示默认容器
	Namespace     string         `gorm:"size:100;not null" json:"namespace"`
	ClientIP      string         `gorm:"size:50" json:"clientIp"`
	Status        string         `gorm:"type:enum('active','closed','error');default:'active'" json:"status"`
	StartedAt     time.Time      `json:"startedAt"`
	EndedAt       *time.Time     `json:"endedAt"`
	Duration      int            `gorm:"default:0" json:"duration"` // 会话时长（秒）
	CloseReason   string         `gorm:"size:200" json:"closeReason"`
	CreatedAt     time.Time      `json:"createdAt"`

	// 关联数据（不存储在数据库）
	Commands []K8sCommand `gorm:"-" json:"commands,omitempty"`
}

// TableName 指定表名
func (K8sSession) TableName() string {
	return "k8s_sessions"
}

// K8sCommand K8s命令审计
type K8sCommand struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	SessionID     uint       `gorm:"not null" json:"sessionId"`
	Session       *K8sSession `gorm:"foreignKey:SessionID" json:"session,omitempty"`
	Command       string     `gorm:"type:text;not null" json:"command"`
	ExecutedAt    time.Time  `json:"executedAt"`
	ExitCode      *int       `json:"exitCode"`
	RiskLevel     string     `gorm:"type:enum('safe','low','medium','high','critical');default:'safe'" json:"riskLevel"`
	OutputSummary string     `gorm:"type:text" json:"outputSummary"`
	CreatedAt     time.Time  `json:"createdAt"`
}

// TableName 指定表名
func (K8sCommand) TableName() string {
	return "k8s_commands"
}

// K8sClusterFilter 集群筛选条件
type K8sClusterFilter struct {
	Name       *string `json:"name"`
	Status     *int    `json:"status"`
	ClusterType *string `json:"clusterType"`
	UserID     *uint   `json:"userId"` // 用于筛选用户有权限的集群
}

// K8sSessionFilter 会话筛选条件
type K8sSessionFilter struct {
	ClusterID *uint   `json:"clusterId"`
	UserID    *uint   `json:"userId"`
	Status    *string `json:"status"`
	StartDate *string `json:"startDate"`
	EndDate   *string `json:"endDate"`
	Namespace *string `json:"namespace"`
}

// K8sCommandFilter 命令筛选条件
type K8sCommandFilter struct {
	SessionID   *uint   `json:"sessionId"`
	RiskLevel   *string `json:"riskLevel"`
	StartDate   *string `json:"startDate"`
	EndDate     *string `json:"endDate"`
	CommandLike *string `json:"commandLike"`
}
