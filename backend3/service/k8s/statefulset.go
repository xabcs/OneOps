package k8s

import (
	"encoding/json"
	"fmt"
	"strings"

	"oneops/backend3/pkg/logger"

	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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

	statefulset, err := clientset.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 StatefulSet 失败: %w", err)
	}

	var labelSelector string
	if len(statefulset.Spec.Selector.MatchLabels) > 0 {
		selectors := []string{}
		for k, v := range statefulset.Spec.Selector.MatchLabels {
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

	daemonset, err := clientset.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 DaemonSet 失败: %w", err)
	}

	var labelSelector string
	if len(daemonset.Spec.Selector.MatchLabels) > 0 {
		selectors := []string{}
		for k, v := range daemonset.Spec.Selector.MatchLabels {
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

	labelSelector := fmt.Sprintf("job-name=%s", jobName)

	list, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("获取 Pod 列表失败: %w", err)
	}

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

	labelSelector := fmt.Sprintf("cronjob-name=%s", cronJobName)

	list, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("获取 Pod 列表失败: %w", err)
	}

	result := make([]map[string]interface{}, len(list.Items))
	for i, item := range list.Items {
		podData := s.formatPod(&item)
		podData["ownerCronJob"] = cronJobName
		result[i] = podData
	}

	return result, nil
}

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
