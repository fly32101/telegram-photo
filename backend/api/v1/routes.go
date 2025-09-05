package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/telegram-photo/middleware"
)

// RegisterRoutes 注册API路由
func RegisterRoutes(r *gin.Engine) {
	// API版本分组
	v1 := r.Group("/api/v1")

	// 认证相关路由
	auth := v1.Group("/auth")
	{
		auth.GET("/github", redirectToGitHub)
		auth.GET("/github/callback", githubCallback)
	}

	// 图片相关路由（需要认证）
	image := v1.Group("/image")
	image.Use(middleware.JWTAuth())
	{
		image.POST("/upload", uploadImage)
		image.GET("/list", listImages)
		image.DELETE("/:id", deleteImage)
	}

	// 管理员路由
	admin := v1.Group("/admin")
	admin.Use(middleware.JWTAuth(), middleware.AdminAuth())
	{
		admin.GET("/images", adminListImages)
		admin.GET("/stats", getStats)
	}

	// 代理访问路由
	proxy := r.Group("/proxy")
	{
		proxy.GET("/image/:file_id", proxyImage)
	}
}