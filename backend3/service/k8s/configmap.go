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

// ListConfigMaps 获取 ConfigMap 列表（支持分页）
func (s *K8sResourceService) ListConfigMaps(clusterID uint, namespace string, page, pageSize int) ([]map[string]interface{}, int64, error) {
	clientset, err := s.getScopedClientset(clusterID)
	if err != nil {
		return nil, 0, err
	}

	ctx, cancel := s.createContextWithTimeout()
	defer cancel()
	list, err := clientset.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, 0, fmt.Errorf("获取 ConfigMap 列表失败: %w", err)
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
		result[i] = s.formatConfigMap(&item)
	}

	return result, total, nil
}

// GetConfigMap 获取 ConfigMap 详情
func (s *K8sResourceService) GetConfigMap(clusterID uint, namespace, name string) (map[string]interface{}, error) {
	clientset, err := s.getScopedClientset(clusterID)
	if err != nil {
		return nil, err
	}

	ctx, cancel := s.createContextWithTimeout()
	defer cancel()
	cm, err := clientset.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 ConfigMap 详情失败: %w", err)
	}

	return s.formatConfigMapDetail(cm), nil
}

// CreateConfigMap 创建 ConfigMap
func (s *K8sResourceService) CreateConfigMap(clusterID uint, namespace string, manifest map[string]interface{}) error {
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
		Resource: "configmaps",
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
		return fmt.Errorf("创建 ConfigMap 失败: %w", err)
	}

	logger.Info("创建 ConfigMap 成功",
		zap.Uint("cluster_id", clusterID),
		zap.String("namespace", namespace),
		zap.String("name", uns.GetName()))

	return nil
}

// UpdateConfigMap 更新 ConfigMap
func (s *K8sResourceService) UpdateConfigMap(clusterID uint, namespace string, manifest map[string]interface{}) error {
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
		Resource: "configmaps",
	}

	uns := &unstructured.Unstructured{}
	uns.SetUnstructuredContent(manifest)

	ctx, cancel := s.createContextWithTimeout()
	defer cancel()
	_, err = dynamicClient.Resource(gvr).Namespace(namespace).Update(ctx, uns, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("更新 ConfigMap 失败: %w", err)
	}

	logger.Info("更新 ConfigMap 成功",
		zap.Uint("cluster_id", clusterID),
		zap.String("namespace", namespace),
		zap.String("name", uns.GetName()))

	return nil
}

// DeleteConfigMap 删除 ConfigMap
func (s *K8sResourceService) DeleteConfigMap(clusterID uint, namespace, name string) error {
	clientset, err := s.getScopedClientset(clusterID)
	if err != nil {
		return err
	}

	ctx, cancel := s.createContextWithTimeout()
	defer cancel()
	err = clientset.CoreV1().ConfigMaps(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("删除 ConfigMap 失败: %w", err)
	}

	logger.Warn("删除 ConfigMap 成功",
		zap.Uint("cluster_id", clusterID),
		zap.String("namespace", namespace),
		zap.String("name", name))

	return nil
}

// ListSecrets 获取 Secret 列表（支持分页）
func (s *K8sResourceService) ListSecrets(clusterID uint, namespace string, page, pageSize int) ([]map[string]interface{}, int64, error) {
	clientset, err := s.getScopedClientset(clusterID)
	if err != nil {
		return nil, 0, err
	}

	ctx, cancel := s.createContextWithTimeout()
	defer cancel()
	list, err := clientset.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, 0, fmt.Errorf("获取 Secret 列表失败: %w", err)
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
		result[i] = s.formatSecret(&item)
	}

	return result, total, nil
}

// GetSecret 获取 Secret 详情
func (s *K8sResourceService) GetSecret(clusterID uint, namespace, name string) (map[string]interface{}, error) {
	clientset, err := s.getScopedClientset(clusterID)
	if err != nil {
		return nil, err
	}

	ctx, cancel := s.createContextWithTimeout()
	defer cancel()
	secret, err := clientset.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 Secret 详情失败: %w", err)
	}

	return s.formatSecretDetail(secret), nil
}

// CreateSecret 创建 Secret
func (s *K8sResourceService) CreateSecret(clusterID uint, namespace string, manifest map[string]interface{}) error {
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
		Resource: "secrets",
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
		return fmt.Errorf("创建 Secret 失败: %w", err)
	}

	logger.Info("创建 Secret 成功",
		zap.Uint("cluster_id", clusterID),
		zap.String("namespace", namespace),
		zap.String("name", uns.GetName()))

	return nil
}

// UpdateSecret 更新 Secret
func (s *K8sResourceService) UpdateSecret(clusterID uint, namespace string, manifest map[string]interface{}) error {
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
		Resource: "secrets",
	}

	uns := &unstructured.Unstructured{}
	uns.SetUnstructuredContent(manifest)

	ctx, cancel := s.createContextWithTimeout()
	defer cancel()
	_, err = dynamicClient.Resource(gvr).Namespace(namespace).Update(ctx, uns, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("更新 Secret 失败: %w", err)
	}

	logger.Info("更新 Secret 成功",
		zap.Uint("cluster_id", clusterID),
		zap.String("namespace", namespace),
		zap.String("name", uns.GetName()))

	return nil
}

// DeleteSecret 删除 Secret
func (s *K8sResourceService) DeleteSecret(clusterID uint, namespace, name string) error {
	clientset, err := s.getScopedClientset(clusterID)
	if err != nil {
		return err
	}

	ctx, cancel := s.createContextWithTimeout()
	defer cancel()
	err = clientset.CoreV1().Secrets(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("删除 Secret 失败: %w", err)
	}

	logger.Warn("删除 Secret 成功",
		zap.Uint("cluster_id", clusterID),
		zap.String("namespace", namespace),
		zap.String("name", name))

	return nil
}

func (s *K8sResourceService) formatConfigMap(cm *corev1.ConfigMap) map[string]interface{} {
	return map[string]interface{}{
		"name":      cm.Name,
		"namespace": cm.Namespace,
		"age":       cm.CreationTimestamp.Format("2006-01-02 15:04:05"),
		"labels":    cm.Labels,
		"dataKeys":  getConfigMapDataKeys(cm),
	}
}

func (s *K8sResourceService) formatConfigMapDetail(cm *corev1.ConfigMap) map[string]interface{} {
	manifest, _ := json.Marshal(cm)

	detail := s.formatConfigMap(cm)
	detail["manifest"] = string(manifest)
	detail["data"] = cm.Data
	return detail
}

func (s *K8sResourceService) formatSecret(secret *corev1.Secret) map[string]interface{} {
	return map[string]interface{}{
		"name":      secret.Name,
		"namespace": secret.Namespace,
		"type":      string(secret.Type),
		"age":       secret.CreationTimestamp.Format("2006-01-02 15:04:05"),
		"labels":    secret.Labels,
		"dataKeys":  getSecretDataKeys(secret),
	}
}

func (s *K8sResourceService) formatSecretDetail(secret *corev1.Secret) map[string]interface{} {
	manifest, _ := json.Marshal(secret)

	detail := s.formatSecret(secret)
	detail["manifest"] = string(manifest)
	detail["data"] = maskSecretData(secret.Data)
	return detail
}

// getConfigMapDataKeys 获取 ConfigMap 数据键列表
func getConfigMapDataKeys(cm *corev1.ConfigMap) []string {
	keys := make([]string, 0, len(cm.Data))
	for k := range cm.Data {
		keys = append(keys, k)
	}
	return keys
}

// getSecretDataKeys 获取 Secret 数据键列表
func getSecretDataKeys(secret *corev1.Secret) []string {
	keys := make([]string, 0, len(secret.Data))
	for k := range secret.Data {
		keys = append(keys, k)
	}
	return keys
}

// maskSecretData 隐藏 Secret 敏感数据
func maskSecretData(data map[string][]byte) map[string]string {
	result := make(map[string]string)
	for k := range data {
		result[k] = "******"
	}
	return result
}
