package middleware

import (
	"log"
	"net"
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
		// 记录原始请求信息，用于调试
		originalRemoteAddr := c.Request.RemoteAddr
		xRealIP := c.Request.Header.Get("X-Real-IP")
		xForwardedFor := c.Request.Header.Get("X-Forwarded-For")
		
		log.Printf("[IP调试] 原始RemoteAddr: %s, X-Real-IP: %s, X-Forwarded-For: %s", 
			originalRemoteAddr, xRealIP, xForwardedFor)
		
		// 优先使用X-Forwarded-For的第一个IP
		if xForwardedFor != "" {
			// 如果有多个IP，取第一个（最原始的客户端IP）
			ips := strings.Split(xForwardedFor, ",")
			if len(ips) > 0 {
				ip := strings.TrimSpace(ips[0])
				// 确保IP格式正确
				if net.ParseIP(ip) != nil {
					c.Request.RemoteAddr = ip
					log.Printf("[IP调试] 使用X-Forwarded-For设置RemoteAddr: %s", ip)
				}
			}
		} else if xRealIP != "" {
			// 确保IP格式正确
			if net.ParseIP(xRealIP) != nil {
				c.Request.RemoteAddr = xRealIP
				log.Printf("[IP调试] 使用X-Real-IP设置RemoteAddr: %s", xRealIP)
			}
		}
		
		// 记录最终使用的IP
		log.Printf("[IP调试] 最终RemoteAddr: %s, ClientIP(): %s", c.Request.RemoteAddr, c.ClientIP())
		
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