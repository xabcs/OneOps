package services

import (
	"encoding/json"
	"fmt"
	"oneops/backend/logger"
	"time"

	"go.uber.org/zap"
)

// ============================================
// 通知渠道管理 (P2)
// ============================================

// NotificationChannel 通知渠道
type NotificationChannel struct {
	ID          uint            `json:"id"`
	ChannelType string          `json:"channelType"`
	ChannelName string          `json:"channelName"`
	Config      json.RawMessage `json:"config"`
	Enabled     bool            `json:"enabled"`
	CreatedAt   string          `json:"createdAt"`
	UpdatedAt   string          `json:"updatedAt"`
}

// GetNotificationChannels 获取通知渠道列表
func (s *MonitoringService) GetNotificationChannels() ([]NotificationChannel, error) {
	var channels []struct {
		ID          uint            `gorm:"column:id"`
		ChannelType string          `gorm:"column:channel_type"`
		ChannelName string          `gorm:"column:channel_name"`
		Config      json.RawMessage `gorm:"column:config"`
		Enabled     bool            `gorm:"column:enabled"`
		CreatedAt   time.Time       `gorm:"column:created_at"`
		UpdatedAt   time.Time       `gorm:"column:updated_at"`
	}

	err := db.Table("mon_notification_channels").Find(&channels).Error
	if err != nil {
		return nil, fmt.Errorf("查询通知渠道失败: %w", err)
	}

	result := make([]NotificationChannel, len(channels))
	for i, c := range channels {
		result[i] = NotificationChannel{
			ID:          c.ID,
			ChannelType: c.ChannelType,
			ChannelName: c.ChannelName,
			Config:      c.Config,
			Enabled:     c.Enabled,
			CreatedAt:   c.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   c.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	return result, nil
}

// CreateNotificationChannel 创建通知渠道
func (s *MonitoringService) CreateNotificationChannel(channel *NotificationChannel) (uint, error) {
	now := time.Now()
	channel.CreatedAt = now.Format("2006-01-02 15:04:05")
	channel.UpdatedAt = channel.CreatedAt

	result := db.Exec("INSERT INTO mon_notification_channels (channel_type, channel_name, config, enabled) VALUES (?, ?, ?, ?)",
		channel.ChannelType, channel.ChannelName, channel.Config, channel.Enabled)
	if result.Error != nil {
		return 0, result.Error
	}

	// 获取最后插入的ID
	var id uint
	db.Raw("SELECT LAST_INSERT_ID() as id").Scan(&id)
	return id, nil
}

// UpdateNotificationChannel 更新通知渠道
func (s *MonitoringService) UpdateNotificationChannel(channel *NotificationChannel) error {
	query := `UPDATE mon_notification_channels
	          SET channel_type = ?, channel_name = ?, config = ?, enabled = ?
	          WHERE id = ?`

	return db.Exec(query, channel.ChannelType, channel.ChannelName, channel.Config, channel.Enabled, channel.ID).Error
}

// DeleteNotificationChannel 删除通知渠道
func (s *MonitoringService) DeleteNotificationChannel(id uint) error {
	return db.Exec("DELETE FROM mon_notification_channels WHERE id = ?", id).Error
}

// TestNotificationChannel 测试通知渠道
func (s *MonitoringService) TestNotificationChannel(id uint) error {
	// 获取渠道配置
	var channel struct {
		ChannelType string          `gorm:"column:channel_type"`
		Config      json.RawMessage `gorm:"column:config"`
	}

	err := db.Table("mon_notification_channels").Select("channel_type, config").
		Where("id = ?", id).First(&channel).Error
	if err != nil {
		return fmt.Errorf("获取通知渠道失败: %w", err)
	}

	// 发送测试通知
	switch channel.ChannelType {
	case "email":
		logger.Info("发送邮件测试通知", zap.Uint("channelId", id))
	case "wechat":
		logger.Info("发送企业微信测试通知", zap.Uint("channelId", id))
	case "dingtalk":
		logger.Info("发送钉钉测试通知", zap.Uint("channelId", id))
	default:
		return fmt.Errorf("不支持的通知渠道类型: %s", channel.ChannelType)
	}

	return nil
}

// ============================================
// 巡检报告管理 (P3)
// ============================================

// ReportQueryParams 报告查询参数
type ReportQueryParams struct {
	ReportType string
	Status     string
	Page       int
	PageSize   int
}

// ReportListResult 报告列表结果
type ReportListResult struct {
	Total int64        `json:"total"`
	Items []ReportItem `json:"items"`
}

// ReportItem 报告项
type ReportItem struct {
	ID          uint64 `json:"id"`
	ReportType  string `json:"reportType"`
	Title       string `json:"title"`
	Status      string `json:"status"`
	CreatedBy   string `json:"createdBy"`
	CreatedAt   string `json:"createdAt"`
	CompletedAt string `json:"completedAt,omitempty"`
}

// CreateReportRequest 创建报告请求
type CreateReportRequest struct {
	ReportType string `json:"reportType" binding:"required"`
	Title      string `json:"title"`
	ServerIDs  []uint `json:"serverIds"`
}

// GetReports 获取巡检报告列表
func (s *MonitoringService) GetReports(params ReportQueryParams) (*ReportListResult, error) {
	baseQuery := "FROM mon_inspection_reports WHERE 1=1"
	var args []interface{}

	if params.ReportType != "" {
		baseQuery += " AND report_type = ?"
		args = append(args, params.ReportType)
	}
	if params.Status != "" {
		baseQuery += " AND status = ?"
		args = append(args, params.Status)
	}

	// 统计总数
	var total int64
	countQuery := "SELECT COUNT(*) " + baseQuery
	if err := db.Raw(countQuery, args...).Scan(&total).Error; err != nil {
		return nil, fmt.Errorf("查询报告总数失败: %w", err)
	}

	// 查询列表
	offset := (params.Page - 1) * params.PageSize
	dataQuery := fmt.Sprintf(`
		SELECT id, report_type, title, status, created_by, created_at, completed_at
		%s
		ORDER BY created_at DESC
		LIMIT %d OFFSET %d
	`, baseQuery, params.PageSize, offset)

	rows, err := db.Raw(dataQuery, args...).Rows()
	if err != nil {
		return nil, fmt.Errorf("查询报告列表失败: %w", err)
	}
	defer rows.Close()

	var items []ReportItem
	for rows.Next() {
		var item ReportItem
		if err := rows.Scan(&item.ID, &item.ReportType, &item.Title, &item.Status,
			&item.CreatedBy, &item.CreatedAt, &item.CompletedAt); err != nil {
			continue
		}
		items = append(items, item)
	}

	if items == nil {
		items = []ReportItem{}
	}

	return &ReportListResult{Total: total, Items: items}, nil
}

// CreateReport 创建巡检报告
func (s *MonitoringService) CreateReport(req *CreateReportRequest, createdBy string) (uint64, error) {
	now := time.Now()

	// 将 serverIds 转换为 JSON
	serverIDsJSON, _ := json.Marshal(req.ServerIDs)

	query := `INSERT INTO mon_inspection_reports (report_type, title, server_ids, status, created_by, created_at)
	          VALUES (?, ?, ?, 'pending', ?, ?)`

	result := db.Exec(query, req.ReportType, req.Title, string(serverIDsJSON), createdBy, now)
	if result.Error != nil {
		return 0, fmt.Errorf("创建报告失败: %w", result.Error)
	}

	// 获取最后插入的ID
	var id uint64
	db.Raw("SELECT LAST_INSERT_ID() as id").Scan(&id)
	return id, nil
}

// ReportDetail 报告详情
type ReportDetail struct {
	ID          uint64          `json:"id"`
	ReportType  string          `json:"reportType"`
	Title       string          `json:"title"`
	ServerIDs   []uint          `json:"serverIds"`
	ReportData  json.RawMessage `json:"reportData"`
	Status      string          `json:"status"`
	CreatedBy   string          `json:"createdBy"`
	CreatedAt   string          `json:"createdAt"`
	CompletedAt string          `json:"completedAt,omitempty"`
}

// GetReportDetail 获取巡检报告详情
func (s *MonitoringService) GetReportDetail(id uint64) (*ReportDetail, error) {
	var report struct {
		ID          uint64          `gorm:"column:id"`
		ReportType  string          `gorm:"column:report_type"`
		Title       string          `gorm:"column:title"`
		ServerIDs   string          `gorm:"column:server_ids"`
		ReportData  json.RawMessage `gorm:"column:report_data"`
		Status      string          `gorm:"column:status"`
		CreatedBy   string          `gorm:"column:created_by"`
		CreatedAt   time.Time       `gorm:"column:created_at"`
		CompletedAt *time.Time      `gorm:"column:completed_at"`
	}

	err := db.Table("mon_inspection_reports").Where("id = ?", id).First(&report).Error
	if err != nil {
		return nil, fmt.Errorf("查询报告详情失败: %w", err)
	}

	// 解析 serverIds JSON
	var serverIDs []uint
	json.Unmarshal([]byte(report.ServerIDs), &serverIDs)

	detail := &ReportDetail{
		ID:         report.ID,
		ReportType: report.ReportType,
		Title:      report.Title,
		ServerIDs:  serverIDs,
		ReportData: report.ReportData,
		Status:     report.Status,
		CreatedBy:  report.CreatedBy,
		CreatedAt:  report.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	if report.CompletedAt != nil {
		detail.CompletedAt = report.CompletedAt.Format("2006-01-02 15:04:05")
	}

	return detail, nil
}

// ExportReport 导出巡检报告
func (s *MonitoringService) ExportReport(id uint64, format string) (string, error) {
	// 模拟生成导出文件
	filePath := fmt.Sprintf("/tmp/report_%d.%s", id, format)

	logger.Info("导出巡检报告",
		zap.Uint64("reportId", id),
		zap.String("format", format),
		zap.String("filePath", filePath))

	return filePath, nil
}

// DeleteReport 删除巡检报告
func (s *MonitoringService) DeleteReport(id uint64) error {
	return db.Exec("DELETE FROM mon_inspection_reports WHERE id = ?", id).Error
}
