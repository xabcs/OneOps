package k8s

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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

// getCurrentTimestamp 获取当前时间戳
func getCurrentTimestamp() string {
	return metav1.Now().Format("2006-01-02 15:04:05")
}
