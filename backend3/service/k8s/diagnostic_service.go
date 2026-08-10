package k8s

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	modelk8s "oneops/backend3/model/k8s"
	"oneops/backend3/pkg/logger"
	repok8s "oneops/backend3/repository/k8s"

	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DiagnosticService 诊断服务
type DiagnosticService struct {
	clusterSvc     *K8sClusterService
	diagnosticRepo *repok8s.DiagnosticRepository
}

// NewDiagnosticService 创建诊断服务
func NewDiagnosticService(clusterSvc *K8sClusterService, diagnosticRepo *repok8s.DiagnosticRepository) *DiagnosticService {
	return &DiagnosticService{
		clusterSvc:     clusterSvc,
		diagnosticRepo: diagnosticRepo,
	}
}

// DiagnosticExecRequest 诊断执行请求
type DiagnosticExecRequest struct {
	ClusterID    uint
	ClusterIDStr string
	Namespace    string
	PodName      string
	Command      string
	Args         map[string]string
	Timeout      int
	UserID       uint
	Username     string
}

// DiagnosticExecResult 诊断执行结果
type DiagnosticExecResult struct {
	Output   string
	Duration int
	Status   string // "success" 或 "error"
	ErrMsg   string // Status 为 "error" 时的错误信息
}

// GetDiagnosticCommands 获取诊断命令列表
func (s *DiagnosticService) GetDiagnosticCommands() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"id":          "thread",
			"name":        "线程信息",
			"description": "查看当前JVM线程信息，包括线程状态、CPU使用率等",
			"category":    "性能分析",
			"defaultArgs": map[string]string{"threads": "10"},
		},
		{
			"id":          "heap",
			"name":        "堆内存信息",
			"description": "查看JVM堆内存使用情况，包括各区域的大小和使用率",
			"category":    "内存分析",
			"defaultArgs": map[string]string{},
		},
		{
			"id":          "jvm",
			"name":        "JVM信息",
			"description": "查看JVM版本、启动参数、类路径等基础信息",
			"category":    "基础信息",
			"defaultArgs": map[string]string{},
		},
		{
			"id":          "monitor",
			"name":        "监控信息",
			"description": "实时监控JVM各项指标，包括CPU、内存、GC等",
			"category":    "监控",
			"defaultArgs": map[string]string{"interval": "5", "duration": "60"},
		},
		{
			"id":          "classloader",
			"name":        "类加载器",
			"description": "查看类加载器统计信息",
			"category":    "类加载",
			"defaultArgs": map[string]string{},
		},
		{
			"id":          "gc",
			"name":        "GC统计",
			"description": "查看垃圾回收统计信息",
			"category":    "内存分析",
			"defaultArgs": map[string]string{},
		},
	}
}

// GetJavaPods 获取Java应用Pod列表
func (s *DiagnosticService) GetJavaPods(clusterID uint, namespace string) ([]map[string]interface{}, error) {
	client, err := s.clusterSvc.GetClient(clusterID)
	if err != nil {
		return nil, fmt.Errorf("获取K8s客户端失败: %w", err)
	}

	pods, err := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取Pod列表失败: %w", err)
	}

	var javaPods []map[string]interface{}
	for i := range pods.Items {
		javaInfo := s.analyzeJavaPod(&pods.Items[i])
		if javaInfo != nil {
			javaPods = append(javaPods, javaInfo)
		}
	}

	return javaPods, nil
}

// GetNamespaces 获取命名空间列表
func (s *DiagnosticService) GetNamespaces(clusterID uint) ([]map[string]interface{}, error) {
	client, err := s.clusterSvc.GetClient(clusterID)
	if err != nil {
		return nil, fmt.Errorf("获取K8s客户端失败: %w", err)
	}

	namespaces, err := client.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取命名空间列表失败: %w", err)
	}

	var result []map[string]interface{}
	for _, ns := range namespaces.Items {
		result = append(result, map[string]interface{}{
			"name":   ns.Name,
			"status": string(ns.Status.Phase),
		})
	}

	return result, nil
}

// ExecuteDiagnostic 执行诊断
// 返回的 error 表示客户端或 Pod 获取阶段的失败（不记录历史）；
// 当诊断 Agent 执行完成（无论成功或失败），历史已记录，结果通过 *DiagnosticExecResult 返回。
func (s *DiagnosticService) ExecuteDiagnostic(req *DiagnosticExecRequest) (*DiagnosticExecResult, error) {
	if req.Timeout == 0 {
		req.Timeout = 60
	}

	client, err := s.clusterSvc.GetClient(req.ClusterID)
	if err != nil {
		return nil, fmt.Errorf("获取K8s客户端失败: %w", err)
	}

	pod, err := client.CoreV1().Pods(req.Namespace).Get(context.TODO(), req.PodName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取Pod信息失败: %w", err)
	}

	startTime := time.Now()
	result, agentErr := s.callDiagnosticAgent(pod, req.Command, req.Args, req.Timeout)
	duration := int(time.Since(startTime).Milliseconds())

	clusterName := req.ClusterIDStr
	if cluster, err := s.clusterSvc.GetClusterByID(req.ClusterID, req.UserID); err == nil && cluster != nil {
		clusterName = cluster.Name
	}

	history := modelk8s.DiagnosticHistory{
		ClusterID:      req.ClusterIDStr,
		ClusterName:    clusterName,
		Namespace:      req.Namespace,
		PodName:        req.PodName,
		Command:        req.Command,
		Args:           argsToString(req.Args),
		ResultStatus:   "success",
		ResultOutput:   result,
		ResultMethod:   "daemonset",
		ResultDuration: duration,
		UserID:         req.UserID,
		Username:       req.Username,
		Timestamp:      time.Now(),
	}

	execResult := &DiagnosticExecResult{
		Output:   result,
		Duration: duration,
		Status:   "success",
	}

	if agentErr != nil {
		history.ResultStatus = "error"
		history.ResultError = agentErr.Error()
		execResult.Status = "error"
		execResult.ErrMsg = agentErr.Error()
	}

	if dbErr := s.diagnosticRepo.Create(&history); dbErr != nil {
		logger.Error("保存诊断历史失败", zap.Error(dbErr))
	}

	return execResult, nil
}

// GetDiagnosticHistory 获取诊断历史
func (s *DiagnosticService) GetDiagnosticHistory(clusterID, namespace, podName string, page, pageSize int) ([]modelk8s.DiagnosticHistory, int64, error) {
	histories, total, err := s.diagnosticRepo.FindHistoryWithPagination(repok8s.DiagnosticHistoryQuery{
		ClusterID: clusterID,
		Namespace: namespace,
		PodName:   podName,
		Page:      page,
		PageSize:  pageSize,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("查询诊断历史失败: %w", err)
	}

	return histories, total, nil
}

// ========== 辅助方法 ==========

// analyzeJavaPod 分析Pod是否为Java应用
func (s *DiagnosticService) analyzeJavaPod(pod *v1.Pod) map[string]interface{} {
	for _, container := range pod.Spec.Containers {
		image := strings.ToLower(container.Image)

		javaKeywords := []string{
			"java", "jdk", "openjdk", "oraclejdk",
			"tomcat", "jetty", "jboss", "wildfly",
			"spring", "springboot",
			"gradle", "maven",
			"kafka", "zookeeper", "elasticsearch",
			"hadoop", "spark", "flink",
		}

		for _, keyword := range javaKeywords {
			if strings.Contains(image, keyword) {
				return s.buildPodInfo(pod, container.Name)
			}
		}
	}

	for _, container := range pod.Spec.Containers {
		for _, env := range container.Env {
			envName := strings.ToUpper(env.Name)
			javaEnvPatterns := []string{
				"JAVA_HOME", "JDK_HOME", "JRE_HOME",
				"JAVA_OPTS", "JAVA_OPTIONS", "JAVA_TOOL_OPTIONS",
				"JVM_OPTS", "JVM_OPTIONS",
				"CATALINA_OPTS",
				"JAVA_ARGS", "JAVA_ARGS_APPEND",
			}
			for _, pattern := range javaEnvPatterns {
				if strings.Contains(envName, pattern) {
					return s.buildPodInfo(pod, container.Name)
				}
			}
		}
	}

	for _, container := range pod.Spec.Containers {
		for _, cmd := range container.Command {
			cmdLower := strings.ToLower(cmd)
			if strings.Contains(cmdLower, "java") ||
				strings.Contains(cmdLower, "javaw") ||
				strings.Contains(cmdLower, ".jar") {
				return s.buildPodInfo(pod, container.Name)
			}
		}

		for _, arg := range container.Args {
			argLower := strings.ToLower(arg)
			if strings.Contains(argLower, ".jar") ||
				strings.Contains(argLower, "-xmx") ||
				strings.Contains(argLower, "-xms") ||
				strings.Contains(argLower, "-xx:") {
				return s.buildPodInfo(pod, container.Name)
			}
		}
	}

	javaLabels := []string{"app-type", "runtime", "language"}
	for _, label := range javaLabels {
		if value, exists := pod.Labels[label]; exists {
			v := strings.ToLower(value)
			if v == "java" || v == "spring" || v == "springboot" {
				return s.buildPodInfo(pod, "")
			}
		}
	}

	if value, exists := pod.Annotations["diagnostic.oneops.io/java-app"]; exists {
		if strings.ToLower(value) == "true" {
			return s.buildPodInfo(pod, "")
		}
	}

	return nil
}

// buildPodInfo 构建Pod信息返回结构
func (s *DiagnosticService) buildPodInfo(pod *v1.Pod, containerName string) map[string]interface{} {
	return map[string]interface{}{
		"podName":       pod.Name,
		"namespace":     pod.Namespace,
		"ip":            pod.Status.PodIP,
		"phase":         string(pod.Status.Phase),
		"hasAgent":      s.checkDiagnosticAvailability(pod),
		"nodeName":      pod.Spec.NodeName,
		"containerName": containerName,
	}
}

// checkDiagnosticAvailability 检查Pod的诊断可用性
func (s *DiagnosticService) checkDiagnosticAvailability(pod *v1.Pod) bool {
	for _, container := range pod.Spec.Containers {
		containerName := strings.ToLower(container.Name)
		if strings.Contains(containerName, "diagnostic") ||
			strings.Contains(containerName, "arthas") ||
			strings.Contains(containerName, "java-agent") {
			return true
		}
	}

	if value, exists := pod.Annotations["diagnostic.oneops.io/enabled"]; exists {
		if strings.ToLower(value) == "true" {
			return true
		}
	}

	return false
}

// callDiagnosticAgent 调用诊断Agent执行诊断
func (s *DiagnosticService) callDiagnosticAgent(pod *v1.Pod, command string, args map[string]string, timeout int) (string, error) {
	agentRequest := map[string]interface{}{
		"pod_name":  pod.Name,
		"namespace": pod.Namespace,
		"command":   command,
		"args":      args,
		"timeout":   timeout,
	}

	requestBody, _ := json.Marshal(agentRequest)

	agentURL := getAgentURL()

	logger.Info("调用诊断Agent",
		zap.String("url", agentURL),
		zap.String("pod", pod.Name),
		zap.String("namespace", pod.Namespace),
		zap.String("command", command))

	httpClient := &http.Client{
		Timeout: time.Duration(timeout+10) * time.Second,
	}

	resp, err := httpClient.Post(agentURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("调用诊断Agent失败 (%s): %v", agentURL, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取Agent响应失败: %v", err)
	}

	var result struct {
		Status    string `json:"status"`
		Output    string `json:"output"`
		Error     string `json:"error"`
		Timestamp int64  `json:"timestamp"`
		Duration  int64  `json:"duration"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("解析Agent响应失败: %v (响应: %s)", err, string(body))
	}

	if result.Status == "error" {
		return "", fmt.Errorf("诊断执行失败: %s", result.Error)
	}

	return result.Output, nil
}

// getAgentURL 获取诊断Agent的URL
func getAgentURL() string {
	if url := os.Getenv("DIAGNOSTIC_AGENT_URL"); url != "" {
		return url
	}
	return "http://diagnostic-agent.yourdomain.com/api/diagnostic"
}

// argsToString 将参数map转换为JSON字符串
func argsToString(args map[string]string) string {
	if args == nil {
		return ""
	}
	b, _ := json.Marshal(args)
	return string(b)
}
