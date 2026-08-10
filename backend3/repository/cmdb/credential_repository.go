package cmdb

import (
	modelcmdb "oneops/backend3/model/cmdb"

	"gorm.io/gorm"
)

// CredentialRepository SSH凭证数据访问层
type CredentialRepository struct {
	db *gorm.DB
}

// NewCredentialRepository 创建凭证仓库
func NewCredentialRepository(db *gorm.DB) *CredentialRepository {
	return &CredentialRepository{db: db}
}

// DB 返回底层 *gorm.DB
func (r *CredentialRepository) DB() *gorm.DB {
	return r.db
}

// ========== SSH凭证管理 ==========

// FindSSHCredentials 获取SSH凭证列表
func (r *CredentialRepository) FindSSHCredentials(credentialType string) ([]modelcmdb.SSHCredential, error) {
	var credentials []modelcmdb.SSHCredential
	tx := r.db.Order("sort_order ASC, id ASC")
	if credentialType == "user" || credentialType == "system" {
		tx = tx.Where("credential_type = ?", credentialType)
	}
	err := tx.Find(&credentials).Error
	return credentials, err
}

// FindSSHCredentialByID 根据ID获取SSH凭证
func (r *CredentialRepository) FindSSHCredentialByID(id uint) (*modelcmdb.SSHCredential, error) {
	var credential modelcmdb.SSHCredential
	err := r.db.First(&credential, id).Error
	if err != nil {
		return nil, err
	}
	return &credential, nil
}

// CreateSSHCredential 创建SSH凭证
func (r *CredentialRepository) CreateSSHCredential(credential *modelcmdb.SSHCredential) error {
	return r.db.Create(credential).Error
}

// UpdateSSHCredential 更新SSH凭证
func (r *CredentialRepository) UpdateSSHCredential(id uint, updates map[string]interface{}) error {
	return r.db.Model(&modelcmdb.SSHCredential{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteSSHCredential 删除SSH凭证
func (r *CredentialRepository) DeleteSSHCredential(id uint) error {
	return r.db.Delete(&modelcmdb.SSHCredential{}, id).Error
}
