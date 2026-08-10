package k8s

import (
	. "oneops/backend3/service/k8s"
)

// K8sResourceController K8s资源管理控制器
type K8sResourceController struct {
	svc        *K8sResourceService
	clusterSvc *K8sClusterService
}

// NewK8sResourceController 创建K8s资源管理控制器
func NewK8sResourceController(svc *K8sResourceService, clusterSvc *K8sClusterService) *K8sResourceController {
	return &K8sResourceController{svc: svc, clusterSvc: clusterSvc}
}
