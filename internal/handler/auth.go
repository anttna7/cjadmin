package handler

import (
	"github.com/anttna7/cjadmin/internal/config"
	"github.com/anttna7/cjadmin/internal/service"
	"github.com/anttna7/cjadmin/internal/utils"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		authService: service.NewAuthService(cfg),
	}
}

// Login 登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.CodeInvalidParams, "参数错误: "+err.Error())
		return
	}

	resp, err := h.authService.Login(&req)
	if err != nil {
		utils.Error(c, utils.CodeUnauthorized, err.Error())
		return
	}

	utils.Success(c, resp)
}

// Register 注册（创建用户）
func (h *AuthHandler) Register(c *gin.Context) {
	var req service.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.CodeInvalidParams, "参数错误: "+err.Error())
		return
	}

	user, err := h.authService.Register(&req)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "用户创建成功", user)
}

// GetCurrentUser 获取当前登录用户信息
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")
	tenantID, _ := c.Get("tenant_id")
	isPlatformAdmin, _ := c.Get("is_platform_admin")

	utils.Success(c, gin.H{
		"id":                userID,
		"username":          username,
		"tenant_id":         tenantID,
		"is_platform_admin": isPlatformAdmin,
	})
}
