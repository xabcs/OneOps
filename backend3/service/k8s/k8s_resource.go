package k8s

import (
	"context"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// K8sResourceService K8s资源管理服务
type K8sResourceService struct {
	clientPool *K8sClientPool
	// userID 当前作用域用户：0=平台身份（共享凭据，④级平台校验负责拦截）；
	// >0 时若该用户在此集群存在原生 RBAC 绑定，则资源操作自动切换为 impersonation 执行
	userID uint
}

// NewK8sResourceService 创建K8s资源管理服务
func NewK8sResourceService(clientPool *K8sClientPool) *K8sResourceService {
	return &K8sResourceService{
		clientPool: clientPool,
	}
}

// ForUser 返回绑定到指定用户作用域的服务副本（浅拷贝，方法签名不变）
// controller 层每个请求调用一次，副本仅在本请求内使用
func (s *K8sResourceService) ForUser(userID uint) *K8sResourceService {
	return &K8sResourceService{
		clientPool: s.clientPool,
		userID:     userID,
	}
}

// getScopedConfig 按当前作用域获取集群连接配置（见 K8sClientPool.GetScopedConfig）
func (s *K8sResourceService) getScopedConfig(clusterID uint) (*rest.Config, error) {
	return s.clientPool.GetScopedConfig(clusterID, s.userID)
}

// getScopedClientset 按当前作用域获取集群客户端
func (s *K8sResourceService) getScopedClientset(clusterID uint) (*kubernetes.Clientset, error) {
	return s.clientPool.GetScopedClientset(clusterID, s.userID)
}

// createContextWithTimeout 创建带超时的context
func (s *K8sResourceService) createContextWithTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 30*time.Second)
}
