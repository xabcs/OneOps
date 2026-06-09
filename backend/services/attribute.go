package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"oneops/backend/models"
	"strings"
)

// AttributeService 属性服务
type AttributeService struct{}

// NewAttributeService 创建属性服务
func NewAttributeService() *AttributeService {
	return &AttributeService{}
}

// GetAttributeDefinitions 获取属性定义列表
func (s *AttributeService) GetAttributeDefinitions(category string) ([]models.AttributeDefinition, error) {
	var attributes []models.AttributeDefinition
	query := db.Model(&models.AttributeDefinition{}).Where("status = 1")

	if category != "" {
		query = query.Where("category = ?", category)
	}

	err := query.Order("sort_order ASC, id ASC").Find(&attributes).Error
	return attributes, err
}

// GetAttributeDefinitionByID 根据ID获取属性定义
func (s *AttributeService) GetAttributeDefinitionByID(id uint) (*models.AttributeDefinition, error) {
	var attr models.AttributeDefinition
	err := db.First(&attr, id).Error
	if err != nil {
		return nil, err
	}
	return &attr, nil
}

// GetAttributeDefinitionByKey 根据Key获取属性定义
func (s *AttributeService) GetAttributeDefinitionByKey(key string) (*models.AttributeDefinition, error) {
	var attr models.AttributeDefinition
	err := db.Where("`key` = ? AND status = 1", key).First(&attr).Error
	if err != nil {
		return nil, err
	}
	return &attr, nil
}

// CreateAttributeDefinition 创建属性定义
func (s *AttributeService) CreateAttributeDefinition(attr *models.AttributeDefinition) error {
	// 检查key是否已存在
	var count int64
	if err := db.Model(&models.AttributeDefinition{}).Where("`key` = ?", attr.Key).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("属性键 %s 已存在", attr.Key)
	}

	// 验证类型
	if !isValidAttributeType(attr.Type) {
		return fmt.Errorf("无效的属性类型: %s", attr.Type)
	}

	// 验证分类
	if !isValidAttributeCategory(attr.Category) {
		return fmt.Errorf("无效的属性分类: %s", attr.Category)
	}

	// 如果是select/multiselect类型，必须有选项
	if (attr.Type == models.AttrTypeSelect || attr.Type == models.AttrTypeMultiselect) && attr.Options == "" {
		return fmt.Errorf("%s 类型必须配置选项", attr.Type)
	}

	return db.Create(attr).Error
}

// UpdateAttributeDefinition 更新属性定义
func (s *AttributeService) UpdateAttributeDefinition(id uint, updates map[string]interface{}) error {
	var attr models.AttributeDefinition
	if err := db.First(&attr, id).Error; err != nil {
		return err
	}

	// 如果要更新key，检查是否冲突
	if newKey, ok := updates["key"].(string); ok && newKey != attr.Key {
		var count int64
		if err := db.Model(&models.AttributeDefinition{}).Where("`key` = ? AND id != ?", newKey, id).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return fmt.Errorf("属性键 %s 已存在", newKey)
		}
	}

	// 如果要更新类型，验证新类型
	if newType, ok := updates["type"].(string); ok && !isValidAttributeType(newType) {
		return fmt.Errorf("无效的属性类型: %s", newType)
	}

	// 如果要更新分类，验证新分类
	if newCategory, ok := updates["category"].(string); ok && !isValidAttributeCategory(newCategory) {
		return fmt.Errorf("无效的属性分类: %s", newCategory)
	}

	return db.Model(&attr).Updates(updates).Error
}

// DeleteAttributeDefinition 删除属性定义
func (s *AttributeService) DeleteAttributeDefinition(id uint) error {
	// 检查是否有主机使用此属性
	var count int64
	if err := db.Model(&models.ServerAttribute{}).Where("attribute_id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("有 %d 台主机使用此属性，无法删除", count)
	}

	return db.Delete(&models.AttributeDefinition{}, id).Error
}

// ValidateAttributeValue 验证属性值
func (s *AttributeService) ValidateAttributeValue(attrID uint, value string) error {
	attr, err := s.GetAttributeDefinitionByID(attrID)
	if err != nil {
		return err
	}

	// 必填验证
	if attr.Required && strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s 不能为空", attr.Name)
	}

	// 如果值为空，通过验证
	if strings.TrimSpace(value) == "" {
		return nil
	}

	// 类型验证
	switch attr.Type {
	case models.AttrTypeNumber:
		if _, err := parseFloat(value); err != nil {
			return fmt.Errorf("%s 必须是数字", attr.Name)
		}
	case models.AttrTypeBoolean:
		if value != "true" && value != "false" {
			return fmt.Errorf("%s 必须是 true 或 false", attr.Name)
		}
	case models.AttrTypeSelect, models.AttrTypeMultiselect:
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
func (s *AttributeService) GetServerAttributes(serverID uint) ([]models.ServerAttribute, error) {
	// 性能优化：使用JOIN代替Preload，避免N+1查询
	// 先查询属性和定义信息
	rows, err := db.Table("server_attributes").
		Select(`
			server_attributes.id,
			server_attributes.server_id,
			server_attributes.attribute_id,
			server_attributes.attribute_key,
			server_attributes.attribute_value,
			server_attributes.value_type,
			server_attributes.category,
			server_attributes.created_at,
			server_attributes.updated_at,
			attribute_definitions.id as def_id,
			attribute_definitions.name as def_name,
			attribute_definitions.type as def_type,
			attribute_definitions.category as def_category,
			attribute_definitions.description as def_description
		`).
		Joins("LEFT JOIN attribute_definitions ON attribute_definitions.id = server_attributes.attribute_id").
		Where("server_attributes.server_id = ?", serverID).
		Order("server_attributes.attribute_id ASC").
		Rows()

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attributes []models.ServerAttribute
	for rows.Next() {
		var attr models.ServerAttribute
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

		// 手动构建Definition对象
		if defID > 0 {
			attr.Definition = &models.AttributeDefinition{
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
func (s *AttributeService) SaveServerAttributes(serverID uint, attributes []models.ServerAttribute) error {
	// 开始事务
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 删除旧的属性关联
	if err := tx.Where("server_id = ?", serverID).Delete(&models.ServerAttribute{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 创建新的属性关联
	for _, attr := range attributes {
		attr.ServerID = serverID

		// 设置分类和值类型
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

// GetServersByAttributes 按属性筛选主机
func (s *AttributeService) GetServersByAttributes(filters map[string]string, page, pageSize int) ([]models.Server, int64, error) {
	var servers []models.Server
	var total int64

	// 构建查询
	query := db.Model(&models.Server{})

	// 多属性AND查询：多次JOIN
	index := 0
	for key, value := range filters {
		alias := fmt.Sprintf("sa%d", index)
		query = query.Joins(
			fmt.Sprintf("JOIN server_attributes %s ON servers.id = %s.server_id", alias, alias),
		).Where(fmt.Sprintf("%s.attribute_key = ? AND %s.attribute_value = ?", alias, alias), key, value)
		index++
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取数据
	err := query.
		Preload("Attributes.Definition").
		Order("servers.id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&servers).Error

	return servers, total, err
}

// 辅助函数

func isValidAttributeType(attrType string) bool {
	validTypes := []string{
		models.AttrTypeText,
		models.AttrTypeSelect,
		models.AttrTypeMultiselect,
		models.AttrTypeNumber,
		models.AttrTypeDate,
		models.AttrTypeBoolean,
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
		models.AttrCategorySystem,
		models.AttrCategoryLocation,
		models.AttrCategoryEnvironment,
		models.AttrCategoryHardware,
		models.AttrCategoryCustom,
	}
	for _, c := range validCategories {
		if c == category {
			return true
		}
	}
	return false
}

func parseAttributeOptions(optionsStr string) models.AttributeOptions {
	if optionsStr == "" {
		return models.AttributeOptions{}
	}

	var options models.AttributeOptions
	// JSON解析
	if err := json.Unmarshal([]byte(optionsStr), &options); err != nil {
		// 解析失败，返回空数组
		return models.AttributeOptions{}
	}
	return options
}

func getValueType(attrType string) string {
	switch attrType {
	case models.AttrTypeNumber:
		return "number"
	case models.AttrTypeBoolean:
		return "boolean"
	case models.AttrTypeDate:
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
