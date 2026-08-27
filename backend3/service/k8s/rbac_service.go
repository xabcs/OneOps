package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// K8sRbacService K8s 原生 RBAC 对象代管服务（A 模式：平台页面对集群内 RBAC 资源的查看与编辑）
type K8sRbacService struct {
	clientPool *K8sClientPool
}

// NewK8sRbacService 创建 RBAC 代管服务
func NewK8sRbacService(clientPool *K8sClientPool) *K8sRbacService {
	return &K8sRbacService{clientPool: clientPool}
}

// client 获取集群客户端（H3：按操作者身份 scoped 化——
// 超管/平台任务落平台凭据；普通用户走 impersonation，由集群原生 RBAC 终判可读/可写的 RBAC 对象）
func (s *K8sRbacService) client(clusterID uint, userID uint) (*kubernetes.Clientset, error) {
	return s.clientPool.GetScopedClientset(clusterID, userID)
}

func ctxTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 30*time.Second)
}

// managedBy 判断对象是否由外部工具管理（如 Helm），提示前端谨慎编辑
func managedBy(labels map[string]string, annotations map[string]string) string {
	if v, ok := labels["app.kubernetes.io/managed-by"]; ok {
		return v
	}
	if _, ok := annotations["meta.helm.sh/release-name"]; ok {
		return "Helm"
	}
	return ""
}

// ========== ClusterRole ==========

// ListClusterRoles 列出集群内全部 ClusterRole（精简视图）
func (s *K8sRbacService) ListClusterRoles(clusterID uint, userID uint, search string) ([]map[string]interface{}, error) {
	clientset, err := s.client(clusterID, userID)
	if err != nil {
		return nil, err
	}
	ctx, cancel := ctxTimeout()
	defer cancel()

	list, err := clientset.RbacV1().ClusterRoles().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 ClusterRole 列表失败: %w", err)
	}

	result := make([]map[string]interface{}, 0, len(list.Items))
	for i := range list.Items {
		r := &list.Items[i]
		if search != "" && !strings.Contains(r.Name, search) {
			continue
		}
		result = append(result, map[string]interface{}{
			"name":      r.Name,
			"labels":    r.Labels,
			"managedBy": managedBy(r.Labels, r.Annotations),
			"rules":     len(r.Rules),
			"createdAt": r.CreationTimestamp.Format("2006-01-02 15:04:05"),
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i]["name"].(string) < result[j]["name"].(string)
	})
	return result, nil
}

// GetClusterRole 获取 ClusterRole 详情（含完整 rules）
func (s *K8sRbacService) GetClusterRole(clusterID uint, userID uint, name string) (map[string]interface{}, error) {
	clientset, err := s.client(clusterID, userID)
	if err != nil {
		return nil, err
	}
	ctx, cancel := ctxTimeout()
	defer cancel()

	r, err := clientset.RbacV1().ClusterRoles().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 ClusterRole 失败: %w", err)
	}
	return clusterRoleView(r), nil
}

// UpdateClusterRole 更新 ClusterRole（manifest 需含 metadata.name，rules 整体替换）
func (s *K8sRbacService) UpdateClusterRole(clusterID uint, userID uint, manifest map[string]interface{}) error {
	clientset, err := s.client(clusterID, userID)
	if err != nil {
		return err
	}

	target := &rbacv1.ClusterRole{}
	if err := decodeManifest(manifest, target); err != nil {
		return err
	}

	ctx, cancel := ctxTimeout()
	defer cancel()
	if _, err := clientset.RbacV1().ClusterRoles().Update(ctx, target, metav1.UpdateOptions{}); err != nil {
		return fmt.Errorf("更新 ClusterRole 失败: %w", err)
	}
	return nil
}

// ========== Role（命名空间级） ==========

// ListRoles 列出命名空间（或全集群）的 Role
func (s *K8sRbacService) ListRoles(clusterID uint, userID uint, namespace string, search string) ([]map[string]interface{}, error) {
	clientset, err := s.client(clusterID, userID)
	if err != nil {
		return nil, err
	}
	ctx, cancel := ctxTimeout()
	defer cancel()

	list, err := clientset.RbacV1().Roles(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 Role 列表失败: %w", err)
	}

	result := make([]map[string]interface{}, 0, len(list.Items))
	for i := range list.Items {
		r := &list.Items[i]
		if search != "" && !strings.Contains(r.Name, search) {
			continue
		}
		result = append(result, map[string]interface{}{
			"name":      r.Name,
			"namespace": r.Namespace,
			"labels":    r.Labels,
			"managedBy": managedBy(r.Labels, r.Annotations),
			"rules":     len(r.Rules),
			"createdAt": r.CreationTimestamp.Format("2006-01-02 15:04:05"),
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i]["name"].(string) < result[j]["name"].(string)
	})
	return result, nil
}

// GetRole 获取 Role 详情
func (s *K8sRbacService) GetRole(clusterID uint, userID uint, namespace string, name string) (map[string]interface{}, error) {
	clientset, err := s.client(clusterID, userID)
	if err != nil {
		return nil, err
	}
	ctx, cancel := ctxTimeout()
	defer cancel()

	r, err := clientset.RbacV1().Roles(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 Role 失败: %w", err)
	}
	return roleView(r), nil
}

// UpdateRole 更新 Role
func (s *K8sRbacService) UpdateRole(clusterID uint, userID uint, namespace string, manifest map[string]interface{}) error {
	clientset, err := s.client(clusterID, userID)
	if err != nil {
		return err
	}

	target := &rbacv1.Role{}
	if err := decodeManifest(manifest, target); err != nil {
		return err
	}
	if target.Namespace == "" {
		target.Namespace = namespace
	}

	ctx, cancel := ctxTimeout()
	defer cancel()
	if _, err := clientset.RbacV1().Roles(namespace).Update(ctx, target, metav1.UpdateOptions{}); err != nil {
		return fmt.Errorf("更新 Role 失败: %w", err)
	}
	return nil
}

// ========== ClusterRoleBinding ==========

// ListClusterRoleBindings 列出集群内全部 ClusterRoleBinding
func (s *K8sRbacService) ListClusterRoleBindings(clusterID uint, userID uint, search string) ([]map[string]interface{}, error) {
	clientset, err := s.client(clusterID, userID)
	if err != nil {
		return nil, err
	}
	ctx, cancel := ctxTimeout()
	defer cancel()

	list, err := clientset.RbacV1().ClusterRoleBindings().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 ClusterRoleBinding 列表失败: %w", err)
	}

	result := make([]map[string]interface{}, 0, len(list.Items))
	for i := range list.Items {
		b := &list.Items[i]
		if search != "" && !matchBinding(b.Name, b.Subjects, search) {
			continue
		}
		result = append(result, clusterRoleBindingSummary(b))
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i]["name"].(string) < result[j]["name"].(string)
	})
	return result, nil
}

// GetClusterRoleBinding 获取 ClusterRoleBinding 详情
func (s *K8sRbacService) GetClusterRoleBinding(clusterID uint, userID uint, name string) (map[string]interface{}, error) {
	clientset, err := s.client(clusterID, userID)
	if err != nil {
		return nil, err
	}
	ctx, cancel := ctxTimeout()
	defer cancel()

	b, err := clientset.RbacV1().ClusterRoleBindings().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 ClusterRoleBinding 失败: %w", err)
	}
	return clusterRoleBindingSummary(b), nil
}

// ========== RoleBinding ==========

// ListRoleBindings 列出命名空间（或全集群）的 RoleBinding
func (s *K8sRbacService) ListRoleBindings(clusterID uint, userID uint, namespace string, search string) ([]map[string]interface{}, error) {
	clientset, err := s.client(clusterID, userID)
	if err != nil {
		return nil, err
	}
	ctx, cancel := ctxTimeout()
	defer cancel()

	list, err := clientset.RbacV1().RoleBindings(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 RoleBinding 列表失败: %w", err)
	}

	result := make([]map[string]interface{}, 0, len(list.Items))
	for i := range list.Items {
		b := &list.Items[i]
		if search != "" && !matchBinding(b.Name, b.Subjects, search) {
			continue
		}
		item := roleBindingSummary(b)
		result = append(result, item)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i]["name"].(string) < result[j]["name"].(string)
	})
	return result, nil
}

// GetRoleBinding 获取 RoleBinding 详情
func (s *K8sRbacService) GetRoleBinding(clusterID uint, userID uint, namespace string, name string) (map[string]interface{}, error) {
	clientset, err := s.client(clusterID, userID)
	if err != nil {
		return nil, err
	}
	ctx, cancel := ctxTimeout()
	defer cancel()

	b, err := clientset.RbacV1().RoleBindings(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 RoleBinding 失败: %w", err)
	}
	return roleBindingSummary(b), nil
}

// ========== 视图辅助 ==========

func clusterRoleView(r *rbacv1.ClusterRole) map[string]interface{} {
	return map[string]interface{}{
		"name":      r.Name,
		"labels":    r.Labels,
		"managedBy": managedBy(r.Labels, r.Annotations),
		"rules":     ruleViews(r.Rules),
		"createdAt": r.CreationTimestamp.Format("2006-01-02 15:04:05"),
	}
}

func roleView(r *rbacv1.Role) map[string]interface{} {
	v := clusterRoleView(&rbacv1.ClusterRole{
		ObjectMeta: r.ObjectMeta,
		Rules:      r.Rules,
	})
	v["namespace"] = r.Namespace
	return v
}

func ruleViews(rules []rbacv1.PolicyRule) []map[string]interface{} {
	views := make([]map[string]interface{}, 0, len(rules))
	for _, rule := range rules {
		views = append(views, map[string]interface{}{
			"apiGroups": rule.APIGroups,
			"resources": rule.Resources,
			"verbs":     rule.Verbs,
			"nonResourceURLs": func() []string {
				if len(rule.NonResourceURLs) == 0 {
					return []string{}
				}
				return rule.NonResourceURLs
			}(),
		})
	}
	return views
}

func subjectViews(subjects []rbacv1.Subject) []map[string]interface{} {
	views := make([]map[string]interface{}, 0, len(subjects))
	for _, sub := range subjects {
		views = append(views, map[string]interface{}{
			"kind":      sub.Kind,
			"name":      sub.Name,
			"namespace": sub.Namespace,
		})
	}
	return views
}

func clusterRoleBindingSummary(b *rbacv1.ClusterRoleBinding) map[string]interface{} {
	return map[string]interface{}{
		"name":      b.Name,
		"labels":    b.Labels,
		"managedBy": managedBy(b.Labels, b.Annotations),
		"subjects":  subjectViews(b.Subjects),
		"roleKind":  b.RoleRef.Kind,
		"roleName":  b.RoleRef.Name,
		"createdAt": b.CreationTimestamp.Format("2006-01-02 15:04:05"),
	}
}

func roleBindingSummary(b *rbacv1.RoleBinding) map[string]interface{} {
	v := clusterRoleBindingSummary(&rbacv1.ClusterRoleBinding{
		ObjectMeta: b.ObjectMeta,
		Subjects:   b.Subjects,
		RoleRef:    b.RoleRef,
	})
	v["namespace"] = b.Namespace
	return v
}

func matchBinding(name string, subjects []rbacv1.Subject, search string) bool {
	if strings.Contains(name, search) {
		return true
	}
	for _, sub := range subjects {
		if strings.Contains(sub.Name, search) {
			return true
		}
	}
	return false
}

// decodeManifest 将 manifest map 解析为指定 RBAC 对象
func decodeManifest(manifest map[string]interface{}, target interface{}) error {
	data, err := json.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("序列化 manifest 失败: %w", err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("解析 manifest 失败: %w", err)
	}
	return nil
}
