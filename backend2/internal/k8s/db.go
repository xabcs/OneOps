package k8s

import (
	"oneops/backend2/pkg/database"

	"gorm.io/gorm"
)

// db 包级别数据库连接（由 main.go 在启动时通过 SetDB 注入）
var db *gorm.DB

// SetDB 设置包级别数据库连接
func SetDB(d *gorm.DB) { db = d }

// GetDB 获取数据库连接（懒加载：若未注入则从 pkg/database 获取）
func GetDB() *gorm.DB {
	if db == nil {
		db = database.GetDB()
	}
	return db
}
