package cmdb

import (
	repocmdb "oneops/backend3/repository/cmdb"
)

// CMDBService CMDB服务
type CMDBService struct {
	repo     *repocmdb.ServerRepository
	tagRepo  *repocmdb.TagRepository
	credRepo *repocmdb.CredentialRepository
}

// NewCMDBService 创建CMDB服务
func NewCMDBService(repo *repocmdb.ServerRepository) *CMDBService {
	return &CMDBService{
		repo:     repo,
		tagRepo:  repocmdb.NewTagRepository(repo.DB()),
		credRepo: repocmdb.NewCredentialRepository(repo.DB()),
	}
}
