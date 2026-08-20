package k8s

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	modelk8s "oneops/backend3/model/k8s"
	"oneops/backend3/pkg/logger"
	"oneops/backend3/pkg/utils"
	repok8s "oneops/backend3/repository/k8s"

	"go.uber.org/zap"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// K8sClientPool K8s客户端池
type K8sClientPool struct {
	clients     map[string]*k8sClientHolder // clusterID -> client holder
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	clusterRepo *repok8s.ClusterRepository
}

// k8sClientHolder K8s客户端持有者
type k8sClientHolder struct {
	clientset    *kubernetes.Clientset
	config       *rest.Config
	cluster      *modelk8s.K8sCluster
	lastUsed     time.Time
	healthStatus string // "healthy", "unhealthy", "unknown"
}

// NewK8sClientPool 创建K8s客户端池
func NewK8sClientPool(clusterRepo *repok8s.ClusterRepository) *K8sClientPool {
	ctx, cancel := context.WithCancel(context.Background())
	pool := &K8sClientPool{
		clients:     make(map[string]*k8sClientHolder),
		ctx:         ctx,
		cancel:      cancel,
		clusterRepo: clusterRepo,
	}

	// 启动健康检查和清理goroutine
	go pool.healthCheckLoop()

	return pool
}

// GetClient 获取集群的K8s客户端
func (p *K8sClientPool) GetClient(clusterID uint) (*kubernetes.Clientset, *rest.Config, error) {
	clusterKey := fmt.Sprintf("%d", clusterID)

	// 先尝试读缓存
	p.mu.RLock()
	holder, exists := p.clients[clusterKey]
	isHealthy := exists && holder != nil && holder.healthStatus == "healthy"
	p.mu.RUnlock()

	if isHealthy {
		// 更新 lastUsed 需要加写锁
		p.mu.Lock()
		if holder != nil {
			holder.lastUsed = time.Now()
		}
		p.mu.Unlock()
		return holder.clientset, holder.config, nil
	}

	// 缓存未命中或连接不健康，创建新连接（需要加锁防止并发创建）
	p.mu.Lock()
	defer p.mu.Unlock()

	// 再次检查，因为可能有其他 goroutine 已经创建了连接
	holder, exists = p.clients[clusterKey]
	if exists && holder != nil && holder.healthStatus == "healthy" {
		holder.lastUsed = time.Now()
		return holder.clientset, holder.config, nil
	}

	return p.createClientLocked(clusterID)
}

// GetScopedConfig 获取按用户作用域的连接配置（交集模型执行层）：
// userID=0（平台管理动作）或平台超管 → 共享平台凭据；
// 其他用户必须存在原生 RBAC 绑定 → impersonation 由 kube-apiserver 判定；
// 无绑定则直接拒绝（fail-closed），不再以平台身份代执行。
func (p *K8sClientPool) GetScopedConfig(clusterID uint, userID uint) (*rest.Config, error) {
	_, config, err := p.GetClient(clusterID)
	if err != nil {
		return nil, err
	}
	if userID == 0 || p.clusterRepo.IsSuperAdmin(userID) {
		return config, nil
	}

	username, groups, has, err := p.clusterRepo.FindNativeImpersonation(userID, clusterID)
	if err != nil {
		// 交集模型：查询失败视为无权限（fail-closed），不降级为平台身份
		logger.Warn("查询原生绑定身份失败，拒绝执行（fail-closed）",
			zap.Uint("cluster_id", clusterID),
			zap.Uint("user_id", userID),
			zap.Error(err))
		return nil, fmt.Errorf("该用户在此集群的原生 RBAC 绑定校验失败，已拒绝执行: %w", err)
	}
	if !has {
		return nil, fmt.Errorf("该用户在此集群未授予原生 RBAC 角色（集群详情-原生授权），交集模型下平台身份不代执行")
	}

	impersonated := rest.CopyConfig(config)
	impersonated.Impersonate = rest.ImpersonationConfig{
		UserName: username,
		Groups:   groups,
	}
	impersonated.Timeout = 30 * time.Second

	logger.Debug("使用 impersonation 访问集群",
		zap.Uint("cluster_id", clusterID),
		zap.Uint("user_id", userID),
		zap.String("impersonate_user", username),
		zap.Strings("impersonate_groups", groups))

	return impersonated, nil
}

// GetScopedClientset 获取按用户作用域的客户端（可能为 impersonated 客户端）
// impersonated 客户端构建成本低（纯内存），按请求构建、不做长期缓存
func (p *K8sClientPool) GetScopedClientset(clusterID uint, userID uint) (*kubernetes.Clientset, error) {
	config, err := p.GetScopedConfig(clusterID, userID)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

// createClient 创建新的K8s客户端连接（公共方法，带锁）
func (p *K8sClientPool) createClient(clusterID uint) (*kubernetes.Clientset, *rest.Config, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.createClientLocked(clusterID)
}

// createClientLocked 创建新的K8s客户端连接（内部方法，假设已持有锁）
func (p *K8sClientPool) createClientLocked(clusterID uint) (*kubernetes.Clientset, *rest.Config, error) {
	// 从数据库获取集群信息
	cluster, err := p.clusterRepo.FindActiveByID(clusterID)
	if err != nil {
		return nil, nil, fmt.Errorf("获取集群信息失败: %w", err)
	}

	// 解密kubeconfig
	kubeconfigContent, err := utils.DecryptString(cluster.Kubeconfig)
	if err != nil {
		return nil, nil, fmt.Errorf("解密kubeconfig失败: %w", err)
	}

	// 创建临时文件存储kubeconfig内容
	tmpFile, err := os.CreateTemp("", "kubeconfig-*")
	if err != nil {
		return nil, nil, fmt.Errorf("创建临时文件失败: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(kubeconfigContent); err != nil {
		tmpFile.Close()
		return nil, nil, fmt.Errorf("写入kubeconfig失败: %w", err)
	}
	tmpFile.Close()

	// 使用client-go加载kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", tmpFile.Name())
	if err != nil {
		return nil, nil, fmt.Errorf("加载kubeconfig失败: %w", err)
	}

	// 设置客户端超时时间
	config.Timeout = 30 * time.Second

	// 创建clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("创建K8s客户端失败: %w", err)
	}

	// 验证连接
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{Limit: 1}); err != nil {
		return nil, nil, fmt.Errorf("K8s连接验证失败: %w", err)
	}

	// 创建holder并缓存
	holder := &k8sClientHolder{
		clientset:    clientset,
		config:       config,
		cluster:      cluster,
		lastUsed:     time.Now(),
		healthStatus: "healthy",
	}

	clusterKey := fmt.Sprintf("%d", clusterID)
	p.clients[clusterKey] = holder

	logger.Info("K8s客户端连接成功",
		zap.Uint("cluster_id", clusterID),
		zap.String("cluster_name", cluster.Name))

	return clientset, config, nil
}

// RemoveClient 移除集群的客户端连接
func (p *K8sClientPool) RemoveClient(clusterID uint) {
	clusterKey := fmt.Sprintf("%d", clusterID)
	p.mu.Lock()
	delete(p.clients, clusterKey)
	p.mu.Unlock()

	logger.Info("移除K8s客户端连接", zap.Uint("cluster_id", clusterID))
}

// InvalidateCluster 失效集群缓存（当集群配置更新时调用）
func (p *K8sClientPool) InvalidateCluster(clusterID uint) {
	p.RemoveClient(clusterID)
	logger.Info("失效K8s集群缓存", zap.Uint("cluster_id", clusterID))
}

// healthCheckLoop 健康检查循环
func (p *K8sClientPool) healthCheckLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			p.performHealthCheck()
		}
	}
}

// performHealthCheck 执行健康检查
func (p *K8sClientPool) performHealthCheck() {
	p.mu.RLock()
	clusterKeys := make([]string, 0, len(p.clients))
	for key := range p.clients {
		clusterKeys = append(clusterKeys, key)
	}
	p.mu.RUnlock()

	for _, key := range clusterKeys {
		p.mu.RLock()
		holder := p.clients[key]
		p.mu.RUnlock()

		if holder == nil || holder.healthStatus != "healthy" {
			continue
		}

		// 检查是否超过30分钟未使用
		if time.Since(holder.lastUsed) > 30*time.Minute {
			p.mu.Lock()
			delete(p.clients, key)
			p.mu.Unlock()
			logger.Info("清理长时间未使用的K8s客户端",
				zap.String("cluster_id", key),
				zap.Duration("idle_time", time.Since(holder.lastUsed)))
			continue
		}

		// 执行健康检查
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if _, err := holder.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{Limit: 1}); err != nil {
			p.mu.Lock()
			holder.healthStatus = "unhealthy"
			p.mu.Unlock()
			logger.Warn("K8s客户端健康检查失败",
				zap.String("cluster_id", key),
				zap.Error(err))
		}
	}
}

// Close 关闭客户端池
func (p *K8sClientPool) Close() {
	p.cancel()
	p.mu.Lock()
	defer p.mu.Unlock()

	for key := range p.clients {
		delete(p.clients, key)
	}

	logger.Info("K8s客户端池已关闭")
}

// GetClusterInfo 获取集群信息（用于显示）
func (p *K8sClientPool) GetClusterInfo(clusterID uint) (*modelk8s.K8sCluster, error) {
	clusterKey := fmt.Sprintf("%d", clusterID)

	p.mu.RLock()
	holder, exists := p.clients[clusterKey]
	p.mu.RUnlock()

	if exists && holder != nil {
		return holder.cluster, nil
	}

	// 从数据库获取
	return p.clusterRepo.FindActiveByID(clusterID)
}

// TestConnection 测试集群连接
func (p *K8sClientPool) TestConnection(clusterID uint) error {
	clientset, _, err := p.GetClient(clusterID)
	if err != nil {
		return err
	}

	// 执行简单的健康检查
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{Limit: 1}); err != nil {
		return fmt.Errorf("连接测试失败: %w", err)
	}

	return nil
}
