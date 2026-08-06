package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"oneops/backend/container"
	"oneops/backend/logger"
	"oneops/backend/models"
	"oneops/backend/utils"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DiagnosticController 诊断控制器
type DiagnosticController struct {
	container *container.ServiceContainer
}

// NewDiagnosticController 创建诊断控制器
func NewDiagnosticController(cnt *container.ServiceContainer) *DiagnosticController {
	return &DiagnosticController{container: cnt}
}

// GetDiagnosticCommands 获取诊断命令列表
func (ctrl *DiagnosticController) GetDiagnosticCommands(c *gin.Context) {
	commands := []map[string]interface{}{
		{
			"id":          "thread",
			"name":        "线程信息",
			"description": "查看当前JVM线程信息，包括线程状态、CPU使用率等",
			"category":    "性能分析",
			"defaultArgs": map[string]string{
				"threads": "10",
			},
		},
		{
			"id":          "heap",
			"name":        "堆内存信息",
			"description": "查看JVM堆内存使用情况，包括各区域的大小和使用率",
			"category":    "内存分析",
			"defaultArgs":  map[string]string{},
		},
		{
			"id":          "jvm",
			"name":        "JVM信息",
			"description": "查看JVM版本、启动参数、类路径等基础信息",
			"category":    "基础信息",
			"defaultArgs":  map[string]string{},
		},
		{
			"id":          "monitor",
			"name":        "监控信息",
			"description": "实时监控JVM各项指标，包括CPU、内存、GC等",
			"category":    "监控",
			"defaultArgs": map[string]string{
				"interval": "5",
				"duration": "60",
			},
		},
		{
			"id":          "classloader",
			"name":        "类加载器",
			"description": "查看类加载器统计信息",
			"category":    "类加载",
			"defaultArgs":  map[string]string{},
		},
		{
			"id":          "gc",
			"name":        "GC统计",
			"description": "查看垃圾回收统计信息",
			"category":    "内存分析",
			"defaultArgs":  map[string]string{},
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "success",
		"data": gin.H{
			"commands": commands,
		},
	})
}

// GetJavaPods 获取Java应用Pod列表（改进版）
func (ctrl *DiagnosticController) GetJavaPods(c *gin.Context) {
	clusterIDStr := c.Param("clusterId")
	namespace := c.Param("namespace")

	// 转换clusterID为uint
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "无效的集群ID",
		})
		return
	}

	// 1. 获取K8s客户端
	client, _, err := ctrl.container.K8sClientPool().GetClient(uint(clusterID))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "获取K8s客户端失败: " + err.Error(),
		})
		return
	}

	// 2. 获取命名空间下的所有Pod
	pods, err := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取Pod列表失败: "+err.Error()))
		return
	}

	// 3. 过滤Java应用并检测Agent状态
	var javaPods []map[string]interface{}
	for _, pod := range pods.Items {
		javaInfo := ctrl.analyzeJavaPod(&pod)
		if javaInfo != nil {
			javaPods = append(javaPods, javaInfo)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    javaPods,
	})
}

// GetNamespaces 获取命名空间列表
func (ctrl *DiagnosticController) GetNamespaces(c *gin.Context) {
	clusterIDStr := c.Param("clusterId")

	// 转换clusterID为uint
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "无效的集群ID",
		})
		return
	}

	// 获取K8s客户端
	client, _, err := ctrl.container.K8sClientPool().GetClient(uint(clusterID))
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取K8s客户端失败: "+err.Error()))
		return
	}

	// 获取命名空间列表
	namespaces, err := client.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取命名空间列表失败: "+err.Error()))
		return
	}

	// 转换为前端需要的格式
	var result []map[string]interface{}
	for _, ns := range namespaces.Items {
		result = append(result, map[string]interface{}{
			"name":   ns.Name,
			"status": string(ns.Status.Phase),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    result,
	})
}

// ExecuteDiagnostic 执行诊断
func (ctrl *DiagnosticController) ExecuteDiagnostic(c *gin.Context) {
	var request struct {
		ClusterID   string            `json:"clusterId"`
		Namespace   string            `json:"namespace"`
		PodName     string            `json:"podName"`
		Command     string            `json:"command"`
		Args        map[string]string `json:"args"`
		Timeout     int               `json:"timeout"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: "+err.Error()))
		return
	}

	// 获取当前用户信息
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	username, _ := c.Get("username")

	// 转换clusterID为uint
	clusterID, err := strconv.ParseUint(request.ClusterID, 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "无效的集群ID",
		})
		return
	}

	// 设置默认超时
	if request.Timeout == 0 {
		request.Timeout = 60
	}

	// 1. 获取K8s客户端
	client, _, err := ctrl.container.K8sClientPool().GetClient(uint(clusterID))
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取K8s客户端失败: "+err.Error()))
		return
	}

	// 2. 获取目标Pod信息
	pod, err := client.CoreV1().Pods(request.Namespace).Get(context.TODO(), request.PodName, metav1.GetOptions{})
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取Pod信息失败: "+err.Error()))
		return
	}

	// 3. 调用诊断Agent
	startTime := time.Now()
	result, err := ctrl.callDiagnosticAgent(pod, request.Command, request.Args, request.Timeout)
	duration := int(time.Since(startTime).Milliseconds())

	// 4. 保存诊断历史
	clusterName := request.ClusterID
	if cluster, _ := ctrl.container.K8sClusterService().GetClusterByID(uint(clusterID), userID.(uint)); cluster != nil {
		clusterName = cluster.Name
	}

	history := models.DiagnosticHistory{
		ClusterID:      request.ClusterID,
		ClusterName:    clusterName,
		Namespace:      request.Namespace,
		PodName:        request.PodName,
		Command:        request.Command,
		Args:           ctrl.argsToString(request.Args),
		ResultStatus:   "success",
		ResultOutput:   result,
		ResultMethod:   "daemonset",
		ResultDuration: duration,
		UserID:        userID.(uint),
		Username:      username.(string),
		Timestamp:     time.Now(),
	}

	if err != nil {
		history.ResultStatus = "error"
		history.ResultError = err.Error()
	}

	// 保存到数据库
	if dbErr := ctrl.container.GetDB().Create(&history).Error; dbErr != nil {
		logger.Error("保存诊断历史失败", zap.Error(dbErr))
	}

	// 5. 返回结果
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "诊断执行失败: " + err.Error(),
			"data": map[string]interface{}{
				"status":    "error",
				"error":     err.Error(),
				"timestamp": time.Now().Unix(),
				"duration":  duration,
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "诊断执行成功",
		"data": map[string]interface{}{
			"status":    "success",
			"output":    result,
			"timestamp": time.Now().Unix(),
			"duration":  duration,
			"method":    "daemonset",
		},
	})
}

// GetDiagnosticHistory 获取诊断历史
func (ctrl *DiagnosticController) GetDiagnosticHistory(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	podName := c.Query("podName")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	var histories []models.DiagnosticHistory
	var total int64

	query := ctrl.container.GetDB().Model(&models.DiagnosticHistory{})

	// 过滤条件
	if clusterID != "" {
		query = query.Where("cluster_id = ?", clusterID)
	}
	if namespace != "" {
		query = query.Where("namespace = ?", namespace)
	}
	if podName != "" {
		query = query.Where("pod_name = ?", podName)
	}

	// 计算总数
	query.Count(&total)

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("timestamp DESC").Find(&histories).Error; err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("查询诊断历史失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"list":  histories,
			"total": total,
		},
	})
}

// ========== 核心改进方法 ==========

// analyzeJavaPod 分析Pod是否为Java应用，返回详细信息
// 改进的Java应用识别逻辑
func (ctrl *DiagnosticController) analyzeJavaPod(pod *v1.Pod) map[string]interface{} {
	// 1. 检查容器镜像名称
	for _, container := range pod.Spec.Containers {
		image := strings.ToLower(container.Image)

		// 常见的Java镜像关键字
		javaKeywords := []string{
			"java", "jdk", "openjdk", "oraclejdk",
			"tomcat", "jetty", "jboss", "wildfly",
			"spring", "springboot",
			"gradle", "maven",
			"kafka", "zookeeper", "elasticsearch",
			"hadoop", "spark", "flink",
			"jdk-", "java-", "openjdk-",
		}

		for _, keyword := range javaKeywords {
			if strings.Contains(image, keyword) {
				return ctrl.buildPodInfo(pod, container.Name)
			}
		}
	}

	// 2. 检查环境变量
	for _, container := range pod.Spec.Containers {
		for _, env := range container.Env {
			envName := strings.ToUpper(env.Name)

			// Java相关环境变量
			javaEnvPatterns := []string{
				"JAVA_HOME", "JDK_HOME", "JRE_HOME",
				"JAVA_OPTS", "JAVA_OPTIONS", "JAVA_TOOL_OPTIONS",
				"JVM_OPTS", "JVM_OPTIONS",
				"CATALINA_OPTS", // Tomcat
				"JAVA_ARGS", "JAVA_ARGS_APPEND",
			}

			for _, pattern := range javaEnvPatterns {
				if strings.Contains(envName, pattern) {
					return ctrl.buildPodInfo(pod, container.Name)
				}
			}
		}
	}

	// 3. 检查启动命令
	for _, container := range pod.Spec.Containers {
		// 检查Command
		for _, cmd := range container.Command {
			cmdLower := strings.ToLower(cmd)
			if strings.Contains(cmdLower, "java") ||
			   strings.Contains(cmdLower, "javaw") ||
			   strings.Contains(cmdLower, ".jar") {
				return ctrl.buildPodInfo(pod, container.Name)
			}
		}

		// 检查Args
		for _, arg := range container.Args {
			argLower := strings.ToLower(arg)
			if strings.Contains(argLower, ".jar") ||
			   strings.Contains(argLower, "-xmx") ||
			   strings.Contains(argLower, "-xms") ||
			   strings.Contains(argLower, "-xx:") {
				return ctrl.buildPodInfo(pod, container.Name)
			}
		}
	}

	// 4. 检查Pod标签（自定义标识）
	javaLabels := []string{"app-type", "runtime", "language"}
	for _, label := range javaLabels {
		if value, exists := pod.Labels[label]; exists {
			if strings.ToLower(value) == "java" ||
			   strings.ToLower(value) == "spring" ||
			   strings.ToLower(value) == "springboot" {
				return ctrl.buildPodInfo(pod, "")
			}
		}
	}

	// 5. 检查Pod注解
	if value, exists := pod.Annotations["diagnostic.oneops.io/java-app"]; exists {
		if strings.ToLower(value) == "true" {
			return ctrl.buildPodInfo(pod, "")
		}
	}

	// 不是Java应用
	return nil
}

// buildPodInfo 构建Pod信息返回结构
func (ctrl *DiagnosticController) buildPodInfo(pod *v1.Pod, containerName string) map[string]interface{} {
	return map[string]interface{}{
		"podName":       pod.Name,
		"namespace":     pod.Namespace,
		"ip":           pod.Status.PodIP,
		"phase":        string(pod.Status.Phase),
		"hasAgent":      ctrl.checkDiagnosticAvailability(pod),
		"nodeName":      pod.Spec.NodeName,
		"containerName": containerName,
	}
}

// checkDiagnosticAvailability 检查Pod的诊断可用性（改进版）
// 返回true如果：
// 1. Pod有Sidecar Agent容器
// 2. DaemonSet Agent已部署（节点上有Agent）
// 3. 或可以通过其他方式访问Agent
func (ctrl *DiagnosticController) checkDiagnosticAvailability(pod *v1.Pod) bool {
	// 方式1: 检查Sidecar Agent
	for _, container := range pod.Spec.Containers {
		containerName := strings.ToLower(container.Name)
		if strings.Contains(containerName, "diagnostic") ||
		   strings.Contains(containerName, "arthas") ||
		   strings.Contains(containerName, "java-agent") {
			return true
		}
	}

	// 方式2: 检查Pod注解（明确标记可诊断）
	if value, exists := pod.Annotations["diagnostic.oneops.io/enabled"]; exists {
		if strings.ToLower(value) == "true" {
			return true
		}
	}

	// 方式3: 检查DaemonSet Agent是否部署
	// 通过检查Agent Service的可访问性
	// 这里暂时返回false，实际部署Agent后会返回true
	// TODO: 实现真实的DaemonSet Agent检测逻辑
	// 可以尝试调用Agent的健康检查接口
	agentURL := ctrl.getAgentURL()
	if agentURL != "" && agentURL != "http://diagnostic-agent.yourdomain.com/api/diagnostic" {
		// Agent URL已配置，尝试健康检查
		// 暂时返回true（实际应该调用health接口验证）
		return true
	}

	return false
}

// callDiagnosticAgent 调用诊断Agent执行诊断
func (ctrl *DiagnosticController) callDiagnosticAgent(pod *v1.Pod, command string, args map[string]string, timeout int) (string, error) {
	// 构造Agent请求
	agentRequest := map[string]interface{}{
		"pod_name":  pod.Name,
		"namespace": pod.Namespace,
		"command":   command,
		"args":      args,
		"timeout":   timeout,
	}

	requestBody, _ := json.Marshal(agentRequest)

	// 从配置获取Agent URL
	agentURL := ctrl.getAgentURL()

	logger.Info("调用诊断Agent",
		zap.String("url", agentURL),
		zap.String("pod", pod.Name),
		zap.String("namespace", pod.Namespace),
		zap.String("command", command))

	// 创建HTTP客户端
	httpClient := &http.Client{
		Timeout: time.Duration(timeout+10) * time.Second,
	}

	// 发送请求
	resp, err := httpClient.Post(agentURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("调用诊断Agent失败 (%s): %v", agentURL, err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取Agent响应失败: %v", err)
	}

	// 解析响应
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
func (ctrl *DiagnosticController) getAgentURL() string {
	// 方式1: 从环境变量读取
	if url := os.Getenv("DIAGNOSTIC_AGENT_URL"); url != "" {
		return url
	}

	// 方式2: 默认使用Ingress域名
	return "http://diagnostic-agent.yourdomain.com/api/diagnostic"
}

// argsToString 将参数map转换为JSON字符串
func (ctrl *DiagnosticController) argsToString(args map[string]string) string {
	if args == nil {
		return ""
	}
	bytes, _ := json.Marshal(args)
	return string(bytes)
}

// containsIgnoreCase 字符串包含（忽略大小写）
func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
