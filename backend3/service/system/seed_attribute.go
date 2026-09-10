package system

import (
	modelsystem "oneops/backend3/model/system"
	"oneops/backend3/pkg/database"
	"oneops/backend3/pkg/logger"

	"go.uber.org/zap"
)

// initAttributes 初始化属性定义数据
func (i *Initializer) initAttributes() error {
	logger.Info("开始初始化属性定义数据...")

	db := database.GetDB()

	attributes := []modelsystem.AttributeDefinition{
		{
			Name:        "业务系统",
			Key:         "business_system",
			Category:    "system",
			Type:        "select",
			Options:     `[{"value":"ecommerce","label":"电商系统"},{"value":"crm","label":"CRM系统"},{"value":"erp","label":"ERP系统"},{"value":"monitor","label":"监控系统"}]`,
			SortOrder:   1,
			Status:      1,
			Description: "主机所属的业务系统",
		},
		{
			Name:        "机房",
			Key:         "room",
			Category:    "location",
			Type:        "select",
			Options:     `[{"value":"hz","label":"杭州机房"},{"value":"bj","label":"北京机房"},{"value":"sh","label":"上海机房"},{"value":"sz","label":"深圳机房"}]`,
			SortOrder:   2,
			Status:      1,
			Description: "主机所在的机房",
		},
		{
			Name:        "机柜",
			Key:         "cabinet",
			Category:    "location",
			Type:        "text",
			SortOrder:   3,
			Status:      1,
			Description: "主机所在的机柜",
		},
		{
			Name:         "环境",
			Key:          "env",
			Category:     "environment",
			Type:         "select",
			Options:      `[{"value":"prod","label":"生产"},{"value":"test","label":"测试"},{"value":"dev","label":"开发"}]`,
			DefaultValue: "test",
			Required:     false,
			SortOrder:    4,
			Status:       1,
			Description:  "主机运行环境",
		},
		{
			Name:        "标签",
			Key:         "tags",
			Category:    "system",
			Type:        "multiselect",
			Options:     `[{"value":"important","label":"重要"},{"value":"backup","label":"备份节点"},{"value":"monitor","label":"监控节点"},{"value":"web","label":"Web服务器"},{"value":"db","label":"数据库服务器"}]`,
			SortOrder:   5,
			Status:      1,
			Description: "主机的标签分类",
		},
		{
			Name:        "所属项目",
			Key:         "project",
			Category:    "system",
			Type:        "text",
			SortOrder:   6,
			Status:      1,
			Description: "主机所属的项目",
		},
		{
			Name:        "购买日期",
			Key:         "purchase_date",
			Category:    "hardware",
			Type:        "date",
			SortOrder:   7,
			Status:      1,
			Description: "主机购买日期",
		},
		{
			Name:        "过保日期",
			Key:         "warranty_date",
			Category:    "hardware",
			Type:        "date",
			SortOrder:   8,
			Status:      1,
			Description: "主机过保日期",
		},
		{
			Name:        "责任人",
			Key:         "owner",
			Category:    "system",
			Type:        "text",
			SortOrder:   9,
			Status:      1,
			Description: "主机责任人",
		},
		{
			Name:        "联系方式",
			Key:         "contact",
			Category:    "system",
			Type:        "text",
			SortOrder:   10,
			Status:      1,
			Description: "责任人联系方式",
		},
		{
			Name:        "备注",
			Key:         "remark",
			Category:    "custom",
			Type:        "text",
			SortOrder:   11,
			Status:      1,
			Description: "主机备注信息",
		},
	}

	for _, attr := range attributes {
		if err := db.Create(&attr).Error; err != nil {
			logger.Error("创建属性定义失败",
				zap.String("name", attr.Name),
				zap.String("key", attr.Key),
				zap.Any("error", err))
			return err
		}
		logger.Info("创建属性定义",
			zap.String("name", attr.Name),
			zap.String("key", attr.Key),
			zap.Uint("id", attr.ID))
	}

	logger.Info("属性定义初始化完成", zap.Int("total", len(attributes)))
	return nil
}

// syncAttributes 同步属性定义（增量更新）
func (i *Initializer) syncAttributes() error {
	logger.Info("开始同步属性定义数据...")

	db := database.GetDB()

	// 定义预置属性（通过 key 唯一标识）
	// 仅保留不与 Server 内置字段/实体重复的扩展属性
	builtinAttributes := []struct {
		name        string
		key         string
		category    string
		attrType    string
		options     string
		defaultVal  string
		required    bool
		sortOrder   int
		description string
	}{
		{"环境", "env", "environment", "select", `[{"value":"prod","label":"生产"},{"value":"test","label":"测试"},{"value":"dev","label":"开发"}]`, "test", false, 1, "主机运行环境"},
		{"责任人", "owner", "system", "text", "", "", false, 2, "主机责任人"},
		{"联系方式", "contact", "system", "text", "", "", false, 3, "责任人联系方式"},
	}

	// 删除已废弃的属性定义（与 Server 内置字段重复或类型不再支持）
	deprecatedKeys := []string{"business_system", "room", "cabinet", "tags", "purchase_date", "warranty_date", "remark", "project"}
	// 列名 attr_key（原 key 列为 MySQL 保留字，已由 initializer 迁移改名）
	if err := db.Where("attr_key IN ?", deprecatedKeys).Delete(&modelsystem.AttributeDefinition{}).Error; err != nil {
		logger.Error("删除废弃属性定义失败", zap.Any("error", err))
	}

	addedCount := 0
	updatedCount := 0

	for _, builtinAttr := range builtinAttributes {
		var existingAttr modelsystem.AttributeDefinition
		err := db.Where(&modelsystem.AttributeDefinition{Key: builtinAttr.key}).First(&existingAttr).Error

		if err == nil {
			// 属性已存在，更新数据（保持数据同步）
			db.Model(&existingAttr).Updates(map[string]interface{}{
				"name":          builtinAttr.name,
				"category":      builtinAttr.category,
				"type":          builtinAttr.attrType,
				"options":       builtinAttr.options,
				"default_value": builtinAttr.defaultVal,
				"required":      builtinAttr.required,
				"sort_order":    builtinAttr.sortOrder,
				"status":        1,
				"description":   builtinAttr.description,
			})
			updatedCount++
			logger.Debug("更新属性定义",
				zap.String("name", builtinAttr.name),
				zap.String("key", builtinAttr.key))
		} else {
			// 属性不存在，添加新属性
			newAttr := modelsystem.AttributeDefinition{
				Name:         builtinAttr.name,
				Key:          builtinAttr.key,
				Category:     builtinAttr.category,
				Type:         builtinAttr.attrType,
				Options:      builtinAttr.options,
				DefaultValue: builtinAttr.defaultVal,
				Required:     builtinAttr.required,
				SortOrder:    builtinAttr.sortOrder,
				Status:       1,
				Description:  builtinAttr.description,
			}
			if err := db.Create(&newAttr).Error; err != nil {
				logger.Error("添加属性定义失败",
					zap.String("name", builtinAttr.name),
					zap.String("key", builtinAttr.key),
					zap.Any("error", err))
				return err
			}
			addedCount++
			logger.Info("添加属性定义",
				zap.String("name", builtinAttr.name),
				zap.String("key", builtinAttr.key),
				zap.Uint("id", newAttr.ID))
		}
	}

	logger.Info("属性定义同步完成",
		zap.Int("added", addedCount),
		zap.Int("updated", updatedCount),
		zap.Int("total", len(builtinAttributes)))

	return nil
}
