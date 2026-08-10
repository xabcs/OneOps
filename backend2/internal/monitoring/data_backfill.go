package monitoring

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"oneops/backend2/internal/cmdb"
	"oneops/backend2/pkg/logger"

	"go.uber.org/zap"
)

const (
	// MaxBackfillHours 最大回填时长（小时）- 防止回填过多数据
	MaxBackfillHours = 24
	// BackfillBatchSize 回填批次大小（每次请求的分钟数）
	BackfillBatchSize = 60
)

// DataBackfillService 数据回填服务
type DataBackfillService struct{}

// NewDataBackfillService 创建数据回填服务
func NewDataBackfillService() *DataBackfillService {
	return &DataBackfillService{}
}

// BackfillRequest 回填请求结构
type BackfillRequest struct {
	StartTime string `json:"startTime"` // ISO 8601 格式
	EndTime   string `json:"endTime"`   // ISO 8601 格式
}

// BackfillResponse Agent 回填响应
type BackfillResponse struct {
	Success   bool               `json:"success"`
	Message   string             `json:"message,omitempty"`
	Data      []HistoricalMetric `json:"data,omitempty"`
	StartTime string             `json:"startTime"`
	EndTime   string             `json:"endTime"`
}

// HistoricalMetric 历史指标数据
type HistoricalMetric struct {
	Timestamp string          `json:"timestamp"`
	Type      string          `json:"type"` // performance, system, hardware, etc.
	Data      json.RawMessage `json:"data"`
}

// CheckAndBackfill 检查并回填缺失数据（当 Agent 重新上线时调用）
func (s *DataBackfillService) CheckAndBackfill(server *cmdb.Server) error {
	if server.AgentStatus != "running" {
		return fmt.Errorf("Agent 未运行，无法回填数据")
	}

	// 获取最后一条指标的时间
	var lastMetric struct {
		ReportTime time.Time `gorm:"column:report_time"`
	}
	err := db.Table("mon_agent_metrics").
		Select("report_time").
		Where("server_id = ?", server.ID).
		Order("report_time DESC").
		First(&lastMetric).Error

	if err != nil {
		// 没有历史数据，不需要回填
		logger.Debug("主机无历史指标数据，跳过回填", zap.Uint("serverID", server.ID))
		return nil
	}

	// 计算数据缺失时长
	now := time.Now()
	gapDuration := now.Sub(lastMetric.ReportTime)

	// 如果缺失时长小于 5 分钟，不需要回填
	if gapDuration < 5*time.Minute {
		logger.Debug("数据缺失时长过短，无需回填",
			zap.Uint("serverID", server.ID),
			zap.Duration("gap", gapDuration))
		return nil
	}

	// 如果缺失时长超过最大回填时长，只回填最近的部分
	backfillStart := lastMetric.ReportTime
	if gapDuration.Hours() > MaxBackfillHours {
		backfillStart = now.Add(-time.Duration(MaxBackfillHours) * time.Hour)
		logger.Info("数据缺失时长超过最大回填时长，仅回填最近部分",
			zap.Uint("serverID", server.ID),
			zap.Duration("gap", gapDuration),
			zap.Duration("backfillWindow", time.Duration(MaxBackfillHours)*time.Hour))
	}

	logger.Info("开始回填主机缺失数据",
		zap.Uint("serverID", server.ID),
		zap.String("hostname", server.Hostname),
		zap.Time("from", backfillStart),
		zap.Time("to", now))

	// 执行回填
	return s.BackfillData(server.ID, backfillStart, now)
}

// BackfillData 回填指定时间段的数据
func (s *DataBackfillService) BackfillData(serverID uint, startTime, endTime time.Time) error {
	var server cmdb.Server
	if err := db.First(&server, serverID).Error; err != nil {
		return fmt.Errorf("主机不存在: %w", err)
	}

	// 分批回填（每次回填 1 小时数据）
	currentStart := startTime
	for currentStart.Before(endTime) {
		currentEnd := currentStart.Add(time.Duration(BackfillBatchSize) * time.Minute)
		if currentEnd.After(endTime) {
			currentEnd = endTime
		}

		// 请求 Agent 提供该时间段的历史数据
		metrics, err := s.fetchHistoricalMetrics(&server, currentStart, currentEnd)
		if err != nil {
			logger.Warn("获取历史数据失败",
				zap.Uint("serverID", serverID),
				zap.Time("from", currentStart),
				zap.Time("to", currentEnd),
				zap.Error(err))
			// 继续下一批次
			currentStart = currentEnd
			continue
		}

		// 存储回填数据
		insertCount := 0
		for _, metric := range metrics {
			// 检查是否已存在该时间点的数据
			var existing struct {
				ID uint `gorm:"column:id"`
			}
			err := db.Table("mon_agent_metrics").
				Select("id").
				Where("server_id = ? AND metric_type = ? AND report_time = ?",
					serverID, metric.Type, metric.Timestamp).
				First(&existing).Error

			if err == nil {
				// 数据已存在，跳过
				continue
			}

			// 解析时间戳
			reportTime, err := time.Parse(time.RFC3339, metric.Timestamp)
			if err != nil {
				logger.Warn("解析时间戳失败",
					zap.String("timestamp", metric.Timestamp),
					zap.Error(err))
				continue
			}

			// 插入数据
			insertSQL := `INSERT INTO mon_agent_metrics (server_id, metric_type, metric_data, report_time, received_at)
			              VALUES (?, ?, ?, ?, NOW())`
			if err := db.Exec(insertSQL, serverID, metric.Type, string(metric.Data), reportTime).Error; err != nil {
				logger.Warn("插入回填数据失败",
					zap.Uint("serverID", serverID),
					zap.String("type", metric.Type),
					zap.String("timestamp", metric.Timestamp),
					zap.Error(err))
			} else {
				insertCount++
			}
		}

		logger.Debug("批次回填完成",
			zap.Uint("serverID", serverID),
			zap.Time("from", currentStart),
			zap.Time("to", currentEnd),
			zap.Int("count", insertCount))

		// 移动到下一批次
		currentStart = currentEnd

		// 避免请求过快
		time.Sleep(500 * time.Millisecond)
	}

	// 使 Redis 缓存失效
	cache := NewRedisCache()
	cache.InvalidateServerCache(serverID)

	logger.Info("数据回填完成",
		zap.Uint("serverID", serverID),
		zap.String("hostname", server.Hostname))

	return nil
}

// fetchHistoricalMetrics 从 Agent 获取历史指标数据
func (s *DataBackfillService) fetchHistoricalMetrics(server *cmdb.Server, startTime, endTime time.Time) ([]HistoricalMetric, error) {
	agentPort := server.AgentPort
	if agentPort == 0 {
		agentPort = 9100
	}

	ip := server.InnerIP
	if ip == "" {
		ip = server.IP
	}

	// 构建 Agent 回填 API URL
	url := fmt.Sprintf("http://%s:%d/api/v1/backfill", ip, agentPort)

	// 构建请求体
	reqBody := BackfillRequest{
		StartTime: startTime.Format(time.RFC3339),
		EndTime:   endTime.Format(time.RFC3339),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 发送 HTTP POST 请求
	reqData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(reqData))
	if err != nil {
		return nil, fmt.Errorf("构建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求 Agent 失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Agent 返回错误状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var response BackfillResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("Agent 回填失败: %s", response.Message)
	}

	return response.Data, nil
}

// ScheduleBackfillOnAgentRecovery 在 Agent 恢复时调度回填任务
func ScheduleBackfillOnAgentRecovery(serverID uint) {
	go func() {
		// 等待一段时间确保 Agent 完全启动
		time.Sleep(30 * time.Second)

		service := NewDataBackfillService()

		var server cmdb.Server
		if err := db.First(&server, serverID).Error; err != nil {
			logger.Error("查询主机失败，无法执行回填",
				zap.Uint("serverID", serverID),
				zap.Error(err))
			return
		}

		if err := service.CheckAndBackfill(&server); err != nil {
			logger.Error("数据回填失败",
				zap.Uint("serverID", serverID),
				zap.Error(err))
		}
	}()
}
