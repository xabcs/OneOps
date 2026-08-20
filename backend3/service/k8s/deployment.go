package k8s

import (
	"encoding/json"
	"fmt"
	"strings"

	"oneops/backend3/pkg/logger"

	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// ListDeployments 获取 Deployment 列表（支持分页）
func (s *K8sResourceService) ListDeployments(clusterID uint, namespace string, page, pageSize int) ([]map[string]interface{}, int64, error) {
	clientset, err := s.getScopedClientset(clusterID)
	if err != nil {
		return nil, 0, err
	}

	ctx, cancel := s.createContextWithTimeout()
	defer cancel()

	list, err := clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, 0, fmt.Errorf("获取 Deployment 列表失败: %w", err)
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
		result[i] = s.formatDeployment(&item)
	}

	return result, total, nil
}

// GetDeployment 获取 Deployment 详情
func (s *K8sResourceService) GetDeployment(clusterID uint, namespace, name string) (map[string]interface{}, error) {
	clientset, err := s.getScopedClientset(clusterID)
	if err != nil {
		return nil, err
	}

	ctx, cancel := s.createContextWithTimeout()
	defer cancel()
	deployment, err := clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 Deployment 详情失败: %w", err)
	}

	return s.formatDeploymentDetail(deployment), nil
}

// GetDeploymentPods 获取 Deployment 管理的 Pods
func (s *K8sResourceService) GetDeploymentPods(clusterID uint, namespace, deploymentName string) ([]map[string]interface{}, error) {
	clientset, err := s.getScopedClientset(clusterID)
	if err != nil {
		return nil, err
	}

	ctx, cancel := s.createContextWithTimeout()
	defer cancel()

	deployment, err := clientset.AppsV1().Deployments(namespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 Deployment 失败: %w", err)
	}

	var labelSelector string
	if len(deployment.Spec.Selector.MatchLabels) > 0 {
		selectors := []string{}
		for k, v := range deployment.Spec.Selector.MatchLabels {
			selectors = append(selectors, fmt.Sprintf("%s=%s", k, v))
		}
		labelSelector = strings.Join(selectors, ",")
	}

	list, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("获取 Pod 列表失败: %w", err)
	}

	result := make([]map[string]interface{}, len(list.Items))
	for i, item := range list.Items {
		podData := s.formatPod(&item)
		podData["ownerDeployment"] = deploymentName
		result[i] = podData
	}

	return result, nil
}

// CreateDeployment 创建 Deployment
func (s *K8sResourceService) CreateDeployment(clusterID uint, namespace string, manifest map[string]interface{}) error {
	config, err := s.getScopedConfig(clusterID)
	if err != nil {
		return err
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("创建 dynamic client 失败: %w", err)
	}

	gvr := schema.GroupVersionResource{
		Group:    "apps",
		Version:  "v1",
		Resource: "deployments",
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
		return fmt.Errorf("创建 Deployment 失败: %w", err)
	}

	logger.Info("创建 Deployment 成功",
		zap.Uint("cluster_id", clusterID),
		zap.String("namespace", namespace),
		zap.String("name", uns.GetName()))

	return nil
}

// UpdateDeployment 更新 Deployment
func (s *K8sResourceService) UpdateDeployment(clusterID uint, namespace string, manifest map[string]interface{}) error {
	config, err := s.getScopedConfig(clusterID)
	if err != nil {
		return err
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("创建 dynamic client 失败: %w", err)
	}

	gvr := schema.GroupVersionResource{
		Group:    "apps",
		Version:  "v1",
		Resource: "deployments",
	}

	uns := &unstructured.Unstructured{}
	uns.SetUnstructuredContent(manifest)

	ctx, cancel := s.createContextWithTimeout()
	defer cancel()
	_, err = dynamicClient.Resource(gvr).Namespace(namespace).Update(ctx, uns, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("更新 Deployment 失败: %w", err)
	}

	logger.Info("更新 Deployment 成功",
		zap.Uint("cluster_id", clusterID),
		zap.String("namespace", namespace),
		zap.String("name", uns.GetName()))

	return nil
}

// DeleteDeployment 删除 Deployment
func (s *K8sResourceService) DeleteDeployment(clusterID uint, namespace, name string) error {
	clientset, err := s.getScopedClientset(clusterID)
	if err != nil {
		return err
	}

	ctx, cancel := s.createContextWithTimeout()
	defer cancel()
	err = clientset.AppsV1().Deployments(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("删除 Deployment 失败: %w", err)
	}

	logger.Warn("删除 Deployment 成功",
		zap.Uint("cluster_id", clusterID),
		zap.String("namespace", namespace),
		zap.String("name", name))

	return nil
}

// ScaleDeployment 扩缩容 Deployment
func (s *K8sResourceService) ScaleDeployment(clusterID uint, namespace, name string, replicas int32) error {
	clientset, err := s.getScopedClientset(clusterID)
	if err != nil {
		return err
	}

	ctx, cancel := s.createContextWithTimeout()
	defer cancel()
	deployment, err := clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("获取 Deployment 失败: %w", err)
	}

	deployment.Spec.Replicas = &replicas
	_, err = clientset.AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("扩缩容 Deployment 失败: %w", err)
	}

	logger.Info("扩缩容 Deployment 成功",
		zap.Uint("cluster_id", clusterID),
		zap.String("namespace", namespace),
		zap.String("name", name),
		zap.Int32("replicas", replicas))

	return nil
}

// RestartDeployment 重启 Deployment
func (s *K8sResourceService) RestartDeployment(clusterID uint, namespace, name string) error {
	clientset, err := s.getScopedClientset(clusterID)
	if err != nil {
		return err
	}

	ctx, cancel := s.createContextWithTimeout()
	defer cancel()
	deployment, err := clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("获取 Deployment 失败: %w", err)
	}

	if deployment.Spec.Template.Annotations == nil {
		deployment.Spec.Template.Annotations = make(map[string]string)
	}
	deployment.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = getCurrentTimestamp()

	_, err = clientset.AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("重启 Deployment 失败: %w", err)
	}

	logger.Info("重启 Deployment 成功",
		zap.Uint("cluster_id", clusterID),
		zap.String("namespace", namespace),
		zap.String("name", name))

	return nil
}

func (s *K8sResourceService) formatDeployment(deployment *appsv1.Deployment) map[string]interface{} {
	return map[string]interface{}{
		"name":       deployment.Name,
		"namespace":  deployment.Namespace,
		"replicas":   deployment.Status.Replicas,
		"ready":      deployment.Status.ReadyReplicas,
		"upToDate":   deployment.Status.UpdatedReplicas,
		"available":  deployment.Status.AvailableReplicas,
		"age":        deployment.CreationTimestamp.Format("2006-01-02 15:04:05"),
		"labels":     deployment.Labels,
		"conditions": s.formatDeploymentConditions(deployment),
	}
}

func (s *K8sResourceService) formatDeploymentDetail(deployment *appsv1.Deployment) map[string]interface{} {
	manifest, _ := json.Marshal(deployment)

	detail := s.formatDeployment(deployment)
	detail["manifest"] = string(manifest)
	detail["images"] = extractContainerImages(deployment.Spec.Template.Spec.Containers)
	detail["selector"] = deployment.Spec.Selector.MatchLabels
	detail["annotations"] = deployment.Annotations

	strategy := map[string]interface{}{
		"type": string(deployment.Spec.Strategy.Type),
	}
	if ru := deployment.Spec.Strategy.RollingUpdate; ru != nil {
		if ru.MaxUnavailable != nil {
			strategy["maxUnavailable"] = ru.MaxUnavailable.String()
		}
		if ru.MaxSurge != nil {
			strategy["maxSurge"] = ru.MaxSurge.String()
		}
	}
	detail["strategy"] = strategy
	return detail
}

func (s *K8sResourceService) formatDeploymentConditions(deployment *appsv1.Deployment) []map[string]interface{} {
	conditions := make([]map[string]interface{}, 0)

	for _, cond := range deployment.Status.Conditions {
		conditions = append(conditions, map[string]interface{}{
			"type":    string(cond.Type),
			"status":  string(cond.Status),
			"reason":  cond.Reason,
			"message": cond.Message,
		})
	}

	if deployment.Status.Replicas == deployment.Status.ReadyReplicas && deployment.Status.Replicas > 0 {
		conditions = append([]map[string]interface{}{{
			"type":    "Ready",
			"status":  "True",
			"reason":  "",
			"message": "Deployment is ready",
		}}, conditions...)
	}

	return conditions
}
