package cmdb

import (
	"encoding/json"
	"fmt"
	"strings"

	modelcmdb "oneops/backend3/model/cmdb"
	modelsystem "oneops/backend3/model/system"
	repocmdb "oneops/backend3/repository/cmdb"
)

// AttributeService 属性服务
type AttributeService struct {
	repo *repocmdb.AttributeRepository
}

// NewAttributeService 创建属性服务
func NewAttributeService(repo *repocmdb.AttributeRepository) *AttributeService {
	return &AttributeService{repo: repo}
}

// GetAttributeDefinitions 获取属性定义列表
func (s *AttributeService) GetAttributeDefinitions(category string) ([]modelsystem.AttributeDefinition, error) {
	return s.repo.FindAttributeDefinitions(category)
}

// GetAttributeDefinitionByID 根据ID获取属性定义
func (s *AttributeService) GetAttributeDefinitionByID(id uint) (*modelsystem.AttributeDefinition, error) {
	return s.repo.FindAttributeDefinitionByID(id)
}

// GetAttributeDefinitionByKey 根据Key获取属性定义
func (s *AttributeService) GetAttributeDefinitionByKey(key string) (*modelsystem.AttributeDefinition, error) {
	return s.repo.FindAttributeDefinitionByKey(key)
}

// CreateAttributeDefinition 创建属性定义
func (s *AttributeService) CreateAttributeDefinition(attr *modelsystem.AttributeDefinition) error {
	count, err := s.repo.CountAttributeDefinitionByKey(attr.Key)
	if err != nil {
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

	if attr.Type == modelsystem.AttrTypeSelect && attr.Options == "" {
		return fmt.Errorf("%s 类型必须配置选项", attr.Type)
	}

	return s.repo.CreateAttributeDefinition(attr)
}

// UpdateAttributeDefinition 更新属性定义
func (s *AttributeService) UpdateAttributeDefinition(id uint, updates map[string]interface{}) error {
	attr, err := s.repo.FindAttributeDefinitionByID(id)
	if err != nil {
		return err
	}

	if newKey, ok := updates["key"].(string); ok && newKey != attr.Key {
		count, err := s.repo.CountAttributeDefinitionByKeyExcludeID(newKey, id)
		if err != nil {
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

	return s.repo.UpdateAttributeDefinition(attr, updates)
}

// DeleteAttributeDefinition 删除属性定义
func (s *AttributeService) DeleteAttributeDefinition(id uint) error {
	count, err := s.repo.CountServerAttributesByAttrID(id)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("有 %d 台主机使用此属性，无法删除", count)
	}

	return s.repo.DeleteAttributeDefinition(id)
}

// ValidateAttributeValue 验证属性值
func (s *AttributeService) ValidateAttributeValue(attrID uint, value string) error {
	attr, err := s.repo.FindAttributeDefinitionByID(attrID)
	if err != nil {
		return err
	}
	return s.validateAttributeValueByDef(attr, value)
}

// validateAttributeValueByDef 根据属性定义验证值
func (s *AttributeService) validateAttributeValueByDef(attr *modelsystem.AttributeDefinition, value string) error {
	if attr.Required && strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s 不能为空", attr.Name)
	}

	if strings.TrimSpace(value) == "" {
		return nil
	}

	switch attr.Type {
	case modelsystem.AttrTypeNumber:
		if _, err := parseFloat(value); err != nil {
			return fmt.Errorf("%s 必须是数字", attr.Name)
		}
	case modelsystem.AttrTypeBoolean:
		if value != "true" && value != "false" {
			return fmt.Errorf("%s 必须是 true 或 false", attr.Name)
		}
	case modelsystem.AttrTypeSelect:
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
func (s *AttributeService) GetServerAttributes(serverID uint) ([]modelcmdb.ServerAttribute, error) {
	return s.repo.FindServerAttributes(serverID)
}

// SaveServerAttributes 保存主机属性
func (s *AttributeService) SaveServerAttributes(serverID uint, attributes []modelcmdb.ServerAttribute) error {
	for i := range attributes {
		def, err := s.repo.FindAttributeDefinitionByID(attributes[i].AttributeID)
		if err != nil {
			return fmt.Errorf("属性ID %d 不存在", attributes[i].AttributeID)
		}

		// 验证属性值
		if err := s.validateAttributeValueByDef(def, attributes[i].AttributeValue); err != nil {
			return err
		}

		attributes[i].Category = def.Category
		attributes[i].ValueType = getValueType(def.Type)
	}

	return s.repo.SaveServerAttributes(serverID, attributes)
}

// GetServersByAttributes 按属性筛选主机
func (s *AttributeService) GetServersByAttributes(filters map[string]string, page, pageSize int) ([]modelcmdb.Server, int64, error) {
	return s.repo.FindServersByAttributes(filters, page, pageSize)
}

// ========== 辅助函数 ==========

func isValidAttributeType(attrType string) bool {
	validTypes := []string{
		modelsystem.AttrTypeText,
		modelsystem.AttrTypeSelect,
		modelsystem.AttrTypeNumber,
		modelsystem.AttrTypeBoolean,
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
		modelsystem.AttrCategorySystem,
		modelsystem.AttrCategoryLocation,
		modelsystem.AttrCategoryEnvironment,
		modelsystem.AttrCategoryHardware,
		modelsystem.AttrCategoryCustom,
	}
	for _, c := range validCategories {
		if c == category {
			return true
		}
	}
	return false
}

func parseAttributeOptions(optionsStr string) modelsystem.AttributeOptions {
	if optionsStr == "" {
		return modelsystem.AttributeOptions{}
	}

	var options modelsystem.AttributeOptions
	if err := json.Unmarshal([]byte(optionsStr), &options); err != nil {
		return modelsystem.AttributeOptions{}
	}
	return options
}

func getValueType(attrType string) string {
	switch attrType {
	case modelsystem.AttrTypeNumber:
		return "number"
	case modelsystem.AttrTypeBoolean:
		return "boolean"
	default:
		return "string"
	}
}

func parseFloat(s string) (float64, error) {
	var result float64
	_, err := fmt.Sscanf(s, "%f", &result)
	return result, err
}
