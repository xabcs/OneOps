package modelauth

import (
	"time"
)

// ApplicationOperationLog 应用操作日志
type ApplicationOperationLog struct {
	ID           uint        `json:"id" gorm:"primaryKey"`
	AppID        uint        `json:"appId" gorm:"not null;index"`
	AppIDField   Application `json:"-" gorm:"foreignKey:AppID"`
	Operation    string      `json:"operation" gorm:"size:50;not null"`
	Target       string      `json:"target" gorm:"size:100"`
	RequestData  string      `json:"requestData" gorm:"type:text"`
	ResponseData string      `json:"responseData" gorm:"type:text"`
	Status       string      `json:"status" gorm:"size:20"`
	ErrorMsg     string      `json:"errorMsg" gorm:"type:text"`
	Operator     string      `json:"operator" gorm:"size:50"`
	CreatedAt    time.Time   `json:"createdAt"`
}

func (ApplicationOperationLog) TableName() string {
	return "auth_operation_logs"
}
