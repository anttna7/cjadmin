package middleware

import (
	"bytes"
	"io"
	"time"

	"cjadmin/internal/models"
	"cjadmin/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// responseWriter 包装gin的ResponseWriter以捕获响应数据
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// OperationLogger 创建详细的操作日志中间件
func OperationLogger(db *gorm.DB) gin.HandlerFunc {
	systemService := service.NewSystemService(db)

	return func(c *gin.Context) {
		// 开始时间
		startTime := time.Now()

		// 读取请求体（如果有）
		var requestBody string
		if c.Request.Body != nil {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			requestBody = string(bodyBytes)
			// 重新设置请求体，以便后续处理器可以读取
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			// 限制请求体长度
			if len(requestBody) > 5000 {
				requestBody = requestBody[:5000] + "...(truncated)"
			}
		}

		// 包装ResponseWriter以捕获响应
		blw := &responseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
		}
		c.Writer = blw

		// 处理请求
		c.Next()

		// 计算耗时
		duration := time.Since(startTime)

		// 获取当前用户信息
		var userID int64
		var username string
		if userIDVal, exists := c.Get("user_id"); exists {
			userID = userIDVal.(int64)
		}
		if usernameVal, exists := c.Get("username"); exists {
			username = usernameVal.(string)
		}

		// 获取租户ID
		var tenantID int64
		if tenantIDVal, exists := c.Get("tenant_id"); exists {
			tenantID = tenantIDVal.(int64)
		}

		// 获取响应数据
		responseBody := blw.body.String()
		if len(responseBody) > 5000 {
			responseBody = responseBody[:5000] + "...(truncated)"
		}

		// 确定操作类型
		operationType := determineOperationType(c.Request.Method, c.Request.URL.Path)

		// 创建审计日志
		req := &service.CreateAuditLogRequest{
			OperationType: operationType,
			OperationDesc: buildOperationDescription(c.Request.Method, c.Request.URL.Path, c.Writer.Status()),
			ResourceType:  extractResourceType(c.Request.URL.Path),
			ResourceID:    extractResourceID(c),
			RequestMethod: c.Request.Method,
			RequestPath:   c.Request.URL.Path,
			RequestParams: c.Request.URL.RawQuery,
			RequestBody:   requestBody,
			ResponseCode:  c.Writer.Status(),
			ResponseBody:  responseBody,
			Duration:      int(duration.Milliseconds()),
			UserAgent:     c.Request.UserAgent(),
		}

		// 记录日志（异步，不影响主流程）
		go func() {
			_ = systemService.CreateAuditLog(
				tenantID,
				userID,
				req,
				c.ClientIP(),
				c.Request.UserAgent(),
			)
		}()
	}
}

// determineOperationType 根据HTTP方法和路径确定操作类型
func determineOperationType(method, path string) string {
	switch method {
	case "POST":
		return "create"
	case "PUT", "PATCH":
		return "update"
	case "DELETE":
		return "delete"
	case "GET":
		if containsID(path) {
			return "view"
		}
		return "query"
	default:
		return "other"
	}
}

// buildOperationDescription 构建操作描述
func buildOperationDescription(method, path string, statusCode int) string {
	action := ""
	switch method {
	case "POST":
		action = "创建"
	case "PUT", "PATCH":
		action = "更新"
	case "DELETE":
		action = "删除"
	case "GET":
		action = "查询"
	default:
		action = "操作"
	}

	resource := extractResourceType(path)
	status := "成功"
	if statusCode >= 400 {
		status = "失败"
	}

	return action + resource + status
}

// extractResourceType 从路径中提取资源类型
func extractResourceType(path string) string {
	// /api/customers -> customers
	// /api/orders/:id -> orders
	parts := splitPath(path)
	if len(parts) >= 2 && parts[0] == "api" {
		return parts[1]
	}
	return "unknown"
}

// extractResourceID 从上下文中提取资源ID
func extractResourceID(c *gin.Context) *string {
	// 尝试从路径参数获取ID
	if id := c.Param("id"); id != "" {
		return &id
	}

	// 尝试从查询参数获取ID
	if id := c.Query("id"); id != "" {
		return &id
	}

	return nil
}

// containsID 检查路径是否包含ID参数
func containsID(path string) bool {
	parts := splitPath(path)
	for _, part := range parts {
		if part == ":id" || isNumeric(part) {
			return true
		}
	}
	return false
}

// splitPath 分割路径
func splitPath(path string) []string {
	var parts []string
	current := ""
	for _, char := range path {
		if char == '/' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

// isNumeric 检查字符串是否为数字
func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, char := range s {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}
