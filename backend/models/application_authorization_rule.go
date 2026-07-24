package models

import (
	"time"
)

// ApplicationAuthorizationRule 应用授权规则模型（针对 JumpServer 等使用授权规则的系统）
type ApplicationAuthorizationRule struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	AppID        uint      `json:"appId" gorm:"not null;index"`
	AppIDField   Application `json:"-" gorm:"foreignKey:AppID"`
	RuleID       string    `json:"ruleId" gorm:"size:100;not null;index"`        // 外部系统的规则ID
	RuleName     string    `json:"ruleName" gorm:"size:200;not null"`           // 规则名称
	RuleType     string    `json:"ruleType" gorm:"size:50;default:user"`         // 规则类型: user(用户), group(用户组)

	// 授权主体
	SubjectType  string `json:"subjectType" gorm:"size:50;not null"`             // 主体类型: user, group
	SubjectID    string `json:"subjectId" gorm:"size:100;not null"`              // 用户ID或用户组ID
	SubjectName  string `json:"subjectName" gorm:"size:100"`                    // 用户名或用户组名（用于显示）

	// 授权对象
	ObjectType   string `json:"objectType" gorm:"size:50;not null"`              // 对象类型: asset(资产), node(节点), asset_category(资产类别), system(全部)
	ObjectID     string `json:"objectId" gorm:"size:100"`                        // 资产/节点ID（如果是all则为空）
	ObjectName   string `json:"objectName" gorm:"size:200"`                      // 资产/节点名称（用于显示）

	// 操作权限
	Actions      string `json:"actions" gorm:"type:text"`                         // 操作权限列表(JSON): ["connect", "upload", "download", "command"]

	// 优先级和状态
	Priority    int       `json:"priority" gorm:"default:0"`                      // 优先级（数字越大优先级越高）
	IsEnabled    bool      `json:"isEnabled" gorm:"default:true"`                  // 是否启用
	IsExpired    bool      `json:"isExpired" gorm:"default:false"`                  // 是否过期

	// 时间信息
	SyncTime     time.Time `json:"syncTime"`
	ExpireTime   *time.Time `json:"expireTime"`                                  // 过期时间
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

func (ApplicationAuthorizationRule) TableName() string {
	return "application_authorization_rules"
}
