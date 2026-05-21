package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// AssetAccessPolicy 资产访问策略
type AssetAccessPolicy struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	Name             string    `gorm:"size:100;not null" json:"name"`
	SubjectType      string    `gorm:"type:enum('user','role','user_group');not null" json:"subjectType"`
	SubjectID        uint      `gorm:"not null" json:"subjectId"`
	AssetScopeType   string    `gorm:"type:enum('server','group','business','tag','all');not null" json:"assetScopeType"`
	AssetScopeID     uint      `gorm:"default:0" json:"assetScopeId"`
	LoginAccounts    StringArray `gorm:"type:json" json:"loginAccounts"`
	Protocols        StringArray `gorm:"type:json" json:"protocols"`
	AllowFileTransfer bool      `gorm:"default:true" json:"allowFileTransfer"`
	AllowSudo        bool      `gorm:"default:false" json:"allowSudo"`
	RequireApproval  bool      `gorm:"default:false" json:"requireApproval"`
	TimeWindow       TimeWindow `gorm:"type:json" json:"timeWindow"`
	HighRiskCommands StringArray `gorm:"type:json" json:"highRiskCommands"`
	Status           int       `gorm:"default:1" json:"status"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

// TimeWindow 时间窗口
type TimeWindow struct {
	Start string `json:"start"` // HH:mm格式
	End   string `json:"end"`   // HH:mm格式
	Days  []int  `json:"days"`  // 1-7, 1=周一
}

// StringArray 用于存储JSON数组
type StringArray []string

// Scan 实现 sql.Scanner 接口
func (sa *StringArray) Scan(value interface{}) error {
	if value == nil {
		*sa = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, sa)
}

// Value 实现 driver.Valuer 接口
func (sa StringArray) Value() (driver.Value, error) {
	if len(sa) == 0 {
		return nil, nil
	}
	return json.Marshal(sa)
}

// BastionSession 堡垒机会话
type BastionSession struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	ServerID        uint           `gorm:"not null" json:"serverId"`
	Server          *Server        `gorm:"foreignKey:ServerID" json:"server,omitempty"`
	UserID          uint           `gorm:"not null" json:"userId"`
	User            *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Username        string         `gorm:"size:50;not null" json:"username"`
	LoginAccount    string         `gorm:"size:50;not null" json:"loginAccount"`
	ClientIP        string         `gorm:"size:50" json:"clientIp"`
	Protocol        string         `gorm:"type:enum('ssh','sftp');default:'ssh'" json:"protocol"`
	SSHCredentialID uint           `gorm:"null" json:"sshCredentialId"`
	SSHCredential   *SSHCredential `gorm:"foreignKey:SSHCredentialID" json:"sshCredential,omitempty"`
	StartedAt       *time.Time     `json:"startedAt"`
	EndedAt         *time.Time     `json:"endedAt"`
	Duration        int            `gorm:"default:0" json:"duration"`
	Status          string         `gorm:"type:enum('active','closed','error','terminated');default:'active'" json:"status"`
	CloseReason     string         `gorm:"size:200" json:"closeReason"`
	CreatedAt       time.Time      `json:"createdAt"`

	// 关联数据（不存储在数据库）
	Commands       []BastionCommand       `gorm:"-" json:"commands,omitempty"`
	FileTransfers  []BastionFileTransfer  `gorm:"-" json:"fileTransfers,omitempty"`
}

// BastionCommand 命令审计
type BastionCommand struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	SessionID    uint       `gorm:"not null" json:"sessionId"`
	Session      *BastionSession `gorm:"foreignKey:SessionID" json:"session,omitempty"`
	Command      string     `gorm:"type:text;not null" json:"command"`
	ExecutedAt   *time.Time `json:"executedAt"`
	ExitCode     *int       `json:"exitCode"`
	RiskLevel    string     `gorm:"type:enum('safe','low','medium','high','critical');default:'safe'" json:"riskLevel"`
	Blocked      bool       `gorm:"default:false" json:"blocked"`
	OutputSummary string    `gorm:"type:text" json:"outputSummary"`
	CreatedAt    time.Time  `json:"createdAt"`
}

// BastionFileTransfer 文件传输审计
type BastionFileTransfer struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	SessionID    uint       `gorm:"not null" json:"sessionId"`
	Session      *BastionSession `gorm:"foreignKey:SessionID" json:"session,omitempty"`
	Direction    string     `gorm:"type:enum('upload','download');not null" json:"direction"`
	RemotePath   string     `gorm:"size:500;not null" json:"remotePath"`
	LocalPath    string     `gorm:"size:500" json:"localPath"`
	FileSize     int64      `gorm:"default:0" json:"fileSize"`
	Status       string     `gorm:"type:enum('pending','transferring','success','failed');default:'pending'" json:"status"`
	ErrorMessage string     `gorm:"size:500" json:"errorMessage"`
	StartedAt    *time.Time `json:"startedAt"`
	CompletedAt  *time.Time `json:"completedAt"`
	CreatedAt    time.Time  `json:"createdAt"`
}

// BastionApproval 连接审批
type BastionApproval struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	RequesterID  uint       `gorm:"not null" json:"requesterId"`
	Requester    *User      `gorm:"foreignKey:RequesterID" json:"requester,omitempty"`
	ApproverID   *uint      `gorm:"null" json:"approverId"`
	Approver     *User      `gorm:"foreignKey:ApproverID" json:"approver,omitempty"`
	ServerID     uint       `gorm:"not null" json:"serverId"`
	Server       *Server    `gorm:"foreignKey:ServerID" json:"server,omitempty"`
	LoginAccount string     `gorm:"size:50;not null" json:"loginAccount"`
	Reason       string     `gorm:"type:text" json:"reason"`
	Status       string     `gorm:"type:enum('pending','approved','rejected','expired');default:'pending'" json:"status"`
	RequestedAt  *time.Time `json:"requestedAt"`
	ApprovedAt   *time.Time `json:"approvedAt"`
	ExpiredAt    *time.Time `json:"expiredAt"`
	Comment      string     `gorm:"type:text" json:"comment"`
	CreatedAt    time.Time  `json:"createdAt"`
}

// ConnectRequest 连接请求
type ConnectRequest struct {
	Protocol     string `json:"protocol" binding:"required,oneof=ssh sftp"`
	LoginAccount string `json:"loginAccount" binding:"required"`
}

// ConnectResponse 连接响应
type ConnectResponse struct {
	SessionID    uint   `json:"sessionId"`
	WebSocketURL string `json:"websocketUrl"`
	ServerName   string `json:"serverName"`
	ServerIP     string `json:"serverIp"`
}

// SessionFilter 会话筛选条件
type SessionFilter struct {
	ServerID     *uint    `json:"serverId"`
	UserID       *uint    `json:"userId"`
	Status       *string  `json:"status"`
	Protocol     *string  `json:"protocol"`
	StartDate    *string  `json:"startDate"`
	EndDate      *string  `json:"endDate"`
	ClientIP     *string  `json:"clientIp"`
	LoginAccount *string  `json:"loginAccount"`
}

// CommandFilter 命令筛选条件
type CommandFilter struct {
	SessionID   *uint   `json:"sessionId"`
	RiskLevel   *string `json:"riskLevel"`
	Blocked     *bool   `json:"blocked"`
	StartDate   *string `json:"startDate"`
	EndDate     *string `json:"endDate"`
	CommandLike *string `json:"commandLike"`
}

// FileTransferFilter 文件传输筛选条件
type FileTransferFilter struct {
	SessionID *uint   `json:"sessionId"`
	Direction *string `json:"direction"`
	Status    *string `json:"status"`
	StartDate *string `json:"startDate"`
	EndDate   *string `json:"endDate"`
}

// ApprovalFilter 审批筛选条件
type ApprovalFilter struct {
	RequesterID *uint   `json:"requesterId"`
	ApproverID  *uint   `json:"approverId"`
	ServerID    *uint   `json:"serverId"`
	Status      *string `json:"status"`
	StartDate   *string `json:"startDate"`
	EndDate     *string `json:"endDate"`
}
