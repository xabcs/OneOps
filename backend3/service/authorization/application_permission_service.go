package authorization

import (
	repoauth "oneops/backend3/repository/authorization"
)

// ApplicationPermissionService 应用权限服务
type ApplicationPermissionService struct {
	repo           *repoauth.AuthorizationRepository
	adapterFactory *AdapterFactory
}

// NewApplicationPermissionService 创建服务实例
func NewApplicationPermissionService(repo *repoauth.AuthorizationRepository) *ApplicationPermissionService {
	return &ApplicationPermissionService{
		repo:           repo,
		adapterFactory: NewAdapterFactory(),
	}
}
