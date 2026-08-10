package k8s

import (
	"context"
	"time"
)

// K8sResourceService K8s资源管理服务
type K8sResourceService struct {
	clientPool *K8sClientPool
}

// NewK8sResourceService 创建K8s资源管理服务
func NewK8sResourceService(clientPool *K8sClientPool) *K8sResourceService {
	return &K8sResourceService{
		clientPool: clientPool,
	}
}

// createContextWithTimeout 创建带超时的context
func (s *K8sResourceService) createContextWithTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 30*time.Second)
}
