package service

import (
	"testing"

	"cjadmin/internal/models"
	"cjadmin/internal/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOAuthService_GetFeishuAuthURL(t *testing.T) {
	testDB := test.SetupTestDB(t)
	defer testDB.TearDown(t)

	oauthService := NewOAuthService(testDB.DB)

	config := &models.OAuthConfig{
		AppID:       "cli_test123456",
		AppSecret:   "secret123",
		RedirectURI: "https://example.com/oauth/callback/feishu",
	}

	state := "random_state_string"
	authURL := oauthService.GetFeishuAuthURL(config, state)

	assert.Contains(t, authURL, "https://open.feishu.cn/open-apis/authen/v1/authorize")
	assert.Contains(t, authURL, "app_id=cli_test123456")
	assert.Contains(t, authURL, "redirect_uri=")
	assert.Contains(t, authURL, "state=random_state_string")
}

func TestOAuthService_GetWechatAuthURL(t *testing.T) {
	testDB := test.SetupTestDB(t)
	defer testDB.TearDown(t)

	oauthService := NewOAuthService(testDB.DB)

	config := &models.OAuthConfig{
		AppID:       "wx_test123456",
		AppSecret:   "secret123",
		RedirectURI: "https://example.com/oauth/callback/wechat",
	}

	state := "random_state_string"
	authURL := oauthService.GetWechatAuthURL(config, state)

	assert.Contains(t, authURL, "https://open.weixin.qq.com/connect/qrconnect")
	assert.Contains(t, authURL, "appid=wx_test123456")
	assert.Contains(t, authURL, "redirect_uri=")
	assert.Contains(t, authURL, "state=random_state_string")
	assert.Contains(t, authURL, "response_type=code")
	assert.Contains(t, authURL, "scope=snsapi_login")
	assert.Contains(t, authURL, "#wechat_redirect")
}

func TestOAuthService_GetAuthURL(t *testing.T) {
	testDB := test.SetupTestDB(t)
	defer testDB.TearDown(t)

	oauthService := NewOAuthService(testDB.DB)
	tenant := testDB.CreateTestTenant(t, "test-tenant")

	// 创建飞书OAuth配置
	feishuConfig := &models.OAuthConfig{
		TenantID:    tenant.ID,
		Provider:    models.OAuthProviderFeishu,
		AppID:       "cli_feishu123",
		AppSecret:   "secret_feishu",
		RedirectURI: "https://example.com/oauth/callback/feishu",
		IsActive:    true,
	}
	require.NoError(t, testDB.DB.Create(feishuConfig).Error)

	// 创建微信OAuth配置
	wechatConfig := &models.OAuthConfig{
		TenantID:    tenant.ID,
		Provider:    models.OAuthProviderWechat,
		AppID:       "wx_wechat123",
		AppSecret:   "secret_wechat",
		RedirectURI: "https://example.com/oauth/callback/wechat",
		IsActive:    true,
	}
	require.NoError(t, testDB.DB.Create(wechatConfig).Error)

	tests := []struct {
		name        string
		provider    string
		wantErr     bool
		errContains string
		urlContains string
	}{
		{
			name:        "feishu auth URL",
			provider:    models.OAuthProviderFeishu,
			wantErr:     false,
			urlContains: "feishu.cn",
		},
		{
			name:        "wechat auth URL",
			provider:    models.OAuthProviderWechat,
			wantErr:     false,
			urlContains: "weixin.qq.com",
		},
		{
			name:        "unsupported provider",
			provider:    "github",
			wantErr:     true,
			errContains: "未配置",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authURL, err := oauthService.GetAuthURL(tenant.ID, tt.provider, "test_state")

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, authURL)
				assert.Contains(t, authURL, tt.urlContains)
			}
		})
	}
}

func TestOAuthService_FindOrCreateUser_ExistingBinding(t *testing.T) {
	testDB := test.SetupTestDB(t)
	defer testDB.TearDown(t)

	oauthService := NewOAuthService(testDB.DB)
	tenant := testDB.CreateTestTenant(t, "test-tenant")
	existingUser := testDB.CreateTestUser(t, tenant.ID, "existinguser", "")

	// 创建已有的OAuth绑定
	securityService := NewSecurityService(testDB.DB)
	binding := &models.OAuthBinding{
		TenantID:       tenant.ID,
		UserID:         existingUser.ID,
		Provider:       models.OAuthProviderFeishu,
		ProviderUserID: "feishu_existing_user",
		IsActive:       true,
	}
	require.NoError(t, securityService.CreateOAuthBinding(binding))

	// 使用相同的OAuth信息查找用户
	userInfo := &OAuthUserInfo{
		ProviderUserID: "feishu_existing_user",
		Username:       "飞书用户",
		Email:          "feishu@example.com",
		AccessToken:    "new_access_token",
	}

	user, isNew, err := oauthService.FindOrCreateUser(tenant.ID, models.OAuthProviderFeishu, userInfo)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.False(t, isNew)
	assert.Equal(t, existingUser.ID, user.ID)
}

func TestOAuthService_FindOrCreateUser_NewUser(t *testing.T) {
	testDB := test.SetupTestDB(t)
	defer testDB.TearDown(t)

	oauthService := NewOAuthService(testDB.DB)
	tenant := testDB.CreateTestTenant(t, "test-tenant")

	// 新用户信息
	userInfo := &OAuthUserInfo{
		ProviderUserID: "feishu_new_user_12345678",
		Username:       "新飞书用户",
		Email:          "newfeishu@example.com",
		Avatar:         "https://avatar.example.com/user.jpg",
		AccessToken:    "access_token",
		RefreshToken:   "refresh_token",
		ExpiresIn:      7200,
	}

	user, isNew, err := oauthService.FindOrCreateUser(tenant.ID, models.OAuthProviderFeishu, userInfo)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.True(t, isNew)
	assert.Equal(t, "新飞书用户", user.Nickname)
	assert.Equal(t, "newfeishu@example.com", user.Email)
	assert.Equal(t, models.LoginMethodOAuth, user.LoginMethod)
	assert.Equal(t, "active", user.Status)

	// 验证绑定已创建
	securityService := NewSecurityService(testDB.DB)
	binding, err := securityService.GetOAuthBindingByProviderID(tenant.ID, models.OAuthProviderFeishu, "feishu_new_user_12345678")
	assert.NoError(t, err)
	assert.NotNil(t, binding)
	assert.Equal(t, user.ID, binding.UserID)
}

func TestOAuthService_HandleCallback_NoConfig(t *testing.T) {
	testDB := test.SetupTestDB(t)
	defer testDB.TearDown(t)

	oauthService := NewOAuthService(testDB.DB)
	tenant := testDB.CreateTestTenant(t, "test-tenant")

	// 没有配置时应该返回错误
	_, err := oauthService.HandleCallback(tenant.ID, models.OAuthProviderFeishu, "test_code")
	assert.Error(t, err)
}

func TestOAuthService_GetAuthURL_InactiveConfig(t *testing.T) {
	testDB := test.SetupTestDB(t)
	defer testDB.TearDown(t)

	oauthService := NewOAuthService(testDB.DB)
	tenant := testDB.CreateTestTenant(t, "test-tenant")

	// 创建未激活的配置
	config := &models.OAuthConfig{
		TenantID:    tenant.ID,
		Provider:    models.OAuthProviderFeishu,
		AppID:       "cli_inactive",
		AppSecret:   "secret",
		RedirectURI: "https://example.com/callback",
		IsActive:    false, // 未激活
	}
	require.NoError(t, testDB.DB.Create(config).Error)

	// 应该返回错误
	_, err := oauthService.GetAuthURL(tenant.ID, models.OAuthProviderFeishu, "test_state")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "未配置")
}

func TestOAuthService_MultipleProviders(t *testing.T) {
	testDB := test.SetupTestDB(t)
	defer testDB.TearDown(t)

	oauthService := NewOAuthService(testDB.DB)
	tenant := testDB.CreateTestTenant(t, "test-tenant")

	// 创建多个OAuth配置
	providers := []struct {
		provider string
		appID    string
	}{
		{models.OAuthProviderFeishu, "cli_feishu"},
		{models.OAuthProviderWechat, "wx_wechat"},
	}

	for _, p := range providers {
		config := &models.OAuthConfig{
			TenantID:    tenant.ID,
			Provider:    p.provider,
			AppID:       p.appID,
			AppSecret:   "secret",
			RedirectURI: "https://example.com/callback/" + p.provider,
			IsActive:    true,
		}
		require.NoError(t, testDB.DB.Create(config).Error)
	}

	// 验证可以获取各个提供商的授权URL
	for _, p := range providers {
		authURL, err := oauthService.GetAuthURL(tenant.ID, p.provider, "state")
		assert.NoError(t, err)
		assert.NotEmpty(t, authURL)
	}
}

func TestOAuthUserInfo_TokenExpiration(t *testing.T) {
	testDB := test.SetupTestDB(t)
	defer testDB.TearDown(t)

	oauthService := NewOAuthService(testDB.DB)
	tenant := testDB.CreateTestTenant(t, "test-tenant")

	// 带过期时间的用户信息
	userInfo := &OAuthUserInfo{
		ProviderUserID: "wechat_user_12345678",
		Username:       "微信用户",
		AccessToken:    "access_token",
		RefreshToken:   "refresh_token",
		ExpiresIn:      7200, // 2小时
	}

	user, _, err := oauthService.FindOrCreateUser(tenant.ID, models.OAuthProviderWechat, userInfo)
	require.NoError(t, err)

	// 验证绑定中保存了Token过期时间
	var binding models.OAuthBinding
	err = testDB.DB.Where("user_id = ? AND provider = ?", user.ID, models.OAuthProviderWechat).First(&binding).Error
	assert.NoError(t, err)
	assert.NotNil(t, binding.TokenExpiresAt)
}
