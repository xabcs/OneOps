package k8s

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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
