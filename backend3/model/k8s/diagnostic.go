package modelk8s

import (
	"time"
)

// DiagnosticHistory 诊断历史记录
type DiagnosticHistory struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	ClusterID      string    `json:"clusterId" gorm:"size:50;not null;index:idx_diagnostic_cluster"`
	ClusterName    string    `json:"clusterName" gorm:"size:100;not null"`
	Namespace      string    `json:"namespace" gorm:"size:100;not null;index:idx_diagnostic_pod"`
	PodName        string    `json:"podName" gorm:"size:100;not null;index:idx_diagnostic_pod"`
	ContainerName  string    `json:"containerName" gorm:"size:100"`
	Command        string    `json:"command" gorm:"size:50;not null;index:idx_diagnostic_command"`
	Args           string    `json:"args" gorm:"type:text"`
	ResultStatus   string    `json:"resultStatus" gorm:"size:20;not null"` // success, error
	ResultOutput   string    `json:"resultOutput" gorm:"type:longtext"`
	ResultError    string    `json:"resultError" gorm:"type:text"`
	ResultMethod   string    `json:"resultMethod" gorm:"size:20"` // sidecar, daemonset
	ResultDuration int       `json:"resultDuration"`              // milliseconds
	UserID         uint      `json:"userId" gorm:"not null;index:idx_diagnostic_user"`
	Username       string    `json:"username" gorm:"size:50;not null"`
	Timestamp      time.Time `json:"timestamp" gorm:"not null;index:idx_diagnostic_timestamp"`
	CreatedAt      time.Time `json:"createdAt" gorm:"autoUpdateTime"`
	UpdatedAt      time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

// TableName 指定表名
func (DiagnosticHistory) TableName() string {
	return "k8s_diagnostic_history"
}

// DiagnosticConfig 诊断配置
type DiagnosticConfig struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	ClusterID   string    `json:"clusterId" gorm:"size:50;not null;index:idx_diagnostic_config_cluster"`
	Namespace   string    `json:"namespace" gorm:"size:100;not null;index:idx_diagnostic_config_cluster"`
	ConfigKey   string    `json:"configKey" gorm:"size:100;not null;index:idx_diagnostic_config_unique"`
	ConfigValue string    `json:"configValue" gorm:"type:text;not null"`
	Description string    `json:"description" gorm:"size:255"`
	CreatedAt   time.Time `json:"createdAt" gorm:"autoUpdateTime"`
	UpdatedAt   time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

// TableName 指定表名
func (DiagnosticConfig) TableName() string {
	return "k8s_diagnostic_config"
}

// DiagnosticPermission 诊断权限定义
type DiagnosticPermission struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"size:50;not null"`
	Code        string    `json:"code" gorm:"size:50;uniqueIndex;not null"`
	Category    string    `json:"category" gorm:"size:50;not null;index:idx_diagnostic_permission_category"`
	Description string    `json:"description" gorm:"size:200"`
	RiskLevel   string    `json:"riskLevel" gorm:"size:20"` // low, medium, high
	Status      int       `json:"status" gorm:"default:1"`
	CreatedAt   time.Time `json:"createdAt" gorm:"autoUpdateTime"`
	UpdatedAt   time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

// TableName 指定表名
func (DiagnosticPermission) TableName() string {
	return "k8s_diagnostic_permissions"
}
