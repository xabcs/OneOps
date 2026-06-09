package services

import (
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"oneops/backend/config"
	"oneops/backend/logger"
)

// DBManager 数据库管理器
// 管理数据库连接池的生命周期和配置
type DBManager struct {
	// 主数据库连接
	masterDB *gorm.DB

	// 从数据库连接（读写分离）
	slaveDBs []*gorm.DB
	slaveIdx int

	// 配置
	config config.DatabaseConfig

	// 统计信息
	stats *DBStats

	// 互斥锁
	mu sync.RWMutex
}

// DBStats 数据库统计信息
type DBStats struct {
	TotalConnections     int32     // 总连接数
	ActiveConnections   int32     // 活跃连接数
	IdleConnections      int32     // 空闲连接数
	WaitCount           int64     // 等待次数
	WaitDuration        time.Duration // 总等待时长
	MaxWaitDuration     time.Duration // 最大等待时长
	TotalQueries        int64     // 总查询数
	SlowQueries         int64     // 慢查询数
	LastError           error     // 最后一次错误
	LastErrorTime       time.Time // 最后一次错误时间
	LastHealthCheckTime time.Time // 最后一次健康检查时间
	HealthCheckPassed   bool      // 健康检查是否通过
}

// DBManagerConfig 数据库管理器配置
type DBManagerConfig struct {
	// MaxOpenConns 最大打开连接数（默认：100）
	MaxOpenConns int
	// MaxIdleConns 最大空闲连接数（默认：10）
	MaxIdleConns int
	// ConnMaxLifetime 连接最大生命周期（默认：1小时）
	ConnMaxLifetime time.Duration
	// ConnMaxIdleTime 连接最大空闲时间（默认：10分钟）
	ConnMaxIdleTime time.Duration
	// SlowThreshold 慢查询阈值（默认：1秒）
	SlowThreshold time.Duration
	// HealthCheckInterval 健康检查间隔（默认：30秒）
	HealthCheckInterval time.Duration
	// EnableReadWriteSeparation 是否启用读写分离
	EnableReadWriteSeparation bool
	// SlaveConfigs 从库配置列表
	SlaveConfigs []config.DatabaseConfig
}

// DefaultDBManagerConfig 默认配置
var DefaultDBManagerConfig = DBManagerConfig{
	MaxOpenConns:              100,
	MaxIdleConns:              10,
	ConnMaxLifetime:           time.Hour,
	ConnMaxIdleTime:           10 * time.Minute,
	SlowThreshold:             time.Second,
	HealthCheckInterval:       30 * time.Second,
	EnableReadWriteSeparation: false,
}

// NewDBManager 创建数据库管理器
func NewDBManager(cfg config.DatabaseConfig, dbManagerCfg ...DBManagerConfig) (*DBManager, error) {
	// 合并配置
	managerCfg := DefaultDBManagerConfig
	if len(dbManagerCfg) > 0 {
		managerCfg = dbManagerCfg[0]
	}

	mgr := &DBManager{
		config: cfg,
		stats:  &DBStats{},
	}

	// 初始化主数据库连接
	if err := mgr.initMasterDB(cfg, managerCfg); err != nil {
		return nil, fmt.Errorf("初始化主数据库失败: %w", err)
	}

	// 初始化从数据库连接（读写分离）
	if managerCfg.EnableReadWriteSeparation && len(managerCfg.SlaveConfigs) > 0 {
		if err := mgr.initSlaveDBs(managerCfg); err != nil {
			logger.Logger.Warn("初始化从数据库失败，将只使用主数据库", zap.Error(err))
		}
	}

	// 启动健康检查
	go mgr.startHealthCheck(managerCfg.HealthCheckInterval)

	logger.Logger.Info("数据库管理器初始化成功",
		zap.Int("max_open_conns", managerCfg.MaxOpenConns),
		zap.Int("max_idle_conns", managerCfg.MaxIdleConns),
		zap.Bool("read_write_separation", managerCfg.EnableReadWriteSeparation),
		zap.Int("slave_count", len(mgr.slaveDBs)),
	)

	return mgr, nil
}

// initMasterDB 初始化主数据库连接
func (mgr *DBManager) initMasterDB(cfg config.DatabaseConfig, managerCfg DBManagerConfig) error {
	// 获取现有数据库连接（由InitDB创建）
	if db == nil {
		return fmt.Errorf("数据库未初始化")
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("获取底层数据库连接失败: %w", err)
	}

	// 配置连接池
	sqlDB.SetMaxOpenConns(managerCfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(managerCfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(managerCfg.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(managerCfg.ConnMaxIdleTime)

	mgr.masterDB = db
	return nil
}

// initSlaveDBs 初始化从数据库连接
func (mgr *DBManager) initSlaveDBs(managerCfg DBManagerConfig) error {
	for _, slaveCfg := range managerCfg.SlaveConfigs {
		// TODO: 根据从库配置创建连接
		// 这里需要实现独立的数据库连接逻辑
		logger.Logger.Info("从数据库连接待实现", zap.String("host", slaveCfg.Host))
	}

	return nil
}

// GetMaster 获取主数据库连接（用于写操作）
func (mgr *DBManager) GetMaster() *gorm.DB {
	mgr.mu.RLock()
	defer mgr.mu.RUnlock()
	return mgr.masterDB
}

// GetSlave 获取从数据库连接（用于读操作）
// 使用轮询策略选择从库
func (mgr *DBManager) GetSlave() *gorm.DB {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	// 如果没有从库，返回主库
	if len(mgr.slaveDBs) == 0 {
		return mgr.masterDB
	}

	// 轮询选择从库
	idx := mgr.slaveIdx % len(mgr.slaveDBs)
	mgr.slaveIdx++
	return mgr.slaveDBs[idx]
}

// GetDB 获取数据库连接（自动选择主库或从库）
// 对于写操作，应该使用GetMaster()明确指定
func (mgr *DBManager) GetDB(readOnly bool) *gorm.DB {
	if readOnly {
		return mgr.GetSlave()
	}
	return mgr.GetMaster()
}

// startHealthCheck 启动健康检查
func (mgr *DBManager) startHealthCheck(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		mgr.healthCheck()
	}
}

// healthCheck 执行健康检查
func (mgr *DBManager) healthCheck() {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	now := time.Now()
	mgr.stats.LastHealthCheckTime = now

	// 检查主库
	if err := mgr.pingDB(mgr.masterDB); err != nil {
		mgr.stats.HealthCheckPassed = false
		mgr.stats.LastError = err
		mgr.stats.LastErrorTime = now
		logger.Logger.Error("主数据库健康检查失败", zap.Error(err))
		return
	}

	// 检查从库
	for i, slave := range mgr.slaveDBs {
		if err := mgr.pingDB(slave); err != nil {
			logger.Logger.Error("从数据库健康检查失败",
				zap.Int("slave_index", i),
				zap.Error(err))
		}
	}

	mgr.stats.HealthCheckPassed = true
}

// pingDB Ping数据库连接
func (mgr *DBManager) pingDB(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// GetStats 获取统计信息
func (mgr *DBManager) GetStats() *DBStats {
	mgr.mu.RLock()
	defer mgr.mu.RUnlock()

	// 更新连接池统计
	if sqlDB, err := mgr.masterDB.DB(); err == nil {
		stats := sqlDB.Stats()
		mgr.stats.TotalConnections = int32(stats.MaxOpenConnections)
		mgr.stats.ActiveConnections = int32(stats.InUse)
		mgr.stats.IdleConnections = int32(stats.Idle)
	}

	return mgr.stats
}

// Close 关闭数据库连接
func (mgr *DBManager) Close() error {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	var lastErr error

	// 关闭主库
	if sqlDB, err := mgr.masterDB.DB(); err == nil {
		if err := sqlDB.Close(); err != nil {
			lastErr = err
		}
	}

	// 关闭从库
	for _, slave := range mgr.slaveDBs {
		if sqlDB, err := slave.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				lastErr = err
			}
		}
	}

	logger.Logger.Info("数据库连接已关闭")
	return lastErr
}

// RecordSlowQuery 记录慢查询
func (mgr *DBManager) RecordSlowQuery(sql string, duration time.Duration) {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	mgr.stats.SlowQueries++
	mgr.stats.TotalQueries++

	logger.Logger.Warn("数据库慢查询",
		zap.String("sql", sql),
		zap.Duration("duration", duration),
		zap.Duration("threshold", time.Second),
		zap.Int64("duration_ms", duration.Milliseconds()),
	)
}

// ExecuteWithStats 执行查询并记录统计
func (mgr *DBManager) ExecuteWithStats(db *gorm.DB, query string, args ...interface{}) *gorm.DB {
	start := time.Now()

	result := db.Raw(query, args...)
	duration := time.Since(start)

	// 记录统计
	mgr.mu.Lock()
	mgr.stats.TotalQueries++
	if duration > time.Second {
		mgr.stats.SlowQueries++
	}
	mgr.mu.Unlock()

	// 记录慢查询
	if duration > time.Second {
		mgr.RecordSlowQuery(query, duration)
	}

	return result
}

// GetHealthStatus 获取健康状态
func (mgr *DBManager) GetHealthStatus() map[string]interface{} {
	stats := mgr.GetStats()

	return map[string]interface{}{
		"healthy":             stats.HealthCheckPassed,
		"last_health_check":   stats.LastHealthCheckTime,
		"total_connections":   stats.TotalConnections,
		"active_connections":  stats.ActiveConnections,
		"idle_connections":    stats.IdleConnections,
		"total_queries":       stats.TotalQueries,
		"slow_queries":        stats.SlowQueries,
		"last_error":          stats.LastError,
		"last_error_time":     stats.LastErrorTime,
		"master_available":    mgr.masterDB != nil,
		"slave_count":         len(mgr.slaveDBs),
	}
}

// IsHealthy 检查数据库是否健康
func (mgr *DBManager) IsHealthy() bool {
	stats := mgr.GetStats()
	return stats.HealthCheckPassed && stats.ActiveConnections > 0
}

// GetConnectionInfo 获取连接信息（用于调试）
func (mgr *DBManager) GetConnectionInfo() map[string]interface{} {
	mgr.mu.RLock()
	defer mgr.mu.RUnlock()

	info := map[string]interface{}{
		"master_available": mgr.masterDB != nil,
		"slave_count":      len(mgr.slaveDBs),
		"read_write_separation": false,
	}

	if mgr.masterDB != nil {
		if sqlDB, err := mgr.masterDB.DB(); err == nil {
			stats := sqlDB.Stats()
			info["master_stats"] = map[string]interface{}{
				"max_open_connections": stats.MaxOpenConnections,
				"in_use":              stats.InUse,
				"idle":                stats.Idle,
				"wait_count":          stats.WaitCount,
				"wait_duration":       stats.WaitDuration.String(),
			}
		}
	}

	return info
}

// ResetStats 重置统计信息
func (mgr *DBManager) ResetStats() {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	mgr.stats = &DBStats{
		LastHealthCheckTime: time.Now(),
	}
}
