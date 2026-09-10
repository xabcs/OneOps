package monitoring

import (
	"time"

	"oneops/backend3/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// DataLifecycleManager 数据生命周期管理器
type DataLifecycleManager struct {
	stopChan chan bool
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

	// 每天固定时刻调度：使用一次性 Timer 触发后按当天时刻重新计算间隔；
	// Ticker 只会按创建时的固定间隔重复触发，无法对齐"每天 02:00/03:00"的调度语义
	go m.scheduleLoop(2, 0, m.ArchiveOldMetrics)
	go m.scheduleLoop(3, 0, m.CleanupExpiredMetrics)

	logger.Info("数据生命周期管理器已启动",
		zap.String("archive_schedule", "每天 02:00"),
		zap.String("cleanup_schedule", "每天 03:00"))
}

// scheduleLoop 每天在指定时刻执行任务，直到收到停止信号
func (m *DataLifecycleManager) scheduleLoop(hour, minute int, task func()) {
	timer := time.NewTimer(nextRunDelay(hour, minute))
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			task()
			// 一次性 Timer 触发后需 Reset 为下一次执行时刻
			timer.Reset(nextRunDelay(hour, minute))
		case <-m.stopChan:
			return
		}
	}
}

// Stop 停止数据生命周期管理
func (m *DataLifecycleManager) Stop() {
	logger.Info("停止数据生命周期管理器")
	close(m.stopChan)
}

// archivedMetricRow 归档表写入行（与 mon_agent_metrics_archive 列对应，id 由数据库自增）
type archivedMetricRow struct {
	ServerID   uint      `gorm:"column:server_id"`
	MetricType string    `gorm:"column:metric_type"`
	MetricData string    `gorm:"column:metric_data"`
	ReceivedAt time.Time `gorm:"column:received_at"`
	ArchivedAt time.Time `gorm:"column:archived_at"`
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
	totalArchived := 0
	// 基于 ID 游标分页：源表数据随删除不断前移，offset 分页会跳过数据导致漏归档；
	// 游标始终从上一批最大 ID 之后继续，保证不重不漏
	lastID := uint(0)

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
			Where("received_at < ? AND id > ?", thresholdTime, lastID).
			Order("id ASC").
			Limit(batchSize).
			Find(&metrics).Error
		if err != nil {
			logger.Error("查询待归档数据失败", zap.Error(err))
			break
		}

		if len(metrics) == 0 {
			break
		}

		// 组装归档行（逐行改批量写入）
		archivedAt := time.Now()
		ids := make([]uint, len(metrics))
		archiveRows := make([]archivedMetricRow, len(metrics))
		for i, metric := range metrics {
			ids[i] = metric.ID
			archiveRows[i] = archivedMetricRow{
				ServerID:   metric.ServerID,
				MetricType: metric.MetricType,
				MetricData: metric.MetricData,
				ReceivedAt: metric.ReceivedAt,
				ArchivedAt: archivedAt,
			}
		}

		// 归档语义：先写入归档表、成功后再删源表，任一步失败事务回滚，源数据不删
		err = getDB().Transaction(func(tx *gorm.DB) error {
			if err := tx.Table("mon_agent_metrics_archive").CreateInBatches(archiveRows, batchSize).Error; err != nil {
				return err
			}
			return tx.Table("mon_agent_metrics").Where("id IN ?", ids).Delete(nil).Error
		})
		if err != nil {
			logger.Error("归档批次失败（事务已回滚，源数据未删除）", zap.Error(err))
			break
		}

		// 推进游标到本批最大 ID
		lastID = metrics[len(metrics)-1].ID
		totalArchived += len(metrics)

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

// nextRunDelay 计算距离下一个 hour:minute 时刻的等待时长
func nextRunDelay(hour, minute int) time.Duration {
	now := time.Now()
	next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())

	if next.Before(now) {
		next = next.Add(24 * time.Hour)
	}

	duration := next.Sub(now)
	logger.Info("计划下次执行",
		zap.Time("next_run", next),
		zap.Duration("wait_duration", duration))

	return duration
}
