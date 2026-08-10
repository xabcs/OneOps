package modelauth

import (
	"time"
)

// AuthUser 授权中心用户模型（独立于系统登录用户）
type AuthUser struct {
	ID          uint        `json:"id" gorm:"primaryKey"`
	Username    string      `json:"username" gorm:"size:50;not null;uniqueIndex"`
	Password    string      `json:"-" gorm:"size:255;comment:'初始密码（加密存储）'"` // 不返回给前端
	Nickname    string      `json:"nickname" gorm:"size:50"`
	Email       string      `json:"email" gorm:"size:100"`
	Phone       string      `json:"phone" gorm:"size:20"`
	Description string      `json:"description" gorm:"size:200"`
	Status      int         `json:"status" gorm:"default:1;comment:1=启用,0=禁用"`
	Groups      []AuthGroup `json:"groups" gorm:"many2many:auth_user_groups;joinForeignKey:UserID;joinReferences:GroupID;association_foreignkey:ID;references:ID"`
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt"`
}

func (AuthUser) TableName() string {
	return "auth_users"
}

// AuthGroup 授权中心用户组模型（独立于系统角色）
type AuthGroup struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"size:50;not null;uniqueIndex"`
	Code        string    `json:"code" gorm:"size:50;not null;uniqueIndex"`
	Description string    `json:"description" gorm:"size:200"`
	Status      int       `json:"status" gorm:"default:1;comment:1=启用,0=禁用"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (AuthGroup) TableName() string {
	return "auth_groups"
}

// AuthUserGroup 授权中心用户组成员关联模型
type AuthUserGroup struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	UserID       uint      `json:"userId" gorm:"not null;uniqueIndex:idx_user_group;comment:授权中心用户ID"`
	UserIDField  AuthUser  `json:"user,omitempty" gorm:"foreignKey:UserID"`
	GroupID      uint      `json:"groupId" gorm:"not null;uniqueIndex:idx_user_group;comment:授权中心用户组ID"`
	GroupIDField AuthGroup `json:"group,omitempty" gorm:"foreignKey:GroupID"`
	GrantedBy    string    `json:"grantedBy" gorm:"size:50"`
	GrantedAt    time.Time `json:"grantedAt"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

func (AuthUserGroup) TableName() string {
	return "auth_user_groups"
}
