package service

import (
	"cjadmin/internal/models"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gorm.io/gorm"
)

// OAuthService OAuth服务
type OAuthService struct {
	db *gorm.DB
}

// NewOAuthService 创建OAuth服务实例
func NewOAuthService(db *gorm.DB) *OAuthService {
	return &OAuthService{db: db}
}

// OAuthUserInfo OAuth用户信息
type OAuthUserInfo struct {
	ProviderUserID string
	Username       string
	Email          string
	Avatar         string
	AccessToken    string
	RefreshToken   string
	ExpiresIn      int
}

// ========== 飞书OAuth ==========

// GetFeishuAuthURL 获取飞书授权URL
func (s *OAuthService) GetFeishuAuthURL(config *models.OAuthConfig, state string) string {
	baseURL := "https://open.feishu.cn/open-apis/authen/v1/authorize"
	params := url.Values{}
	params.Add("app_id", config.AppID)
	params.Add("redirect_uri", config.RedirectURI)
	params.Add("state", state)
	return baseURL + "?" + params.Encode()
}

// GetFeishuUserInfo 通过授权码获取飞书用户信息
func (s *OAuthService) GetFeishuUserInfo(config *models.OAuthConfig, code string) (*OAuthUserInfo, error) {
	// 1. 获取app_access_token
	appToken, err := s.getFeishuAppToken(config)
	if err != nil {
		return nil, fmt.Errorf("获取飞书应用Token失败: %v", err)
	}

	// 2. 获取user_access_token
	userToken, err := s.getFeishuUserToken(appToken, code)
	if err != nil {
		return nil, fmt.Errorf("获取飞书用户Token失败: %v", err)
	}

	// 3. 获取用户信息
	userInfo, err := s.getFeishuUserInfoByToken(userToken)
	if err != nil {
		return nil, fmt.Errorf("获取飞书用户信息失败: %v", err)
	}

	return userInfo, nil
}

func (s *OAuthService) getFeishuAppToken(config *models.OAuthConfig) (string, error) {
	url := "https://open.feishu.cn/open-apis/auth/v3/app_access_token/internal"
	body := fmt.Sprintf(`{"app_id":"%s","app_secret":"%s"}`, config.AppID, config.AppSecret)

	resp, err := http.Post(url, "application/json", strings.NewReader(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Code           int    `json:"code"`
		Msg            string `json:"msg"`
		AppAccessToken string `json:"app_access_token"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if result.Code != 0 {
		return "", errors.New(result.Msg)
	}

	return result.AppAccessToken, nil
}

func (s *OAuthService) getFeishuUserToken(appToken, code string) (string, error) {
	url := "https://open.feishu.cn/open-apis/authen/v1/oidc/access_token"
	body := fmt.Sprintf(`{"grant_type":"authorization_code","code":"%s"}`, code)

	req, _ := http.NewRequest("POST", url, strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+appToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if result.Code != 0 {
		return "", errors.New(result.Msg)
	}

	return result.Data.AccessToken, nil
}

func (s *OAuthService) getFeishuUserInfoByToken(userToken string) (*OAuthUserInfo, error) {
	url := "https://open.feishu.cn/open-apis/authen/v1/user_info"

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+userToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			UserID   string `json:"user_id"`
			OpenID   string `json:"open_id"`
			UnionID  string `json:"union_id"`
			Name     string `json:"name"`
			EnName   string `json:"en_name"`
			Email    string `json:"email"`
			AvatarURL string `json:"avatar_url"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Code != 0 {
		return nil, errors.New(result.Msg)
	}

	return &OAuthUserInfo{
		ProviderUserID: result.Data.UnionID, // 使用UnionID作为唯一标识
		Username:       result.Data.Name,
		Email:          result.Data.Email,
		Avatar:         result.Data.AvatarURL,
		AccessToken:    userToken,
	}, nil
}

// ========== 微信OAuth ==========

// GetWechatAuthURL 获取微信授权URL
func (s *OAuthService) GetWechatAuthURL(config *models.OAuthConfig, state string) string {
	baseURL := "https://open.weixin.qq.com/connect/qrconnect"
	params := url.Values{}
	params.Add("appid", config.AppID)
	params.Add("redirect_uri", config.RedirectURI)
	params.Add("response_type", "code")
	params.Add("scope", "snsapi_login")
	params.Add("state", state)
	return baseURL + "?" + params.Encode() + "#wechat_redirect"
}

// GetWechatUserInfo 通过授权码获取微信用户信息
func (s *OAuthService) GetWechatUserInfo(config *models.OAuthConfig, code string) (*OAuthUserInfo, error) {
	// 1. 获取access_token
	tokenURL := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code",
		config.AppID, config.AppSecret, code,
	)

	resp, err := http.Get(tokenURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var tokenResult struct {
		AccessToken  string `json:"access_token"`
		ExpiresIn    int    `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
		OpenID       string `json:"openid"`
		UnionID      string `json:"unionid"`
		ErrCode      int    `json:"errcode"`
		ErrMsg       string `json:"errmsg"`
	}

	if err := json.Unmarshal(body, &tokenResult); err != nil {
		return nil, err
	}

	if tokenResult.ErrCode != 0 {
		return nil, errors.New(tokenResult.ErrMsg)
	}

	// 2. 获取用户信息
	userURL := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/userinfo?access_token=%s&openid=%s",
		tokenResult.AccessToken, tokenResult.OpenID,
	)

	resp2, err := http.Get(userURL)
	if err != nil {
		return nil, err
	}
	defer resp2.Body.Close()

	var userResult struct {
		OpenID     string `json:"openid"`
		UnionID    string `json:"unionid"`
		Nickname   string `json:"nickname"`
		HeadImgURL string `json:"headimgurl"`
		ErrCode    int    `json:"errcode"`
		ErrMsg     string `json:"errmsg"`
	}

	if err := json.NewDecoder(resp2.Body).Decode(&userResult); err != nil {
		return nil, err
	}

	if userResult.ErrCode != 0 {
		return nil, errors.New(userResult.ErrMsg)
	}

	// 优先使用UnionID，如果没有则使用OpenID
	providerUserID := userResult.UnionID
	if providerUserID == "" {
		providerUserID = userResult.OpenID
	}

	return &OAuthUserInfo{
		ProviderUserID: providerUserID,
		Username:       userResult.Nickname,
		Avatar:         userResult.HeadImgURL,
		AccessToken:    tokenResult.AccessToken,
		RefreshToken:   tokenResult.RefreshToken,
		ExpiresIn:      tokenResult.ExpiresIn,
	}, nil
}

// ========== 通用OAuth处理 ==========

// GetAuthURL 获取授权URL
func (s *OAuthService) GetAuthURL(tenantID int64, provider, state string) (string, error) {
	config, err := s.getOAuthConfig(tenantID, provider)
	if err != nil {
		return "", fmt.Errorf("未配置%s登录", provider)
	}

	switch provider {
	case models.OAuthProviderFeishu:
		return s.GetFeishuAuthURL(config, state), nil
	case models.OAuthProviderWechat:
		return s.GetWechatAuthURL(config, state), nil
	default:
		return "", fmt.Errorf("不支持的OAuth提供商: %s", provider)
	}
}

// HandleCallback 处理OAuth回调
func (s *OAuthService) HandleCallback(tenantID int64, provider, code string) (*OAuthUserInfo, error) {
	config, err := s.getOAuthConfig(tenantID, provider)
	if err != nil {
		return nil, err
	}

	switch provider {
	case models.OAuthProviderFeishu:
		return s.GetFeishuUserInfo(config, code)
	case models.OAuthProviderWechat:
		return s.GetWechatUserInfo(config, code)
	default:
		return nil, fmt.Errorf("不支持的OAuth提供商: %s", provider)
	}
}

// FindOrCreateUser 根据OAuth信息查找或创建用户
func (s *OAuthService) FindOrCreateUser(tenantID int64, provider string, userInfo *OAuthUserInfo) (*models.User, bool, error) {
	securityService := NewSecurityService(s.db)

	// 检查是否已绑定
	binding, err := securityService.GetOAuthBindingByProviderID(tenantID, provider, userInfo.ProviderUserID)
	if err == nil && binding.UserID > 0 {
		// 已绑定，返回用户
		var user models.User
		if err := s.db.Where("id = ?", binding.UserID).First(&user).Error; err != nil {
			return nil, false, err
		}

		// 更新Token信息
		binding.AccessToken = userInfo.AccessToken
		binding.RefreshToken = userInfo.RefreshToken
		if userInfo.ExpiresIn > 0 {
			expiresAt := time.Now().Add(time.Duration(userInfo.ExpiresIn) * time.Second)
			binding.TokenExpiresAt = &expiresAt
		}
		s.db.Save(binding)

		return &user, false, nil
	}

	// 未绑定，创建新用户
	username := fmt.Sprintf("%s_%s", provider, userInfo.ProviderUserID[:8])
	user := &models.User{
		TenantID:    tenantID,
		Username:    username,
		Nickname:    userInfo.Username,
		Email:       userInfo.Email,
		Avatar:      userInfo.Avatar,
		Status:      "active",
		LoginMethod: models.LoginMethodOAuth,
	}

	if err := s.db.Create(user).Error; err != nil {
		return nil, false, err
	}

	// 创建绑定关系
	var expiresAt *time.Time
	if userInfo.ExpiresIn > 0 {
		t := time.Now().Add(time.Duration(userInfo.ExpiresIn) * time.Second)
		expiresAt = &t
	}

	newBinding := &models.OAuthBinding{
		TenantID:         tenantID,
		UserID:           user.ID,
		Provider:         provider,
		ProviderUserID:   userInfo.ProviderUserID,
		ProviderUsername: userInfo.Username,
		ProviderAvatar:   userInfo.Avatar,
		ProviderEmail:    userInfo.Email,
		AccessToken:      userInfo.AccessToken,
		RefreshToken:     userInfo.RefreshToken,
		TokenExpiresAt:   expiresAt,
		IsActive:         true,
	}

	if err := securityService.CreateOAuthBinding(newBinding); err != nil {
		return nil, false, err
	}

	return user, true, nil
}

func (s *OAuthService) getOAuthConfig(tenantID int64, provider string) (*models.OAuthConfig, error) {
	var config models.OAuthConfig
	err := s.db.Where("tenant_id = ? AND provider = ? AND is_active = ? AND deleted_at IS NULL",
		tenantID, provider, true).First(&config).Error
	return &config, err
}
