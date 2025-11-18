package handler

import (
	"cjadmin/internal/database"
	"cjadmin/internal/models"
	"cjadmin/internal/service"
	"cjadmin/internal/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// SecurityHandler 安全相关处理器
type SecurityHandler struct {
	securityService *service.SecurityService
	oauthService    *service.OAuthService
}

// NewSecurityHandler 创建安全处理器实例
func NewSecurityHandler() *SecurityHandler {
	db := database.GetDB()
	return &SecurityHandler{
		securityService: service.NewSecurityService(db),
		oauthService:    service.NewOAuthService(db),
	}
}

// ========== 验证码相关 ==========

// SendSMSCodeRequest 发送验证码请求
type SendSMSCodeRequest struct {
	Phone    string `json:"phone" binding:"required"`
	CodeType string `json:"code_type" binding:"required"`
}

// SendSMSCode 发送短信验证码
func (h *SecurityHandler) SendSMSCode(c *gin.Context) {
	var req SendSMSCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	// 获取租户ID
	var tenantID *int64
	if tid, exists := c.Get("tenant_id"); exists {
		id := tid.(int64)
		tenantID = &id
	}

	// 获取客户端IP
	ip := service.GetClientIP(
		c.GetHeader("X-Forwarded-For"),
		c.GetHeader("X-Real-IP"),
		c.Request.RemoteAddr,
	)

	// 检查黑名单
	if blocked, reason := h.securityService.CheckBlacklist(tenantID, ip, "", "", req.Phone); blocked {
		utils.Forbidden(c, reason)
		return
	}

	// 生成验证码
	smsCode, err := h.securityService.GenerateSMSCode(
		tenantID,
		req.Phone,
		req.CodeType,
		ip,
		c.GetHeader("User-Agent"),
	)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	// TODO: 调用短信服务发送验证码
	// 这里只是返回成功，实际需要集成短信服务商API

	utils.Success(c, gin.H{
		"message":    "验证码已发送",
		"expires_at": smsCode.ExpiresAt,
		// 开发环境可以返回验证码，生产环境应删除
		"code": smsCode.Code,
	})
}

// LoginBySMSRequest 验证码登录请求
type LoginBySMSRequest struct {
	Phone string `json:"phone" binding:"required"`
	Code  string `json:"code" binding:"required"`
}

// LoginBySMS 验证码登录
func (h *SecurityHandler) LoginBySMS(c *gin.Context) {
	var req LoginBySMSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	// 获取客户端信息
	ip := service.GetClientIP(
		c.GetHeader("X-Forwarded-For"),
		c.GetHeader("X-Real-IP"),
		c.Request.RemoteAddr,
	)

	// 检查黑名单
	if blocked, reason := h.securityService.CheckBlacklist(nil, ip, "", "", req.Phone); blocked {
		h.securityService.RecordLoginAttempt(nil, req.Phone, models.IdentifierTypePhone, ip, c.GetHeader("User-Agent"), "", false, reason)
		utils.Forbidden(c, reason)
		return
	}

	// 检查登录尝试次数
	if blocked, remaining := h.securityService.CheckLoginAttempts(req.Phone, ip, 5, 15*time.Minute); blocked {
		utils.TooManyRequests(c, "登录尝试次数过多，请15分钟后再试")
		return
	} else if remaining <= 2 {
		// 提醒用户剩余尝试次数
	}

	// 验证验证码
	if err := h.securityService.VerifySMSCode(req.Phone, req.Code, models.CodeTypeLogin); err != nil {
		h.securityService.RecordLoginAttempt(nil, req.Phone, models.IdentifierTypePhone, ip, c.GetHeader("User-Agent"), "", false, "验证码错误")
		utils.BadRequest(c, err.Error())
		return
	}

	// 查找或创建用户
	db := database.GetDB()
	var user models.User
	err := db.Where("phone = ? AND deleted_at IS NULL", req.Phone).First(&user).Error
	if err != nil {
		// 用户不存在，返回错误或自动注册
		utils.NotFound(c, "用户不存在，请先注册")
		return
	}

	// 检查用户状态
	if user.Status != "active" {
		h.securityService.RecordLoginAttempt(&user.TenantID, req.Phone, models.IdentifierTypePhone, ip, c.GetHeader("User-Agent"), "", false, "账户已禁用")
		utils.Forbidden(c, "账户已禁用")
		return
	}

	// 生成Token
	token, err := utils.GenerateToken(user.ID, user.TenantID, user.Username, user.IsPlatformAdmin)
	if err != nil {
		utils.ServerError(c, "生成Token失败")
		return
	}

	// 更新最后登录信息
	now := time.Now()
	db.Model(&user).Updates(map[string]interface{}{
		"last_login_at": now,
		"last_login_ip": ip,
		"login_method":  models.LoginMethodPhone,
	})

	// 记录成功登录
	h.securityService.RecordLoginAttempt(&user.TenantID, req.Phone, models.IdentifierTypePhone, ip, c.GetHeader("User-Agent"), "", true, "")

	utils.Success(c, gin.H{
		"token": token,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"nickname": user.Nickname,
			"phone":    user.Phone,
		},
	})
}

// ========== OAuth相关 ==========

// GetOAuthURL 获取OAuth授权URL
func (h *SecurityHandler) GetOAuthURL(c *gin.Context) {
	provider := c.Query("provider")
	if provider == "" {
		utils.BadRequest(c, "请指定OAuth提供商")
		return
	}

	// 获取租户ID
	tenantID, _ := c.Get("tenant_id")
	tid := tenantID.(int64)

	// 生成state（可以用于CSRF防护）
	state := utils.GenerateRandomString(32)
	// TODO: 将state存储到Redis或Session中

	authURL, err := h.oauthService.GetAuthURL(tid, provider, state)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.Success(c, gin.H{
		"auth_url": authURL,
		"state":    state,
	})
}

// OAuthCallback OAuth回调处理
func (h *SecurityHandler) OAuthCallback(c *gin.Context) {
	provider := c.Param("provider")
	code := c.Query("code")
	// state := c.Query("state") // TODO: 验证state

	if code == "" {
		utils.BadRequest(c, "授权码不能为空")
		return
	}

	// 从query获取租户ID（或从state中解析）
	tenantIDStr := c.Query("tenant_id")
	tenantID, _ := strconv.ParseInt(tenantIDStr, 10, 64)

	if tenantID == 0 {
		utils.BadRequest(c, "租户ID不能为空")
		return
	}

	// 获取用户信息
	userInfo, err := h.oauthService.HandleCallback(tenantID, provider, code)
	if err != nil {
		utils.ServerError(c, err.Error())
		return
	}

	// 查找或创建用户
	user, isNew, err := h.oauthService.FindOrCreateUser(tenantID, provider, userInfo)
	if err != nil {
		utils.ServerError(c, err.Error())
		return
	}

	// 生成Token
	token, err := utils.GenerateToken(user.ID, user.TenantID, user.Username, user.IsPlatformAdmin)
	if err != nil {
		utils.ServerError(c, "生成Token失败")
		return
	}

	// 更新最后登录信息
	ip := service.GetClientIP(
		c.GetHeader("X-Forwarded-For"),
		c.GetHeader("X-Real-IP"),
		c.Request.RemoteAddr,
	)
	now := time.Now()
	database.GetDB().Model(user).Updates(map[string]interface{}{
		"last_login_at": now,
		"last_login_ip": ip,
		"login_method":  models.LoginMethodOAuth,
	})

	utils.Success(c, gin.H{
		"token":  token,
		"is_new": isNew,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"nickname": user.Nickname,
			"avatar":   user.Avatar,
		},
	})
}

// ========== 黑名单管理 ==========

// AddBlacklistRequest 添加黑名单请求
type AddBlacklistRequest struct {
	RuleType  string     `json:"rule_type" binding:"required"`
	RuleValue string     `json:"rule_value" binding:"required"`
	Reason    string     `json:"reason"`
	ExpiresAt *time.Time `json:"expires_at"`
}

// AddBlacklist 添加黑名单规则
func (h *SecurityHandler) AddBlacklist(c *gin.Context) {
	var req AddBlacklistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	// 获取租户ID
	var tenantID *int64
	if tid, exists := c.Get("tenant_id"); exists {
		id := tid.(int64)
		tenantID = &id
	}

	userID, _ := c.Get("user_id")

	rule, err := h.securityService.AddBlacklistRule(
		tenantID,
		req.RuleType,
		req.RuleValue,
		req.Reason,
		req.ExpiresAt,
		userID.(int64),
	)
	if err != nil {
		utils.ServerError(c, "添加黑名单失败")
		return
	}

	utils.Success(c, rule)
}

// ListBlacklist 获取黑名单列表
func (h *SecurityHandler) ListBlacklist(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	var tenantID *int64
	if tid, exists := c.Get("tenant_id"); exists {
		id := tid.(int64)
		tenantID = &id
	}

	rules, total, err := h.securityService.ListBlacklist(tenantID, page, pageSize)
	if err != nil {
		utils.ServerError(c, "获取黑名单列表失败")
		return
	}

	utils.SuccessWithPage(c, rules, total, page, pageSize)
}

// DeleteBlacklist 删除黑名单规则
func (h *SecurityHandler) DeleteBlacklist(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	if err := h.securityService.RemoveBlacklistRule(id); err != nil {
		utils.ServerError(c, "删除黑名单失败")
		return
	}

	utils.Success(c, nil)
}

// ========== 白名单管理 ==========

// AddWhitelistRequest 添加白名单请求
type AddWhitelistRequest struct {
	RuleType    string `json:"rule_type" binding:"required"`
	RuleValue   string `json:"rule_value" binding:"required"`
	Description string `json:"description"`
}

// AddWhitelist 添加白名单规则
func (h *SecurityHandler) AddWhitelist(c *gin.Context) {
	var req AddWhitelistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	var tenantID *int64
	if tid, exists := c.Get("tenant_id"); exists {
		id := tid.(int64)
		tenantID = &id
	}

	userID, _ := c.Get("user_id")

	rule, err := h.securityService.AddWhitelistRule(
		tenantID,
		req.RuleType,
		req.RuleValue,
		req.Description,
		userID.(int64),
	)
	if err != nil {
		utils.ServerError(c, "添加白名单失败")
		return
	}

	utils.Success(c, rule)
}

// ListWhitelist 获取白名单列表
func (h *SecurityHandler) ListWhitelist(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	var tenantID *int64
	if tid, exists := c.Get("tenant_id"); exists {
		id := tid.(int64)
		tenantID = &id
	}

	rules, total, err := h.securityService.ListWhitelist(tenantID, page, pageSize)
	if err != nil {
		utils.ServerError(c, "获取白名单列表失败")
		return
	}

	utils.SuccessWithPage(c, rules, total, page, pageSize)
}

// DeleteWhitelist 删除白名单规则
func (h *SecurityHandler) DeleteWhitelist(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	if err := h.securityService.RemoveWhitelistRule(id); err != nil {
		utils.ServerError(c, "删除白名单失败")
		return
	}

	utils.Success(c, nil)
}

// ========== OAuth绑定管理 ==========

// ListOAuthBindings 获取用户OAuth绑定列表
func (h *SecurityHandler) ListOAuthBindings(c *gin.Context) {
	tenantID, _ := c.Get("tenant_id")
	userID, _ := c.Get("user_id")

	bindings, err := h.securityService.ListUserOAuthBindings(tenantID.(int64), userID.(int64))
	if err != nil {
		utils.ServerError(c, "获取OAuth绑定列表失败")
		return
	}

	utils.Success(c, bindings)
}

// UnbindOAuth 解绑OAuth
func (h *SecurityHandler) UnbindOAuth(c *gin.Context) {
	provider := c.Param("provider")

	tenantID, _ := c.Get("tenant_id")
	userID, _ := c.Get("user_id")

	if err := h.securityService.DeleteOAuthBinding(tenantID.(int64), userID.(int64), provider); err != nil {
		utils.ServerError(c, "解绑失败")
		return
	}

	utils.Success(c, nil)
}
