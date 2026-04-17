package services

import (
	"database/sql"
	"fmt"
	"oneops/backend/config"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB

// InitDB 初始化数据库连接
func InitDB(cfg *config.DatabaseConfig) error {
	// 先连接到 MySQL 服务器（不指定数据库），创建数据库
	dsnWithoutDB := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port)

	sqlDB, err := sql.Open("mysql", dsnWithoutDB)
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	// 创建数据库
	_, err = sqlDB.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", cfg.DBName))
	if err != nil {
		return fmt.Errorf("创建数据库失败: %v", err)
	}

	// 连接到指定数据库
	dsn := cfg.GetDSN()
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	sqlDB2, err := db.DB()
	if err != nil {
		return err
	}

	// 设置连接池
	sqlDB2.SetMaxIdleConns(10)
	sqlDB2.SetMaxOpenConns(100)
	sqlDB2.SetConnMaxLifetime(time.Hour)

	return nil
}

// GetDB 获取数据库连接
func GetDB() *gorm.DB {
	return db
}
