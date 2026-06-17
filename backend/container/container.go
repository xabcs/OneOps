package container

import (
	"sync"

	"gorm.io/gorm"
	"oneops/backend/services"
)

// ServiceContainer 服务容器
// 用于统一管理所有服务的创建和生命周期
type ServiceContainer struct {
	// 数据库连接
	db *gorm.DB

	// 服务实例（懒加载）
	authService        *services.AuthService
	cmdbService        *services.CMDBService
	auditService       *services.AuditService
	rbacService        *services.RBACService
	bastionService     *services.BastionService
	attributeService   *services.AttributeService
	monitoringService  *services.MonitoringService
	agentService       *services.AgentService
	alertRuleService   *services.AlertRuleService
	k8sClientPool      *services.K8sClientPool
	k8sIdentityService *services.K8sIdentityService
	k8sClusterService  *services.K8sClusterService
	k8sResourceService *services.K8sResourceService

	// 互斥锁，确保线程安全
	mu sync.RWMutex
}

// 全局容器实例
var (
	globalContainer *ServiceContainer
	once            sync.Once
)

// Initialize 初始化全局容器
func Initialize(db *gorm.DB) {
	once.Do(func() {
		globalContainer = &ServiceContainer{
			db: db,
		}

		// 预初始化常用服务，避免运行时懒加载导致的锁竞争
		globalContainer.preloadCommonServices()
	})
}

// preloadCommonServices 预加载常用服务
func (c *ServiceContainer) preloadCommonServices() {
	// 预初始化认证相关服务
	c.authService = services.NewAuthService()
	c.auditService = services.NewAuditService()
	c.rbacService = services.NewRBACService()

	// 预初始化K8s相关服务
	c.k8sClientPool = services.NewK8sClientPool()
	c.k8sIdentityService = services.NewK8sIdentityService(c.k8sClientPool)
	c.k8sClusterService = services.NewK8sClusterService(c.k8sClientPool, c.k8sIdentityService)
	c.k8sResourceService = services.NewK8sResourceService(c.k8sClientPool)

	// 预初始化其他常用服务
	c.cmdbService = services.NewCMDBService()
	c.bastionService = services.NewBastionService()
	c.attributeService = services.NewAttributeService()
	c.monitoringService = services.NewMonitoringService()
	c.agentService = services.NewAgentService()
	c.alertRuleService = services.NewAlertRuleService()
}

// GetContainer 获取全局容器实例
func GetContainer() *ServiceContainer {
	if globalContainer == nil {
		panic("ServiceContainer 未初始化，请先调用 Initialize()")
	}
	return globalContainer
}

// ========== 服务获取方法（懒加载，线程安全）==========

// AuthService 获取认证服务
func (c *ServiceContainer) AuthService() *services.AuthService {
	c.mu.RLock()
	if c.authService != nil {
		c.mu.RUnlock()
		return c.authService
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()
	// 双重检查
	if c.authService != nil {
		return c.authService
	}

	c.authService = services.NewAuthService()
	return c.authService
}

// CMDBService 获取CMDB服务
func (c *ServiceContainer) CMDBService() *services.CMDBService {
	c.mu.RLock()
	if c.cmdbService != nil {
		c.mu.RUnlock()
		return c.cmdbService
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cmdbService != nil {
		return c.cmdbService
	}

	c.cmdbService = services.NewCMDBService()
	return c.cmdbService
}

// AuditService 获取审计服务
func (c *ServiceContainer) AuditService() *services.AuditService {
	c.mu.RLock()
	if c.auditService != nil {
		c.mu.RUnlock()
		return c.auditService
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.auditService != nil {
		return c.auditService
	}

	c.auditService = services.NewAuditService()
	return c.auditService
}

// RBACService 获取RBAC服务
func (c *ServiceContainer) RBACService() *services.RBACService {
	c.mu.RLock()
	if c.rbacService != nil {
		c.mu.RUnlock()
		return c.rbacService
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.rbacService != nil {
		return c.rbacService
	}

	c.rbacService = services.NewRBACService()
	return c.rbacService
}

// BastionService 获取堡垒机服务
func (c *ServiceContainer) BastionService() *services.BastionService {
	c.mu.RLock()
	if c.bastionService != nil {
		c.mu.RUnlock()
		return c.bastionService
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.bastionService != nil {
		return c.bastionService
	}

	c.bastionService = services.NewBastionService()
	return c.bastionService
}

// AttributeService 获取属性服务
func (c *ServiceContainer) AttributeService() *services.AttributeService {
	c.mu.RLock()
	if c.attributeService != nil {
		c.mu.RUnlock()
		return c.attributeService
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.attributeService != nil {
		return c.attributeService
	}

	c.attributeService = services.NewAttributeService()
	return c.attributeService
}

// MonitoringService 获取监控服务
func (c *ServiceContainer) MonitoringService() *services.MonitoringService {
	c.mu.RLock()
	if c.monitoringService != nil {
		c.mu.RUnlock()
		return c.monitoringService
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.monitoringService != nil {
		return c.monitoringService
	}

	c.monitoringService = services.NewMonitoringService()
	return c.monitoringService
}

// AgentService 获取Agent服务
func (c *ServiceContainer) AgentService() *services.AgentService {
	c.mu.RLock()
	if c.agentService != nil {
		c.mu.RUnlock()
		return c.agentService
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.agentService != nil {
		return c.agentService
	}

	c.agentService = services.NewAgentService()
	return c.agentService
}

// AlertRuleService 获取告警规则服务
func (c *ServiceContainer) AlertRuleService() *services.AlertRuleService {
	c.mu.RLock()
	if c.alertRuleService != nil {
		c.mu.RUnlock()
		return c.alertRuleService
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.alertRuleService != nil {
		return c.alertRuleService
	}

	c.alertRuleService = services.NewAlertRuleService()
	return c.alertRuleService
}

// K8sClientPool 获取K8s客户端池
func (c *ServiceContainer) K8sClientPool() *services.K8sClientPool {
	c.mu.RLock()
	if c.k8sClientPool != nil {
		c.mu.RUnlock()
		return c.k8sClientPool
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.k8sClientPool != nil {
		return c.k8sClientPool
	}

	c.k8sClientPool = services.NewK8sClientPool()
	return c.k8sClientPool
}

// K8sIdentityService 获取K8s用户身份映射服务
func (c *ServiceContainer) K8sIdentityService() *services.K8sIdentityService {
	c.mu.RLock()
	if c.k8sIdentityService != nil {
		c.mu.RUnlock()
		return c.k8sIdentityService
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.k8sIdentityService != nil {
		return c.k8sIdentityService
	}

	// 直接访问字段初始化K8sClientPool，避免在持锁时调用方法
	if c.k8sClientPool == nil {
		c.k8sClientPool = services.NewK8sClientPool()
	}

	c.k8sIdentityService = services.NewK8sIdentityService(c.k8sClientPool)
	return c.k8sIdentityService
}

// K8sClusterService 获取K8s集群管理服务
func (c *ServiceContainer) K8sClusterService() *services.K8sClusterService {
	c.mu.RLock()
	if c.k8sClusterService != nil {
		c.mu.RUnlock()
		return c.k8sClusterService
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.k8sClusterService != nil {
		return c.k8sClusterService
	}

	// 确保依赖服务已初始化（直接访问字段，避免死锁）
	if c.k8sClientPool == nil {
		c.k8sClientPool = services.NewK8sClientPool()
	}
	if c.k8sIdentityService == nil {
		c.k8sIdentityService = services.NewK8sIdentityService(c.k8sClientPool)
	}

	c.k8sClusterService = services.NewK8sClusterService(c.k8sClientPool, c.k8sIdentityService)
	return c.k8sClusterService
}

	// K8sResourceService 获取K8s资源管理服务
	func (c *ServiceContainer) K8sResourceService() *services.K8sResourceService {
		c.mu.RLock()
		if c.k8sResourceService != nil {
			c.mu.RUnlock()
			return c.k8sResourceService
		}
		c.mu.RUnlock()

		c.mu.Lock()
		defer c.mu.Unlock()
		if c.k8sResourceService != nil {
			return c.k8sResourceService
		}

		// 直接访问字段初始化K8sClientPool，避免在持锁时调用方法
		if c.k8sClientPool == nil {
			c.k8sClientPool = services.NewK8sClientPool()
		}

		c.k8sResourceService = services.NewK8sResourceService(c.k8sClientPool)
		return c.k8sResourceService
	}

// GetDB 获取数据库连接
func (c *ServiceContainer) GetDB() *gorm.DB {
	return c.db
}

// Reset 重置容器（主要用于测试）
func (c *ServiceContainer) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.authService = nil
	c.cmdbService = nil
	c.auditService = nil
	c.rbacService = nil
	c.bastionService = nil
	c.attributeService = nil
	c.monitoringService = nil
	c.agentService = nil
	c.alertRuleService = nil
	c.k8sClientPool = nil
	c.k8sIdentityService = nil
	c.k8sClusterService = nil
	c.k8sResourceService = nil
}
