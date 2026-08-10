package modelk8s

import (
	"time"

	modelsystem "oneops/backend3/model/system"
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
	ID        uint              `gorm:"primaryKey" json:"id"`
	UserID    uint              `gorm:"not null" json:"userId"`
	User      *modelsystem.User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	ClusterID uint              `gorm:"not null" json:"clusterId"`
	Cluster   *K8sCluster       `gorm:"foreignKey:ClusterID" json:"cluster,omitempty"`
	RoleID    uint              `gorm:"not null" json:"roleId"`
	Role      *modelsystem.Role `gorm:"foreignKey:RoleID" json:"role,omitempty"`
	CreatedAt time.Time         `json:"createdAt"`
}

// TableName 指定表名
func (ClusterRoleBinding) TableName() string {
	return "k8s_cluster_role_bindings"
}
