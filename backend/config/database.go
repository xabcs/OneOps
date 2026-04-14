package config

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

// DB 全局数据库连接
var DB *sql.DB

// InitDatabase 初始化数据库连接
func InitDatabase() {
	// 先连接到 MySQL 服务器（不指定数据库）
	dsn := "root:123456@tcp(60.191.116.75:38089)/?charset=utf8mb4&parseTime=True&loc=Local"
	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to MySQL server: %v", err)
	}

	// 测试连接
	err = DB.Ping()
	if err != nil {
		log.Fatalf("Failed to ping MySQL server: %v", err)
	}

	// 创建数据库
	_, err = DB.Exec("CREATE DATABASE IF NOT EXISTS ops")
	if err != nil {
		log.Fatalf("Failed to create database: %v", err)
	}

	// 关闭连接
	DB.Close()

	// 重新连接到 ops 数据库
	dsn = "root:123456@tcp(60.191.116.75:38089)/ops?charset=utf8mb4&parseTime=True&loc=Local"
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 测试连接
	err = DB.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("Database connected successfully")

	// 创建用户表
	createUserTable()
}

// createUserTable 创建用户表
func createUserTable() {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id INT PRIMARY KEY AUTO_INCREMENT,
		username VARCHAR(50) NOT NULL UNIQUE,
		password VARCHAR(100) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	);
	`

	_, err := DB.Exec(query)
	if err != nil {
		log.Fatalf("Failed to create user table: %v", err)
	}

	// 检查是否有默认用户，如果没有则创建
	checkAndCreateDefaultUser()
}

// checkAndCreateDefaultUser 检查并创建默认用户
func checkAndCreateDefaultUser() {
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		log.Fatalf("Failed to check user count: %v", err)
	}

	if count == 0 {
		// 创建默认管理员用户
		_, err := DB.Exec(
			"INSERT INTO users (username, password) VALUES (?, ?)",
			"admin",
			"123456", // 实际生产环境应该使用密码哈希
		)
		if err != nil {
			log.Fatalf("Failed to create default user: %v", err)
		}
		fmt.Println("Default user created: admin/123456")
	}
}