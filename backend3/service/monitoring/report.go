package monitoring

import (
	"encoding/json"
	"fmt"
	"time"

	"oneops/backend3/pkg/logger"

	"go.uber.org/zap"
)

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

	var total int64
	countQuery := "SELECT COUNT(*) " + baseQuery
	if err := getDB().Raw(countQuery, args...).Scan(&total).Error; err != nil {
		return nil, fmt.Errorf("查询报告总数失败: %w", err)
	}

	offset := (params.Page - 1) * params.PageSize
	dataQuery := fmt.Sprintf(`
		SELECT id, report_type, title, status, created_by, created_at, completed_at
		%s
		ORDER BY created_at DESC
		LIMIT %d OFFSET %d
	`, baseQuery, params.PageSize, offset)

	rows, err := getDB().Raw(dataQuery, args...).Rows()
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

	serverIDsJSON, _ := json.Marshal(req.ServerIDs)

	query := `INSERT INTO mon_inspection_reports (report_type, title, server_ids, status, created_by, created_at)
	          VALUES (?, ?, ?, 'pending', ?, ?)`

	result := getDB().Exec(query, req.ReportType, req.Title, string(serverIDsJSON), createdBy, now)
	if result.Error != nil {
		return 0, fmt.Errorf("创建报告失败: %w", result.Error)
	}

	var id uint64
	getDB().Raw("SELECT LAST_INSERT_ID() as id").Scan(&id)
	return id, nil
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

	err := getDB().Table("mon_inspection_reports").Where("id = ?", id).First(&report).Error
	if err != nil {
		return nil, fmt.Errorf("查询报告详情失败: %w", err)
	}

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
	filePath := fmt.Sprintf("/tmp/report_%d.%s", id, format)

	logger.Info("导出巡检报告",
		zap.Uint64("reportId", id),
		zap.String("format", format),
		zap.String("filePath", filePath))

	return filePath, nil
}

// DeleteReport 删除巡检报告
func (s *MonitoringService) DeleteReport(id uint64) error {
	return getDB().Exec("DELETE FROM mon_inspection_reports WHERE id = ?", id).Error
}
