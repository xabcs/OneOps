package cmdb

import (
	modelcmdb "oneops/backend3/model/cmdb"

	"gorm.io/gorm"
)

// TagRepository 服务器标签数据访问层
type TagRepository struct {
	db *gorm.DB
}

// NewTagRepository 创建标签仓库
func NewTagRepository(db *gorm.DB) *TagRepository {
	return &TagRepository{db: db}
}

// DB 返回底层 *gorm.DB
func (r *TagRepository) DB() *gorm.DB {
	return r.db
}

// ========== 标签管理 ==========

// FindServerTags 获取服务器标签列表
func (r *TagRepository) FindServerTags() ([]modelcmdb.ServerTag, error) {
	var tags []modelcmdb.ServerTag
	err := r.db.Order("sort_order ASC").Find(&tags).Error
	return tags, err
}

// CreateServerTag 创建服务器标签
func (r *TagRepository) CreateServerTag(tag *modelcmdb.ServerTag) error {
	return r.db.Create(tag).Error
}

// UpdateServerTag 更新服务器标签
func (r *TagRepository) UpdateServerTag(id uint, updates map[string]interface{}) error {
	return r.db.Model(&modelcmdb.ServerTag{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteServerTag 删除服务器标签
func (r *TagRepository) DeleteServerTag(id uint) error {
	return r.db.Delete(&modelcmdb.ServerTag{}, id).Error
}

// AssignServerTag 为服务器分配标签
func (r *TagRepository) AssignServerTag(serverID, tagID uint) error {
	var count int64
	r.db.Model(&modelcmdb.ServerTagRelation{}).Where("server_id = ? AND tag_id = ?", serverID, tagID).Count(&count)
	if count > 0 {
		return nil
	}
	relation := &modelcmdb.ServerTagRelation{
		ServerID: serverID,
		TagID:    tagID,
	}
	return r.db.Create(relation).Error
}

// RemoveServerTag 移除服务器标签
func (r *TagRepository) RemoveServerTag(serverID, tagID uint) error {
	return r.db.Where("server_id = ? AND tag_id = ?", serverID, tagID).Delete(&modelcmdb.ServerTagRelation{}).Error
}
