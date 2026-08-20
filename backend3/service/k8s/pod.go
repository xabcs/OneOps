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

// ListPods 获取 Pod 列表（支持分页）
func (s *K8sResourceService) ListPods(clusterID uint, namespace string, labelSelector string, page, pageSize int) ([]map[string]interface{}, int64, error) {
	clientset, err := s.getScopedClientset(clusterID)
	if err != nil {
		return nil, 0, err
	}

	ctx, cancel := s.createContextWithTimeout()
	defer cancel()
	list, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("获取 Pod 列表失败: %w", err)
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
		result[i] = s.formatPod(&item)
	}

	return result, total, nil
}

// GetPod 获取 Pod 详情
func (s *K8sResourceService) GetPod(clusterID uint, namespace, name string) (map[string]interface{}, error) {
	clientset, err := s.getScopedClientset(clusterID)
	if err != nil {
		return nil, err
	}

	ctx, cancel := s.createContextWithTimeout()
	defer cancel()
	pod, err := clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 Pod 详情失败: %w", err)
	}

	return s.formatPodDetail(pod), nil
}

// UpdatePod 更新 Pod（使用 dynamic client）
func (s *K8sResourceService) UpdatePod(clusterID uint, namespace string, manifest map[string]interface{}) error {
	config, err := s.getScopedConfig(clusterID)
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
		Resource: "pods",
	}

	name, ok := manifest["metadata"].(map[string]interface{})["name"].(string)
	if !ok || name == "" {
		return fmt.Errorf("无效的资源名称")
	}

	uns := &unstructured.Unstructured{
		Object: manifest,
	}

	ctx, cancel := s.createContextWithTimeout()
	defer cancel()

	_, err = dynamicClient.Resource(gvr).Namespace(namespace).Update(ctx, uns, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("更新 Pod 失败: %w", err)
	}

	logger.Info("更新 Pod 成功",
		zap.Uint("cluster_id", clusterID),
		zap.String("namespace", namespace),
		zap.String("name", name))

	return nil
}

// GetPodLogs 获取 Pod 日志
func (s *K8sResourceService) GetPodLogs(clusterID uint, namespace, name, container string, tailLines int64) (string, error) {
	clientset, err := s.getScopedClientset(clusterID)
	if err != nil {
		return "", err
	}

	ctx, cancel := s.createContextWithTimeout()
	defer cancel()
	req := clientset.CoreV1().Pods(namespace).GetLogs(name, &corev1.PodLogOptions{
		Container: container,
		TailLines: &tailLines,
	})

	logs, err := req.Do(ctx).Raw()
	if err != nil {
		return "", fmt.Errorf("获取 Pod 日志失败: %w", err)
	}

	return string(logs), nil
}

// DeletePod 删除 Pod
func (s *K8sResourceService) DeletePod(clusterID uint, namespace, name string) error {
	clientset, err := s.getScopedClientset(clusterID)
	if err != nil {
		return err
	}

	ctx, cancel := s.createContextWithTimeout()
	defer cancel()
	err = clientset.CoreV1().Pods(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("删除 Pod 失败: %w", err)
	}

	logger.Warn("删除 Pod 成功",
		zap.Uint("cluster_id", clusterID),
		zap.String("namespace", namespace),
		zap.String("name", name))

	return nil
}

func (s *K8sResourceService) formatPod(pod *corev1.Pod) map[string]interface{} {
	return map[string]interface{}{
		"name":       pod.Name,
		"namespace":  pod.Namespace,
		"status":     getPodStatus(pod),
		"phase":      string(pod.Status.Phase),
		"ip":         pod.Status.PodIP,
		"node":       pod.Spec.NodeName,
		"age":        pod.CreationTimestamp.Format("2006-01-02 15:04:05"),
		"labels":     pod.Labels,
		"restarts":   countPodRestarts(pod),
		"containers": s.formatContainers(pod),
	}
}

func (s *K8sResourceService) formatPodDetail(pod *corev1.Pod) map[string]interface{} {
	manifest, _ := json.Marshal(pod)

	detail := s.formatPod(pod)
	detail["manifest"] = string(manifest)
	detail["images"] = extractContainerImages(pod.Spec.Containers)
	detail["containers"] = s.formatContainers(pod)
	return detail
}

func (s *K8sResourceService) formatContainers(pod *corev1.Pod) []map[string]interface{} {
	containers := make([]map[string]interface{}, 0)

	for _, c := range pod.Spec.Containers {
		status := getContainerStatus(pod, c.Name)
		containers = append(containers, map[string]interface{}{
			"name":    c.Name,
			"image":   c.Image,
			"ready":   status.Ready,
			"restart": status.RestartCount,
			"state":   formatContainerState(status.State),
		})
	}

	return containers
}
