package service

import (
	"cjadmin/internal/models"
	"crypto/rand"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"gorm.io/gorm"
)

// SecurityService 安全服务
type SecurityService struct {
	db *gorm.DB
}

// NewSecurityService 创建安全服务实例
func NewSecurityService(db *gorm.DB) *SecurityService {
	return &SecurityService{db: db}
}

// ========== 验证码相关 ==========

// GenerateSMSCode 生成短信验证码
func (s *SecurityService) GenerateSMSCode(tenantID *int64, phone, codeType, ipAddress, userAgent string) (*models.SMSCode, error) {
	// 检查发送频率（同一手机号1分钟内只能发送一次）
	var recentCode models.SMSCode
	oneMinuteAgo := time.Now().Add(-1 * time.Minute)
	err := s.db.Where("phone = ? AND code_type = ? AND created_at > ?", phone, codeType, oneMinuteAgo).
		First(&recentCode).Error
	if err == nil {
		return nil, errors.New("发送过于频繁，请稍后再试")
	}

	// 生成6位数字验证码
	code := s.generateRandomCode(6)

	// 创建验证码记录
	smsCode := &models.SMSCode{
		TenantID:  tenantID,
		Phone:     phone,
		Code:      code,
		CodeType:  codeType,
		ExpiresAt: time.Now().Add(5 * time.Minute), // 5分钟有效期
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}

	if err := s.db.Create(smsCode).Error; err != nil {
		return nil, err
	}

	return smsCode, nil
}

// VerifySMSCode 验证短信验证码
func (s *SecurityService) VerifySMSCode(phone, code, codeType string) error {
	var smsCode models.SMSCode
	err := s.db.Where("phone = ? AND code = ? AND code_type = ? AND is_used = ? AND expires_at > ?",
		phone, code, codeType, false, time.Now()).
		Order("created_at DESC").
		First(&smsCode).Error

	if err != nil {
		return errors.New("验证码错误或已过期")
	}

	// 标记验证码已使用
	now := time.Now()
	smsCode.IsUsed = true
	smsCode.UsedAt = &now
	s.db.Save(&smsCode)

	return nil
}

// generateRandomCode 生成随机数字验证码
func (s *SecurityService) generateRandomCode(length int) string {
	const digits = "0123456789"
	b := make([]byte, length)
	rand.Read(b)
	for i := range b {
		b[i] = digits[int(b[i])%len(digits)]
	}
	return string(b)
}

// ========== 黑白名单相关 ==========

// CheckBlacklist 检查是否在黑名单中
func (s *SecurityService) CheckBlacklist(tenantID *int64, ip, userID, deviceFingerprint, phone string) (bool, string) {
	now := time.Now()

	// 构建查询条件
	query := s.db.Model(&models.LoginBlacklist{}).
		Where("is_active = ? AND deleted_at IS NULL", true).
		Where("(expires_at IS NULL OR expires_at > ?)", now)

	// 租户隔离（包含全局规则）
	if tenantID != nil {
		query = query.Where("(tenant_id IS NULL OR tenant_id = ?)", *tenantID)
	} else {
		query = query.Where("tenant_id IS NULL")
	}

	var blacklist models.LoginBlacklist

	// 检查IP
	if ip != "" {
		if err := query.Where("rule_type = ? AND rule_value = ?", models.RuleTypeIP, ip).First(&blacklist).Error; err == nil {
			return true, fmt.Sprintf("IP地址 %s 已被禁止登录: %s", ip, blacklist.Reason)
		}

		// 检查IP段
		var ipRanges []models.LoginBlacklist
		s.db.Model(&models.LoginBlacklist{}).
			Where("rule_type = ? AND is_active = ? AND deleted_at IS NULL", models.RuleTypeIPRange, true).
			Where("(expires_at IS NULL OR expires_at > ?)", now).
			Find(&ipRanges)

		for _, rule := range ipRanges {
			if s.isIPInRange(ip, rule.RuleValue) {
				return true, fmt.Sprintf("IP地址 %s 在禁止范围内: %s", ip, rule.Reason)
			}
		}
	}

	// 检查用户ID
	if userID != "" {
		if err := query.Where("rule_type = ? AND rule_value = ?", models.RuleTypeUser, userID).First(&blacklist).Error; err == nil {
			return true, fmt.Sprintf("用户已被禁止登录: %s", blacklist.Reason)
		}
	}

	// 检查设备指纹
	if deviceFingerprint != "" {
		if err := query.Where("rule_type = ? AND rule_value = ?", models.RuleTypeDevice, deviceFingerprint).First(&blacklist).Error; err == nil {
			return true, fmt.Sprintf("设备已被禁止登录: %s", blacklist.Reason)
		}
	}

	// 检查手机号
	if phone != "" {
		if err := query.Where("rule_type = ? AND rule_value = ?", models.RuleTypePhone, phone).First(&blacklist).Error; err == nil {
			return true, fmt.Sprintf("手机号已被禁止登录: %s", blacklist.Reason)
		}
	}

	return false, ""
}

// CheckWhitelist 检查是否在白名单中（如果启用白名单模式）
func (s *SecurityService) CheckWhitelist(tenantID *int64, ip, userID string) bool {
	query := s.db.Model(&models.LoginWhitelist{}).
		Where("is_active = ? AND deleted_at IS NULL", true)

	if tenantID != nil {
		query = query.Where("(tenant_id IS NULL OR tenant_id = ?)", *tenantID)
	} else {
		query = query.Where("tenant_id IS NULL")
	}

	var whitelist models.LoginWhitelist

	// 检查IP
	if ip != "" {
		if err := query.Where("rule_type = ? AND rule_value = ?", models.RuleTypeIP, ip).First(&whitelist).Error; err == nil {
			return true
		}
	}

	// 检查用户
	if userID != "" {
		if err := query.Where("rule_type = ? AND rule_value = ?", models.RuleTypeUser, userID).First(&whitelist).Error; err == nil {
			return true
		}
	}

	return false
}

// isIPInRange 检查IP是否在CIDR范围内
func (s *SecurityService) isIPInRange(ip, cidr string) bool {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}
	parsedIP := net.ParseIP(ip)
	return ipNet.Contains(parsedIP)
}

// ========== 登录尝试记录 ==========

// RecordLoginAttempt 记录登录尝试
func (s *SecurityService) RecordLoginAttempt(tenantID *int64, identifier, identifierType, ipAddress, userAgent, deviceFingerprint string, isSuccess bool, failureReason string) error {
	attempt := &models.LoginAttempt{
		TenantID:          tenantID,
		Identifier:        identifier,
		IdentifierType:    identifierType,
		IPAddress:         ipAddress,
		UserAgent:         userAgent,
		DeviceFingerprint: deviceFingerprint,
		IsSuccess:         isSuccess,
		FailureReason:     failureReason,
	}

	return s.db.Create(attempt).Error
}

// CheckLoginAttempts 检查登录尝试次数（限流）
func (s *SecurityService) CheckLoginAttempts(identifier, ipAddress string, maxAttempts int, duration time.Duration) (bool, int) {
	since := time.Now().Add(-duration)

	var count int64
	s.db.Model(&models.LoginAttempt{}).
		Where("(identifier = ? OR ip_address = ?) AND is_success = ? AND created_at > ?",
			identifier, ipAddress, false, since).
		Count(&count)

	remaining := maxAttempts - int(count)
	if remaining < 0 {
		remaining = 0
	}

	return count >= int64(maxAttempts), remaining
}

// ========== 黑白名单CRUD ==========

// AddBlacklistRule 添加黑名单规则
func (s *SecurityService) AddBlacklistRule(tenantID *int64, ruleType, ruleValue, reason string, expiresAt *time.Time, createdBy int64) (*models.LoginBlacklist, error) {
	rule := &models.LoginBlacklist{
		TenantID:  tenantID,
		RuleType:  ruleType,
		RuleValue: ruleValue,
		Reason:    reason,
		ExpiresAt: expiresAt,
		CreatedBy: &createdBy,
		IsActive:  true,
	}

	if err := s.db.Create(rule).Error; err != nil {
		return nil, err
	}

	return rule, nil
}

// RemoveBlacklistRule 移除黑名单规则
func (s *SecurityService) RemoveBlacklistRule(id int64) error {
	now := time.Now()
	return s.db.Model(&models.LoginBlacklist{}).
		Where("id = ?", id).
		Update("deleted_at", now).Error
}

// ListBlacklist 获取黑名单列表
func (s *SecurityService) ListBlacklist(tenantID *int64, page, pageSize int) ([]models.LoginBlacklist, int64, error) {
	var rules []models.LoginBlacklist
	var total int64

	query := s.db.Model(&models.LoginBlacklist{}).Where("deleted_at IS NULL")

	if tenantID != nil {
		query = query.Where("(tenant_id IS NULL OR tenant_id = ?)", *tenantID)
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&rules).Error

	return rules, total, err
}

// AddWhitelistRule 添加白名单规则
func (s *SecurityService) AddWhitelistRule(tenantID *int64, ruleType, ruleValue, description string, createdBy int64) (*models.LoginWhitelist, error) {
	rule := &models.LoginWhitelist{
		TenantID:    tenantID,
		RuleType:    ruleType,
		RuleValue:   ruleValue,
		Description: description,
		CreatedBy:   &createdBy,
		IsActive:    true,
	}

	if err := s.db.Create(rule).Error; err != nil {
		return nil, err
	}

	return rule, nil
}

// RemoveWhitelistRule 移除白名单规则
func (s *SecurityService) RemoveWhitelistRule(id int64) error {
	now := time.Now()
	return s.db.Model(&models.LoginWhitelist{}).
		Where("id = ?", id).
		Update("deleted_at", now).Error
}

// ListWhitelist 获取白名单列表
func (s *SecurityService) ListWhitelist(tenantID *int64, page, pageSize int) ([]models.LoginWhitelist, int64, error) {
	var rules []models.LoginWhitelist
	var total int64

	query := s.db.Model(&models.LoginWhitelist{}).Where("deleted_at IS NULL")

	if tenantID != nil {
		query = query.Where("(tenant_id IS NULL OR tenant_id = ?)", *tenantID)
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&rules).Error

	return rules, total, err
}

// ========== OAuth绑定相关 ==========

// GetOAuthBinding 获取用户的OAuth绑定
func (s *SecurityService) GetOAuthBinding(tenantID int64, userID int64, provider string) (*models.OAuthBinding, error) {
	var binding models.OAuthBinding
	err := s.db.Where("tenant_id = ? AND user_id = ? AND provider = ? AND deleted_at IS NULL",
		tenantID, userID, provider).First(&binding).Error
	return &binding, err
}

// GetOAuthBindingByProviderID 通过第三方ID获取绑定
func (s *SecurityService) GetOAuthBindingByProviderID(tenantID int64, provider, providerUserID string) (*models.OAuthBinding, error) {
	var binding models.OAuthBinding
	err := s.db.Where("tenant_id = ? AND provider = ? AND provider_user_id = ? AND deleted_at IS NULL",
		tenantID, provider, providerUserID).First(&binding).Error
	return &binding, err
}

// CreateOAuthBinding 创建OAuth绑定
func (s *SecurityService) CreateOAuthBinding(binding *models.OAuthBinding) error {
	return s.db.Create(binding).Error
}

// UpdateOAuthBinding 更新OAuth绑定
func (s *SecurityService) UpdateOAuthBinding(binding *models.OAuthBinding) error {
	return s.db.Save(binding).Error
}

// DeleteOAuthBinding 删除OAuth绑定
func (s *SecurityService) DeleteOAuthBinding(tenantID int64, userID int64, provider string) error {
	now := time.Now()
	return s.db.Model(&models.OAuthBinding{}).
		Where("tenant_id = ? AND user_id = ? AND provider = ?", tenantID, userID, provider).
		Update("deleted_at", now).Error
}

// ListUserOAuthBindings 获取用户的所有OAuth绑定
func (s *SecurityService) ListUserOAuthBindings(tenantID int64, userID int64) ([]models.OAuthBinding, error) {
	var bindings []models.OAuthBinding
	err := s.db.Where("tenant_id = ? AND user_id = ? AND deleted_at IS NULL", tenantID, userID).
		Find(&bindings).Error
	return bindings, err
}

// ========== OAuth配置相关 ==========

// GetOAuthConfig 获取租户的OAuth配置
func (s *SecurityService) GetOAuthConfig(tenantID int64, provider string) (*models.OAuthConfig, error) {
	var config models.OAuthConfig
	err := s.db.Where("tenant_id = ? AND provider = ? AND is_active = ? AND deleted_at IS NULL",
		tenantID, provider, true).First(&config).Error
	return &config, err
}

// SaveOAuthConfig 保存OAuth配置
func (s *SecurityService) SaveOAuthConfig(config *models.OAuthConfig) error {
	// 检查是否已存在
	var existing models.OAuthConfig
	err := s.db.Where("tenant_id = ? AND provider = ? AND deleted_at IS NULL",
		config.TenantID, config.Provider).First(&existing).Error

	if err == nil {
		// 更新现有配置
		config.ID = existing.ID
		return s.db.Save(config).Error
	}

	// 创建新配置
	return s.db.Create(config).Error
}

// ========== 辅助方法 ==========

// GetClientIP 从请求中获取客户端IP
func GetClientIP(xForwardedFor, xRealIP, remoteAddr string) string {
	// 优先使用X-Forwarded-For
	if xForwardedFor != "" {
		ips := strings.Split(xForwardedFor, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 其次使用X-Real-IP
	if xRealIP != "" {
		return xRealIP
	}

	// 最后使用RemoteAddr
	ip, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}
	return ip
}
