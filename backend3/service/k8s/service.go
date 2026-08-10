package k8s

import (
	"encoding/json"
	"fmt"

	"oneops/backend3/pkg/logger"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// ListServices 获取 Service 列表（支持分页）
func (s *K8sResourceService) ListServices(clusterID uint, namespace string, page, pageSize int) ([]map[string]interface{}, int64, error) {
	clientset, _, err := s.clientPool.GetClient(clusterID)
	if err != nil {
		return nil, 0, err
	}

	ctx, cancel := s.createContextWithTimeout()
	defer cancel()
	list, err := clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, 0, fmt.Errorf("获取 Service 列表失败: %w", err)
	}

	total := int64(len(list.Items))

	start := (page - 1) * pageSize
	end := start + pageSize

	if start < 0 {
		start = 0
	}
	if start > len(list.Items) {
		start = len(list.Items)
	}
	if end > len(list.Items) {
		end = len(list.Items)
	}

	items := list.Items[start:end]
	result := make([]map[string]interface{}, len(items))
	for i, item := range items {
		result[i] = s.formatService(&item)
	}

	return result, total, nil
}

// GetService 获取 Service 详情
func (s *K8sResourceService) GetService(clusterID uint, namespace, name string) (map[string]interface{}, error) {
	clientset, _, err := s.clientPool.GetClient(clusterID)
	if err != nil {
		return nil, err
	}

	ctx, cancel := s.createContextWithTimeout()
	defer cancel()
	svc, err := clientset.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 Service 详情失败: %w", err)
	}

	endpoints, err := clientset.CoreV1().Endpoints(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		endpoints = &corev1.Endpoints{}
	}

	return s.formatServiceDetail(svc, endpoints), nil
}

// CreateService 创建 Service
func (s *K8sResourceService) CreateService(clusterID uint, namespace string, manifest map[string]interface{}) error {
	_, config, err := s.clientPool.GetClient(clusterID)
	if err != nil {
		return err
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("创建 dynamic client 失败: %w", err)
	}

	gvr := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "services",
	}

	uns := &unstructured.Unstructured{}
	uns.SetUnstructuredContent(manifest)
	if ns := uns.GetNamespace(); ns == "" {
		uns.SetNamespace(namespace)
	}

	ctx, cancel := s.createContextWithTimeout()
	defer cancel()
	_, err = dynamicClient.Resource(gvr).Namespace(namespace).Create(ctx, uns, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("创建 Service 失败: %w", err)
	}

	logger.Info("创建 Service 成功",
		zap.Uint("cluster_id", clusterID),
		zap.String("namespace", namespace),
		zap.String("name", uns.GetName()))

	return nil
}

// UpdateService 更新 Service
func (s *K8sResourceService) UpdateService(clusterID uint, namespace string, manifest map[string]interface{}) error {
	_, config, err := s.clientPool.GetClient(clusterID)
	if err != nil {
		return err
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("创建 dynamic client 失败: %w", err)
	}

	gvr := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "services",
	}

	uns := &unstructured.Unstructured{}
	uns.SetUnstructuredContent(manifest)

	ctx, cancel := s.createContextWithTimeout()
	defer cancel()
	_, err = dynamicClient.Resource(gvr).Namespace(namespace).Update(ctx, uns, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("更新 Service 失败: %w", err)
	}

	logger.Info("更新 Service 成功",
		zap.Uint("cluster_id", clusterID),
		zap.String("namespace", namespace),
		zap.String("name", uns.GetName()))

	return nil
}

// DeleteService 删除 Service
func (s *K8sResourceService) DeleteService(clusterID uint, namespace, name string) error {
	clientset, _, err := s.clientPool.GetClient(clusterID)
	if err != nil {
		return err
	}

	ctx, cancel := s.createContextWithTimeout()
	defer cancel()
	err = clientset.CoreV1().Services(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("删除 Service 失败: %w", err)
	}

	logger.Warn("删除 Service 成功",
		zap.Uint("cluster_id", clusterID),
		zap.String("namespace", namespace),
		zap.String("name", name))

	return nil
}

func (s *K8sResourceService) formatService(svc *corev1.Service) map[string]interface{} {
	return map[string]interface{}{
		"name":       svc.Name,
		"namespace":  svc.Namespace,
		"type":       string(svc.Spec.Type),
		"clusterIP":  svc.Spec.ClusterIP,
		"externalIP": svc.Spec.ExternalIPs,
		"ports":      s.formatServicePorts(svc),
		"age":        svc.CreationTimestamp.Format("2006-01-02 15:04:05"),
		"selector":   svc.Spec.Selector,
	}
}

func (s *K8sResourceService) formatServiceDetail(svc *corev1.Service, endpoints *corev1.Endpoints) map[string]interface{} {
	manifest, _ := json.Marshal(svc)

	detail := s.formatService(svc)
	detail["manifest"] = string(manifest)
	detail["endpoints"] = s.formatEndpoints(endpoints)
	return detail
}

func (s *K8sResourceService) formatServicePorts(svc *corev1.Service) []map[string]interface{} {
	ports := make([]map[string]interface{}, len(svc.Spec.Ports))
	for i, p := range svc.Spec.Ports {
		ports[i] = map[string]interface{}{
			"name":       p.Name,
			"protocol":   string(p.Protocol),
			"port":       p.Port,
			"targetPort": p.TargetPort.String(),
			"nodePort":   p.NodePort,
		}
	}
	return ports
}

func (s *K8sResourceService) formatEndpoints(endpoints *corev1.Endpoints) []map[string]interface{} {
	if endpoints == nil || len(endpoints.Subsets) == 0 {
		return []map[string]interface{}{}
	}

	result := make([]map[string]interface{}, 0)

	for _, subset := range endpoints.Subsets {
		for _, addr := range subset.Addresses {
			result = append(result, map[string]interface{}{
				"ip":        addr.IP,
				"hostname":  addr.Hostname,
				"targetRef": addr.TargetRef,
				"ports":     formatSubsetPorts(subset.Ports),
			})
		}
	}

	return result
}
