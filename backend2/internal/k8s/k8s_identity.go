package k8s

import (
	"context"
	"fmt"
	"oneops/backend2/pkg/logger"

	"go.uber.org/zap"
	authorizationv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// K8sIdentityService K8s用户身份映射服务
type K8sIdentityService struct {
	clientPool *K8sClientPool
}

// NewK8sIdentityService 创建K8s用户身份映射服务
func NewK8sIdentityService(clientPool *K8sClientPool) *K8sIdentityService {
	return &K8sIdentityService{
		clientPool: clientPool,
	}
}

// MapUserIDToK8sUser 将OneOps用户ID映射为K8s用户身份
// 格式: oneops:user:<user_id>
func MapUserIDToK8sUser(userID uint) string {
	return fmt.Sprintf("oneops:user:%d", userID)
}

// MapK8sUserToUserID 将K8s用户身份解析为OneOps用户ID
// 输入格式: oneops:user:<user_id>
func MapK8sUserToUserID(k8sUser string) (uint, error) {
	var userID uint
	if _, err := fmt.Sscanf(k8sUser, "oneops:user:%d", &userID); err != nil {
		return 0, fmt.Errorf("解析K8s用户身份失败: %w", err)
	}
	return userID, nil
}

// CheckPermission 检查用户是否有权限执行指定操作
// 使用K8s SubjectAccessReview API进行权限验证
func (s *K8sIdentityService) CheckPermission(
	clusterID uint,
	userID uint,
	resource string,
	verb string,
	namespace string,
) (bool, error) {
	// 获取K8s客户端
	clientset, _, err := s.clientPool.GetClient(clusterID)
	if err != nil {
		return false, fmt.Errorf("获取K8s客户端失败: %w", err)
	}

	// 映射用户身份
	k8sUser := MapUserIDToK8sUser(userID)

	// 创建SubjectAccessReview请求
	sar := &authorizationv1.SubjectAccessReview{
		Spec: authorizationv1.SubjectAccessReviewSpec{
			ResourceAttributes: &authorizationv1.ResourceAttributes{
				Namespace: namespace,
				Verb:      verb,
				Group:     "",
				Resource:  resource,
			},
			User: k8sUser,
		},
	}

	// 发送SubjectAccessReview请求
	ctx := context.Background()
	response, err := clientset.AuthorizationV1().SubjectAccessReviews().Create(ctx, sar, metav1.CreateOptions{})
	if err != nil {
		return false, fmt.Errorf("SubjectAccessReview请求失败: %w", err)
	}

	// 记录权限检查结果
	logger.Debug("K8s权限检查",
		zap.Uint("cluster_id", clusterID),
		zap.Uint("user_id", userID),
		zap.String("k8s_user", k8sUser),
		zap.String("resource", resource),
		zap.String("verb", verb),
		zap.String("namespace", namespace),
		zap.Bool("allowed", response.Status.Allowed))

	return response.Status.Allowed, nil
}

// CheckNamespaceAccess 检查用户是否有权限访问指定命名空间
func (s *K8sIdentityService) CheckNamespaceAccess(
	clusterID uint,
	userID uint,
	namespace string,
) (bool, error) {
	return s.CheckPermission(clusterID, userID, "namespaces", "get", namespace)
}

// CheckPodAccess 检查用户是否有权限访问Pod
func (s *K8sIdentityService) CheckPodAccess(
	clusterID uint,
	userID uint,
	namespace string,
) (bool, error) {
	return s.CheckPermission(clusterID, userID, "pods", "get", namespace)
}

// CheckPodExecAccess 检查用户是否有权限exec到Pod
func (s *K8sIdentityService) CheckPodExecAccess(
	clusterID uint,
	userID uint,
	namespace string,
) (bool, error) {
	return s.CheckPermission(clusterID, userID, "pods", "exec", namespace)
}

// CheckDeploymentAccess 检查用户是否有权限访问Deployment
func (s *K8sIdentityService) CheckDeploymentAccess(
	clusterID uint,
	userID uint,
	namespace string,
) (bool, error) {
	return s.CheckPermission(clusterID, userID, "deployments", "get", namespace)
}

// CheckServiceAccess 检查用户是否有权限访问Service
func (s *K8sIdentityService) CheckServiceAccess(
	clusterID uint,
	userID uint,
	namespace string,
) (bool, error) {
	return s.CheckPermission(clusterID, userID, "services", "get", namespace)
}

// CheckConfigMapAccess 检查用户是否有权限访问ConfigMap
func (s *K8sIdentityService) CheckConfigMapAccess(
	clusterID uint,
	userID uint,
	namespace string,
) (bool, error) {
	return s.CheckPermission(clusterID, userID, "configmaps", "get", namespace)
}

// CheckSecretAccess 检查用户是否有权限访问Secret
func (s *K8sIdentityService) CheckSecretAccess(
	clusterID uint,
	userID uint,
	namespace string,
) (bool, error) {
	return s.CheckPermission(clusterID, userID, "secrets", "get", namespace)
}

// GetUserClusterPermissions 获取用户在集群中的所有权限（用于显示）
func (s *K8sIdentityService) GetUserClusterPermissions(
	clusterID uint,
	userID uint,
	namespace string,
) (map[string]bool, error) {
	resources := []string{"pods", "deployments", "services", "configmaps", "secrets"}
	verbs := []string{"get", "list", "watch", "create", "update", "delete", "exec"}

	permissions := make(map[string]bool)

	for _, resource := range resources {
		for _, verb := range verbs {
			key := fmt.Sprintf("%s:%s", resource, verb)
			allowed, err := s.CheckPermission(clusterID, userID, resource, verb, namespace)
			if err != nil {
				logger.Warn("权限检查失败",
					zap.String("resource", resource),
					zap.String("verb", verb),
					zap.Error(err))
				permissions[key] = false
			} else {
				permissions[key] = allowed
			}
		}
	}

	return permissions, nil
}

// ValidateClusterAccess 验证用户是否有集群访问权限
// 用于在显示集群列表前的权限检查
func (s *K8sIdentityService) ValidateClusterAccess(clusterID uint, userID uint) (bool, error) {
	// 简单检查：用户能否访问任意命名空间
	return s.CheckPermission(clusterID, userID, "namespaces", "list", "")
}

// CreateK8sUserMapping 创建K8s用户映射配置
// 这个函数生成K8s RoleBinding配置示例，供管理员参考
func CreateK8sUserMapping(userID uint, roleName string, namespace string) map[string]interface{} {
	k8sUser := MapUserIDToK8sUser(userID)

	return map[string]interface{}{
		"apiVersion": "rbac.authorization.k8s.io/v1",
		"kind":       "RoleBinding",
		"metadata": map[string]interface{}{
			"name":      fmt.Sprintf("oneops-user-%d-%s", userID, roleName),
			"namespace": namespace,
		},
		"subjects": []map[string]interface{}{
			{
				"kind": "User",
				"name": k8sUser,
			},
		},
		"roleRef": map[string]interface{}{
			"kind":      "Role",
			"name":      roleName,
			"namespace": namespace,
		},
	}
}

// CreateClusterRoleBindingMapping 创建K8s ClusterRoleBinding配置
func CreateClusterRoleBindingMapping(userID uint, roleClusterName string) map[string]interface{} {
	k8sUser := MapUserIDToK8sUser(userID)

	return map[string]interface{}{
		"apiVersion": "rbac.authorization.k8s.io/v1",
		"kind":       "ClusterRoleBinding",
		"metadata": map[string]interface{}{
			"name": fmt.Sprintf("oneops-user-%d-%s", userID, roleClusterName),
		},
		"subjects": []map[string]interface{}{
			{
				"kind": "User",
				"name": k8sUser,
			},
		},
		"roleRef": map[string]interface{}{
			"kind": "ClusterRole",
			"name": roleClusterName,
		},
	}
}
