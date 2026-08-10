package monitoring

import (
	"oneops/backend2/pkg/logger"
	"time"

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

	// 立即执行一次归档和清理
	go func() {
		m.ArchiveOldMetrics()
		m.CleanupExpiredMetrics()
	}()

	// 设置归档任务：每天凌晨 2 点执行
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

	// 设置清理任务：每天凌晨 3 点执行
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

// ArchiveOldMetrics 归档旧指标数据（30 天前的数据迁移到归档表）
func (m *DataLifecycleManager) ArchiveOldMetrics() {
	startTime := time.Now()
	logger.Info("开始归档旧指标数据")

	// 计算归档阈值时间（30 天前）
	thresholdTime := time.Now().AddDate(0, 0, -30)

	// 统计需要归档的数据量
	var count int64
	err := db.Table("mon_agent_metrics").
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

	// 分批归档数据（每批 10000 条）
	batchSize := 10000
	offset := 0
	totalArchived := 0

	for {
		// 查询一批数据
		var metrics []struct {
			ID         uint      `gorm:"column:id"`
			ServerID   uint      `gorm:"column:server_id"`
			MetricType string    `gorm:"column:metric_type"`
			MetricData string    `gorm:"column:metric_data"`
			ReceivedAt time.Time `gorm:"column:received_at"`
		}

		err := db.Table("mon_agent_metrics").
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

		// 插入到归档表
		for _, metric := range metrics {
			insertSQL := `INSERT INTO mon_agent_metrics_archive
				(server_id, metric_type, metric_data, received_at, archived_at)
				VALUES (?, ?, ?, ?, ?)`
			if err := db.Exec(insertSQL, metric.ServerID, metric.MetricType,
				metric.MetricData, metric.ReceivedAt, time.Now()).Error; err != nil {
				logger.Error("插入归档数据失败",
					zap.Uint("id", metric.ID),
					zap.Error(err))
				continue
			}
		}

		// 从主表删除已归档的数据
		ids := make([]uint, len(metrics))
		for i, metric := range metrics {
			ids[i] = metric.ID
		}

		if err := db.Table("mon_agent_metrics").Where("id IN ?", ids).Delete(nil).Error; err != nil {
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

// CleanupExpiredMetrics 清理过期归档数据（365 天前的归档数据删除）
func (m *DataLifecycleManager) CleanupExpiredMetrics() {
	startTime := time.Now()
	logger.Info("开始清理过期归档数据")

	// 计算清理阈值时间（365 天前）
	thresholdTime := time.Now().AddDate(0, 0, -365)

	// 统计需要删除的数据量
	var count int64
	err := db.Table("mon_agent_metrics_archive").
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

	// 分批删除数据（每批 10000 条）
	batchSize := 10000
	totalDeleted := 0

	for {
		result := db.Table("mon_agent_metrics_archive").
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

		// 避免一次性删除太多数据导致锁表
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

	// 主表数据量
	var hotCount int64
	db.Table("mon_agent_metrics").Count(&hotCount)
	stats["hot_data_count"] = hotCount

	// 归档表数据量
	var archiveCount int64
	db.Table("mon_agent_metrics_archive").Count(&archiveCount)
	stats["archive_data_count"] = archiveCount

	// 30 天前数据量（待归档）
	thresholdTime := time.Now().AddDate(0, 0, -30)
	var pendingArchiveCount int64
	db.Table("mon_agent_metrics").
		Where("received_at < ?", thresholdTime).
		Count(&pendingArchiveCount)
	stats["pending_archive_count"] = pendingArchiveCount

	// 365 天前归档数据量（待清理）
	expiredTime := time.Now().AddDate(0, 0, -365)
	var expiredArchiveCount int64
	db.Table("mon_agent_metrics_archive").
		Where("received_at < ?", expiredTime).
		Count(&expiredArchiveCount)
	stats["expired_archive_count"] = expiredArchiveCount

	// 数据库表大小
	var tableSize []struct {
		TableName string `gorm:"column:table_name"`
		DataSize  string `gorm:"column:data_size"`
		IndexSize string `gorm:"column:index_size"`
	}

	db.Raw(`
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

// scheduleAtTime 计算下次执行时间（每天指定时间执行）
func (m *DataLifecycleManager) scheduleAtTime(hour, minute int) *time.Ticker {
	now := time.Now()
	next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())

	// 如果今天的时间已过，设置为明天
	if next.Before(now) {
		next = next.Add(24 * time.Hour)
	}

	duration := next.Sub(now)
	logger.Info("计划下次执行",
		zap.Time("next_run", next),
		zap.Duration("wait_duration", duration))

	return time.NewTicker(duration)
}
