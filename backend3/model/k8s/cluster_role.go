package modelk8s

import (
	"time"

	modelsystem "oneops/backend3/model/system"
)

// K8sClusterRole 集群角色（集群内授权档位：cluster-admin / cluster-operator / cluster-viewer）
// 与全局功能角色 sys_roles 分离：sys_roles 决定"岗位能在平台做什么"（经 Casbin），
// 集群角色决定"用户在某个集群内能做什么"（操作集 = sys_permissions 中 k8s.* 权限码的命名集合）
type K8sClusterRole struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Code        string    `json:"code" gorm:"size:50;not null;uniqueIndex"`
	Name        string    `json:"name" gorm:"size:50;not null"`
	Description string    `json:"description" gorm:"size:200"`
	Status      int       `json:"status" gorm:"default:1;comment:1=启用,0=停用"`
	CreatedAt   time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

func (K8sClusterRole) TableName() string { return "k8s_cluster_roles" }

// K8sClusterRolePermission 集群角色操作集：引用 sys_permissions 权限码（仅 k8s.*）
// seed 预置三档，页面可维护；权限码目录本身仍由权限管理页统一管理
type K8sClusterRolePermission struct {
	ID             uint   `json:"id" gorm:"primaryKey"`
	ClusterRoleID  uint   `json:"clusterRoleId" gorm:"not null;uniqueIndex:uk_cluster_role_perm"`
	PermissionCode string `json:"permissionCode" gorm:"type:varchar(100);not null;uniqueIndex:uk_cluster_role_perm"`
}

func (K8sClusterRolePermission) TableName() string { return "k8s_cluster_role_permissions" }

// ClusterGroupBinding 用户组与集群的绑定（批量授权：组内所有成员获得该集群的集群角色）
type ClusterGroupBinding struct {
	ID            uint                   `json:"id" gorm:"primaryKey"`
	GroupID       uint                   `json:"groupId" gorm:"not null;uniqueIndex:uk_cluster_group_binding"`
	Group         *modelsystem.UserGroup `json:"group,omitempty" gorm:"foreignKey:GroupID"`
	ClusterID     uint                   `json:"clusterId" gorm:"not null;uniqueIndex:uk_cluster_group_binding"`
	ClusterRoleID uint                   `json:"clusterRoleId" gorm:"not null"`
	ClusterRole   *K8sClusterRole        `json:"clusterRole,omitempty" gorm:"foreignKey:ClusterRoleID"`
	CreatedAt     time.Time              `json:"createdAt" gorm:"autoCreateTime"`
}

func (ClusterGroupBinding) TableName() string { return "k8s_cluster_group_bindings" }

// K8sClusterRole / ClusterGroupBinding 等表保留历史数据：
// 三档（cluster-viewer/operator/admin）授权入口已下线，集群授权统一走原生 RBAC 绑定
// （k8s_native_role_bindings），可见性 UNION 仍包含历史直绑/组绑段。
