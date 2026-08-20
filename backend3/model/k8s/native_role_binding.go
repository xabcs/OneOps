package modelk8s

import (
	"time"

	modelsystem "oneops/backend3/model/system"
)

// K8sNativeRoleBinding OneOps 主体与 K8s 原生 RBAC 角色的绑定（B 模式：授权执行层）
// 授权时除落库外，还会在目标集群创建真实的 ClusterRoleBinding / RoleBinding，
// 用户操作通过 impersonation 以真实身份访问 kube-apiserver，由原生 RBAC 最终判定。
type K8sNativeRoleBinding struct {
	ID          uint                   `gorm:"primaryKey" json:"id"`
	SubjectType string                 `gorm:"size:10;not null;uniqueIndex:uk_native_binding" json:"subjectType"` // user=用户直绑 / group=用户组绑定
	UserID      uint                   `gorm:"uniqueIndex:uk_native_binding" json:"userId"`                       // SubjectType=user 时有效
	User        *modelsystem.User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	GroupID     uint                   `gorm:"uniqueIndex:uk_native_binding" json:"groupId"` // SubjectType=group 时有效
	Group       *modelsystem.UserGroup `gorm:"foreignKey:GroupID" json:"group,omitempty"`
	ClusterID   uint                   `gorm:"not null;uniqueIndex:uk_native_binding" json:"clusterId"`
	Cluster     *K8sCluster            `gorm:"foreignKey:ClusterID" json:"cluster,omitempty"`
	RoleKind    string                 `gorm:"size:20;not null;uniqueIndex:uk_native_binding" json:"roleKind"` // ClusterRole / Role
	RoleName    string                 `gorm:"size:253;not null;uniqueIndex:uk_native_binding" json:"roleName"`
	Namespace   string                 `gorm:"size:63;uniqueIndex:uk_native_binding" json:"namespace"` // RoleKind=Role 时必填
	// ImpersonationName 集群内模拟身份名：user → oneops-{username}；group → oneops-group-{groupCode}
	ImpersonationName string `gorm:"size:100;not null" json:"impersonationName"`
	// K8sBindingName 集群内创建的 Binding 资源名：oneops-native-{ID}
	K8sBindingName string    `gorm:"size:253;not null" json:"k8sBindingName"`
	CreatedAt      time.Time `json:"createdAt"`
}

// TableName 指定表名
func (K8sNativeRoleBinding) TableName() string {
	return "k8s_native_role_bindings"
}

// AssignNativeRoleBindingRequest 创建原生角色绑定请求
type AssignNativeRoleBindingRequest struct {
	SubjectType string   `json:"subjectType" binding:"required,oneof=user group"`
	UserID      uint     `json:"userId"`
	GroupID     uint     `json:"groupId"`
	RoleKind    string   `json:"roleKind" binding:"required,oneof=ClusterRole Role"`
	RoleName    string   `json:"roleName" binding:"required"`
	Namespace   string   `json:"namespace"`  // 兼容单命名空间；与 Namespaces 合并去重
	Namespaces  []string `json:"namespaces"` // 多命名空间批量授权：每个 ns 创建一条绑定 + 集群内 RoleBinding
}
