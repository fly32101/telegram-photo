package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/spf13/viper"
)

// Cors 跨域中间件
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
			c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers")
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// TrustProxyHeaders 信任代理头中间件
func TrustProxyHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置信任所有代理
		// 这样Gin的c.ClientIP()会正确处理X-Forwarded-For和X-Real-IP头
		// 注意：在生产环境中，应该只信任特定的代理
		
		// 直接修改请求的RemoteAddr为X-Real-IP的值（如果存在）
		if realIP := c.Request.Header.Get("X-Real-IP"); realIP != "" {
			c.Request.RemoteAddr = realIP
		} else if forwardedFor := c.Request.Header.Get("X-Forwarded-For"); forwardedFor != "" {
			// 如果有多个IP，取第一个（最原始的客户端IP）
			ips := strings.Split(forwardedFor, ",")
			if len(ips) > 0 {
				c.Request.RemoteAddr = strings.TrimSpace(ips[0])
			}
		}
		
		c.Next()
	}
}

// JWTAuth JWT认证中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供认证令牌"})
			c.Abort()
			return
		}

		parts := strings.SplitN(auth, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "认证格式错误"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(viper.GetString("jwt.secret")), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的认证令牌"})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user_id", claims.UserID)
		c.Next()
	}
}

// AdminAuth 管理员认证中间件
func AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
			c.Abort()
			return
		}

		// 从配置中获取管理员ID列表
		adminIDs := viper.GetStringSlice("admin.user_ids")
		isAdmin := false

		for _, adminID := range adminIDs {
			if userID == adminID {
				isAdmin = true
				break
			}
		}

		if !isAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "无管理员权限"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Claims JWT声明结构
type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}