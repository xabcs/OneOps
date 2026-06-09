package models

import "time"

// ServerMetrics 服务器监控指标（分离的监控表）
type ServerMetrics struct {
	ServerID         uint            `json:"serverId" gorm:"primaryKey"`
	CPUUsage         float64         `json:"cpuUsage" gorm:"default:0"`
	MemoryUsage      float64         `json:"memoryUsage" gorm:"default:0"`
	DiskUsage        float64         `json:"diskUsage" gorm:"default:0"`
	Load1            float64         `json:"load1" gorm:"default:0"`
	Load5            float64         `json:"load5" gorm:"default:0"`
	Load15           float64         `json:"load15" gorm:"default:0"`
	DiskPartitions   DiskPartitions  `json:"diskPartitions,omitempty" gorm:"type:json"`
	MetricsUpdatedAt *time.Time      `json:"metricsUpdatedAt"`
	CreatedAt        time.Time       `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt        time.Time       `json:"updatedAt" gorm:"autoUpdateTime"`

	// 关联
	Server *Server `json:"server,omitempty" gorm:"foreignKey:ServerID;constraint:OnDelete:CASCADE"`
}

// TableName 指定表名
func (ServerMetrics) TableName() string {
	return "server_metrics"
}
