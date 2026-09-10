package database

import (
	"database/sql"
	"fmt"
	"time"

	"oneops/backend3/config"

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
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		// 迁移时不自动创建外键约束，避免因表顺序或已有约束导致 AutoMigrate 失败
		DisableForeignKeyConstraintWhenMigrating: true,
		// 单语句写不再自动包 BEGIN/COMMIT（省 2 个网络往返）；需要原子性的
		// 多语句写仍走显式 db.Transaction，不受此开关影响
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return err
	}

	sqlDB2, err := db.DB()
	if err != nil {
		return err
	}

	// 设置连接池
	sqlDB2.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB2.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB2.SetConnMaxLifetime(time.Hour)
	sqlDB2.SetConnMaxIdleTime(30 * time.Minute) // 空闲连接最大存活时间

	return nil
}

// GetDB 获取数据库连接
func GetDB() *gorm.DB {
	return db
}
