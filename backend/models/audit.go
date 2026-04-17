package models

import (
	"time"
)

// LoginLog 登录日志模型
type LoginLog struct {
	ID         uint       `json:"id" gorm:"primaryKey"`
	UserID     uint       `json:"userId" gorm:"index;not null"`
	Username   string     `json:"username" gorm:"size:50;not null;index"`
	Nickname   string     `json:"nickname" gorm:"size:50"`
	IP         string     `json:"ip" gorm:"size:45;not null"`
	UserAgent  string     `json:"userAgent" gorm:"size:500"`
	Location   string     `json:"location" gorm:"size:100"`
	Status     string     `json:"status" gorm:"size:20;not null;default:'success'"` // success, failed
	FailReason string     `json:"failReason" gorm:"size:200"`
	LoginTime  time.Time  `json:"loginTime" gorm:"not null;index"`
	LogoutTime *time.Time `json:"logoutTime" gorm:"default:null"`
	Duration   int        `json:"duration" gorm:"default:0"` // 会话时长(秒)
	CreatedAt  time.Time  `json:"createdAt" gorm:"autoCreateTime"`
}

// TableName 指定表名
func (LoginLog) TableName() string {
	return "login_logs"
}

// OperationLog 操作日志模型
type OperationLog struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	UserID      uint      `json:"userId" gorm:"index;not null"`
	Username    string    `json:"username" gorm:"size:50;not null;index"`
	Nickname    string    `json:"nickname" gorm:"size:50"`
	Module      string    `json:"module" gorm:"size:50;not null;index"` // 模块名称
	Action      string    `json:"action" gorm:"size:100;not null"`      // 操作动作
	Description string    `json:"description" gorm:"size:500"`          // 操作描述
	Method      string    `json:"method" gorm:"size:10;not null"`       // HTTP方法
	Path        string    `json:"path" gorm:"size:500;not null"`        // 请求路径
	Params      string    `json:"params" gorm:"type:text"`              // 请求参数
	Response    string    `json:"response" gorm:"type:text"`            // 响应数据
	StatusCode  int       `json:"statusCode" gorm:"not null"`           // HTTP状态码
	IP          string    `json:"ip" gorm:"size:45;not null"`
	UserAgent   string    `json:"userAgent" gorm:"size:500"`
	Duration    int       `json:"duration" gorm:"default:0"`                        // 请求耗时(毫秒)
	Status      string    `json:"status" gorm:"size:20;not null;default:'success'"` // success, failed
	ErrorMsg    string    `json:"errorMsg" gorm:"size:500"`
	OperateTime time.Time `json:"operateTime" gorm:"not null;index"`
	CreatedAt   time.Time `json:"createdAt" gorm:"autoCreateTime"`
}

// TableName 指定表名
func (OperationLog) TableName() string {
	return "operation_logs"
}

// SystemEventLog 系统事件日志模型
type SystemEventLog struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Level     string    `json:"level" gorm:"size:20;not null;index"`   // info, warning, error, critical
	Source    string    `json:"source" gorm:"size:100;not null;index"` // 事件来源
	Category  string    `json:"category" gorm:"size:50;index"`         // 事件分类
	Message   string    `json:"message" gorm:"size:1000;not null"`     // 事件消息
	Details   string    `json:"details" gorm:"type:text"`              // 详细数据
	IP        string    `json:"ip" gorm:"size:45"`
	EventTime time.Time `json:"eventTime" gorm:"not null;index"`
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
}

// TableName 指定表名
func (SystemEventLog) TableName() string {
	return "system_event_logs"
}
