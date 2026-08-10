package modelcmdb

import "time"

// AgentVersion Agent 版本模型
type AgentVersion struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Version      string    `json:"version" gorm:"size:50;not null;uniqueIndex"` // 版本号
	ReleaseNotes string    `json:"releaseNotes" gorm:"type:text"`               // 发布说明
	Changelog    string    `json:"changelog" gorm:"type:text"`                  // 更新日志
	ReleasedAt   time.Time `json:"releasedAt"`                                  // 发布时间

	// 二进制文件信息
	AMD64BinaryPath string `json:"amd64BinaryPath" gorm:"size:255"` // AMD64 二进制路径
	AMD64BinaryHash string `json:"amd64BinaryHash" gorm:"size:64"`  // AMD64 文件哈希
	AMD64BinarySize int64  `json:"amd64BinarySize"`                 // AMD64 文件大小
	ARM64BinaryPath string `json:"arm64BinaryPath" gorm:"size:255"` // ARM64 二进制路径
	ARM64BinaryHash string `json:"arm64BinaryHash" gorm:"size:64"`  // ARM64 文件哈希
	ARM64BinarySize int64  `json:"arm64BinarySize"`                 // ARM64 文件大小

	// 版本状态
	IsLatest     bool `json:"isLatest" gorm:"default:0"`     // 是否最新版本
	IsDeprecated bool `json:"isDeprecated" gorm:"default:0"` // 是否弃用

	// 功能支持
	Features             string `json:"features" gorm:"type:json"`           // 功能列表 JSON
	MinCompatibleVersion string `json:"minCompatibleVersion" gorm:"size:50"` // 最小兼容版本
	MaxCompatibleVersion string `json:"maxCompatibleVersion" gorm:"size:50"` // 最大兼容版本

	// 统计信息
	DownloadCount int `json:"downloadCount" gorm:"default:0"` // 下载次数
	DeployCount   int `json:"deployCount" gorm:"default:0"`   // 部署次数

	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

// TableName 指定表名
func (AgentVersion) TableName() string {
	return "cmdb_agent_versions"
}

// AgentUpgradeTask Agent 升级任务模型
type AgentUpgradeTask struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	TaskName string `json:"taskName" gorm:"size:100"`

	// 升级目标信息
	TargetVersion   string `json:"targetVersion" gorm:"size:50;not null"` // 目标版本
	TargetServerIDs string `json:"targetServerIds" gorm:"type:json"`      // 目标主机ID列表

	// 任务状态
	Status      string `json:"status" gorm:"type:varchar(20);default:'pending';index"` // pending|running|completed|failed|cancelled
	CurrentStep int    `json:"currentStep" gorm:"default:0"`
	TotalSteps  int    `json:"totalSteps" gorm:"default:0"`

	// 进度统计
	TotalCount   int `json:"totalCount" gorm:"default:0"`
	SuccessCount int `json:"successCount" gorm:"default:0"`
	FailedCount  int `json:"failedCount" gorm:"default:0"`
	SkippedCount int `json:"skippedCount" gorm:"default:0"`

	// 时间记录
	StartedAt   *time.Time `json:"startedAt"`
	CompletedAt *time.Time `json:"completedAt"`

	// 详细日志
	ErrorMessage string `json:"errorMessage" gorm:"type:text"`
	OperationLog string `json:"operationLog" gorm:"type:json"`

	CreatedBy string    `json:"createdBy" gorm:"size:50"`
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

// TableName 指定表名
func (AgentUpgradeTask) TableName() string {
	return "cmdb_agent_upgrade_tasks"
}

// AgentVersionFeature 版本功能支持结构
type AgentVersionFeature struct {
	ExtendedMetrics bool `json:"extendedMetrics"` // 支持扩展指标
	CustomConfigs   bool `json:"customConfigs"`   // 支持自定义配置
}
