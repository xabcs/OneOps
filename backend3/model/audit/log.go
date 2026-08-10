package modelaudit

import "time"

// LoginLog 登录日志模型
type LoginLog struct {
	ID         uint       `json:"id" gorm:"primaryKey"`
	UserID     uint       `json:"userId" gorm:"index;not null"`
	Username   string     `json:"username" gorm:"size:50;not null;index"`
	Nickname   string     `json:"nickname" gorm:"size:50"`
	IP         string     `json:"ip" gorm:"size:45;not null"`
	UserAgent  string     `json:"userAgent" gorm:"size:500"`
	Location   string     `json:"location" gorm:"size:100"`
	Status     string     `json:"status" gorm:"size:20;not null;default:'success'"`
	FailReason string     `json:"failReason" gorm:"size:200"`
	LoginTime  time.Time  `json:"loginTime" gorm:"not null;index"`
	Time       string     `json:"time" gorm:"-"`
	LogoutTime *time.Time `json:"logoutTime" gorm:"default:null"`
	Duration   int        `json:"duration" gorm:"default:0"`
	CreatedAt  time.Time  `json:"createdAt" gorm:"autoCreateTime"`
}

func (LoginLog) TableName() string { return "audit_login_logs" }

// OperationLog 操作日志模型
type OperationLog struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	UserID      uint      `json:"userId" gorm:"index;not null"`
	Username    string    `json:"username" gorm:"size:50;not null;index"`
	Nickname    string    `json:"nickname" gorm:"size:50"`
	Module      string    `json:"module" gorm:"size:50;not null;index"`
	Action      string    `json:"action" gorm:"size:100;not null"`
	Description string    `json:"description" gorm:"size:500"`
	Method      string    `json:"method" gorm:"size:10;not null"`
	Path        string    `json:"path" gorm:"size:500;not null"`
	Params      string    `json:"params" gorm:"type:text"`
	Response    string    `json:"response" gorm:"type:text"`
	StatusCode  int       `json:"statusCode" gorm:"not null"`
	IP          string    `json:"ip" gorm:"size:45;not null"`
	UserAgent   string    `json:"userAgent" gorm:"size:500"`
	Duration    int       `json:"duration" gorm:"default:0"`
	Status      string    `json:"status" gorm:"size:20;not null;default:'success'"`
	ErrorMsg    string    `json:"errorMsg" gorm:"size:500"`
	OperateTime time.Time `json:"operateTime" gorm:"not null;index"`
	Time        string    `json:"time" gorm:"-"`
	CreatedAt   time.Time `json:"createdAt" gorm:"autoCreateTime"`
}

func (OperationLog) TableName() string { return "audit_operation_logs" }

// SystemEventLog 系统事件日志模型
type SystemEventLog struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Level     string    `json:"level" gorm:"size:20;not null;index"`
	Source    string    `json:"source" gorm:"size:100;not null;index"`
	Category  string    `json:"category" gorm:"size:50;index"`
	Message   string    `json:"message" gorm:"size:1000;not null"`
	Details   string    `json:"details" gorm:"type:text"`
	IP        string    `json:"ip" gorm:"size:45"`
	EventTime time.Time `json:"eventTime" gorm:"not null;index"`
	Time      string    `json:"time" gorm:"-"`
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
}

func (SystemEventLog) TableName() string { return "audit_system_event_logs" }
