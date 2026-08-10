package modelsystem

import "time"

// User 用户模型
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"uniqueIndex;size:50;not null"`
	Password  string    `json:"-" gorm:"size:255;not null"`
	Nickname  string    `json:"nickname" gorm:"size:50"`
	Avatar    string    `json:"avatar" gorm:"size:255"`
	Email     string    `json:"email" gorm:"size:100"`
	RoleIDs   string    `json:"roleIds" gorm:"type:json"`
	Status    string    `json:"status" gorm:"size:20;default:active"`
	HomePath  string    `json:"homePath" gorm:"size:100;default:/"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (User) TableName() string { return "sys_users" }
