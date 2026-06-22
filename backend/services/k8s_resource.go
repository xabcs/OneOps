package services

import (
	"context"
	"encoding/json"
	"fmt"
	"oneops/backend/logger"
	"strings"
	"time"

	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
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

// ========== Workloads 管理 ==========

// ListDeployments 获取 Deployment 列表（支持分页）
func (s *K8sResourceService) ListDeployments(clusterID uint, namespace string, page, pageSize int) ([]map[string]interface{}, int64, error) {
	clientset, _, err := s.clientPool.GetClient(clusterID)
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

	// 计算分页范围
	start := (page - 1) * pageSize
	end := start + pageSize

	// 边界检查
	if start < 0 {
		start = 0
	}
	if start > len(list.Items) {
		start = len(list.Items)
	}
	if end > len(list.Items) {
		end = len(list.Items)
	}

	// 提取分页数据
	items := list.Items[start:end]
	result := make([]map[string]interface{}, len(items))
	for i, item := range items {
		result[i] = s.formatDeployment(&item)
	}

	return result, total, nil
}

// GetDeployment 获取 Deployment 详情
func (s *K8sResourceService) GetDeployment(clusterID uint, namespace, name string) (map[string]interface{}, error) {
	clientset, _, err := s.clientPool.GetClient(clusterID)
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
	clientset, _, err := s.clientPool.GetClient(clusterID)
	if err != nil {
		return nil, err
	}

	// 首先获取 Deployment，提取其 label selector
	ctx, cancel := s.createContextWithTimeout()
	defer cancel()

	deployment, err := clientset.AppsV1().Deployments(namespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 Deployment 失败: %w", err)
	}

	// 构建 label selector
	var labelSelector string
	if len(deployment.Spec.Selector.MatchLabels) > 0 {
		selectors := []string{}
		for k, v := range deployment.Spec.Selector.MatchLabels {
			selectors = append(selectors, fmt.Sprintf("%s=%s", k, v))
		}
		labelSelector = strings.Join(selectors, ",")
	}

	// 使用 label selector 查询 Pods
	list, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("获取 Pod 列表失败: %w", err)
	}

	// 格式化 Pod 数据
	result := make([]map[string]interface{}, len(list.Items))
	for i, item := range list.Items {
		podData := s.formatPod(&item)
		// 添加是否属于当前 Deployment 的标记
		podData["ownerDeployment"] = deploymentName
		result[i] = podData
	}

	return result, nil
}

// CreateDeployment 创建 Deployment
func (s *K8sResourceService) CreateDeployment(clusterID uint, namespace string, manifest map[string]interface{}) error {
	_, config, err := s.clientPool.GetClient(clusterID)
	if err != nil {
		return err
	}

	// 创建 dynamic client
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("创建 dynamic client 失败: %w", err)
	}

	// 获取资源的 GVR
	gvr := schema.GroupVersionResource{
		Group:    "apps",
		Version:  "v1",
		Resource: "deployments",
	}

	// 创建 unstructured 对象
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
	_, config, err := s.clientPool.GetClient(clusterID)

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
	clientset, _, err := s.clientPool.GetClient(clusterID)
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
	clientset, _, err := s.clientPool.GetClient(clusterID)
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
	clientset, _, err := s.clientPool.GetClient(clusterID)
	if err != nil {
		return err
	}

	ctx, cancel := s.createContextWithTimeout()
	defer cancel()
	deployment, err := clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("获取 Deployment 失败: %w", err)
	}

	// 通过更新 annotation 触发滚动重启
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

// ListStatefulSets 获取 StatefulSet 列表（支持分页）
func (s *K8sResourceService) ListStatefulSets(clusterID uint, namespace string, page, pageSize int) ([]map[string]interface{}, int64, error) {
	clientset, _, err := s.clientPool.GetClient(clusterID)
	if err != nil {
		return nil, 0, err
	}

	ctx, cancel := s.createContextWithTimeout()
	defer cancel()
	list, err := clientset.AppsV1().StatefulSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, 0, fmt.Errorf("获取 StatefulSet 列表失败: %w", err)
	}

	total := int64(len(list.Items))

	// 计算分页范围
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
		result[i] = s.formatStatefulSet(&item)
	}

	return result, total, nil
}

// GetStatefulSet 获取 StatefulSet 详情
func (s *K8sResourceService) GetStatefulSet(clusterID uint, namespace, name string) (map[string]interface{}, error) {
	clientset, _, err := s.clientPool.GetClient(clusterID)
	if err != nil {
		return nil, err
	}

	ctx, cancel := s.createContextWithTimeout()
	defer cancel()
	sts, err := clientset.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 StatefulSet 详情失败: %w", err)
	}

	return s.formatStatefulSetDetail(sts), nil
}

// GetStatefulSetPods 获取 StatefulSet 管理的 Pods
func (s *K8sResourceService) GetStatefulSetPods(clusterID uint, namespace, name string) ([]map[string]interface{}, error) {
	clientset, _, err := s.clientPool.GetClient(clusterID)
	if err != nil {
		return nil, err
	}

	ctx, cancel := s.createContextWithTimeout()
	defer cancel()

	// 首先获取 StatefulSet，提取其 label selector
	statefulset, err := clientset.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 StatefulSet 失败: %w", err)
	}

	// 构建 label selector
	var labelSelector string
	if len(statefulset.Spec.Selector.MatchLabels) > 0 {
		selectors := []string{}
		for k, v := range statefulset.Spec.Selector.MatchLabels {
			selectors = append(selectors, fmt.Sprintf("%s=%s", k, v))
		}
		labelSelector = strings.Join(selectors, ",")
	}

	// 使用 label selector 查询 Pods
	list, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("获取 Pod 列表失败: %w", err)
	}

	// 格式化 Pod 数据
	result := make([]map[string]interface{}, len(list.Items))
	for i, item := range list.Items {
		podData := s.formatPod(&item)
		// 添加是否属于当前 StatefulSet 的标记
		podData["ownerStatefulSet"] = name
		result[i] = podData
	}

	return result, nil
}

// ListDaemonSets 获取 DaemonSet 列表（支持分页）
func (s *K8sResourceService) ListDaemonSets(clusterID uint, namespace string, page, pageSize int) ([]map[string]interface{}, int64, error) {
	clientset, _, err := s.clientPool.GetClient(clusterID)
	if err != nil {
		return nil, 0, err
	}

	ctx, cancel := s.createContextWithTimeout()
	defer cancel()
	list, err := clientset.AppsV1().DaemonSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, 0, fmt.Errorf("获取 DaemonSet 列表失败: %w", err)
	}

	total := int64(len(list.Items))

	// 计算分页范围
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
		result[i] = s.formatDaemonSet(&item)
	}

	return result, total, nil
}

// GetDaemonSet 获取 DaemonSet 详情
func (s *K8sResourceService) GetDaemonSet(clusterID uint, namespace, name string) (map[string]interface{}, error) {
	clientset, _, err := s.clientPool.GetClient(clusterID)
	if err != nil {
		return nil, err
	}

	ctx, cancel := s.createContextWithTimeout()
	defer cancel()
	ds, err := clientset.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 DaemonSet 详情失败: %w", err)
	}

	return s.formatDaemonSetDetail(ds), nil
}

// GetDaemonSetPods 获取 DaemonSet 管理的 Pods
func (s *K8sResourceService) GetDaemonSetPods(clusterID uint, namespace, name string) ([]map[string]interface{}, error) {
	clientset, _, err := s.clientPool.GetClient(clusterID)
	if err != nil {
		return nil, err
	}

	ctx, cancel := s.createContextWithTimeout()
	defer cancel()

	// 首先获取 DaemonSet，提取其 label selector
	daemonset, err := clientset.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 DaemonSet 失败: %w", err)
	}

	// 构建 label selector
	var labelSelector string
	if len(daemonset.Spec.Selector.MatchLabels) > 0 {
		selectors := []string{}
		for k, v := range daemonset.Spec.Selector.MatchLabels {
			selectors = append(selectors, fmt.Sprintf("%s=%s", k, v))
		}
		labelSelector = strings.Join(selectors, ",")
	}

	// 使用 label selector 查询 Pods
	list, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("获取 Pod 列表失败: %w", err)
	}

	// 格式化 Pod 数据
	result := make([]map[string]interface{}, len(list.Items))
	for i, item := range list.Items {
		podData := s.formatPod(&item)
		// 添加是否属于当前 DaemonSet 的标记
		podData["ownerDaemonSet"] = name
		result[i] = podData
	}

	return result, nil
}

// GetJobPods 获取 Job 关联的 Pods
func (s *K8sResourceService) GetJobPods(clusterID uint, namespace, jobName string) ([]map[string]interface{}, error) {
	clientset, _, err := s.clientPool.GetClient(clusterID)
	if err != nil {
		return nil, err
	}

	ctx, cancel := s.createContextWithTimeout()
	defer cancel()

	// 构建 label selector: job-name=<job-name>
	labelSelector := fmt.Sprintf("job-name=%s", jobName)

	// 使用 label selector 查询 Pods
	list, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("获取 Pod 列表失败: %w", err)
	}

	// 格式化 Pod 数据
	result := make([]map[string]interface{}, len(list.Items))
	for i, item := range list.Items {
		podData := s.formatPod(&item)
		podData["ownerJob"] = jobName
		result[i] = podData
	}

	return result, nil
}

// GetCronJobPods 获取 CronJob 关联的 Pods
func (s *K8sResourceService) GetCronJobPods(clusterID uint, namespace, cronJobName string) ([]map[string]interface{}, error) {
	clientset, _, err := s.clientPool.GetClient(clusterID)
	if err != nil {
		return nil, err
	}

	ctx, cancel := s.createContextWithTimeout()
	defer cancel()

	// 构建 label selector: cronjob-name=<cronjob-name>
	labelSelector := fmt.Sprintf("cronjob-name=%s", cronJobName)

	// 使用 label selector 查询 Pods
	list, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("获取 Pod 列表失败: %w", err)
	}

	// 格式化 Pod 数据
	result := make([]map[string]interface{}, len(list.Items))
	for i, item := range list.Items {
		podData := s.formatPod(&item)
		podData["ownerCronJob"] = cronJobName
		result[i] = podData
	}

	return result, nil
}

// ========== Workloads - Jobs ==========

// ListJobs 获取 Job 列表（支持分页）
func (s *K8sResourceService) ListJobs(clusterID uint, namespace string, page, pageSize int) ([]map[string]interface{}, int64, error) {
	clientset, _, err := s.clientPool.GetClient(clusterID)
	if err != nil {
		return nil, 0, err
	}

	ctx, cancel := s.createContextWithTimeout()
	defer cancel()
	list, err := clientset.BatchV1().Jobs(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, 0, fmt.Errorf("获取 Job 列表失败: %w", err)
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
		result[i] = s.formatJob(&item)
	}

	return result, total, nil
}

// GetJob 获取 Job 详情
func (s *K8sResourceService) GetJob(clusterID uint, namespace, name string) (map[string]interface{}, error) {
	clientset, _, err := s.clientPool.GetClient(clusterID)
	if err != nil {
		return nil, err
	}

	ctx, cancel := s.createContextWithTimeout()
	defer cancel()
	job, err := clientset.BatchV1().Jobs(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 Job 详情失败: %w", err)
	}

	return s.formatJobDetail(job), nil
}

// ========== Workloads - CronJobs ==========

// ListCronJobs 获取 CronJob 列表（支持分页）
func (s *K8sResourceService) ListCronJobs(clusterID uint, namespace string, page, pageSize int) ([]map[string]interface{}, int64, error) {
	clientset, _, err := s.clientPool.GetClient(clusterID)
	if err != nil {
		return nil, 0, err
	}

	ctx, cancel := s.createContextWithTimeout()
	defer cancel()
	list, err := clientset.BatchV1().CronJobs(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, 0, fmt.Errorf("获取 CronJob 列表失败: %w", err)
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
		result[i] = s.formatCronJob(&item)
	}

	return result, total, nil
}

// GetCronJob 获取 CronJob 详情
func (s *K8sResourceService) GetCronJob(clusterID uint, namespace, name string) (map[string]interface{}, error) {
	clientset, _, err := s.clientPool.GetClient(clusterID)
	if err != nil {
		return nil, err
	}

	ctx, cancel := s.createContextWithTimeout()
	defer cancel()
	cronJob, err := clientset.BatchV1().CronJobs(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 CronJob 详情失败: %w", err)
	}

	return s.formatCronJobDetail(cronJob), nil
}

// ========== Services 管理 ==========

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

	// 计算分页范围
	start := (page - 1) * pageSize
	end := start + pageSize

	// 边界检查
	if start < 0 {
		start = 0
	}
	if start > len(list.Items) {
		start = len(list.Items)
	}
	if end > len(list.Items) {
		end = len(list.Items)
	}

	// 提取分页数据
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

	// 获取 Endpoints
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

// ========== Ingresses 管理 ==========

// ListIngress 获取 Ingress 列表（支持分页）
func (s *K8sResourceService) ListIngress(clusterID uint, namespace string, page, pageSize int) ([]map[string]interface{}, int64, error) {
	clientset, _, err := s.clientPool.GetClient(clusterID)
	if err != nil {
		return nil, 0, err
	}

	ctx, cancel := s.createContextWithTimeout()
	defer cancel()
	list, err := clientset.NetworkingV1().Ingresses(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, 0, fmt.Errorf("获取 Ingress 列表失败: %w", err)
	}

	total := int64(len(list.Items))

	// 计算分页范围
	start := (page - 1) * pageSize
	end := start + pageSize

	// 边界检查
	if start < 0 {
		start = 0
	}
	totalInt := int(total)
	if end > totalInt {
		end = totalInt
	}

	// 分页
	items := list.Items
	if start >= len(items) {
		return []map[string]interface{}{}, total, nil
	}
	pagedItems := items[start:end]

	// 转换格式
	result := make([]map[string]interface{}, 0, len(pagedItems))
	for _, item := range pagedItems {
		result = append(result, s.formatIngress(&item))
	}

	return result, total, nil
}

// GetIngress 获取 Ingress 详情
func (s *K8sResourceService) GetIngress(clusterID uint, namespace, name string) (map[string]interface{}, error) {
	clientset, _, err := s.clientPool.GetClient(clusterID)
	if err != nil {
		return nil, err
	}

	ctx, cancel := s.createContextWithTimeout()
	defer cancel()
	ingress, err := clientset.NetworkingV1().Ingresses(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 Ingress 详情失败: %w", err)
	}

	return s.formatIngressDetail(ingress), nil
}

// CreateIngress 创建 Ingress
func (s *K8sResourceService) CreateIngress(clusterID uint, namespace string, manifest map[string]interface{}) error {
	clientset, config, err := s.clientPool.GetClient(clusterID)
	if err != nil {
		return err
	}

	// 尝试使用 typed client
	ingress := &networkingv1.Ingress{}
	jsonData, err := json.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("序列化 Ingress manifest 失败: %w", err)
	}
	if err := json.Unmarshal(jsonData, ingress); err != nil {
		return fmt.Errorf("解析 Ingress manifest 失败: %w", err)
	}

	ctx, cancel := s.createContextWithTimeout()
	defer cancel()
	_, err = clientset.NetworkingV1().Ingresses(namespace).Create(ctx, ingress, metav1.CreateOptions{})
	if err != nil {
		// 如果 typed client 失败，尝试使用 dynamic client
		dynamicClient, dynErr := dynamic.NewForConfig(config)
		if dynErr != nil {
			return fmt.Errorf("创建 Ingress 失败（typed client）: %w", err)
		}

		gvr := schema.GroupVersionResource{
			Group:    "networking.k8s.io",
			Version:  "v1",
			Resource: "ingresses",
		}

		uns := &unstructured.Unstructured{}
		uns.SetUnstructuredContent(manifest)

		ctx2, cancel2 := s.createContextWithTimeout()
		defer cancel2()
		_, err = dynamicClient.Resource(gvr).Namespace(namespace).Create(ctx2, uns, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("创建 Ingress 失败（dynamic client）: %w", err)
		}
	}

	logger.Info("创建 Ingress 成功",
		zap.Uint("cluster_id", clusterID),
		zap.String("namespace", namespace),
		zap.String("name", ingress.Name))

	return nil
}

// UpdateIngress 更新 Ingress
func (s *K8sResourceService) UpdateIngress(clusterID uint, namespace string, manifest map[string]interface{}) error {
	// 获取 config
	_, config, err := s.clientPool.GetClient(clusterID)
	if err != nil {
		return err
	}

	// 创建 dynamic client
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("创建 dynamic client 失败: %w", err)
	}

	gvr := schema.GroupVersionResource{
		Group:    "networking.k8s.io",
		Version:  "v1",
		Resource: "ingresses",
	}

	uns := &unstructured.Unstructured{}
	uns.SetUnstructuredContent(manifest)

	ctx, cancel := s.createContextWithTimeout()
	defer cancel()
	_, err = dynamicClient.Resource(gvr).Namespace(namespace).Update(ctx, uns, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("更新 Ingress 失败: %w", err)
	}

	logger.Info("更新 Ingress 成功",
		zap.Uint("cluster_id", clusterID),
		zap.String("namespace", namespace),
		zap.String("name", uns.GetName()))

	return nil
}

// DeleteIngress 删除 Ingress
func (s *K8sResourceService) DeleteIngress(clusterID uint, namespace, name string) error {
	clientset, _, err := s.clientPool.GetClient(clusterID)
	if err != nil {
		return err
	}

	ctx, cancel := s.createContextWithTimeout()
	defer cancel()
	err = clientset.NetworkingV1().Ingresses(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("删除 Ingress 失败: %w", err)
	}

	logger.Warn("删除 Ingress 成功",
		zap.Uint("cluster_id", clusterID),
		zap.String("namespace", namespace),
		zap.String("name", name))

	return nil
}

// formatIngress 格式化 Ingress 列表数据
func (s *K8sResourceService) formatIngress(ingress *networkingv1.Ingress) map[string]interface{} {
	result := map[string]interface{}{
		"name":             ingress.Name,
		"namespace":        ingress.Namespace,
		"age":              ingress.CreationTimestamp.Format("2006-01-02 15:04:05"),
		"hosts":            []string{},
		"addresses":        []string{},
		"ports":            []string{},
		"annotations":      ingress.Annotations,
		"ingressClassName": ingress.Spec.IngressClassName,
	}

	// 提取 hosts
	if ingress.Spec.Rules != nil {
		hosts := make([]string, 0)
		for _, rule := range ingress.Spec.Rules {
			if rule.Host != "" {
				hosts = append(hosts, rule.Host)
			}
		}
		result["hosts"] = hosts
	}

	// 提取 TLS
	if ingress.Spec.TLS != nil {
		tlsHosts := make([]string, 0)
		for _, tls := range ingress.Spec.TLS {
			tlsHosts = append(tlsHosts, tls.Hosts...)
		}
		result["tlsHosts"] = tlsHosts
	}

	// 提取 LoadBalancer addresses
	if ingress.Status.LoadBalancer.Ingress != nil {
		addresses := make([]string, 0)
		for _, lb := range ingress.Status.LoadBalancer.Ingress {
			if lb.IP != "" {
				addresses = append(addresses, lb.IP)
			}
			if lb.Hostname != "" {
				addresses = append(addresses, lb.Hostname)
			}
		}
		result["addresses"] = addresses
	}

	return result
}

// formatIngressDetail 格式化 Ingress 详情数据
func (s *K8sResourceService) formatIngressDetail(ingress *networkingv1.Ingress) map[string]interface{} {
	result := s.formatIngress(ingress)

	// 添加完整 manifest
	manifestBytes, _ := json.Marshal(ingress)
	result["manifest"] = string(manifestBytes)

	// 添加规则详情
	if ingress.Spec.Rules != nil {
		rules := make([]map[string]interface{}, 0)
		for _, rule := range ingress.Spec.Rules {
			ruleMap := map[string]interface{}{
				"host": rule.Host,
			}
			if rule.HTTP != nil {
				paths := make([]map[string]interface{}, 0)
				for _, path := range rule.HTTP.Paths {
					paths = append(paths, map[string]interface{}{
						"path":        path.Path,
						"pathType":    path.PathType,
						"serviceName": path.Backend.Service.Name,
						"servicePort": path.Backend.Service.Port,
					})
				}
				ruleMap["http"] = map[string]interface{}{
					"paths": paths,
				}
			}
			rules = append(rules, ruleMap)
		}
		result["rules"] = rules
	}

	// 添加 TLS 详情
	if ingress.Spec.TLS != nil {
		tls := make([]map[string]interface{}, 0)
		for _, t := range ingress.Spec.TLS {
			tls = append(tls, map[string]interface{}{
				"hosts":      t.Hosts,
				"secretName": t.SecretName,
			})
		}
		result["tls"] = tls
	}

	// 添加 IngressClass 信息
	if ingress.Spec.IngressClassName != nil {
		result["ingressClassName"] = *ingress.Spec.IngressClassName
	}

	return result
}

// ========== Pods 管理 ==========

// ListPods 获取 Pod 列表（支持分页）
func (s *K8sResourceService) ListPods(clusterID uint, namespace string, labelSelector string, page, pageSize int) ([]map[string]interface{}, int64, error) {
	clientset, _, err := s.clientPool.GetClient(clusterID)
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

	// 计算分页范围
	start := (page - 1) * pageSize
	end := start + pageSize

	// 边界检查
	if start < 0 {
		start = 0
	}
	if start > len(list.Items) {
		start = len(list.Items)
	}
	if end > len(list.Items) {
		end = len(list.Items)
	}

	// 提取分页数据
	items := list.Items[start:end]
	result := make([]map[string]interface{}, len(items))
	for i, item := range items {
		result[i] = s.formatPod(&item)
	}

	return result, total, nil
}

// GetPod 获取 Pod 详情
func (s *K8sResourceService) GetPod(clusterID uint, namespace, name string) (map[string]interface{}, error) {
	clientset, _, err := s.clientPool.GetClient(clusterID)
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
		Resource: "pods",
	}

	// 检查资源名称
	name, ok := manifest["metadata"].(map[string]interface{})["name"].(string)
	if !ok || name == "" {
		return fmt.Errorf("无效的资源名称")
	}

	// 创建 Unstructured 对象
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
	clientset, _, err := s.clientPool.GetClient(clusterID)
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
	clientset, _, err := s.clientPool.GetClient(clusterID)
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

// ========== 配置管理 ==========

// ListConfigMaps 获取 ConfigMap 列表（支持分页）
func (s *K8sResourceService) ListConfigMaps(clusterID uint, namespace string, page, pageSize int) ([]map[string]interface{}, int64, error) {
	clientset, _, err := s.clientPool.GetClient(clusterID)
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

	// 计算分页范围
	start := (page - 1) * pageSize
	end := start + pageSize

	// 边界检查
	if start < 0 {
		start = 0
	}
	if start > len(list.Items) {
		start = len(list.Items)
	}
	if end > len(list.Items) {
		end = len(list.Items)
	}

	// 提取分页数据
	items := list.Items[start:end]
	result := make([]map[string]interface{}, len(items))
	for i, item := range items {
		result[i] = s.formatConfigMap(&item)
	}

	return result, total, nil
}

// GetConfigMap 获取 ConfigMap 详情
func (s *K8sResourceService) GetConfigMap(clusterID uint, namespace, name string) (map[string]interface{}, error) {
	clientset, _, err := s.clientPool.GetClient(clusterID)
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
	clientset, _, err := s.clientPool.GetClient(clusterID)
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
	clientset, _, err := s.clientPool.GetClient(clusterID)
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

	// 计算分页范围
	start := (page - 1) * pageSize
	end := start + pageSize

	// 边界检查
	if start < 0 {
		start = 0
	}
	if start > len(list.Items) {
		start = len(list.Items)
	}
	if end > len(list.Items) {
		end = len(list.Items)
	}

	// 提取分页数据
	items := list.Items[start:end]
	result := make([]map[string]interface{}, len(items))
	for i, item := range items {
		result[i] = s.formatSecret(&item)
	}

	return result, total, nil
}

// GetSecret 获取 Secret 详情
func (s *K8sResourceService) GetSecret(clusterID uint, namespace, name string) (map[string]interface{}, error) {
	clientset, _, err := s.clientPool.GetClient(clusterID)
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
	clientset, _, err := s.clientPool.GetClient(clusterID)
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

// ========== Events 管理 ==========

// ListEvents 获取 Event 列表
func (s *K8sResourceService) ListEvents(clusterID uint, namespace string, fieldSelector string) ([]map[string]interface{}, error) {
	clientset, _, err := s.clientPool.GetClient(clusterID)
	if err != nil {
		return nil, err
	}

	ctx, cancel := s.createContextWithTimeout()
	defer cancel()
	list, err := clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fieldSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("获取 Event 列表失败: %w", err)
	}

	result := make([]map[string]interface{}, len(list.Items))
	for i, item := range list.Items {
		result[i] = s.formatEvent(&item)
	}

	return result, nil
}

// ========== 格式化方法 ==========

// formatDeployment 格式化 Deployment 基本信息
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

// formatDeploymentDetail 格式化 Deployment 详情
func (s *K8sResourceService) formatDeploymentDetail(deployment *appsv1.Deployment) map[string]interface{} {
	// 获取完整的 YAML manifest
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

// formatDeploymentConditions 格式化 Deployment 状态
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

	// 简化状态
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

// formatStatefulSet 格式化 StatefulSet 基本信息
func (s *K8sResourceService) formatStatefulSet(sts *appsv1.StatefulSet) map[string]interface{} {
	return map[string]interface{}{
		"name":       sts.Name,
		"namespace":  sts.Namespace,
		"replicas":   sts.Status.Replicas,
		"ready":      sts.Status.ReadyReplicas,
		"current":    sts.Status.CurrentReplicas,
		"updated":    sts.Status.UpdatedReplicas,
		"age":        sts.CreationTimestamp.Format("2006-01-02 15:04:05"),
		"labels":     sts.Labels,
		"conditions": s.formatStatefulSetConditions(sts),
	}
}

// formatStatefulSetDetail 格式化 StatefulSet 详情
func (s *K8sResourceService) formatStatefulSetDetail(sts *appsv1.StatefulSet) map[string]interface{} {
	manifest, _ := json.Marshal(sts)

	detail := s.formatStatefulSet(sts)
	detail["manifest"] = string(manifest)
	detail["images"] = extractContainerImages(sts.Spec.Template.Spec.Containers)
	detail["selector"] = sts.Spec.Selector.MatchLabels
	detail["serviceName"] = sts.Spec.ServiceName
	detail["annotations"] = sts.Annotations

	strategy := map[string]interface{}{
		"type": string(sts.Spec.UpdateStrategy.Type),
	}
	if ru := sts.Spec.UpdateStrategy.RollingUpdate; ru != nil {
		if ru.Partition != nil {
			strategy["partition"] = *ru.Partition
		}
		if ru.MaxUnavailable != nil {
			strategy["maxUnavailable"] = ru.MaxUnavailable.String()
		}
	}
	detail["strategy"] = strategy
	return detail
}

// formatStatefulSetConditions 格式化 StatefulSet 状态
func (s *K8sResourceService) formatStatefulSetConditions(sts *appsv1.StatefulSet) []map[string]interface{} {
	conditions := make([]map[string]interface{}, 0)

	for _, cond := range sts.Status.Conditions {
		conditions = append(conditions, map[string]interface{}{
			"type":    string(cond.Type),
			"status":  string(cond.Status),
			"reason":  cond.Reason,
			"message": cond.Message,
		})
	}

	return conditions
}

// formatDaemonSet 格式化 DaemonSet 基本信息
func (s *K8sResourceService) formatDaemonSet(ds *appsv1.DaemonSet) map[string]interface{} {
	return map[string]interface{}{
		"name":       ds.Name,
		"namespace":  ds.Namespace,
		"desired":    ds.Status.DesiredNumberScheduled,
		"current":    ds.Status.CurrentNumberScheduled,
		"ready":      ds.Status.NumberReady,
		"updated":    ds.Status.UpdatedNumberScheduled,
		"available":  ds.Status.NumberAvailable,
		"age":        ds.CreationTimestamp.Format("2006-01-02 15:04:05"),
		"labels":     ds.Labels,
		"conditions": s.formatDaemonSetConditions(ds),
	}
}

// formatDaemonSetDetail 格式化 DaemonSet 详情
func (s *K8sResourceService) formatDaemonSetDetail(ds *appsv1.DaemonSet) map[string]interface{} {
	manifest, _ := json.Marshal(ds)

	detail := s.formatDaemonSet(ds)
	detail["manifest"] = string(manifest)
	detail["images"] = extractContainerImages(ds.Spec.Template.Spec.Containers)
	detail["selector"] = ds.Spec.Selector.MatchLabels
	detail["annotations"] = ds.Annotations

	strategy := map[string]interface{}{
		"type": string(ds.Spec.UpdateStrategy.Type),
	}
	if ru := ds.Spec.UpdateStrategy.RollingUpdate; ru != nil {
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

// formatDaemonSetConditions 格式化 DaemonSet 状态
func (s *K8sResourceService) formatDaemonSetConditions(ds *appsv1.DaemonSet) []map[string]interface{} {
	conditions := make([]map[string]interface{}, 0)

	for _, cond := range ds.Status.Conditions {
		conditions = append(conditions, map[string]interface{}{
			"type":    string(cond.Type),
			"status":  string(cond.Status),
			"reason":  cond.Reason,
			"message": cond.Message,
		})
	}

	return conditions
}

// formatService 格式化 Service 基本信息
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

// formatServiceDetail 格式化 Service 详情
func (s *K8sResourceService) formatServiceDetail(svc *corev1.Service, endpoints *corev1.Endpoints) map[string]interface{} {
	manifest, _ := json.Marshal(svc)

	detail := s.formatService(svc)
	detail["manifest"] = string(manifest)
	detail["endpoints"] = s.formatEndpoints(endpoints)
	return detail
}

// formatServicePorts 格式化 Service 端口
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

// formatEndpoints 格式化 Endpoints
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

// formatPod 格式化 Pod 基本信息
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

// formatPodDetail 格式化 Pod 详情
func (s *K8sResourceService) formatPodDetail(pod *corev1.Pod) map[string]interface{} {
	manifest, _ := json.Marshal(pod)

	detail := s.formatPod(pod)
	detail["manifest"] = string(manifest)
	detail["images"] = extractContainerImages(pod.Spec.Containers)
	detail["containers"] = s.formatContainers(pod)
	return detail
}

// formatContainers 格式化容器列表
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

// formatConfigMap 格式化 ConfigMap 基本信息
func (s *K8sResourceService) formatConfigMap(cm *corev1.ConfigMap) map[string]interface{} {
	return map[string]interface{}{
		"name":      cm.Name,
		"namespace": cm.Namespace,
		"age":       cm.CreationTimestamp.Format("2006-01-02 15:04:05"),
		"labels":    cm.Labels,
		"dataKeys":  getConfigMapDataKeys(cm),
	}
}

// formatConfigMapDetail 格式化 ConfigMap 详情
func (s *K8sResourceService) formatConfigMapDetail(cm *corev1.ConfigMap) map[string]interface{} {
	manifest, _ := json.Marshal(cm)

	detail := s.formatConfigMap(cm)
	detail["manifest"] = string(manifest)
	detail["data"] = cm.Data
	return detail
}

// formatSecret 格式化 Secret 基本信息（隐藏敏感数据）
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

// formatSecretDetail 格式化 Secret 详情（隐藏敏感数据）
func (s *K8sResourceService) formatSecretDetail(secret *corev1.Secret) map[string]interface{} {
	manifest, _ := json.Marshal(secret)

	detail := s.formatSecret(secret)
	detail["manifest"] = string(manifest)
	// 详情中也不返回实际数据，只显示 key
	detail["data"] = maskSecretData(secret.Data)
	return detail
}

// formatEvent 格式化 Event
func (s *K8sResourceService) formatEvent(event *corev1.Event) map[string]interface{} {
	return map[string]interface{}{
		"type":      event.Type,
		"reason":    event.Reason,
		"message":   event.Message,
		"firstSeen": event.FirstTimestamp.Format("2006-01-02 15:04:05"),
		"lastSeen":  event.LastTimestamp.Format("2006-01-02 15:04:05"),
		"count":     event.Count,
		"involvedObject": map[string]interface{}{
			"kind":      event.InvolvedObject.Kind,
			"name":      event.InvolvedObject.Name,
			"namespace": event.InvolvedObject.Namespace,
		},
	}
}

// ========== 辅助函数 ==========

// extractContainerImages 提取容器镜像列表
func extractContainerImages(containers []corev1.Container) []string {
	images := make([]string, len(containers))
	for i, c := range containers {
		images[i] = c.Image
	}
	return images
}

// getPodStatus 获取 Pod 状态
func getPodStatus(pod *corev1.Pod) string {
	// 检查容器状态
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.State.Terminated != nil && cs.State.Terminated.ExitCode == 0 {
			return "Succeeded"
		}
		if cs.State.Terminated != nil {
			return "Failed"
		}
	}

	// 检查初始化容器
	for _, cs := range pod.Status.InitContainerStatuses {
		if cs.State.Terminated != nil && cs.State.Terminated.ExitCode != 0 {
			return "Failed"
		}
	}

	return string(pod.Status.Phase)
}

// countPodRestarts 统计 Pod 重启次数
func countPodRestarts(pod *corev1.Pod) int32 {
	var count int32
	for _, cs := range pod.Status.ContainerStatuses {
		count += cs.RestartCount
	}
	for _, cs := range pod.Status.InitContainerStatuses {
		count += cs.RestartCount
	}
	return count
}

// getContainerStatus 获取容器状态
func getContainerStatus(pod *corev1.Pod, name string) corev1.ContainerStatus {
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.Name == name {
			return cs
		}
	}
	return corev1.ContainerStatus{}
}

// formatContainerState 格式化容器状态
func formatContainerState(state corev1.ContainerState) string {
	if state.Running != nil {
		return "Running"
	}
	if state.Waiting != nil {
		return "Waiting: " + state.Waiting.Reason
	}
	if state.Terminated != nil {
		return "Terminated: " + state.Terminated.Reason
	}
	return "Unknown"
}

// formatSubsetPorts 格式化 Subset 端口
func formatSubsetPorts(ports []corev1.EndpointPort) []map[string]interface{} {
	result := make([]map[string]interface{}, len(ports))
	for i, p := range ports {
		result[i] = map[string]interface{}{
			"name":     p.Name,
			"protocol": string(p.Protocol),
			"port":     p.Port,
		}
	}
	return result
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

// getCurrentTimestamp 获取当前时间戳
func getCurrentTimestamp() string {
	return metav1.Now().Format("2006-01-02 15:04:05")
}

// ========== Job 格式化 ==========

// formatJob 格式化 Job 列表项
func (s *K8sResourceService) formatJob(job *batchv1.Job) map[string]interface{} {
	result := map[string]interface{}{
		"name":        job.Name,
		"namespace":   job.Namespace,
		"age":         job.CreationTimestamp.Format("2006-01-02 15:04:05"),
		"labels":      job.Labels,
		"completions": job.Spec.Completions,
		"parallelism": job.Spec.Parallelism,
	}

	if job.Status.Succeeded > 0 {
		result["succeeded"] = job.Status.Succeeded
	}
	if job.Status.Failed > 0 {
		result["failed"] = job.Status.Failed
	}
	if job.Status.Active > 0 {
		result["active"] = job.Status.Active
	}

	// 计算运行时长
	if job.Status.StartTime != nil {
		duration := time.Since(job.Status.StartTime.Time)
		result["duration"] = formatJobDuration(int(duration.Seconds()))
		if job.Status.CompletionTime != nil {
			duration = job.Status.CompletionTime.Sub(job.Status.StartTime.Time)
			result["duration"] = formatJobDuration(int(duration.Seconds()))
		}
	}

	// 状态
	if job.Status.Succeeded > 0 {
		total := int32(1)
		if job.Spec.Completions != nil {
			total = *job.Spec.Completions
		}
		if job.Status.Succeeded >= total {
			result["status"] = "完成"
		} else {
			result["status"] = "运行中"
		}
	} else if job.Status.Failed > 0 {
		result["status"] = "失败"
	} else if job.Status.Active > 0 {
		result["status"] = "运行中"
	} else {
		result["status"] = "等待中"
	}

	return result
}

// formatJobDetail 格式化 Job 详情
func (s *K8sResourceService) formatJobDetail(job *batchv1.Job) map[string]interface{} {
	result := s.formatJob(job)

	// 添加 manifest
	manifest, _ := json.Marshal(job)
	result["manifest"] = string(manifest)

	// 提取镜像列表
	images := make([]string, 0)
	for _, c := range job.Spec.Template.Spec.Containers {
		images = append(images, c.Image)
	}
	result["images"] = images

	// 选择器
	if job.Spec.Selector != nil {
		result["selector"] = job.Spec.Selector.MatchLabels
	}

	// 状态条件
	conditions := make([]map[string]interface{}, len(job.Status.Conditions))
	for i, c := range job.Status.Conditions {
		conditions[i] = map[string]interface{}{
			"type":   string(c.Type),
			"status": string(c.Status),
			"reason": c.Reason,
		}
	}
	result["conditions"] = conditions

	return result
}

// ========== CronJob 格式化 ==========

// formatCronJob 格式化 CronJob 列表项
func (s *K8sResourceService) formatCronJob(cronJob *batchv1.CronJob) map[string]interface{} {
	result := map[string]interface{}{
		"name":      cronJob.Name,
		"namespace": cronJob.Namespace,
		"schedule":  cronJob.Spec.Schedule,
		"suspend":   cronJob.Spec.Suspend != nil && *cronJob.Spec.Suspend,
		"age":       cronJob.CreationTimestamp.Format("2006-01-02 15:04:05"),
		"labels":    cronJob.Labels,
	}

	if cronJob.Status.LastScheduleTime != nil {
		result["lastSchedule"] = cronJob.Status.LastScheduleTime.Format("2006-01-02 15:04:05")
	}

	if len(cronJob.Status.Active) > 0 {
		result["activeJobs"] = len(cronJob.Status.Active)
	}

	if cronJob.Spec.ConcurrencyPolicy != "" {
		result["concurrencyPolicy"] = string(cronJob.Spec.ConcurrencyPolicy)
	}

	return result
}

// formatCronJobDetail 格式化 CronJob 详情
func (s *K8sResourceService) formatCronJobDetail(cronJob *batchv1.CronJob) map[string]interface{} {
	result := s.formatCronJob(cronJob)

	// 添加 manifest
	manifest, _ := json.Marshal(cronJob)
	result["manifest"] = string(manifest)

	// 提取镜像列表
	images := make([]string, 0)
	for _, c := range cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers {
		images = append(images, c.Image)
	}
	result["images"] = images

	return result
}

// formatJobDuration 格式化时间间隔（秒）
func formatJobDuration(seconds int) string {
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	if seconds < 3600 {
		return fmt.Sprintf("%dm", seconds/60)
	}
	return fmt.Sprintf("%dh%dm", seconds/3600, (seconds%3600)/60)
}

// ========== StatefulSet Lifecycle ==========

// RestartStatefulSet 重启 StatefulSet
func (s *K8sResourceService) RestartStatefulSet(clusterID uint, namespace, name string) error {
	clientset, _, err := s.clientPool.GetClient(clusterID)
	if err != nil {
		return err
	}

	ctx, cancel := s.createContextWithTimeout()
	defer cancel()
	sts, err := clientset.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("获取 StatefulSet 失败: %w", err)
	}

	// 通过更新 annotation 触发滚动重启
	if sts.Spec.Template.Annotations == nil {
		sts.Spec.Template.Annotations = make(map[string]string)
	}
	sts.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = getCurrentTimestamp()

	_, err = clientset.AppsV1().StatefulSets(namespace).Update(ctx, sts, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("重启 StatefulSet 失败: %w", err)
	}

	logger.Info("重启 StatefulSet 成功",
		zap.Uint("cluster_id", clusterID),
		zap.String("namespace", namespace),
		zap.String("name", name))

	return nil
}

// DeleteStatefulSet 删除 StatefulSet
func (s *K8sResourceService) DeleteStatefulSet(clusterID uint, namespace, name string) error {
	clientset, _, err := s.clientPool.GetClient(clusterID)
	if err != nil {
		return err
	}

	ctx, cancel := s.createContextWithTimeout()
	defer cancel()
	err = clientset.AppsV1().StatefulSets(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("删除 StatefulSet 失败: %w", err)
	}

	logger.Info("删除 StatefulSet 成功",
		zap.Uint("cluster_id", clusterID),
		zap.String("namespace", namespace),
		zap.String("name", name))

	return nil
}

// ========== DaemonSet Lifecycle ==========

// RestartDaemonSet 重启 DaemonSet
func (s *K8sResourceService) RestartDaemonSet(clusterID uint, namespace, name string) error {
	clientset, _, err := s.clientPool.GetClient(clusterID)
	if err != nil {
		return err
	}

	ctx, cancel := s.createContextWithTimeout()
	defer cancel()
	ds, err := clientset.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("获取 DaemonSet 失败: %w", err)
	}

	// 通过更新 annotation 触发滚动重启
	if ds.Spec.Template.Annotations == nil {
		ds.Spec.Template.Annotations = make(map[string]string)
	}
	ds.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = getCurrentTimestamp()

	_, err = clientset.AppsV1().DaemonSets(namespace).Update(ctx, ds, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("重启 DaemonSet 失败: %w", err)
	}

	logger.Info("重启 DaemonSet 成功",
		zap.Uint("cluster_id", clusterID),
		zap.String("namespace", namespace),
		zap.String("name", name))

	return nil
}

// DeleteDaemonSet 删除 DaemonSet
func (s *K8sResourceService) DeleteDaemonSet(clusterID uint, namespace, name string) error {
	clientset, _, err := s.clientPool.GetClient(clusterID)
	if err != nil {
		return err
	}

	ctx, cancel := s.createContextWithTimeout()
	defer cancel()
	err = clientset.AppsV1().DaemonSets(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("删除 DaemonSet 失败: %w", err)
	}

	logger.Info("删除 DaemonSet 成功",
		zap.Uint("cluster_id", clusterID),
		zap.String("namespace", namespace),
		zap.String("name", name))

	return nil
}

// ========== Job Lifecycle ==========

// DeleteJob 删除 Job
func (s *K8sResourceService) DeleteJob(clusterID uint, namespace, name string) error {
	clientset, _, err := s.clientPool.GetClient(clusterID)
	if err != nil {
		return err
	}

	ctx, cancel := s.createContextWithTimeout()
	defer cancel()
	err = clientset.BatchV1().Jobs(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("删除 Job 失败: %w", err)
	}

	logger.Info("删除 Job 成功",
		zap.Uint("cluster_id", clusterID),
		zap.String("namespace", namespace),
		zap.String("name", name))

	return nil
}

// ========== CronJob Lifecycle ==========

// DeleteCronJob 删除 CronJob
func (s *K8sResourceService) DeleteCronJob(clusterID uint, namespace, name string) error {
	clientset, _, err := s.clientPool.GetClient(clusterID)
	if err != nil {
		return err
	}

	ctx, cancel := s.createContextWithTimeout()
	defer cancel()
	err = clientset.BatchV1().CronJobs(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("删除 CronJob 失败: %w", err)
	}

	logger.Info("删除 CronJob 成功",
		zap.Uint("cluster_id", clusterID),
		zap.String("namespace", namespace),
		zap.String("name", name))

	return nil
}

// SuspendCronJob 暂停/恢复 CronJob
func (s *K8sResourceService) SuspendCronJob(clusterID uint, namespace, name string, suspend bool) error {
	clientset, _, err := s.clientPool.GetClient(clusterID)
	if err != nil {
		return err
	}

	ctx, cancel := s.createContextWithTimeout()
	defer cancel()
	cronJob, err := clientset.BatchV1().CronJobs(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("获取 CronJob 失败: %w", err)
	}

	cronJob.Spec.Suspend = &suspend

	_, err = clientset.BatchV1().CronJobs(namespace).Update(ctx, cronJob, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("%s CronJob 失败: %w", map[bool]string{true: "暂停", false: "恢复"}[suspend], err)
	}

	logger.Info("更新 CronJob 暂停状态成功",
		zap.Uint("cluster_id", clusterID),
		zap.String("namespace", namespace),
		zap.String("name", name),
		zap.Bool("suspend", suspend))

	return nil
}
