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

// DiagnosticAgent tunnel agent 注册表（由目标 JVM 内 Arthas agent 反向注册到 tunnel-server 产生）
type DiagnosticAgent struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	ClusterID     string    `json:"clusterId" gorm:"size:50;index:idx_diag_agent_cluster"` // tunnel 归属集群
	AppName       string    `json:"appName" gorm:"size:100;not null;index:idx_diag_agent_app"` // appName（应用分组）
	AgentID       string    `json:"agentId" gorm:"size:200;uniqueIndex:uk_diag_agent_id;not null"` // 约定格式: POD_NAME-NAMESPACE
	Namespace     string    `json:"namespace" gorm:"size:100;index:idx_diag_agent_pod"`
	PodName       string    `json:"podName" gorm:"size:200"`
	AgentVersion  string    `json:"agentVersion" gorm:"size:50"` // Arthas 版本
	Online        bool      `json:"online" gorm:"default:false;index:idx_diag_agent_online"`
	LastSeen      time.Time `json:"lastSeen" gorm:"index:idx_diag_agent_last_seen"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// TableName 指定表名
func (DiagnosticAgent) TableName() string { return "k8s_diag_agent" }

// DiagnosticSession 诊断交互会话（专家终端），每 agent 同时仅允许一个活跃会话
type DiagnosticSession struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	SessionKey  string     `json:"sessionKey" gorm:"size:100;uniqueIndex:uk_diag_session_key;not null"` // agentID（互斥键）
	ClusterID   string     `json:"clusterId" gorm:"size:50"`
	AppName     string     `json:"appName" gorm:"size:100"`
	AgentID     string     `json:"agentId" gorm:"size:200;index:idx_diag_session_agent"`
	Status      string     `json:"status" gorm:"size:20;index:idx_diag_session_status"` // active / closed / timeout / terminated
	UserID      uint       `json:"userId" gorm:"index:idx_diag_session_user"`
	Username    string     `json:"username" gorm:"size:50"`
	IOLog       string     `json:"ioLog" gorm:"type:longtext"` // I/O 录制（回放数据源）
	StartAt     time.Time  `json:"startAt" gorm:"index:idx_diag_session_start"`
	EndAt       *time.Time `json:"endAt"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

// TableName 指定表名
func (DiagnosticSession) TableName() string { return "k8s_diag_session" }

// DiagnosticExecution 诊断执行记录（one-shot 命令与终端会话内的命令均落此表）
type DiagnosticExecution struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	SessionID     uint      `json:"sessionId" gorm:"index:idx_diag_exec_session"` // 0 表示 one-shot 执行
	ClusterID     string    `json:"clusterId" gorm:"size:50;index:idx_diag_exec_cluster"`
	AppName       string    `json:"appName" gorm:"size:100;index:idx_diag_exec_app"`
	Namespace     string    `json:"namespace" gorm:"size:100"`
	PodName       string    `json:"podName" gorm:"size:200"`
	AgentID       string    `json:"agentId" gorm:"size:200;index:idx_diag_exec_agent"`
	Command       string    `json:"command" gorm:"size:500;not null"` // 完整命令行
	RiskLevel     string    `json:"riskLevel" gorm:"size:10;index:idx_diag_exec_risk"` // L0-L5
	Channel       string    `json:"channel" gorm:"size:20"` // oneshot / terminal
	ResultStatus  string    `json:"resultStatus" gorm:"size:20;not null"` // success / error
	ResultOutput  string    `json:"resultOutput" gorm:"type:longtext"`
	ResultError   string    `json:"resultError" gorm:"type:text"`
	Duration      int       `json:"duration"` // 毫秒
	UserID        uint      `json:"userId" gorm:"index:idx_diag_exec_user"`
	Username      string    `json:"username" gorm:"size:50;index:idx_diag_exec_username"`
	Timestamp     time.Time `json:"timestamp" gorm:"index:idx_diag_exec_time"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// TableName 指定表名
func (DiagnosticExecution) TableName() string { return "k8s_diag_execution" }

// DiagnosticCommandOverride 命令风险覆盖表（平台维护的唯一命令管控数据；内置目录为默认值）
type DiagnosticCommandOverride struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Command     string    `json:"command" gorm:"size:100;uniqueIndex:uk_diag_cmd_override;not null"` // 命令名（首 token）
	RiskLevel   string    `json:"riskLevel" gorm:"size:10"` // 覆盖后的风险级 L0-L5，disabled 表示拉黑
	Description string    `json:"description" gorm:"size:255"`
	UpdatedBy   string    `json:"updatedBy" gorm:"size:50"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// TableName 指定表名
func (DiagnosticCommandOverride) TableName() string { return "k8s_diag_command_override" }

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
