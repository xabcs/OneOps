package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
)

// DiagnosticAgent 诊断agent结构体
type DiagnosticAgent struct {
	k8sClient *kubernetes.Clientset
	namespace string
	nodeName  string
}

// DiagnosticRequest 诊断请求
type DiagnosticRequest struct {
	PodName   string   `json:"pod_name"`
	Namespace string   `json:"namespace"`
	Container string   `json:"container"`
	Command   string   `json:"command"`   // thread, heap, dashboard, etc.
	Args      []string `json:"args"`
}

// DiagnosticResult 诊断结果
type DiagnosticResult struct {
	Status    string      `json:"status"`
	Output    string      `json:"output"`
	Error     string      `json:"error,omitempty"`
	Timestamp int64       `json:"timestamp"`
	Duration  int64       `json:"duration_ms"`
	Metadata  PodMetadata `json:"metadata,omitempty"`
}

// PodMetadata Pod元数据
type PodMetadata struct {
	PodName       string `json:"pod_name"`
	Namespace     string `json:"namespace"`
	ContainerName string `json:"container_name"`
	NodeName      string `json:"node_name"`
	PodIP         string `json:"pod_ip"`
}

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status      string   `json:"status"`
	Version     string   `json:"version"`
	NodeName    string   `json:"node_name"`
	Namespace   string   `json:"namespace"`
	PodIP       string   `json:"pod_ip"`
	Capabilities []string `json:"capabilities"`
}

const (
	AGENT_VERSION = "1.0.0"
	AGENT_PORT     = "8888"
)

func main() {
	// 1. 初始化K8s客户端
	k8sClient, err := initK8sClient()
	if err != nil {
		log.Fatalf("Failed to create k8s client: %v", err)
	}

	// 2. 获取当前pod的信息
	namespace := getEnv("NAMESPACE", "default")
	nodeName := getEnv("NODE_NAME", "")

	agent := &DiagnosticAgent{
		k8sClient: k8sClient,
		namespace: namespace,
		nodeName:  nodeName,
	}

	// 3. 设置Gin模式
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// 4. 注册路由
	r.GET("/health", agent.healthHandler)
	r.POST("/diagnostic", agent.diagnosticHandler)
	r.GET("/pods", agent.listPodsHandler)
	r.GET("/pods/:namespace/:name", agent.getPodHandler)

	// 5. 启动HTTP服务
	addr := ":" + AGENT_PORT
	log.Printf("Diagnostic Agent v%s starting on %s", AGENT_VERSION, addr)
	log.Printf("Running on node: %s, namespace: %s", nodeName, namespace)

	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// initK8sClient 初始化K8s客户端
func initK8sClient() (*kubernetes.Clientset, error) {
	// 优先使用集群内配置
	config, err := rest.InClusterConfig()
	if err != nil {
		// 如果在集群外，尝试使用kubeconfig
		kubeconfig := os.Getenv("KUBECONFIG")
		if kubeconfig == "" {
			kubeconfig = os.Getenv("HOME") + "/.kube/config"
		}

		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create k8s config: %w", err)
		}
	}

	return kubernetes.NewForConfig(config)
}

// healthHandler 健康检查
func (a *DiagnosticAgent) healthHandler(c *gin.Context) {
	podIP := getEnv("POD_IP", "unknown")

	response := HealthResponse{
		Status:      "healthy",
		Version:     AGENT_VERSION,
		NodeName:    a.nodeName,
		Namespace:   a.namespace,
		PodIP:       podIP,
		Capabilities: []string{"thread", "heap", "jvm", "dashboard", "logger"},
	}

	c.JSON(http.StatusOK, response)
}

// diagnosticHandler 诊断处理器
func (a *DiagnosticAgent) diagnosticHandler(c *gin.Context) {
	var req DiagnosticRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  fmt.Sprintf("Invalid request: %v", err),
		})
		return
	}

	startTime := time.Now()

	// 验证请求
	if req.PodName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "pod_name is required",
		})
		return
	}

	if req.Namespace == "" {
		req.Namespace = a.namespace // 默认使用当前命名空间
	}

	if req.Container == "" {
		req.Container = req.PodName // 默认容器名
	}

	log.Printf("Diagnostic request: pod=%s/%s, command=%s", req.Namespace, req.PodName, req.Command)

	// 执行诊断
	result, err := a.executeDiagnostic(&req)
	if err != nil {
		result = &DiagnosticResult{
			Status:    "error",
			Error:     err.Error(),
			Timestamp: time.Now().Unix(),
			Duration:  time.Since(startTime).Milliseconds(),
		}
		c.JSON(http.StatusInternalServerError, result)
		return
	}

	result.Duration = time.Since(startTime).Milliseconds()
	c.JSON(http.StatusOK, result)
}

// executeDiagnostic 执行诊断命令
func (a *DiagnosticAgent) executeDiagnostic(req *DiagnosticRequest) (*DiagnosticResult, error) {
	// 1. 获取目标pod信息
	pod, err := a.k8sClient.CoreV1().Pods(req.Namespace).Get(
		context.TODO(), req.PodName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod %s/%s: %w", req.Namespace, req.PodName, err)
	}

	// 2. 检查pod是否在当前节点（DaemonSet模式下只处理同节点pod）
	if pod.Spec.NodeName != a.nodeName {
		return nil, fmt.Errorf("pod %s is on different node %s (current: %s)",
			pod.Name, pod.Spec.NodeName, a.nodeName)
	}

	// 3. 构建Arthas命令
	arthasCmd := a.buildArthasCommand(req.Command, req.Args)

	// 4. 执行命令
	output, err := a.execPodCommand(pod, req.Container, arthasCmd)
	if err != nil {
		return nil, fmt.Errorf("failed to execute command: %w", err)
	}

	result := &DiagnosticResult{
		Status:    "success",
		Output:    output,
		Timestamp: time.Now().Unix(),
		Metadata: PodMetadata{
			PodName:       pod.Name,
			Namespace:     pod.Namespace,
			ContainerName: req.Container,
			NodeName:      pod.Spec.NodeName,
			PodIP:         pod.Status.PodIP,
		},
	}

	return result, nil
}

// buildArthasCommand 构建Arthas命令
func (a *DiagnosticAgent) buildArthasCommand(command string, args []string) []string {
	baseCmd := []string{
		"sh", "-c",
		// 简化版Arthas启动和执行脚本
		fmt.Sprintf(`
			# 下载Arthas（如果不存在）
			if [ ! -f /tmp/arthas-boot.jar ]; then
				curl -o /tmp/arthas-boot.jar https://arthas.aliyun.com/arthas-boot.jar
			fi

			# 执行Arthas命令
			java -jar /tmp/arthas-boot.jar --target-ip %s --telnet-port %s --http-port %s -c "%s %s"
		`,
			"127.0.0.1", // 本地连接
			"3658",      // telnet端口
			"8563",      // http端口
			command,
			strings.Join(args, " "),
		),
	}

	return baseCmd
}

// execPodCommand 在pod中执行命令
func (a *DiagnosticAgent) execPodCommand(pod *corev1.Pod, container string, command []string) (string, error) {
	// 配置exec请求
	req := a.k8sClient.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(pod.Name).
		Namespace(pod.Namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: container,
			Command:   command,
			Stdout:    true,
			Stderr:    true,
			Stdin:     false,
			TTY:       false,
		}, metav1.ParameterCodec)

	// 创建执行器
	config, err := rest.InClusterConfig()
	if err != nil {
		return "", fmt.Errorf("failed to get in-cluster config: %w", err)
	}

	executor, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return "", fmt.Errorf("failed to create executor: %w", err)
	}

	// 执行命令并获取输出
	var stdout, stderr bytes.Buffer
	err = executor.Stream(remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	})

	if err != nil {
		return "", fmt.Errorf("command execution failed: %w, stderr: %s", err, stderr.String())
	}

	return stdout.String(), nil
}

// listPodsHandler 列出当前节点的pods
func (a *DiagnosticAgent) listPodsHandler(c *gin.Context) {
	namespace := c.Query("namespace")
	if namespace == "" {
		namespace = a.namespace
	}

	// 获取所有pods
	pods, err := a.k8sClient.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to list pods: %v", err),
		})
		return
	}

	// 过滤只显示当前节点的pods
	var nodePods []corev1.Pod
	for _, pod := range pods.Items {
		if pod.Spec.NodeName == a.nodeName {
			nodePods = append(nodePods, pod)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"node_name": a.nodeName,
		"count":     len(nodePods),
		"pods":      nodePods,
	})
}

// getPodHandler 获取pod详情
func (a *DiagnosticAgent) getPodHandler(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	pod, err := a.k8sClient.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("Pod not found: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, pod)
}

// getEnv 获取环境变量，支持默认值
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}