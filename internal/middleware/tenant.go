package middleware

import (
	"github.com/anttna7/cjadmin/internal/utils"
	"github.com/gin-gonic/gin"
)

// TenantMiddleware 多租户中间件
// 确保非平台管理员只能访问自己租户的数据
func TenantMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 平台管理员可以访问所有租户数据，跳过检查
		if IsPlatformAdmin(c) {
			c.Next()
			return
		}

		// 非平台管理员必须属于某个租户
		tenantID := GetTenantID(c)
		if tenantID == nil {
			utils.ErrorWithStatus(c, 403, utils.CodeForbidden, "用户未关联到任何租户")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequirePlatformAdmin 要求平台管理员权限
func RequirePlatformAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !IsPlatformAdmin(c) {
			utils.ErrorWithStatus(c, 403, utils.CodeForbidden, "需要平台管理员权限")
			c.Abort()
			return
		}
		c.Next()
	}
}
