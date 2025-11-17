package middleware

import (
	"strings"

	"github.com/anttna7/cjadmin/internal/utils"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware JWT认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.ErrorWithStatus(c, 401, utils.CodeUnauthorized, "未提供认证令牌")
			c.Abort()
			return
		}

		// 解析token
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			utils.ErrorWithStatus(c, 401, utils.CodeUnauthorized, "认证令牌格式错误")
			c.Abort()
			return
		}

		claims, err := utils.ParseToken(parts[1])
		if err != nil {
			utils.ErrorWithStatus(c, 401, utils.CodeUnauthorized, "无效的认证令牌")
			c.Abort()
			return
		}

		// 将用户信息存储到上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("tenant_id", claims.TenantID)
		c.Set("is_platform_admin", claims.IsPlatformAdmin)

		c.Next()
	}
}

// GetUserID 从上下文获取用户ID
func GetUserID(c *gin.Context) int64 {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0
	}
	return userID.(int64)
}

// GetTenantID 从上下文获取租户ID
func GetTenantID(c *gin.Context) *int64 {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		return nil
	}
	if tenantID == nil {
		return nil
	}
	id := tenantID.(int64)
	return &id
}

// IsPlatformAdmin 判断是否平台管理员
func IsPlatformAdmin(c *gin.Context) bool {
	isPlatformAdmin, exists := c.Get("is_platform_admin")
	if !exists {
		return false
	}
	return isPlatformAdmin.(bool)
}
