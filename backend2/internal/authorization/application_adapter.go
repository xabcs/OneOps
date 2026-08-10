package authorization

import (
	"fmt"
)

// ApplicationAdapter 应用适配器接口
// 每种应用类型（Jenkins、Jumpserver等）都需要实现此接口
type ApplicationAdapter interface {
	// ValidateConfig 验证应用配置是否完整
	ValidateConfig(config map[string]interface{}) error

	// FetchUsers 获取用户列表
	FetchUsers(baseURL string, authConfig map[string]interface{}) ([]ApplicationUser, error)

	// FetchRoles 获取角色列表
	FetchRoles(baseURL string, authConfig map[string]interface{}) ([]ApplicationRole, error)

	// FetchGroups 获取用户组列表（可选实现）
	FetchGroups(baseURL string, authConfig map[string]interface{}) ([]ApplicationGroup, error)

	// CreateUser 在外部系统中创建用户，返回创建的用户ID
	CreateUser(baseURL string, authConfig map[string]interface{}, user *UserCreateRequest) (string, error)

	// AssignRole 为用户分配角色
	AssignRole(baseURL string, authConfig map[string]interface{}, username, roleCode string) error

	// GetConfigTemplate 获取配置模板（用于前端表单）
	GetConfigTemplate() map[string]interface{}

	// GetDisplayName 获取应用类型显示名称
	GetDisplayName() string
}

// UserCreateRequest 用户创建请求
type UserCreateRequest struct {
	Username    string
	Password    string
	FullName    string
	Email       string
	Description string
}

// AdapterFactory 适配器工厂
type AdapterFactory struct {
	adapters map[string]ApplicationAdapter
}

// NewAdapterFactory 创建适配器工厂
func NewAdapterFactory() *AdapterFactory {
	factory := &AdapterFactory{
		adapters: make(map[string]ApplicationAdapter),
	}

	// 注册内置适配器
	factory.RegisterAdapter("jenkins", &JenkinsAdapter{})
	factory.RegisterAdapter("jumpserver", &JumpserverAdapter{})
	factory.RegisterAdapter("gitlab", &GitLabAdapter{})
	factory.RegisterAdapter("generic", &GenericAdapter{})

	return factory
}

// RegisterAdapter 注册适配器
func (f *AdapterFactory) RegisterAdapter(appType string, adapter ApplicationAdapter) {
	f.adapters[appType] = adapter
}

// GetAdapter 获取指定类型的适配器
func (f *AdapterFactory) GetAdapter(appType string) (ApplicationAdapter, error) {
	adapter, ok := f.adapters[appType]
	if !ok {
		return nil, fmt.Errorf("不支持的应用类型: %s", appType)
	}
	return adapter, nil
}

// GetSupportedTypes 获取支持的应用类型列表
func (f *AdapterFactory) GetSupportedTypes() []string {
	types := make([]string, 0, len(f.adapters))
	for appType := range f.adapters {
		types = append(types, appType)
	}
	return types
}
