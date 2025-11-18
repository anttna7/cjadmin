package models

import (
	"time"
)

// SMSCode 短信验证码
type SMSCode struct {
	ID        int64      `gorm:"primaryKey" json:"id"`
	TenantID  *int64     `gorm:"index" json:"tenant_id,omitempty"`
	Phone     string     `gorm:"size:20;not null;index" json:"phone"`
	Code      string     `gorm:"size:10;not null" json:"code"`
	CodeType  string     `gorm:"size:50;not null;default:login" json:"code_type"`
	IsUsed    bool       `gorm:"default:false" json:"is_used"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	ExpiresAt time.Time  `gorm:"not null" json:"expires_at"`
	IPAddress string     `gorm:"size:50" json:"ip_address,omitempty"`
	UserAgent string     `gorm:"type:text" json:"user_agent,omitempty"`
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`

	Tenant *Tenant `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
}

// OAuthBinding 第三方OAuth绑定
type OAuthBinding struct {
	ID               int64      `gorm:"primaryKey" json:"id"`
	TenantID         int64      `gorm:"not null;index" json:"tenant_id"`
	UserID           int64      `gorm:"not null;index" json:"user_id"`
	Provider         string     `gorm:"size:50;not null;index" json:"provider"`
	ProviderUserID   string     `gorm:"size:200;not null" json:"provider_user_id"`
	ProviderUsername string     `gorm:"size:200" json:"provider_username,omitempty"`
	ProviderAvatar   string     `gorm:"size:500" json:"provider_avatar,omitempty"`
	ProviderEmail    string     `gorm:"size:200" json:"provider_email,omitempty"`
	AccessToken      string     `gorm:"type:text" json:"-"`
	RefreshToken     string     `gorm:"type:text" json:"-"`
	TokenExpiresAt   *time.Time `json:"token_expires_at,omitempty"`
	ExtraData        JSONB      `gorm:"type:jsonb" json:"extra_data,omitempty"`
	IsActive         bool       `gorm:"default:true" json:"is_active"`
	CreatedAt        time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt        *time.Time `gorm:"index" json:"deleted_at,omitempty"`

	Tenant *Tenant `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
	User   *User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// LoginBlacklist 登录黑名单
type LoginBlacklist struct {
	ID        int64      `gorm:"primaryKey" json:"id"`
	TenantID  *int64     `gorm:"index" json:"tenant_id,omitempty"`
	RuleType  string     `gorm:"size:50;not null;index" json:"rule_type"`
	RuleValue string     `gorm:"size:500;not null" json:"rule_value"`
	Reason    string     `gorm:"size:500" json:"reason,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CreatedBy *int64     `gorm:"index" json:"created_by,omitempty"`
	IsActive  bool       `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`

	Tenant  *Tenant `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
	Creator *User   `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

// LoginWhitelist 登录白名单
type LoginWhitelist struct {
	ID          int64      `gorm:"primaryKey" json:"id"`
	TenantID    *int64     `gorm:"index" json:"tenant_id,omitempty"`
	RuleType    string     `gorm:"size:50;not null;index" json:"rule_type"`
	RuleValue   string     `gorm:"size:500;not null" json:"rule_value"`
	Description string     `gorm:"size:500" json:"description,omitempty"`
	CreatedBy   *int64     `gorm:"index" json:"created_by,omitempty"`
	IsActive    bool       `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   *time.Time `gorm:"index" json:"deleted_at,omitempty"`

	Tenant  *Tenant `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
	Creator *User   `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

// LoginAttempt 登录尝试记录
type LoginAttempt struct {
	ID                int64     `gorm:"primaryKey" json:"id"`
	TenantID          *int64    `gorm:"index" json:"tenant_id,omitempty"`
	Identifier        string    `gorm:"size:200;not null;index" json:"identifier"`
	IdentifierType    string    `gorm:"size:50;not null" json:"identifier_type"`
	IsSuccess         bool      `gorm:"default:false" json:"is_success"`
	FailureReason     string    `gorm:"size:200" json:"failure_reason,omitempty"`
	IPAddress         string    `gorm:"size:50;index" json:"ip_address,omitempty"`
	UserAgent         string    `gorm:"type:text" json:"user_agent,omitempty"`
	DeviceFingerprint string    `gorm:"size:100" json:"device_fingerprint,omitempty"`
	CreatedAt         time.Time `gorm:"autoCreateTime;index" json:"created_at"`

	Tenant *Tenant `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
}

// OAuthConfig OAuth应用配置
type OAuthConfig struct {
	ID          int64      `gorm:"primaryKey" json:"id"`
	TenantID    int64      `gorm:"not null;index" json:"tenant_id"`
	Provider    string     `gorm:"size:50;not null;index" json:"provider"`
	AppID       string     `gorm:"size:200;not null" json:"app_id"`
	AppSecret   string     `gorm:"type:text;not null" json:"-"`
	RedirectURI string     `gorm:"size:500" json:"redirect_uri,omitempty"`
	ExtraConfig JSONB      `gorm:"type:jsonb" json:"extra_config,omitempty"`
	IsActive    bool       `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   *time.Time `gorm:"index" json:"deleted_at,omitempty"`

	Tenant *Tenant `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
}

// TableName 指定表名
func (SMSCode) TableName() string {
	return "sms_codes"
}

func (OAuthBinding) TableName() string {
	return "oauth_bindings"
}

func (LoginBlacklist) TableName() string {
	return "login_blacklist"
}

func (LoginWhitelist) TableName() string {
	return "login_whitelist"
}

func (LoginAttempt) TableName() string {
	return "login_attempts"
}

func (OAuthConfig) TableName() string {
	return "oauth_configs"
}

// 验证码类型常量
const (
	CodeTypeLogin         = "login"
	CodeTypeRegister      = "register"
	CodeTypeResetPassword = "reset_password"
	CodeTypeBindPhone     = "bind_phone"
)

// OAuth提供商常量
const (
	OAuthProviderFeishu   = "feishu"
	OAuthProviderWechat   = "wechat"
	OAuthProviderDingtalk = "dingtalk"
	OAuthProviderGithub   = "github"
)

// 黑白名单规则类型常量
const (
	RuleTypeIP       = "ip"
	RuleTypeIPRange  = "ip_range"
	RuleTypeUser     = "user"
	RuleTypeDevice   = "device"
	RuleTypePhone    = "phone"
)

// 登录标识类型常量
const (
	IdentifierTypeUsername = "username"
	IdentifierTypePhone    = "phone"
	IdentifierTypeIP       = "ip"
)

// 登录方式常量
const (
	LoginMethodPassword = "password"
	LoginMethodPhone    = "phone"
	LoginMethodOAuth    = "oauth"
)
