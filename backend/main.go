package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/telegram-photo/config"
	"github.com/telegram-photo/model"
	"github.com/telegram-photo/api/v1"
	"github.com/telegram-photo/middleware"
)

func main() {
	// 加载配置
	if err := config.Init(); err != nil {
		log.Fatalf("配置初始化失败: %v", err)
	}

	// 初始化数据库
	if err := model.Init(); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}

	// 创建Gin引擎
	r := gin.Default()

	// 注册中间件
	r.Use(middleware.Cors())

	// 注册API路由
	v1.RegisterRoutes(r)

	// 启动服务器
	port := viper.GetString("server.port")
	if port == "" {
		port = "8080"
	}

	log.Printf("服务器启动在 http://localhost:%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}