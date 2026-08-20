package k8s

import (
	"encoding/json"
	"fmt"
	"time"

	"oneops/backend3/pkg/logger"

	"go.uber.org/zap"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListJobs 获取 Job 列表（支持分页）
func (s *K8sResourceService) ListJobs(clusterID uint, namespace string, page, pageSize int) ([]map[string]interface{}, int64, error) {
	clientset, err := s.getScopedClientset(clusterID)
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
	clientset, err := s.getScopedClientset(clusterID)
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

// ListCronJobs 获取 CronJob 列表（支持分页）
func (s *K8sResourceService) ListCronJobs(clusterID uint, namespace string, page, pageSize int) ([]map[string]interface{}, int64, error) {
	clientset, err := s.getScopedClientset(clusterID)
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
	clientset, err := s.getScopedClientset(clusterID)
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

	if job.Status.StartTime != nil {
		duration := time.Since(job.Status.StartTime.Time)
		result["duration"] = formatJobDuration(int(duration.Seconds()))
		if job.Status.CompletionTime != nil {
			duration = job.Status.CompletionTime.Sub(job.Status.StartTime.Time)
			result["duration"] = formatJobDuration(int(duration.Seconds()))
		}
	}

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

func (s *K8sResourceService) formatJobDetail(job *batchv1.Job) map[string]interface{} {
	result := s.formatJob(job)

	manifest, _ := json.Marshal(job)
	result["manifest"] = string(manifest)

	images := make([]string, 0)
	for _, c := range job.Spec.Template.Spec.Containers {
		images = append(images, c.Image)
	}
	result["images"] = images

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

// DeleteJob 删除 Job
func (s *K8sResourceService) DeleteJob(clusterID uint, namespace, name string) error {
	clientset, err := s.getScopedClientset(clusterID)
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

// DeleteCronJob 删除 CronJob
func (s *K8sResourceService) DeleteCronJob(clusterID uint, namespace, name string) error {
	clientset, err := s.getScopedClientset(clusterID)
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
	clientset, err := s.getScopedClientset(clusterID)
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
