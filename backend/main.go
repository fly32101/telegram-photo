package main

import (
	"log"

	"github.com/telegram-photo/server"
)

func main() {
	router, port, err := server.New()
	if err != nil {
		log.Fatalf("服务器初始化失败: %v", err)
	}

	log.Printf("服务器启动在 http://localhost:%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
