package models

import (
	"time"
)

// PaymentAccount 收款账户表
type PaymentAccount struct {
	ID       int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID int64  `gorm:"not null;index" json:"tenant_id"`

	// 基本信息
	AccountName   string `gorm:"size:200;not null" json:"account_name"`
	AccountType   string `gorm:"size:50;not null;index" json:"account_type"` // public, private
	PaymentMethod string `gorm:"size:50;not null;index" json:"payment_method"` // bank_card, alipay, wechat, api

	// 账户信息
	AccountNumber string `gorm:"size:200" json:"account_number,omitempty"`
	AccountHolder string `gorm:"size:200" json:"account_holder,omitempty"`
	BankName      string `gorm:"size:200" json:"bank_name,omitempty"`
	BankBranch    string `gorm:"size:200" json:"bank_branch,omitempty"`

	// API接口信息
	APIConfigID   *int64 `gorm:"index" json:"api_config_id,omitempty"`
	APIType       string `gorm:"size:50" json:"api_type,omitempty"`
	APIMerchantID string `gorm:"size:200" json:"api_merchant_id,omitempty"`
	APIAppID      string `gorm:"size:200" json:"api_app_id,omitempty"`
	APISecretKey  string `gorm:"type:text" json:"api_secret_key,omitempty"` // 加密存储
	APIPublicKey  string `gorm:"type:text" json:"api_public_key,omitempty"`
	APIEndpoint   string `gorm:"size:500" json:"api_endpoint,omitempty"`

	// 轮询策略配置
	Status         string `gorm:"size:20;default:active;index" json:"status"` // active, inactive, maintenance
	IsAutoRotate   bool   `gorm:"default:true;index" json:"is_auto_rotate"`
	RotateStrategy string `gorm:"size:50;default:weight" json:"rotate_strategy"` // weight, round_robin, balance, frequency, random
	RotateWeight   int    `gorm:"default:1" json:"rotate_weight"`  // 1-100
	RotatePriority int    `gorm:"default:0" json:"rotate_priority"` // 优先级

	// 余额和限额
	CurrentBalance float64 `gorm:"type:decimal(15,2);default:0" json:"current_balance"`
	DailyLimit     float64 `gorm:"type:decimal(15,2)" json:"daily_limit,omitempty"`
	SingleLimit    float64 `gorm:"type:decimal(15,2)" json:"single_limit,omitempty"`
	MonthlyLimit   float64 `gorm:"type:decimal(15,2)" json:"monthly_limit,omitempty"`

	// 使用统计
	UseFrequency   int       `gorm:"default:0" json:"use_frequency"`
	TotalAmount    float64   `gorm:"type:decimal(15,2);default:0" json:"total_amount"`
	TotalCount     int       `gorm:"default:0" json:"total_count"`
	TodayAmount    float64   `gorm:"type:decimal(15,2);default:0" json:"today_amount"`
	TodayCount     int       `gorm:"default:0" json:"today_count"`
	ThisMonthAmount float64  `gorm:"type:decimal(15,2);default:0" json:"this_month_amount"`
	ThisMonthCount  int      `gorm:"default:0" json:"this_month_count"`
	LastUsedAt     *time.Time `json:"last_used_at,omitempty"`

	// 附加信息
	Description string `gorm:"type:text" json:"description,omitempty"`
	Remark      string `gorm:"type:text" json:"remark,omitempty"`

	// 时间戳
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`

	// 关联
	Tenant    *Tenant           `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
	APIConfig *PaymentAPIConfig `gorm:"foreignKey:APIConfigID" json:"api_config,omitempty"`
}

// PaymentAPIConfig 支付接口配置表
type PaymentAPIConfig struct {
	ID       int64 `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID int64 `gorm:"not null;index" json:"tenant_id"`

	// 提供商信息
	ProviderName string `gorm:"size:200;not null" json:"provider_name"` // 中国银行, 工商银行, 支付宝
	ProviderCode string `gorm:"size:100;not null;index" json:"provider_code"` // BOC, ICBC, ALIPAY
	ProviderType string `gorm:"size:50;not null" json:"provider_type"` // bank, payment_platform, acquirer

	// API配置
	APIVersion        string `gorm:"size:50" json:"api_version,omitempty"`
	APIEndpoint       string `gorm:"size:500;not null" json:"api_endpoint"`
	APISandboxEndpoint string `gorm:"size:500" json:"api_sandbox_endpoint,omitempty"`
	IsSandbox         bool   `gorm:"default:false" json:"is_sandbox"`

	// 认证信息
	MerchantID string `gorm:"size:200" json:"merchant_id,omitempty"`
	AppID      string `gorm:"size:200" json:"app_id,omitempty"`
	AppSecret  string `gorm:"type:text" json:"app_secret,omitempty"` // 加密存储
	PublicKey  string `gorm:"type:text" json:"public_key,omitempty"`
	PrivateKey string `gorm:"type:text" json:"private_key,omitempty"` // 加密存储
	CertPath   string `gorm:"size:500" json:"cert_path,omitempty"`

	// 扩展配置
	ExtraConfig JSONB `gorm:"type:jsonb" json:"extra_config,omitempty"`

	// 功能支持
	SupportsPayment  bool `gorm:"default:true" json:"supports_payment"`
	SupportsQuery    bool `gorm:"default:true" json:"supports_query"`
	SupportsRefund   bool `gorm:"default:false" json:"supports_refund"`
	SupportsTransfer bool `gorm:"default:false" json:"supports_transfer"`

	// 状态
	Status string `gorm:"size:20;default:active;index" json:"status"` // active, inactive

	// 附加信息
	Description string `gorm:"type:text" json:"description,omitempty"`
	Remark      string `gorm:"type:text" json:"remark,omitempty"`

	// 时间戳
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`

	// 关联
	Tenant *Tenant `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
}

// CreditRecord 授信记录表
type CreditRecord struct {
	ID         int64 `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID   int64 `gorm:"not null;index" json:"tenant_id"`
	CustomerID int64 `gorm:"not null;index" json:"customer_id"`

	// 授信信息
	RecordType   string  `gorm:"size:50;not null;index" json:"record_type"` // grant, adjust, repay, consume
	CreditAmount float64 `gorm:"type:decimal(15,2);not null" json:"credit_amount"`
	BeforeCredit float64 `gorm:"type:decimal(15,2)" json:"before_credit,omitempty"`
	AfterCredit  float64 `gorm:"type:decimal(15,2)" json:"after_credit,omitempty"`
	BeforeLimit  float64 `gorm:"type:decimal(15,2)" json:"before_limit,omitempty"`
	AfterLimit   float64 `gorm:"type:decimal(15,2)" json:"after_limit,omitempty"`

	// 申请信息
	ApplyReason string     `gorm:"type:text" json:"apply_reason,omitempty"`
	ApplicantID *int64     `gorm:"index" json:"applicant_id,omitempty"`
	ApplyTime   time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"apply_time"`

	// 审批信息
	ApprovalStatus      string     `gorm:"size:50;default:pending;index" json:"approval_status"` // pending, approved, rejected, cancelled
	ApproverID          *int64     `gorm:"index" json:"approver_id,omitempty"`
	AssignedApproverID  *int64     `gorm:"index" json:"assigned_approver_id,omitempty"` // 超级管理员指定
	AutoAssigned        bool       `gorm:"default:false" json:"auto_assigned"`
	ApprovalTime        *time.Time `json:"approval_time,omitempty"`
	ApprovalRemark      string     `gorm:"type:text" json:"approval_remark,omitempty"`

	// 还款信息
	RepayAmount  float64 `gorm:"type:decimal(15,2)" json:"repay_amount,omitempty"`
	RepayMethod  string  `gorm:"size:50" json:"repay_method,omitempty"` // cash, transfer, deduction
	RepayOrderID *int64  `json:"repay_order_id,omitempty"`

	// 附件
	Attachments JSONB `gorm:"type:jsonb" json:"attachments,omitempty"`

	// 附加信息
	Remark string `gorm:"type:text" json:"remark,omitempty"`

	// 时间戳
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`

	// 关联
	Tenant           *Tenant   `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
	Customer         *Customer `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	Applicant        *User     `gorm:"foreignKey:ApplicantID" json:"applicant,omitempty"`
	Approver         *User     `gorm:"foreignKey:ApproverID" json:"approver,omitempty"`
	AssignedApprover *User     `gorm:"foreignKey:AssignedApproverID" json:"assigned_approver,omitempty"`
	RepayOrder       *Order    `gorm:"foreignKey:RepayOrderID" json:"repay_order,omitempty"`
}

// TableName 指定表名
func (PaymentAccount) TableName() string {
	return "payment_accounts"
}

func (PaymentAPIConfig) TableName() string {
	return "payment_api_configs"
}

func (CreditRecord) TableName() string {
	return "credit_records"
}

// PaymentAccountType 收款账户类型常量
const (
	PaymentAccountTypePublic  = "public"  // 对公
	PaymentAccountTypePrivate = "private" // 对私
)

// PaymentMethod 支付方式常量
const (
	PaymentMethodBankCard = "bank_card" // 银行卡
	PaymentMethodAlipay   = "alipay"    // 支付宝
	PaymentMethodWechat   = "wechat"    // 微信
	PaymentMethodAPI      = "api"       // API接口
)

// PaymentAccountStatus 收款账户状态常量
const (
	PaymentAccountStatusActive      = "active"      // 启用
	PaymentAccountStatusInactive    = "inactive"    // 停用
	PaymentAccountStatusMaintenance = "maintenance" // 维护中
)

// RotateStrategy 轮询策略常量
const (
	RotateStrategyWeight     = "weight"      // 权重
	RotateStrategyRoundRobin = "round_robin" // 轮流
	RotateStrategyBalance    = "balance"     // 余额
	RotateStrategyFrequency  = "frequency"   // 频率
	RotateStrategyRandom     = "random"      // 随机
)

// ProviderType 提供商类型常量
const (
	ProviderTypeBank            = "bank"             // 银行
	ProviderTypePaymentPlatform = "payment_platform" // 支付平台
	ProviderTypeAcquirer        = "acquirer"         // 收单机构
)

// CreditRecordType 授信记录类型常量
const (
	CreditRecordTypeGrant   = "grant"   // 授信
	CreditRecordTypeAdjust  = "adjust"  // 调整
	CreditRecordTypeRepay   = "repay"   // 还款
	CreditRecordTypeConsume = "consume" // 消费
)

// CreditApprovalStatus 授信审批状态常量
const (
	CreditApprovalStatusPending   = "pending"   // 待审批
	CreditApprovalStatusApproved  = "approved"  // 已批准
	CreditApprovalStatusRejected  = "rejected"  // 已拒绝
	CreditApprovalStatusCancelled = "cancelled" // 已取消
)

// RepayMethod 还款方式常量
const (
	RepayMethodCash      = "cash"      // 现金
	RepayMethodTransfer  = "transfer"  // 转账
	RepayMethodDeduction = "deduction" // 扣款
)
