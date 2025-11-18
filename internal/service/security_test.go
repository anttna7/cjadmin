package service

import (
	"testing"
	"time"

	"cjadmin/internal/models"
	"cjadmin/internal/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecurityService_GenerateSMSCode(t *testing.T) {
	testDB := test.SetupTestDB(t)
	defer testDB.TearDown(t)

	securityService := NewSecurityService(testDB.DB)
	tenant := testDB.CreateTestTenant(t, "test-tenant")

	tests := []struct {
		name        string
		phone       string
		codeType    string
		wantErr     bool
		errContains string
	}{
		{
			name:     "successful code generation",
			phone:    "13800138001",
			codeType: models.CodeTypeLogin,
			wantErr:  false,
		},
		{
			name:     "register code type",
			phone:    "13800138002",
			codeType: models.CodeTypeRegister,
			wantErr:  false,
		},
		{
			name:     "reset password code type",
			phone:    "13800138003",
			codeType: models.CodeTypeResetPassword,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			smsCode, err := securityService.GenerateSMSCode(
				&tenant.ID,
				tt.phone,
				tt.codeType,
				"127.0.0.1",
				"Test User-Agent",
			)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, smsCode)
				assert.Equal(t, tt.phone, smsCode.Phone)
				assert.Equal(t, tt.codeType, smsCode.CodeType)
				assert.Len(t, smsCode.Code, 6)
				assert.False(t, smsCode.IsUsed)
				assert.True(t, smsCode.ExpiresAt.After(time.Now()))
			}
		})
	}
}

func TestSecurityService_GenerateSMSCode_RateLimit(t *testing.T) {
	testDB := test.SetupTestDB(t)
	defer testDB.TearDown(t)

	securityService := NewSecurityService(testDB.DB)
	tenant := testDB.CreateTestTenant(t, "test-tenant")
	phone := "13800138000"

	// 第一次发送应该成功
	_, err := securityService.GenerateSMSCode(&tenant.ID, phone, models.CodeTypeLogin, "127.0.0.1", "Test")
	assert.NoError(t, err)

	// 立即再次发送应该失败（60秒限制）
	_, err = securityService.GenerateSMSCode(&tenant.ID, phone, models.CodeTypeLogin, "127.0.0.1", "Test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "发送频率过高")
}

func TestSecurityService_VerifySMSCode(t *testing.T) {
	testDB := test.SetupTestDB(t)
	defer testDB.TearDown(t)

	securityService := NewSecurityService(testDB.DB)
	tenant := testDB.CreateTestTenant(t, "test-tenant")
	phone := "13800138000"

	// 生成验证码
	smsCode, err := securityService.GenerateSMSCode(&tenant.ID, phone, models.CodeTypeLogin, "127.0.0.1", "Test")
	require.NoError(t, err)

	tests := []struct {
		name        string
		phone       string
		code        string
		codeType    string
		wantErr     bool
		errContains string
	}{
		{
			name:     "successful verification",
			phone:    phone,
			code:     smsCode.Code,
			codeType: models.CodeTypeLogin,
			wantErr:  false,
		},
		{
			name:        "wrong code",
			phone:       phone,
			code:        "000000",
			codeType:    models.CodeTypeLogin,
			wantErr:     true,
			errContains: "验证码错误",
		},
		{
			name:        "wrong phone",
			phone:       "13900139000",
			code:        smsCode.Code,
			codeType:    models.CodeTypeLogin,
			wantErr:     true,
			errContains: "验证码错误",
		},
		{
			name:        "wrong code type",
			phone:       phone,
			code:        smsCode.Code,
			codeType:    models.CodeTypeRegister,
			wantErr:     true,
			errContains: "验证码错误",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := securityService.VerifySMSCode(tt.phone, tt.code, tt.codeType)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSecurityService_VerifySMSCode_Expired(t *testing.T) {
	testDB := test.SetupTestDB(t)
	defer testDB.TearDown(t)

	securityService := NewSecurityService(testDB.DB)
	tenant := testDB.CreateTestTenant(t, "test-tenant")
	phone := "13800138000"

	// 创建一个已过期的验证码
	expiredCode := &models.SMSCode{
		TenantID:  &tenant.ID,
		Phone:     phone,
		Code:      "123456",
		CodeType:  models.CodeTypeLogin,
		ExpiresAt: time.Now().Add(-1 * time.Hour), // 1小时前过期
		IPAddress: "127.0.0.1",
	}
	require.NoError(t, testDB.DB.Create(expiredCode).Error)

	// 验证应该失败
	err := securityService.VerifySMSCode(phone, "123456", models.CodeTypeLogin)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "验证码已过期")
}

func TestSecurityService_CheckBlacklist(t *testing.T) {
	testDB := test.SetupTestDB(t)
	defer testDB.TearDown(t)

	securityService := NewSecurityService(testDB.DB)
	tenant := testDB.CreateTestTenant(t, "test-tenant")
	user := testDB.CreateTestUser(t, tenant.ID, "testuser", "")

	// 添加IP黑名单
	_, err := securityService.AddBlacklistRule(
		&tenant.ID,
		models.RuleTypeIP,
		"192.168.1.100",
		"测试封禁",
		nil,
		user.ID,
	)
	require.NoError(t, err)

	// 添加手机号黑名单
	_, err = securityService.AddBlacklistRule(
		&tenant.ID,
		models.RuleTypePhone,
		"13800138000",
		"测试封禁手机",
		nil,
		user.ID,
	)
	require.NoError(t, err)

	tests := []struct {
		name        string
		ip          string
		phone       string
		wantBlocked bool
	}{
		{
			name:        "blocked IP",
			ip:          "192.168.1.100",
			phone:       "",
			wantBlocked: true,
		},
		{
			name:        "allowed IP",
			ip:          "192.168.1.200",
			phone:       "",
			wantBlocked: false,
		},
		{
			name:        "blocked phone",
			ip:          "",
			phone:       "13800138000",
			wantBlocked: true,
		},
		{
			name:        "allowed phone",
			ip:          "",
			phone:       "13900139000",
			wantBlocked: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blocked, _ := securityService.CheckBlacklist(&tenant.ID, tt.ip, "", "", tt.phone)
			assert.Equal(t, tt.wantBlocked, blocked)
		})
	}
}

func TestSecurityService_CheckBlacklist_Expired(t *testing.T) {
	testDB := test.SetupTestDB(t)
	defer testDB.TearDown(t)

	securityService := NewSecurityService(testDB.DB)
	tenant := testDB.CreateTestTenant(t, "test-tenant")
	user := testDB.CreateTestUser(t, tenant.ID, "testuser", "")

	// 添加已过期的黑名单
	expiredTime := time.Now().Add(-1 * time.Hour)
	_, err := securityService.AddBlacklistRule(
		&tenant.ID,
		models.RuleTypeIP,
		"192.168.1.100",
		"已过期封禁",
		&expiredTime,
		user.ID,
	)
	require.NoError(t, err)

	// 应该不被封禁（已过期）
	blocked, _ := securityService.CheckBlacklist(&tenant.ID, "192.168.1.100", "", "", "")
	assert.False(t, blocked)
}

func TestSecurityService_CheckLoginAttempts(t *testing.T) {
	testDB := test.SetupTestDB(t)
	defer testDB.TearDown(t)

	securityService := NewSecurityService(testDB.DB)
	tenant := testDB.CreateTestTenant(t, "test-tenant")
	identifier := "testuser"
	ip := "127.0.0.1"

	// 记录4次失败登录
	for i := 0; i < 4; i++ {
		securityService.RecordLoginAttempt(&tenant.ID, identifier, models.IdentifierTypeUsername, ip, "Test", "", false, "密码错误")
	}

	// 第5次尝试应该还能继续（还有1次机会）
	blocked, remaining := securityService.CheckLoginAttempts(identifier, ip, 5, 15*time.Minute)
	assert.False(t, blocked)
	assert.Equal(t, 1, remaining)

	// 记录第5次失败
	securityService.RecordLoginAttempt(&tenant.ID, identifier, models.IdentifierTypeUsername, ip, "Test", "", false, "密码错误")

	// 第6次应该被封禁
	blocked, _ = securityService.CheckLoginAttempts(identifier, ip, 5, 15*time.Minute)
	assert.True(t, blocked)
}

func TestSecurityService_Whitelist(t *testing.T) {
	testDB := test.SetupTestDB(t)
	defer testDB.TearDown(t)

	securityService := NewSecurityService(testDB.DB)
	tenant := testDB.CreateTestTenant(t, "test-tenant")
	user := testDB.CreateTestUser(t, tenant.ID, "testuser", "")

	// 添加白名单
	rule, err := securityService.AddWhitelistRule(
		&tenant.ID,
		models.RuleTypeIP,
		"10.0.0.0/8",
		"内网IP白名单",
		user.ID,
	)
	require.NoError(t, err)
	assert.NotNil(t, rule)
	assert.Equal(t, models.RuleTypeIP, rule.RuleType)

	// 查询白名单
	rules, total, err := securityService.ListWhitelist(&tenant.ID, 1, 20)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, rules, 1)

	// 删除白名单
	err = securityService.RemoveWhitelistRule(rule.ID)
	assert.NoError(t, err)

	// 验证删除
	rules, total, err = securityService.ListWhitelist(&tenant.ID, 1, 20)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), total)
}

func TestSecurityService_OAuthBinding(t *testing.T) {
	testDB := test.SetupTestDB(t)
	defer testDB.TearDown(t)

	securityService := NewSecurityService(testDB.DB)
	tenant := testDB.CreateTestTenant(t, "test-tenant")
	user := testDB.CreateTestUser(t, tenant.ID, "testuser", "")

	// 创建OAuth绑定
	binding := &models.OAuthBinding{
		TenantID:         tenant.ID,
		UserID:           user.ID,
		Provider:         models.OAuthProviderFeishu,
		ProviderUserID:   "feishu_user_123",
		ProviderUsername: "飞书用户",
		IsActive:         true,
	}
	err := securityService.CreateOAuthBinding(binding)
	require.NoError(t, err)

	// 通过ProviderID查询绑定
	found, err := securityService.GetOAuthBindingByProviderID(tenant.ID, models.OAuthProviderFeishu, "feishu_user_123")
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, user.ID, found.UserID)

	// 查询用户的OAuth绑定列表
	bindings, err := securityService.ListUserOAuthBindings(tenant.ID, user.ID)
	assert.NoError(t, err)
	assert.Len(t, bindings, 1)

	// 删除绑定
	err = securityService.DeleteOAuthBinding(tenant.ID, user.ID, models.OAuthProviderFeishu)
	assert.NoError(t, err)

	// 验证删除
	bindings, err = securityService.ListUserOAuthBindings(tenant.ID, user.ID)
	assert.NoError(t, err)
	assert.Len(t, bindings, 0)
}

func TestSecurityService_RecordLoginAttempt(t *testing.T) {
	testDB := test.SetupTestDB(t)
	defer testDB.TearDown(t)

	securityService := NewSecurityService(testDB.DB)
	tenant := testDB.CreateTestTenant(t, "test-tenant")

	// 记录成功登录
	err := securityService.RecordLoginAttempt(
		&tenant.ID,
		"testuser",
		models.IdentifierTypeUsername,
		"127.0.0.1",
		"Mozilla/5.0",
		"fingerprint123",
		true,
		"",
	)
	assert.NoError(t, err)

	// 记录失败登录
	err = securityService.RecordLoginAttempt(
		&tenant.ID,
		"testuser",
		models.IdentifierTypeUsername,
		"127.0.0.1",
		"Mozilla/5.0",
		"fingerprint123",
		false,
		"密码错误",
	)
	assert.NoError(t, err)

	// 验证记录存在
	var count int64
	testDB.DB.Model(&models.LoginAttempt{}).Where("tenant_id = ?", tenant.ID).Count(&count)
	assert.Equal(t, int64(2), count)
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name       string
		forwarded  string
		realIP     string
		remoteAddr string
		expected   string
	}{
		{
			name:       "use X-Forwarded-For",
			forwarded:  "203.0.113.195, 70.41.3.18, 150.172.238.178",
			realIP:     "",
			remoteAddr: "127.0.0.1:8080",
			expected:   "203.0.113.195",
		},
		{
			name:       "use X-Real-IP",
			forwarded:  "",
			realIP:     "203.0.113.195",
			remoteAddr: "127.0.0.1:8080",
			expected:   "203.0.113.195",
		},
		{
			name:       "use RemoteAddr",
			forwarded:  "",
			realIP:     "",
			remoteAddr: "192.168.1.1:12345",
			expected:   "192.168.1.1",
		},
		{
			name:       "priority: X-Forwarded-For > X-Real-IP",
			forwarded:  "10.0.0.1",
			realIP:     "10.0.0.2",
			remoteAddr: "127.0.0.1:8080",
			expected:   "10.0.0.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetClientIP(tt.forwarded, tt.realIP, tt.remoteAddr)
			assert.Equal(t, tt.expected, result)
		})
	}
}
