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
	authService       *services.AuthService
	cmdbService       *services.CMDBService
	auditService      *services.AuditService
	rbacService       *services.RBACService
	bastionService    *services.BastionService
	attributeService  *services.AttributeService
	monitoringService *services.MonitoringService
	agentService      *services.AgentService
	alertRuleService  *services.AlertRuleService

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
	})
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
}
