package modelauth

import (
	"time"
)

// UserIdentityMapping 用户身份映射表（授权中心用户 ↔ 外部应用用户）
type UserIdentityMapping struct {
	ID               uint        `json:"id" gorm:"primaryKey"`
	AuthUserID       uint        `json:"authUserId" gorm:"not null;index;comment:授权中心用户ID"`
	AuthUserField    AuthUser    `json:"authUser,omitempty" gorm:"foreignKey:AuthUserID"`
	AppID            uint        `json:"appId" gorm:"not null;index;comment:应用ID"`
	AppIDField       Application `json:"-" gorm:"foreignKey:AppID"`
	ExternalUsername string      `json:"externalUsername" gorm:"size:100;not null;index;comment:外部应用用户名"`
	ExternalUserID   string      `json:"externalUserId" gorm:"size:100;comment:外部系统用户ID（如果有的话）"`
	MappingType      string      `json:"mappingType" gorm:"size:20;default:auto;comment:auto=自动创建,manual=手动创建"`
	MappingStatus    string      `json:"mappingStatus" gorm:"size:20;default:active;comment:active=激活,inactive=禁用,deleted=已删除"`
	LastSyncAt       *time.Time  `json:"lastSyncAt" gorm:"comment:最后同步时间"`
	CreatedAt        time.Time   `json:"createdAt"`
	UpdatedAt        time.Time   `json:"updatedAt"`
}

func (UserIdentityMapping) TableName() string {
	return "auth_user_identity_mappings"
}
