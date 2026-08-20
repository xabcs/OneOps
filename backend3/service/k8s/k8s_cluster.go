package k8s

import (
	"context"
	"fmt"
	"os"
	"time"

	modelk8s "oneops/backend3/model/k8s"
	"oneops/backend3/pkg/logger"
	"oneops/backend3/pkg/utils"
	repok8s "oneops/backend3/repository/k8s"

	"go.uber.org/zap"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// K8sClusterService K8s集群管理服务
type K8sClusterService struct {
	clusterRepo *repok8s.ClusterRepository
	clientPool  *K8sClientPool
	identitySvc *K8sIdentityService
}

// NewK8sClusterService 创建K8s集群管理服务
func NewK8sClusterService(clusterRepo *repok8s.ClusterRepository, clientPool *K8sClientPool, identitySvc *K8sIdentityService) *K8sClusterService {
	return &K8sClusterService{
		clusterRepo: clusterRepo,
		clientPool:  clientPool,
		identitySvc: identitySvc,
	}
}

// ========== 集群 CRUD 操作 ==========

// GetClusters 获取集群列表（数据权限下推 SQL，count 与 list 同条件）
func (s *K8sClusterService) GetClusters(userID uint, page, pageSize int, filter *modelk8s.K8sClusterFilter) ([]modelk8s.K8sCluster, int64, error) {
	// 数据权限：非超管只能看到有角色绑定的集群（超管 AuthorizedIDs=nil 不限制）
	var authorizedIDs *[]uint
	if !s.isSuperAdmin(userID) {
		ids, err := s.clusterRepo.FindAuthorizedClusterIDs(userID)
		if err != nil {
			return nil, 0, err
		}
		authorizedIDs = &ids
	}

	return s.clusterRepo.FindWithPagination(repok8s.ClusterQuery{
		Filter:        filter,
		AuthorizedIDs: authorizedIDs,
		Page:          page,
		PageSize:      pageSize,
	})
}

// GetClusterByID 根据ID获取集群详情（带权限检查）
func (s *K8sClusterService) GetClusterByID(clusterID uint, userID uint) (*modelk8s.K8sCluster, error) {
	// 先检查权限
	hasAccess, err := s.CheckUserClusterAccess(userID, clusterID)
	if err != nil {
		return nil, fmt.Errorf("检查集群访问权限失败: %w", err)
	}
	if !hasAccess {
		return nil, fmt.Errorf("无权访问该集群")
	}

	// 查询集群
	cluster, err := s.clusterRepo.FindByID(clusterID)
	if err != nil {
		return nil, fmt.Errorf("集群不存在: %w", err)
	}

	return cluster, nil
}

// CreateCluster 创建新集群
func (s *K8sClusterService) CreateCluster(cluster *modelk8s.K8sCluster, kubeconfigPlaintext string, operatorID uint) error {
	// 加密 kubeconfig
	encryptedConfig, err := utils.EncryptString(kubeconfigPlaintext)
	if err != nil {
		return fmt.Errorf("加密kubeconfig失败: %w", err)
	}

	cluster.Kubeconfig = encryptedConfig

	// 验证 kubeconfig 有效性
	if err := s.validateKubeconfig(kubeconfigPlaintext); err != nil {
		return fmt.Errorf("kubeconfig验证失败: %w", err)
	}

	// 创建集群记录
	if err := s.clusterRepo.Create(cluster); err != nil {
		return fmt.Errorf("创建集群记录失败: %w", err)
	}

	// 记录操作日志
	logger.Info("创建K8s集群",
		zap.Uint("cluster_id", cluster.ID),
		zap.String("cluster_name", cluster.Name),
		zap.Uint("operator_id", operatorID))

	return nil
}

// UpdateCluster 更新集群信息
func (s *K8sClusterService) UpdateCluster(clusterID uint, updates map[string]interface{}, operatorID uint) error {
	cluster, err := s.clusterRepo.FindByID(clusterID)
	if err != nil {
		return fmt.Errorf("集群不存在: %w", err)
	}

	// 如果更新 kubeconfig，需要重新加密
	if kubeconfigPlaintext, exists := updates["kubeconfig"]; exists {
		if kubeconfigStr, ok := kubeconfigPlaintext.(string); ok && kubeconfigStr != "" {
			encryptedConfig, err := utils.EncryptString(kubeconfigStr)
			if err != nil {
				return fmt.Errorf("加密kubeconfig失败: %w", err)
			}
			updates["kubeconfig"] = encryptedConfig

			// 验证 kubeconfig 有效性
			if err := s.validateKubeconfig(kubeconfigStr); err != nil {
				return fmt.Errorf("kubeconfig验证失败: %w", err)
			}
		}
	}

	// 更新集群
	if err := s.clusterRepo.UpdateByID(cluster, updates); err != nil {
		return fmt.Errorf("更新集群失败: %w", err)
	}

	// 失效客户端缓存
	s.clientPool.InvalidateCluster(clusterID)

	// 记录操作日志
	logger.Info("更新K8s集群",
		zap.Uint("cluster_id", clusterID),
		zap.Uint("operator_id", operatorID))

	return nil
}

// DeleteCluster 删除集群（需要二次确认）
func (s *K8sClusterService) DeleteCluster(clusterID uint, confirmName string, operatorID uint) error {
	cluster, err := s.clusterRepo.FindByID(clusterID)
	if err != nil {
		return fmt.Errorf("集群不存在: %w", err)
	}

	// 二次确认：检查输入的集群名称是否匹配
	if cluster.Name != confirmName {
		return fmt.Errorf("集群名称不匹配，取消删除操作")
	}

	// 级联删除集群及其关联数据（事务）
	if err := s.clusterRepo.DeleteClusterCascade(clusterID, cluster); err != nil {
		return fmt.Errorf("删除集群失败: %w", err)
	}

	// 失效客户端缓存
	s.clientPool.RemoveClient(clusterID)

	// 记录操作日志
	logger.Warn("删除K8s集群",
		zap.Uint("cluster_id", clusterID),
		zap.String("cluster_name", cluster.Name),
		zap.Uint("operator_id", operatorID))

	return nil
}

// GetClient 获取 K8s 客户端（委托给 clientPool）
func (s *K8sClusterService) GetClient(clusterID uint) (*kubernetes.Clientset, error) {
	clientset, _, err := s.clientPool.GetClient(clusterID)
	return clientset, err
}

// GetScopedClient 获取按用户作用域的 K8s 客户端（存在原生绑定时为 impersonated 客户端）
func (s *K8sClusterService) GetScopedClient(clusterID uint, userID uint) (*kubernetes.Clientset, error) {
	return s.clientPool.GetScopedClientset(clusterID, userID)
}

// GetScopedClientWithConfig 获取按用户作用域的客户端和 rest.Config（用于 exec 等 SPDY 操作）
func (s *K8sClusterService) GetScopedClientWithConfig(clusterID uint, userID uint) (*kubernetes.Clientset, *rest.Config, error) {
	config, err := s.clientPool.GetScopedConfig(clusterID, userID)
	if err != nil {
		return nil, nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}
	return clientset, config, nil
}

// GetClientWithConfig 获取 K8s 客户端和 rest.Config（用于 exec 等 SPDY 操作）
func (s *K8sClusterService) GetClientWithConfig(clusterID uint) (*kubernetes.Clientset, *rest.Config, error) {
	return s.clientPool.GetClient(clusterID)
}

// ========== 连接测试 ==========

// TestConnection 测试集群连接
func (s *K8sClusterService) TestConnection(clusterID uint) error {
	return s.clientPool.TestConnection(clusterID)
}

// GetClusterNodes 获取集群节点列表（按用户作用域：原生绑定用户经 impersonation 由 apiserver 过滤）
func (s *K8sClusterService) GetClusterNodes(clusterID uint, userID uint) ([]map[string]interface{}, error) {
	clientset, err := s.GetScopedClient(clusterID, userID)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取节点列表失败: %w", err)
	}

	result := make([]map[string]interface{}, len(nodes.Items))
	for i, node := range nodes.Items {
		result[i] = map[string]interface{}{
			"name":    node.Name,
			"status":  getNodeStatus(&node),
			"roles":   getNodeRoles(&node),
			"version": node.Status.NodeInfo.KubeletVersion,
			"created": node.CreationTimestamp.Format("2006-01-02 15:04:05"),
			"capacity": map[string]interface{}{
				"cpu":    node.Status.Capacity.Cpu().String(),
				"memory": node.Status.Capacity.Memory().String(),
			},
		}
	}

	return result, nil
}

// GetClusterNamespaces 获取集群命名空间列表（按用户作用域：原生绑定用户经 impersonation 由 apiserver 过滤）
func (s *K8sClusterService) GetClusterNamespaces(clusterID uint, userID uint) ([]map[string]interface{}, error) {
	clientset, err := s.GetScopedClient(clusterID, userID)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取命名空间列表失败: %w", err)
	}

	result := make([]map[string]interface{}, len(namespaces.Items))
	for i, ns := range namespaces.Items {
		result[i] = map[string]interface{}{
			"name":    ns.Name,
			"status":  getNamespaceStatus(&ns),
			"created": ns.CreationTimestamp.Format("2006-01-02 15:04:05"),
			"labels":  ns.Labels,
		}
	}

	return result, nil
}

// ========== 权限管理 ==========
// 注：平台三档（cluster-viewer/operator/admin）已下线，集群授权统一走原生 RBAC 绑定
// （集群详情-原生授权）；可见性与操作预检见 CheckUserClusterAccess / CheckClusterOperation。

// ========== 辅助方法 ==========

// isSuperAdmin 检查用户是否为超级管理员（委托 repo，与 client pool 白名单共用同一判定）
func (s *K8sClusterService) isSuperAdmin(userID uint) bool {
	return s.clusterRepo.IsSuperAdmin(userID)
}

// CheckUserClusterAccess 数据权限校验（③层）：用户对集群是否可见
// 可见 = 存在任一授权载体：平台直绑/组绑定（历史）∪ 原生 RBAC 绑定（现行）；
// 实际由 FindAuthorizedClusterIDs 的 UNION 实现；超管全可见
func (s *K8sClusterService) CheckUserClusterAccess(userID uint, clusterID uint) (bool, error) {
	// 超级管理员拥有所有权限
	if s.isSuperAdmin(userID) {
		return true, nil
	}

	ids, err := s.clusterRepo.FindAuthorizedClusterIDs(userID)
	if err != nil {
		return false, err
	}
	for _, id := range ids {
		if id == clusterID {
			return true, nil
		}
	}
	return false, nil
}

// CheckClusterOperation 集群内操作级校验（④层）：三档下线后的简化预检
// ②中间件已用系统权限码判定"岗位能否调这类 API"；
// 此处只校验"用户在该集群是否存在原生 RBAC 绑定"（fail-closed 的执行前置），
// 资源粒度终判由执行层 impersonation 交给 kube-apiserver；超管全放行。
// permissionCode 参数保留以兼容既有调用方签名，不再参与比对
func (s *K8sClusterService) CheckClusterOperation(userID uint, clusterID uint, permissionCode string) (bool, error) {
	if s.isSuperAdmin(userID) {
		return true, nil
	}
	_, _, hasNative, err := s.clusterRepo.FindNativeImpersonation(userID, clusterID)
	if err != nil {
		return false, err
	}
	return hasNative, nil
}

// validateKubeconfig 验证 kubeconfig 有效性
func (s *K8sClusterService) validateKubeconfig(kubeconfigContent string) error {
	// 创建临时文件存储 kubeconfig
	tmpFile, err := os.CreateTemp("", "kubeconfig-validate-*")
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(kubeconfigContent); err != nil {
		tmpFile.Close()
		return fmt.Errorf("写入kubeconfig失败: %w", err)
	}
	tmpFile.Close()

	// 尝试加载 kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", tmpFile.Name())
	if err != nil {
		return fmt.Errorf("加载kubeconfig失败: %w", err)
	}

	// 创建客户端测试连接
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("创建K8s客户端失败: %w", err)
	}

	// 简单连接测试
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{Limit: 1}); err != nil {
		return fmt.Errorf("K8s连接测试失败: %w", err)
	}

	return nil
}

// getNodeStatus 获取节点状态
func getNodeStatus(node *v1.Node) string {
	for _, condition := range node.Status.Conditions {
		if condition.Type == "Ready" {
			if condition.Status == "True" {
				return "Ready"
			} else {
				return "NotReady"
			}
		}
	}
	return "Unknown"
}

// getNodeRoles 获取节点角色
func getNodeRoles(node *v1.Node) []string {
	var roles []string
	for label := range node.Labels {
		if label == "node-role.kubernetes.io/master" ||
			label == "node-role.kubernetes.io/control-plane" {
			roles = append(roles, "master")
		}
		if label == "node-role.kubernetes.io/worker" {
			roles = append(roles, "worker")
		}
	}
	if len(roles) == 0 {
		roles = append(roles, "worker")
	}
	return roles
}

// getNamespaceStatus 获取命名空间状态
func getNamespaceStatus(ns *v1.Namespace) string {
	return "Active"
}
