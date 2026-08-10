package cmdb

import (
	"database/sql"
	"fmt"

	modelcmdb "oneops/backend3/model/cmdb"
	modelsystem "oneops/backend3/model/system"

	"gorm.io/gorm"
)

// AttributeRepository 属性数据访问层
type AttributeRepository struct {
	db *gorm.DB
}

// NewAttributeRepository 创建属性仓库
func NewAttributeRepository(db *gorm.DB) *AttributeRepository {
	return &AttributeRepository{db: db}
}

// ========== AttributeDefinition CRUD ==========

// FindAttributeDefinitions 获取属性定义列表
func (r *AttributeRepository) FindAttributeDefinitions(category string) ([]modelsystem.AttributeDefinition, error) {
	var attributes []modelsystem.AttributeDefinition
	query := r.db.Model(&modelsystem.AttributeDefinition{}).Where("status = 1")

	if category != "" {
		query = query.Where("category = ?", category)
	}

	err := query.Order("sort_order ASC, id ASC").Find(&attributes).Error
	return attributes, err
}

// FindAttributeDefinitionByID 根据ID获取属性定义
func (r *AttributeRepository) FindAttributeDefinitionByID(id uint) (*modelsystem.AttributeDefinition, error) {
	var attr modelsystem.AttributeDefinition
	err := r.db.First(&attr, id).Error
	if err != nil {
		return nil, err
	}
	return &attr, nil
}

// FindAttributeDefinitionByKey 根据Key获取属性定义
func (r *AttributeRepository) FindAttributeDefinitionByKey(key string) (*modelsystem.AttributeDefinition, error) {
	var attr modelsystem.AttributeDefinition
	err := r.db.Where("`key` = ? AND status = 1", key).First(&attr).Error
	if err != nil {
		return nil, err
	}
	return &attr, nil
}

// CountAttributeDefinitionByKey 根据Key统计数量（检查重复）
func (r *AttributeRepository) CountAttributeDefinitionByKey(key string) (int64, error) {
	var count int64
	err := r.db.Model(&modelsystem.AttributeDefinition{}).Where("`key` = ?", key).Count(&count).Error
	return count, err
}

// CountAttributeDefinitionByKeyExcludeID 根据Key统计数量（排除指定ID）
func (r *AttributeRepository) CountAttributeDefinitionByKeyExcludeID(key string, excludeID uint) (int64, error) {
	var count int64
	err := r.db.Model(&modelsystem.AttributeDefinition{}).Where("`key` = ? AND id != ?", key, excludeID).Count(&count).Error
	return count, err
}

// CreateAttributeDefinition 创建属性定义
func (r *AttributeRepository) CreateAttributeDefinition(attr *modelsystem.AttributeDefinition) error {
	return r.db.Create(attr).Error
}

// UpdateAttributeDefinition 更新属性定义
func (r *AttributeRepository) UpdateAttributeDefinition(attr *modelsystem.AttributeDefinition, updates map[string]interface{}) error {
	return r.db.Model(attr).Updates(updates).Error
}

// DeleteAttributeDefinition 删除属性定义
func (r *AttributeRepository) DeleteAttributeDefinition(id uint) error {
	return r.db.Delete(&modelsystem.AttributeDefinition{}, id).Error
}

// ========== ServerAttribute CRUD ==========

// CountServerAttributesByAttrID 统计使用指定属性的主机数量
func (r *AttributeRepository) CountServerAttributesByAttrID(attrID uint) (int64, error) {
	var count int64
	err := r.db.Model(&modelcmdb.ServerAttribute{}).Where("attribute_id = ?", attrID).Count(&count).Error
	return count, err
}

// ServerAttributeWithDef 主机属性（含定义信息）
type ServerAttributeWithDef struct {
	modelcmdb.ServerAttribute
	DefID          uint
	DefName        sql.NullString
	DefType        sql.NullString
	DefCategory    sql.NullString
	DefDescription sql.NullString
}

// FindServerAttributes 获取主机的所有属性（含定义信息）
func (r *AttributeRepository) FindServerAttributes(serverID uint) ([]modelcmdb.ServerAttribute, error) {
	rows, err := r.db.Table("cmdb_server_attributes").
		Select(`
			cmdb_server_attributes.id,
			cmdb_server_attributes.server_id,
			cmdb_server_attributes.attribute_id,
			cmdb_server_attributes.attribute_key,
			cmdb_server_attributes.attribute_value,
			cmdb_server_attributes.value_type,
			cmdb_server_attributes.category,
			cmdb_server_attributes.created_at,
			cmdb_server_attributes.updated_at,
			sys_attribute_definitions.id as def_id,
			sys_attribute_definitions.name as def_name,
			sys_attribute_definitions.type as def_type,
			sys_attribute_definitions.category as def_category,
			sys_attribute_definitions.description as def_description
		`).
		Joins("LEFT JOIN sys_attribute_definitions ON sys_attribute_definitions.id = cmdb_server_attributes.attribute_id").
		Where("cmdb_server_attributes.server_id = ?", serverID).
		Order("cmdb_server_attributes.attribute_id ASC").
		Rows()

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attributes []modelcmdb.ServerAttribute
	for rows.Next() {
		var attr modelcmdb.ServerAttribute
		var defID uint
		var defName, defType, defCategory, defDescription sql.NullString

		err := rows.Scan(
			&attr.ID,
			&attr.ServerID,
			&attr.AttributeID,
			&attr.AttributeKey,
			&attr.AttributeValue,
			&attr.ValueType,
			&attr.Category,
			&attr.CreatedAt,
			&attr.UpdatedAt,
			&defID,
			&defName,
			&defType,
			&defCategory,
			&defDescription,
		)
		if err != nil {
			return nil, err
		}

		if defID > 0 {
			attr.Definition = &modelsystem.AttributeDefinition{
				ID:          defID,
				Name:        defName.String,
				Type:        defType.String,
				Category:    defCategory.String,
				Description: defDescription.String,
			}
		}

		attributes = append(attributes, attr)
	}

	return attributes, rows.Err()
}

// SaveServerAttributes 保存主机属性（事务：先删除再创建）
func (r *AttributeRepository) SaveServerAttributes(serverID uint, attributes []modelcmdb.ServerAttribute) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("server_id = ?", serverID).Delete(&modelcmdb.ServerAttribute{}).Error; err != nil {
			return err
		}

		for _, attr := range attributes {
			attr.ServerID = serverID
			if err := tx.Create(&attr).Error; err != nil {
				return fmt.Errorf("保存属性失败: %w", err)
			}
		}
		return nil
	})
}

// FindServersByAttributes 按属性筛选主机
func (r *AttributeRepository) FindServersByAttributes(filters map[string]string, page, pageSize int) ([]modelcmdb.Server, int64, error) {
	var servers []modelcmdb.Server
	var total int64

	query := r.db.Model(&modelcmdb.Server{})

	index := 0
	for key, value := range filters {
		alias := fmt.Sprintf("sa%d", index)
		query = query.Joins(
			fmt.Sprintf("JOIN cmdb_server_attributes %s ON cmdb_servers.id = %s.server_id", alias, alias),
		).Where(fmt.Sprintf("%s.attribute_key = ? AND %s.attribute_value = ?", alias, alias), key, value)
		index++
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("Attributes.Definition").
		Order("cmdb_servers.id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&servers).Error

	return servers, total, err
}
