package monitoring

import (
	"sync"
	"time"

	"oneops/backend3/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PersistMetric 持久化指标数据结构
type PersistMetric struct {
	ServerID   uint
	MetricType string
	MetricData []byte
	ReportTime time.Time
}

// MetricsPersister 指标数据批量写入器
type MetricsPersister struct {
	persistQueue chan *PersistMetric
	isRunning    bool
	mu           sync.Mutex
}

var metricsPersister *MetricsPersister

// GetMetricsPersister 获取批量写入器实例
func GetMetricsPersister() *MetricsPersister {
	if metricsPersister == nil {
		metricsPersister = &MetricsPersister{
			persistQueue: make(chan *PersistMetric, 5000),
			isRunning:    false,
		}
	}
	return metricsPersister
}

// StartPersistWorker 启动批量写入worker
func (p *MetricsPersister) StartPersistWorker() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.isRunning {
		return
	}
	p.isRunning = true

	logger.Info("批量指标写入器已启动")

	ticker := time.NewTicker(30 * time.Second)
	batch := make([]*PersistMetric, 0, 1000)

	go func() {
		for {
			select {
			case metric := <-p.persistQueue:
				batch = append(batch, metric)
				if len(batch) >= 1000 {
					p.persistBatch(batch)
					batch = make([]*PersistMetric, 0, 1000)
				}

			case <-ticker.C:
				if len(batch) > 0 {
					p.persistBatch(batch)
					batch = make([]*PersistMetric, 0, 1000)
				}
			}
		}
	}()
}

// persistBatch 批量写入指标数据
func (p *MetricsPersister) persistBatch(batch []*PersistMetric) {
	if len(batch) == 0 {
		return
	}

	startTime := time.Now()
	logger.Debug("开始批量写入指标", zap.Int("count", len(batch)))

	err := getDB().Transaction(func(tx *gorm.DB) error {
		for _, metric := range batch {
			query := `INSERT INTO mon_agent_metrics (server_id, metric_type, metric_data, report_time, received_at)
			             VALUES (?, ?, ?, ?, NOW())`
			if err := tx.Exec(query, metric.ServerID, metric.MetricType, string(metric.MetricData), metric.ReportTime).Error; err != nil {
				logger.Warn("批量写入指标失败",
					zap.Uint("serverID", metric.ServerID),
					zap.String("metricType", metric.MetricType),
					zap.Error(err))
			}
		}
		return nil
	})

	duration := time.Since(startTime)
	if err != nil {
		logger.Warn("批量指标写入事务失败", zap.Error(err), zap.Duration("duration", duration))
	} else {
		logger.Info("批量指标写入完成",
			zap.Int("count", len(batch)),
			zap.Duration("duration", duration),
			zap.Float64("avg_latency", float64(duration.Milliseconds())/float64(len(batch))))
	}
}

// EnqueueMetric 将指标数据放入持久化队列
func (p *MetricsPersister) EnqueueMetric(serverID uint, metricType string, metricData []byte, reportTime time.Time) {
	metric := &PersistMetric{
		ServerID:   serverID,
		MetricType: metricType,
		MetricData: metricData,
		ReportTime: reportTime,
	}

	select {
	case p.persistQueue <- metric:
	default:
		logger.Warn("持久化队列已满，丢弃指标数据",
			zap.Uint("serverID", serverID),
			zap.String("metricType", metricType))
	}
}
