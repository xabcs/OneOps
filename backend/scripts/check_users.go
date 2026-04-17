package main

import (
	"fmt"
	"oneops/backend/services"
)

func main() {
	// 初始化数据库
	s := services.NewInitService()
	err := s.InitDatabase()
	if err != nil {
		fmt.Printf("数据库初始化失败: %v\n", err)
		return
	}
	fmt.Println("数据库初始化完成")
	
	// 这里可以添加代码来查询用户数据
	// 但由于services包没有导出db变量，我们需要通过其他方式查看
}