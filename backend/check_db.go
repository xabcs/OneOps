package main

import (
	"fmt"
	"log"
	"oneops/backend/config"
	"oneops/backend/models"
	"oneops/backend/services"
)

func main() {
	// 获取配置
	cfg := config.GetConfig()
	
	// 初始化数据库连接
	err := services.InitDB(&cfg.Database)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	fmt.Println("数据库连接成功")
	
	// 检查用户表数据
	var userCount int64
	db := services.GetDB()
	err = db.Model(&models.User{}).Count(&userCount).Error
	if err != nil {
		log.Fatalf("查询用户数量失败: %v", err)
	}
	fmt.Printf("用户表中的记录数量: %d\n", userCount)
	
	// 检查登录日志表数据
	var loginLogCount int64
	err = db.Model(&models.LoginLog{}).Count(&loginLogCount).Error
	if err != nil {
		log.Fatalf("查询登录日志数量失败: %v", err)
	}
	fmt.Printf("登录日志表中的记录数量: %d\n", loginLogCount)
	
	// 如果用户表为空，初始化数据
	if userCount == 0 {
		fmt.Println("用户表为空，开始初始化数据...")
		initService := services.NewInitService()
		err = initService.InitDatabase()
		if err != nil {
			log.Fatalf("初始化数据库失败: %v", err)
		}
		fmt.Println("数据库初始化完成")
	}
	
	// 再次检查用户数量
	err = db.Model(&models.User{}).Count(&userCount).Error
	if err != nil {
		log.Fatalf("查询用户数量失败: %v", err)
	}
	fmt.Printf("初始化后用户表中的记录数量: %d\n", userCount)
	
	// 显示用户信息
	var users []models.User
	err = db.Find(&users).Error
	if err != nil {
		log.Fatalf("查询用户信息失败: %v", err)
	}
	
	for _, user := range users {
		fmt.Printf("用户: %s (ID: %d), 密码: %s\n", user.Username, user.ID, user.Password)
	}
}