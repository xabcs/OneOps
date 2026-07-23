package main

import (
	"context"
	"fmt"
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
	Namespace    string `json:"namespace"`
	ContainerName string `json:"container_name"`
	NodeName     string `json:"node_name"`
	PodIP         string `json:"pod_ip"`
}

// ArthasResponse Arthas响应
type ArthasResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

func main() {
	// 获取环境变量
	namespace := getEnv("NAMESPACE", "default")
	nodeName := getEnv("NODE_NAME", "unknown")
	podName := getEnv("POD_NAME", os.Getenv("HOSTNAME"))
	
	log.Printf("Starting diagnostic agent server...")
	log.Printf("Namespace: %s, Node: %s, Pod: %s", namespace, nodeName, podName)

	// 初始化K8s客户端
	k8sClient, err := initK8sClient()
	if err != nil {
		log.Printf("Failed to initialize K8s client: %v", err)
		log.Printf("Will start in limited mode without K8s integration")
		k8sClient = nil
	} else {
		log.Printf("K8s client initialized successfully")
	}

	agent := &DiagnosticAgent{
		k8sClient: k8sClient,
		namespace: namespace,
		nodeName:  nodeName,
	}

	// 设置路由
	router := gin.Default()
	
	// 健康检查
	router.GET("/health", agent.healthHandler)
	
	// 执行诊断
	router.POST("/diagnostic", agent.diagnosticHandler)
	
	// 启动服务器
	port := getEnv("PORT", "8888")
	log.Printf("Diagnostic agent server starting on port %s...", port)
	
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// getEnv 获取环境变量，带默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
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
	
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"pod_ip": podIP,
		"namespace": a.namespace,
		"node_name": a.nodeName,
	})
}

// diagnosticHandler 诊断处理
func (a *DiagnosticAgent) diagnosticHandler(c *gin.Context) {
	var request DiagnosticRequest
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  fmt.Sprintf("Invalid request: %v", err),
			"timestamp": time.Now().Unix(),
		})
		return
	}

	log.Printf("Diagnostic request: pod=%s, namespace=%s, command=%s", 
		request.PodName, request.Namespace, request.Command)

	// 验证参数
	if request.PodName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error": "Pod name is required",
			"timestamp": time.Now().Unix(),
		})
		return
	}

	// 执行诊断
	startTime := time.Now()
	result := a.executeDiagnostic(&request)
	duration := time.Since(startTime).Milliseconds()

	// 补充元数据
	result.Timestamp = startTime.Unix()
	result.Duration = duration

	if result.Status == "error" {
		c.JSON(http.StatusInternalServerError, result)
	} else {
		c.JSON(http.StatusOK, result)
	}
}

// executeDiagnostic 执行诊断
func (a *DiagnosticAgent) executeDiagnostic(req *DiagnosticRequest) DiagnosticResult {
	if a.k8sClient == nil {
		return DiagnosticResult{
			Status:    "error",
			Error:     "K8s client not initialized",
			Timestamp: time.Now().Unix(),
			Duration:  0,
		}
	}

	// 获取Pod信息
	pod, err := a.k8sClient.CoreV1().Pods(req.Namespace).Get(context.TODO(), req.PodName, metav1.GetOptions{})
	if err != nil {
		return DiagnosticResult{
			Status:    "error",
			Error:     fmt.Sprintf("Failed to get pod: %v", err),
			Timestamp: time.Now().Unix(),
			Duration:  0,
		}
	}

	// 查找容器
	var containerName string
	if req.Container != "" {
		containerName = req.Container
	} else {
		// 使用第一个容器
		if len(pod.Spec.Containers) > 0 {
			containerName = pod.Spec.Containers[0].Name
		} else {
			return DiagnosticResult{
				Status:    "error",
				Error:     "No containers found in pod",
				Timestamp: time.Now().Unix(),
				Duration:  0,
			}
		}
	}

	// 根据命令类型执行不同的诊断操作
	var output string
	switch req.Command {
	case "thread":
		output = a.executeThreadDump(pod, containerName)
	case "heap":
		output = a.executeHeapDump(pod, containerName)
	case "jvm":
		output = a.executeJvmInfo(pod, containerName)
	default:
		output = a.executeGenericCommand(pod, containerName, req.Command, req.Args)
	}

	return DiagnosticResult{
		Status:   "success",
		Output:   output,
		Metadata: PodMetadata{
			PodName:       pod.Name,
			Namespace:    pod.Namespace,
			ContainerName: containerName,
			NodeName:     pod.Spec.NodeName,
			PodIP:         pod.Status.PodIP,
		},
		Timestamp: time.Now().Unix(),
		Duration: 0,
	}
}

// executeThreadDump 执行线程转储
func (a *DiagnosticAgent) executeThreadDump(pod *corev1.Pod, containerName string) string {
	// 简化版本，返回模拟数据
	return fmt.Sprintf("Thread dump for pod %s, container %s", pod.Name, containerName)
}

// executeHeapDump 执行堆转储
func (a *DiagnosticAgent) executeHeapDump(pod *corev1.Pod, containerName string) string {
	// 简化版本，返回模拟数据
	return fmt.Sprintf("Heap dump for pod %s, container %s", pod.Name, containerName)
}

// executeJvmInfo 执行JVM信息
func (a *DiagnosticAgent) executeJvmInfo(pod *corev1.Pod, containerName string) string {
	// 简化版本，返回模拟数据
	return fmt.Sprintf("JVM info for pod %s, container %s", pod.Name, containerName)
}

// executeGenericCommand 执行通用命令
func (a *DiagnosticAgent) executeGenericCommand(pod *corev1.Pod, containerName string, command string, args []string) string {
	// 简化版本，返回模拟数据
	return fmt.Sprintf("Executed command %s on pod %s, container %s", command, pod.Name, containerName)
}
