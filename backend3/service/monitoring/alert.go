package monitoring

import (
	"fmt"
	"time"

	"oneops/backend3/pkg/database"
	"oneops/backend3/pkg/logger"

	"go.uber.org/zap"
)

// alertRule 告警规则定义
type alertRule struct {
	RuleID    string
	Level     string
	Threshold float64
	Message   string
	Value     float64
	Condition string
}

// AlertItem 告警条目
type AlertItem struct {
	ID             uint64  `json:"id"`
	ServerID       uint    `json:"serverId"`
	Hostname       string  `json:"hostname"`
	IP             string  `json:"ip"`
	RuleID         string  `json:"ruleId"`
	Level          string  `json:"level"`
	Message        string  `json:"message"`
	MetricValue    float64 `json:"metricValue"`
	Threshold      float64 `json:"threshold"`
	FirstSeen      string  `json:"firstSeen"`
	LastSeen       string  `json:"lastSeen"`
	Acknowledged   bool    `json:"acknowledged"`
	AcknowledgedBy string  `json:"acknowledgedBy,omitempty"`
	AcknowledgedAt string  `json:"acknowledgedAt,omitempty"`
	ResolvedAt     string  `json:"resolvedAt,omitempty"`
}

// AlertQueryParams 告警查询参数
type AlertQueryParams struct {
	ServerID     string
	Level        string
	Acknowledged string
	Page         int
	PageSize     int
}

// AlertListResult 告警列表结果
type AlertListResult struct {
	Total int64       `json:"total"`
	Items []AlertItem `json:"items"`
}

// AlertStats 告警统计
type AlertStats struct {
	Total          int64             `json:"total"`
	ByLevel        map[string]int64  `json:"byLevel"`
	Acknowledged   int64             `json:"acknowledged"`
	Unacknowledged int64             `json:"unacknowledged"`
	Resolved       int64             `json:"resolved"`
	Active         int64             `json:"active"`
	Trend          []AlertTrendPoint `json:"trend"`
}

// AlertTrendPoint 告警趋势点
type AlertTrendPoint struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

// checkAlertThresholds 检查指标是否触发告警规则
func (s *MonitoringService) checkAlertThresholds(serverID uint, hostname, ip string, metrics *AgentExtendedMetricsResponse) {
	alertRuleService := NewAlertRuleService()
	dbRules, err := alertRuleService.LoadActiveAlertRules()
	if err != nil {
		logger.Error("加载告警规则失败，使用默认规则", zap.Error(err))
		return
	}

	if len(dbRules) == 0 {
		logger.Debug("没有启用的告警规则，跳过告警检测")
		return
	}

	metricValues := map[string]float64{
		"cpu_usage":    metrics.Performance.CPU.UsagePercent,
		"memory_usage": metrics.Performance.Memory.UsedPercent,
		"disk_usage":   metrics.Performance.Disk.UsedPercent,
		"load1":        float64(metrics.Performance.Load.Load1),
		"load5":        float64(metrics.Performance.Load.Load5),
		"load15":       float64(metrics.Performance.Load.Load15),
	}

	now := time.Now()

	for _, dbRule := range dbRules {
		value, exists := metricValues[dbRule.Metric]
		if !exists {
			continue
		}

		triggered := false
		switch dbRule.Condition {
		case ">":
			triggered = value > dbRule.Threshold
		case "<":
			triggered = value < dbRule.Threshold
		case "==":
			triggered = value == dbRule.Threshold
		case "!=":
			triggered = value != dbRule.Threshold
		default:
			logger.Warn("未知的告警条件", zap.String("condition", dbRule.Condition))
			continue
		}

		if triggered {
			message := fmt.Sprintf("%s %s %.1f%% (阈值 %.1f%%)", dbRule.Name, dbRule.Condition, value, dbRule.Threshold)
			if dbRule.Description != "" {
				message = dbRule.Description
			}

			cache := database.NewRedisCache()
			if cache.IsAlertSuppressed(serverID, dbRule.ID) {
				getDB().Exec(`UPDATE mon_agent_alerts SET last_seen = ?, metric_value = ?, message = ?
					         WHERE server_id = ? AND rule_id = ? AND resolved_at IS NULL`,
					now, value, message, serverID, dbRule.ID)
				logger.Debug("告警已被聚合抑制，仅更新时间戳",
					zap.Uint("serverID", serverID),
					zap.String("ruleId", dbRule.ID))
				continue
			}

			var existing struct {
				ID uint `gorm:"column:id"`
			}
			err := getDB().Table("mon_agent_alerts").
				Select("id").
				Where("server_id = ? AND rule_id = ? AND resolved_at IS NULL", serverID, dbRule.ID).
				First(&existing).Error
			if err != nil {
				insertSQL := `INSERT INTO mon_agent_alerts
					(server_id, hostname, ip, rule_id, level, message, metric_value, threshold, first_seen, last_seen)
					VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
				if err := getDB().Exec(insertSQL, serverID, hostname, ip, dbRule.ID, dbRule.Level,
					message, value, dbRule.Threshold, now, now).Error; err != nil {
					logger.Warn("插入告警失败", zap.String("ruleId", dbRule.ID), zap.Error(err))
				} else {
					logger.Info("触发告警",
						zap.Uint("serverID", serverID),
						zap.String("hostname", hostname),
						zap.String("ruleId", dbRule.ID),
						zap.String("level", dbRule.Level),
						zap.Float64("value", value))

					suppressDuration := time.Duration(dbRule.Duration) * time.Second
					if suppressDuration == 0 {
						suppressDuration = 5 * time.Minute
					}
					cache.SetAlertSuppress(serverID, dbRule.ID, suppressDuration)

					alertEvent := map[string]interface{}{
						"type":      "new_alert",
						"serverID":  serverID,
						"hostname":  hostname,
						"ip":        ip,
						"ruleID":    dbRule.ID,
						"level":     dbRule.Level,
						"message":   message,
						"value":     value,
						"threshold": dbRule.Threshold,
						"timestamp": now.Format(time.RFC3339),
					}
					cache.PublishAlert(alertEvent)
				}
			} else {
				getDB().Exec(`UPDATE mon_agent_alerts SET last_seen = ?, metric_value = ?, message = ? WHERE id = ?`,
					now, value, message, existing.ID)
			}
		} else {
			result := getDB().Exec(`UPDATE mon_agent_alerts SET resolved_at = ? WHERE server_id = ? AND rule_id = ? AND resolved_at IS NULL`,
				now, serverID, dbRule.ID)
			if result.RowsAffected > 0 {
				logger.Info("告警已恢复",
					zap.Uint("serverID", serverID),
					zap.String("hostname", hostname),
					zap.String("ruleId", dbRule.ID))
			}
		}
	}
}

// getActiveAlerts 获取活跃告警列表
func (s *MonitoringService) getActiveAlerts(limit int) ([]AlertItem, error) {
	query := `
		SELECT a.id, a.server_id, a.hostname, a.ip, a.rule_id, a.level, a.message,
		       a.metric_value, a.threshold,
		       DATE_FORMAT(a.first_seen, '%Y-%m-%dT%H:%i:%s+08:00') as first_seen,
		       DATE_FORMAT(a.last_seen, '%Y-%m-%dT%H:%i:%s+08:00') as last_seen,
		       a.acknowledged, COALESCE(a.acknowledged_by, '') as acknowledged_by,
		       COALESCE(DATE_FORMAT(a.acknowledged_at, '%Y-%m-%dT%H:%i:%s+08:00'), '') as acknowledged_at,
		       COALESCE(DATE_FORMAT(a.resolved_at, '%Y-%m-%dT%H:%i:%s+08:00'), '') as resolved_at
		FROM mon_agent_alerts a
		WHERE a.resolved_at IS NULL
		ORDER BY
		    CASE a.level
		        WHEN 'critical' THEN 1
		        WHEN 'high' THEN 2
		        WHEN 'medium' THEN 3
		        WHEN 'low' THEN 4
		        ELSE 5
		    END, a.last_seen DESC
		LIMIT ?
	`
	rows, err := getDB().Raw(query, limit).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []AlertItem
	for rows.Next() {
		var a AlertItem
		if err := rows.Scan(
			&a.ID, &a.ServerID, &a.Hostname, &a.IP, &a.RuleID, &a.Level, &a.Message,
			&a.MetricValue, &a.Threshold, &a.FirstSeen, &a.LastSeen,
			&a.Acknowledged, &a.AcknowledgedBy, &a.AcknowledgedAt, &a.ResolvedAt,
		); err != nil {
			continue
		}
		alerts = append(alerts, a)
	}
	return alerts, nil
}

// GetAlerts 查询告警列表
func (s *MonitoringService) GetAlerts(params AlertQueryParams) (*AlertListResult, error) {
	baseQuery := `FROM mon_agent_alerts WHERE 1=1`
	var args []interface{}

	if params.ServerID != "" {
		baseQuery += " AND server_id = ?"
		args = append(args, params.ServerID)
	}
	if params.Level != "" {
		baseQuery += " AND level = ?"
		args = append(args, params.Level)
	}
	if params.Acknowledged == "true" {
		baseQuery += " AND acknowledged = 1"
	} else if params.Acknowledged == "false" {
		baseQuery += " AND acknowledged = 0"
	}

	var total int64
	countQuery := "SELECT COUNT(*) " + baseQuery
	if err := getDB().Raw(countQuery, args...).Scan(&total).Error; err != nil {
		return nil, fmt.Errorf("查询告警总数失败: %w", err)
	}

	offset := (params.Page - 1) * params.PageSize
	dataQuery := fmt.Sprintf(`
		SELECT id, server_id, hostname, ip, rule_id, level, message,
		       metric_value, threshold,
		       DATE_FORMAT(first_seen, '%%Y-%%m-%%dT%%H:%%i:%%s+08:00') as first_seen,
		       DATE_FORMAT(last_seen, '%%Y-%%m-%%dT%%H:%%i:%%s+08:00') as last_seen,
		       acknowledged, COALESCE(acknowledged_by, '') as acknowledged_by,
		       COALESCE(DATE_FORMAT(acknowledged_at, '%%Y-%%m-%%dT%%H:%%i:%%s+08:00'), '') as acknowledged_at,
		       COALESCE(DATE_FORMAT(resolved_at, '%%Y-%%m-%%dT%%H:%%i:%%s+08:00'), '') as resolved_at
		%s
		ORDER BY
		    CASE level
		        WHEN 'critical' THEN 1
		        WHEN 'high' THEN 2
		        WHEN 'medium' THEN 3
		        WHEN 'low' THEN 4
		        ELSE 5
		    END, last_seen DESC
		LIMIT %d OFFSET %d
	`, baseQuery, params.PageSize, offset)

	rows, err := getDB().Raw(dataQuery, args...).Rows()
	if err != nil {
		return nil, fmt.Errorf("查询告警列表失败: %w", err)
	}
	defer rows.Close()

	var items []AlertItem
	for rows.Next() {
		var a AlertItem
		if err := rows.Scan(
			&a.ID, &a.ServerID, &a.Hostname, &a.IP, &a.RuleID, &a.Level, &a.Message,
			&a.MetricValue, &a.Threshold, &a.FirstSeen, &a.LastSeen,
			&a.Acknowledged, &a.AcknowledgedBy, &a.AcknowledgedAt, &a.ResolvedAt,
		); err != nil {
			continue
		}
		items = append(items, a)
	}

	if items == nil {
		items = []AlertItem{}
	}

	return &AlertListResult{Total: total, Items: items}, nil
}

// AcknowledgeAlert 确认告警
func (s *MonitoringService) AcknowledgeAlert(id uint64, acknowledgedBy, comment string) error {
	now := time.Now()
	result := getDB().Exec(
		`UPDATE mon_agent_alerts SET acknowledged = 1, acknowledged_by = ?, acknowledged_at = ? WHERE id = ?`,
		acknowledgedBy, now, id,
	)
	if result.Error != nil {
		return fmt.Errorf("确认告警失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("告警不存在: %d", id)
	}
	return nil
}

// GetAlertStats 获取告警统计
func (s *MonitoringService) GetAlertStats() (*AlertStats, error) {
	stats := &AlertStats{
		ByLevel: map[string]int64{
			"critical": 0,
			"high":     0,
			"medium":   0,
			"low":      0,
			"info":     0,
		},
		Trend: []AlertTrendPoint{},
	}

	type levelCount struct {
		Level string `gorm:"column:level"`
		Count int64  `gorm:"column:count"`
	}
	var levelCounts []levelCount
	getDB().Raw("SELECT level, COUNT(*) as count FROM mon_agent_alerts GROUP BY level").Scan(&levelCounts)
	for _, lc := range levelCounts {
		stats.ByLevel[lc.Level] = lc.Count
		stats.Total += lc.Count
	}

	getDB().Raw("SELECT COUNT(*) FROM mon_agent_alerts WHERE acknowledged = 1").Scan(&stats.Acknowledged)
	getDB().Raw("SELECT COUNT(*) FROM mon_agent_alerts WHERE acknowledged = 0").Scan(&stats.Unacknowledged)
	getDB().Raw("SELECT COUNT(*) FROM mon_agent_alerts WHERE resolved_at IS NOT NULL").Scan(&stats.Resolved)
	getDB().Raw("SELECT COUNT(*) FROM mon_agent_alerts WHERE resolved_at IS NULL").Scan(&stats.Active)

	type trendRow struct {
		Date  string `gorm:"column:date"`
		Count int64  `gorm:"column:count"`
	}
	var trend []trendRow
	getDB().Raw(`
		SELECT DATE_FORMAT(first_seen, '%Y-%m-%d') as date, COUNT(*) as count
		FROM mon_agent_alerts
		WHERE first_seen >= DATE_SUB(NOW(), INTERVAL 7 DAY)
		GROUP BY DATE_FORMAT(first_seen, '%Y-%m-%d')
		ORDER BY date ASC
	`).Scan(&trend)

	for _, t := range trend {
		stats.Trend = append(stats.Trend, AlertTrendPoint{Date: t.Date, Count: t.Count})
	}

	return stats, nil
}
