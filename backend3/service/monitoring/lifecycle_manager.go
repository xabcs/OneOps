package monitoring

import (
	"time"

	"oneops/backend3/pkg/logger"

	"go.uber.org/zap"
)

// DataLifecycleManager 数据生命周期管理器
type DataLifecycleManager struct {
	archiveTicker *time.Ticker
	cleanupTicker *time.Ticker
	stopChan      chan bool
}

// NewDataLifecycleManager 创建数据生命周期管理器
func NewDataLifecycleManager() *DataLifecycleManager {
	return &DataLifecycleManager{
		stopChan: make(chan bool),
	}
}

// Start 启动数据生命周期管理
func (m *DataLifecycleManager) Start() {
	logger.Info("启动数据生命周期管理器")

	go func() {
		m.ArchiveOldMetrics()
		m.CleanupExpiredMetrics()
	}()

	m.archiveTicker = m.scheduleAtTime(2, 0)
	go func() {
		for {
			select {
			case <-m.archiveTicker.C:
				m.ArchiveOldMetrics()
			case <-m.stopChan:
				return
			}
		}
	}()

	m.cleanupTicker = m.scheduleAtTime(3, 0)
	go func() {
		for {
			select {
			case <-m.cleanupTicker.C:
				m.CleanupExpiredMetrics()
			case <-m.stopChan:
				return
			}
		}
	}()

	logger.Info("数据生命周期管理器已启动",
		zap.String("archive_schedule", "每天 02:00"),
		zap.String("cleanup_schedule", "每天 03:00"))
}

// Stop 停止数据生命周期管理
func (m *DataLifecycleManager) Stop() {
	logger.Info("停止数据生命周期管理器")
	close(m.stopChan)

	if m.archiveTicker != nil {
		m.archiveTicker.Stop()
	}
	if m.cleanupTicker != nil {
		m.cleanupTicker.Stop()
	}
}

// ArchiveOldMetrics 归档旧指标数据
func (m *DataLifecycleManager) ArchiveOldMetrics() {
	startTime := time.Now()
	logger.Info("开始归档旧指标数据")

	thresholdTime := time.Now().AddDate(0, 0, -30)

	var count int64
	err := getDB().Table("mon_agent_metrics").
		Where("received_at < ?", thresholdTime).
		Count(&count).Error
	if err != nil {
		logger.Error("统计待归档数据失败", zap.Error(err))
		return
	}

	if count == 0 {
		logger.Info("没有需要归档的数据")
		return
	}

	logger.Info("发现待归档数据", zap.Int64("count", count))

	batchSize := 10000
	offset := 0
	totalArchived := 0

	for {
		var metrics []struct {
			ID         uint      `gorm:"column:id"`
			ServerID   uint      `gorm:"column:server_id"`
			MetricType string    `gorm:"column:metric_type"`
			MetricData string    `gorm:"column:metric_data"`
			ReceivedAt time.Time `gorm:"column:received_at"`
		}

		err := getDB().Table("mon_agent_metrics").
			Select("id, server_id, metric_type, metric_data, received_at").
			Where("received_at < ?", thresholdTime).
			Order("received_at ASC").
			Limit(batchSize).
			Offset(offset).
			Find(&metrics).Error
		if err != nil {
			logger.Error("查询待归档数据失败", zap.Error(err))
			break
		}

		if len(metrics) == 0 {
			break
		}

		for _, metric := range metrics {
			insertSQL := `INSERT INTO mon_agent_metrics_archive
				(server_id, metric_type, metric_data, received_at, archived_at)
				VALUES (?, ?, ?, ?, ?)`
			if err := getDB().Exec(insertSQL, metric.ServerID, metric.MetricType,
				metric.MetricData, metric.ReceivedAt, time.Now()).Error; err != nil {
				logger.Error("插入归档数据失败",
					zap.Uint("id", metric.ID),
					zap.Error(err))
				continue
			}
		}

		ids := make([]uint, len(metrics))
		for i, metric := range metrics {
			ids[i] = metric.ID
		}

		if err := getDB().Table("mon_agent_metrics").Where("id IN ?", ids).Delete(nil).Error; err != nil {
			logger.Error("删除已归档数据失败", zap.Error(err))
			break
		}

		totalArchived += len(metrics)
		offset += batchSize

		logger.Info("归档进度",
			zap.Int("batch", totalArchived),
			zap.Int64("total", count))
	}

	duration := time.Since(startTime)
	logger.Info("归档完成",
		zap.Int("archived_count", totalArchived),
		zap.Duration("duration", duration))
}

// CleanupExpiredMetrics 清理过期归档数据
func (m *DataLifecycleManager) CleanupExpiredMetrics() {
	startTime := time.Now()
	logger.Info("开始清理过期归档数据")

	thresholdTime := time.Now().AddDate(0, 0, -365)

	var count int64
	err := getDB().Table("mon_agent_metrics_archive").
		Where("received_at < ?", thresholdTime).
		Count(&count).Error
	if err != nil {
		logger.Error("统计待清理数据失败", zap.Error(err))
		return
	}

	if count == 0 {
		logger.Info("没有需要清理的数据")
		return
	}

	logger.Info("发现待清理数据", zap.Int64("count", count))

	batchSize := 10000
	totalDeleted := 0

	for {
		result := getDB().Table("mon_agent_metrics_archive").
			Where("received_at < ?", thresholdTime).
			Limit(batchSize).
			Delete(nil)

		if result.Error != nil {
			logger.Error("删除过期数据失败", zap.Error(result.Error))
			break
		}

		if result.RowsAffected == 0 {
			break
		}

		totalDeleted += int(result.RowsAffected)
		logger.Info("清理进度",
			zap.Int("batch", totalDeleted),
			zap.Int64("total", count))

		time.Sleep(100 * time.Millisecond)
	}

	duration := time.Since(startTime)
	logger.Info("清理完成",
		zap.Int("deleted_count", totalDeleted),
		zap.Duration("duration", duration))
}

// GetDataStats 获取数据统计信息
func (m *DataLifecycleManager) GetDataStats() map[string]interface{} {
	stats := make(map[string]interface{})

	var hotCount int64
	getDB().Table("mon_agent_metrics").Count(&hotCount)
	stats["hot_data_count"] = hotCount

	var archiveCount int64
	getDB().Table("mon_agent_metrics_archive").Count(&archiveCount)
	stats["archive_data_count"] = archiveCount

	thresholdTime := time.Now().AddDate(0, 0, -30)
	var pendingArchiveCount int64
	getDB().Table("mon_agent_metrics").
		Where("received_at < ?", thresholdTime).
		Count(&pendingArchiveCount)
	stats["pending_archive_count"] = pendingArchiveCount

	expiredTime := time.Now().AddDate(0, 0, -365)
	var expiredArchiveCount int64
	getDB().Table("mon_agent_metrics_archive").
		Where("received_at < ?", expiredTime).
		Count(&expiredArchiveCount)
	stats["expired_archive_count"] = expiredArchiveCount

	var tableSize []struct {
		TableName string `gorm:"column:table_name"`
		DataSize  string `gorm:"column:data_size"`
		IndexSize string `gorm:"column:index_size"`
	}

	getDB().Raw(`
		SELECT
			table_name,
			ROUND(((data_length + index_length) / 1024 / 1024), 2) AS data_size,
			ROUND((index_length / 1024 / 1024), 2) AS index_size
		FROM information_schema.TABLES
		WHERE table_schema = DATABASE()
			AND table_name IN ('mon_agent_metrics', 'mon_agent_metrics_archive')
		ORDER BY (data_length + index_length) DESC
	`).Scan(&tableSize)

	stats["table_sizes"] = tableSize

	return stats
}

// scheduleAtTime 计算下次执行时间
func (m *DataLifecycleManager) scheduleAtTime(hour, minute int) *time.Ticker {
	now := time.Now()
	next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())

	if next.Before(now) {
		next = next.Add(24 * time.Hour)
	}

	duration := next.Sub(now)
	logger.Info("计划下次执行",
		zap.Time("next_run", next),
		zap.Duration("wait_duration", duration))

	return time.NewTicker(duration)
}
