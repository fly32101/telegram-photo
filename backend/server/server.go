package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/telegram-photo/api/v1"
	"github.com/telegram-photo/config"
	"github.com/telegram-photo/middleware"
	"github.com/telegram-photo/model"
)

// New creates and configures the application server, returning the router and port.
func New() (*gin.Engine, string, error) {
	if err := config.Init(); err != nil {
		return nil, "", fmt.Errorf("配置初始化失败: %w", err)
	}

	if err := model.Init(); err != nil {
		return nil, "", fmt.Errorf("数据库初始化失败: %w", err)
	}

	router := gin.Default()
	registerMiddlewares(router)
	registerRoutes(router)

	port := viper.GetString("server.port")
	if port == "" {
		port = "8080"
	}

	return router, port, nil
}

func registerMiddlewares(r *gin.Engine) {
	r.Use(middleware.Cors(), middleware.TrustProxyHeaders())
}

func registerRoutes(r *gin.Engine) {
	v1.RegisterRoutes(r)

	r.Static("/assets", "./dist/assets")
	r.StaticFile("/favicon.ico", "./dist/favicon.ico")

	r.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") || strings.HasPrefix(c.Request.URL.Path, "/proxy/") {
			c.JSON(http.StatusNotFound, gin.H{"error": "API not found"})
			return
		}

		c.File("./dist/index.html")
	})
}
