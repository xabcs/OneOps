package monitoring

import (
	"fmt"
	"time"

	"oneops/backend3/pkg/logger"

	"go.uber.org/zap"
)

// AlertRuleService 告警规则服务
type AlertRuleService struct{}

// NewAlertRuleService 创建告警规则服务
func NewAlertRuleService() *AlertRuleService {
	return &AlertRuleService{}
}

// AlertRule 告警规则（数据库模型）
type AlertRule struct {
	ID          string    `json:"id" gorm:"column:id"`
	Name        string    `json:"name" gorm:"column:name"`
	Level       string    `json:"level" gorm:"column:level"`
	Metric      string    `json:"metric" gorm:"column:metric"`
	Condition   string    `json:"condition" gorm:"column:condition"`
	Threshold   float64   `json:"threshold" gorm:"column:threshold"`
	Duration    int       `json:"duration" gorm:"column:duration"`
	Description string    `json:"description" gorm:"column:description"`
	Enabled     bool      `json:"enabled" gorm:"column:enabled"`
	CreatedAt   time.Time `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt   time.Time `json:"updatedAt" gorm:"column:updated_at"`
}

// GetAlertRules 获取告警规则列表
func (s *AlertRuleService) GetAlertRules() ([]AlertRule, error) {
	var rules []AlertRule
	err := getDB().Table("mon_agent_alert_rules").
		Select("id, name, level, `condition`, threshold, duration, description, enabled, created_at, updated_at").
		Order("level, created_at DESC").
		Find(&rules).Error

	if err != nil {
		return nil, fmt.Errorf("查询告警规则失败: %w", err)
	}

	return rules, nil
}

// CreateAlertRule 创建告警规则
func (s *AlertRuleService) CreateAlertRule(rule *AlertRule) error {
	if rule.Name == "" {
		return fmt.Errorf("规则名称不能为空")
	}
	if rule.Level == "" {
		return fmt.Errorf("告警级别不能为空")
	}
	if rule.Metric == "" {
		return fmt.Errorf("监控指标不能为空")
	}
	if rule.Condition == "" {
		return fmt.Errorf("判断条件不能为空")
	}
	if rule.Threshold == 0 {
		return fmt.Errorf("阈值不能为空")
	}

	validLevels := map[string]bool{
		"critical": true,
		"high":     true,
		"medium":   true,
		"low":      true,
		"info":     true,
	}
	if !validLevels[rule.Level] {
		return fmt.Errorf("无效的告警级别: %s", rule.Level)
	}

	validConditions := map[string]bool{
		">":  true,
		"<":  true,
		"==": true,
		"!=": true,
	}
	if !validConditions[rule.Condition] {
		return fmt.Errorf("无效的判断条件: %s", rule.Condition)
	}

	validMetrics := map[string]bool{
		"cpu_usage":    true,
		"memory_usage": true,
		"disk_usage":   true,
		"load1":        true,
		"load5":        true,
		"load15":       true,
	}
	if !validMetrics[rule.Metric] {
		return fmt.Errorf("无效的监控指标: %s", rule.Metric)
	}

	if rule.ID == "" {
		rule.ID = fmt.Sprintf("%s_%s", rule.Metric, rule.Level)
	}

	if rule.Duration == 0 {
		rule.Duration = 0
	}
	if rule.Description == "" {
		rule.Description = fmt.Sprintf("%s %s %.1f%%", rule.Metric, rule.Condition, rule.Threshold)
	}
	rule.Enabled = true
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()

	result := getDB().Table("mon_agent_alert_rules").Create(rule)
	if result.Error != nil {
		return fmt.Errorf("创建告警规则失败: %w", result.Error)
	}

	logger.Info("创建告警规则成功",
		zap.String("id", rule.ID),
		zap.String("name", rule.Name),
		zap.String("level", rule.Level))

	return nil
}

// UpdateAlertRule 更新告警规则
func (s *AlertRuleService) UpdateAlertRule(id string, rule *AlertRule) error {
	var existing AlertRule
	err := getDB().Table("mon_agent_alert_rules").Where("id = ?", id).First(&existing).Error
	if err != nil {
		return fmt.Errorf("告警规则不存在: %s", id)
	}

	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}

	if rule.Name != "" {
		updates["name"] = rule.Name
	}
	if rule.Level != "" {
		updates["level"] = rule.Level
	}
	if rule.Metric != "" {
		updates["metric"] = rule.Metric
	}
	if rule.Condition != "" {
		updates["`condition`"] = rule.Condition
	}
	if rule.Threshold > 0 {
		updates["threshold"] = rule.Threshold
	}
	if rule.Duration >= 0 {
		updates["duration"] = rule.Duration
	}
	if rule.Description != "" {
		updates["description"] = rule.Description
	}

	result := getDB().Table("mon_agent_alert_rules").Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("更新告警规则失败: %w", result.Error)
	}

	logger.Info("更新告警规则成功", zap.String("id", id))

	return nil
}

// DeleteAlertRule 删除告警规则
func (s *AlertRuleService) DeleteAlertRule(id string) error {
	var existing AlertRule
	err := getDB().Table("mon_agent_alert_rules").Where("id = ?", id).First(&existing).Error
	if err != nil {
		return fmt.Errorf("告警规则不存在: %s", id)
	}

	result := getDB().Table("mon_agent_alert_rules").Where("id = ?", id).Delete(nil)
	if result.Error != nil {
		return fmt.Errorf("删除告警规则失败: %w", result.Error)
	}

	logger.Info("删除告警规则成功", zap.String("id", id))

	return nil
}

// UpdateAlertRuleStatus 更新告警规则状态（启用/禁用）
func (s *AlertRuleService) UpdateAlertRuleStatus(id string, enabled bool) error {
	var existing AlertRule
	err := getDB().Table("mon_agent_alert_rules").Where("id = ?", id).First(&existing).Error
	if err != nil {
		return fmt.Errorf("告警规则不存在: %s", id)
	}

	result := getDB().Table("mon_agent_alert_rules").Where("id = ?", id).Updates(map[string]interface{}{
		"enabled":    enabled,
		"updated_at": time.Now(),
	})
	if result.Error != nil {
		return fmt.Errorf("更新告警规则状态失败: %w", result.Error)
	}

	logger.Info("更新告警规则状态",
		zap.String("id", id),
		zap.Bool("enabled", enabled))

	return nil
}

// LoadActiveAlertRules 加载启用的告警规则
func (s *AlertRuleService) LoadActiveAlertRules() ([]AlertRule, error) {
	var rules []AlertRule
	err := getDB().Table("mon_agent_alert_rules").
		Where("enabled = 1").
		Find(&rules).Error

	if err != nil {
		return nil, fmt.Errorf("加载告警规则失败: %w", err)
	}

	return rules, nil
}

// InitDefaultAlertRules 初始化默认告警规则
func (s *AlertRuleService) InitDefaultAlertRules() error {
	var count int64
	getDB().Table("mon_agent_alert_rules").Count(&count)
	if count > 0 {
		logger.Info("告警规则已存在，跳过初始化")
		return nil
	}

	defaultRules := []AlertRule{
		{
			ID:          "cpu_critical",
			Name:        "CPU严重告警",
			Level:       "critical",
			Metric:      "cpu_usage",
			Condition:   ">",
			Threshold:   90.0,
			Duration:    600,
			Description: "CPU使用率持续超过90%，可能影响业务性能",
		},
		{
			ID:          "cpu_high",
			Name:        "CPU高告警",
			Level:       "high",
			Metric:      "cpu_usage",
			Condition:   ">",
			Threshold:   80.0,
			Duration:    1800,
			Description: "CPU使用率持续超过80%，需关注",
		},
		{
			ID:          "memory_critical",
			Name:        "内存严重告警",
			Level:       "critical",
			Metric:      "memory_usage",
			Condition:   ">",
			Threshold:   90.0,
			Duration:    300,
			Description: "内存使用率超过90%，可能导致OOM",
		},
		{
			ID:          "memory_high",
			Name:        "内存高告警",
			Level:       "high",
			Metric:      "memory_usage",
			Condition:   ">",
			Threshold:   80.0,
			Duration:    1800,
			Description: "内存使用率超过80%，需关注",
		},
		{
			ID:          "disk_critical",
			Name:        "磁盘严重告警",
			Level:       "critical",
			Metric:      "disk_usage",
			Condition:   ">",
			Threshold:   90.0,
			Duration:    0,
			Description: "磁盘空间不足90%，可能导致服务异常",
		},
		{
			ID:          "disk_high",
			Name:        "磁盘高告警",
			Level:       "high",
			Metric:      "disk_usage",
			Condition:   ">",
			Threshold:   80.0,
			Duration:    0,
			Description: "磁盘空间超过80%，需关注",
		},
	}

	for _, rule := range defaultRules {
		if err := s.CreateAlertRule(&rule); err != nil {
			logger.Error("初始化默认告警规则失败",
				zap.String("id", rule.ID),
				zap.Error(err))
		}
	}

	logger.Info("默认告警规则初始化完成", zap.Int("count", len(defaultRules)))
	return nil
}
