package system

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
)

// AttributeService 属性服务
type AttributeService struct{}

// NewAttributeService 创建属性服务
func NewAttributeService() *AttributeService {
	return &AttributeService{}
}

// GetAttributeDefinitions 获取属性定义列表
func (s *AttributeService) GetAttributeDefinitions(category string) ([]AttributeDefinition, error) {
	var attributes []AttributeDefinition
	query := db.Model(&AttributeDefinition{}).Where("status = 1")

	if category != "" {
		query = query.Where("category = ?", category)
	}

	err := query.Order("sort_order ASC, id ASC").Find(&attributes).Error
	return attributes, err
}

// GetAttributeDefinitionByID 根据ID获取属性定义
func (s *AttributeService) GetAttributeDefinitionByID(id uint) (*AttributeDefinition, error) {
	var attr AttributeDefinition
	err := db.First(&attr, id).Error
	if err != nil {
		return nil, err
	}
	return &attr, nil
}

// GetAttributeDefinitionByKey 根据Key获取属性定义
func (s *AttributeService) GetAttributeDefinitionByKey(key string) (*AttributeDefinition, error) {
	var attr AttributeDefinition
	err := db.Where("`key` = ? AND status = 1", key).First(&attr).Error
	if err != nil {
		return nil, err
	}
	return &attr, nil
}

// CreateAttributeDefinition 创建属性定义
func (s *AttributeService) CreateAttributeDefinition(attr *AttributeDefinition) error {
	var count int64
	if err := db.Model(&AttributeDefinition{}).Where("`key` = ?", attr.Key).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("属性键 %s 已存在", attr.Key)
	}

	if !isValidAttributeType(attr.Type) {
		return fmt.Errorf("无效的属性类型: %s", attr.Type)
	}

	if !isValidAttributeCategory(attr.Category) {
		return fmt.Errorf("无效的属性分类: %s", attr.Category)
	}

	if (attr.Type == AttrTypeSelect || attr.Type == AttrTypeMultiselect) && attr.Options == "" {
		return fmt.Errorf("%s 类型必须配置选项", attr.Type)
	}

	return db.Create(attr).Error
}

// UpdateAttributeDefinition 更新属性定义
func (s *AttributeService) UpdateAttributeDefinition(id uint, updates map[string]interface{}) error {
	var attr AttributeDefinition
	if err := db.First(&attr, id).Error; err != nil {
		return err
	}

	if newKey, ok := updates["key"].(string); ok && newKey != attr.Key {
		var count int64
		if err := db.Model(&AttributeDefinition{}).Where("`key` = ? AND id != ?", newKey, id).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return fmt.Errorf("属性键 %s 已存在", newKey)
		}
	}

	if newType, ok := updates["type"].(string); ok && !isValidAttributeType(newType) {
		return fmt.Errorf("无效的属性类型: %s", newType)
	}

	if newCategory, ok := updates["category"].(string); ok && !isValidAttributeCategory(newCategory) {
		return fmt.Errorf("无效的属性分类: %s", newCategory)
	}

	return db.Model(&attr).Updates(updates).Error
}

// DeleteAttributeDefinition 删除属性定义
func (s *AttributeService) DeleteAttributeDefinition(id uint) error {
	var count int64
	if err := db.Model(&ServerAttribute{}).Where("attribute_id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("有 %d 台主机使用此属性，无法删除", count)
	}

	return db.Delete(&AttributeDefinition{}, id).Error
}

// ValidateAttributeValue 验证属性值
func (s *AttributeService) ValidateAttributeValue(attrID uint, value string) error {
	attr, err := s.GetAttributeDefinitionByID(attrID)
	if err != nil {
		return err
	}

	if attr.Required && strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s 不能为空", attr.Name)
	}

	if strings.TrimSpace(value) == "" {
		return nil
	}

	switch attr.Type {
	case AttrTypeNumber:
		if _, err := parseFloat(value); err != nil {
			return fmt.Errorf("%s 必须是数字", attr.Name)
		}
	case AttrTypeBoolean:
		if value != "true" && value != "false" {
			return fmt.Errorf("%s 必须是 true 或 false", attr.Name)
		}
	case AttrTypeSelect, AttrTypeMultiselect:
		valid := false
		options := parseAttributeOptions(attr.Options)
		for _, opt := range options {
			if opt.Value == value {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("%s 的值 '%s' 不在有效范围内", attr.Name, value)
		}
	}

	return nil
}

// GetServerAttributes 获取主机的所有属性
func (s *AttributeService) GetServerAttributes(serverID uint) ([]ServerAttribute, error) {
	rows, err := db.Table("cmdb_server_attributes").
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

	var attributes []ServerAttribute
	for rows.Next() {
		var attr ServerAttribute
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
			attr.Definition = &AttributeDefinition{
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

// SaveServerAttributes 保存主机属性
func (s *AttributeService) SaveServerAttributes(serverID uint, attributes []ServerAttribute) error {
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Where("server_id = ?", serverID).Delete(&ServerAttribute{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	for _, attr := range attributes {
		attr.ServerID = serverID

		if def, err := s.GetAttributeDefinitionByID(attr.AttributeID); err == nil {
			attr.Category = def.Category
			attr.ValueType = getValueType(def.Type)
		}

		if err := tx.Create(&attr).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("保存属性失败: %w", err)
		}
	}

	return tx.Commit().Error
}

// 辅助函数

func isValidAttributeType(attrType string) bool {
	validTypes := []string{
		AttrTypeText,
		AttrTypeSelect,
		AttrTypeMultiselect,
		AttrTypeNumber,
		AttrTypeDate,
		AttrTypeBoolean,
	}
	for _, t := range validTypes {
		if t == attrType {
			return true
		}
	}
	return false
}

func isValidAttributeCategory(category string) bool {
	validCategories := []string{
		AttrCategorySystem,
		AttrCategoryLocation,
		AttrCategoryEnvironment,
		AttrCategoryHardware,
		AttrCategoryCustom,
	}
	for _, c := range validCategories {
		if c == category {
			return true
		}
	}
	return false
}

func parseAttributeOptions(optionsStr string) AttributeOptions {
	if optionsStr == "" {
		return AttributeOptions{}
	}

	var options AttributeOptions
	if err := json.Unmarshal([]byte(optionsStr), &options); err != nil {
		return AttributeOptions{}
	}
	return options
}

func getValueType(attrType string) string {
	switch attrType {
	case AttrTypeNumber:
		return "number"
	case AttrTypeBoolean:
		return "boolean"
	case AttrTypeDate:
		return "date"
	default:
		return "string"
	}
}

func parseFloat(s string) (float64, error) {
	var result float64
	_, err := fmt.Sscanf(s, "%f", &result)
	return result, err
}
