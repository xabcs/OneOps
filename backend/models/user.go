package models

import (
	"time"
)

// User 用户模型
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"uniqueIndex;size:50;not null"`
	Password  string    `json:"-" gorm:"size:255;not null"` // 不在 JSON 中返回密码
	Nickname  string    `json:"nickname" gorm:"size:50"`
	Avatar    string    `json:"avatar" gorm:"size:255"`
	Email     string    `json:"email" gorm:"size:100"`
	RoleIDs   string    `json:"roleIds" gorm:"type:json"` // 存储为 JSON 数组字符串 "[1,2,3]"
	Status    string    `json:"status" gorm:"size:20;default:active"`
	HomePath  string    `json:"homePath" gorm:"size:100;default:/"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}
