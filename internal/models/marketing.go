package models

import (
	"time"

	"gorm.io/datatypes"
)

// 活动类型常量
const (
	ActivityTypePromotion = "promotion" // 促销活动
	ActivityTypeDiscount  = "discount"  // 折扣活动
	ActivityTypeGift      = "gift"      // 赠品活动
	ActivityTypeLottery   = "lottery"   // 抽奖活动
)

// 活动展示位置常量
const (
	DisplayPositionLogin    = "login"    // 登录页
	DisplayPositionRecharge = "recharge" // 充值页
	DisplayPositionDashboard = "dashboard" // 仪表板
	DisplayPositionPopup    = "popup"    // 弹窗
)

// 横幅位置常量
const (
	BannerPositionTop     = "top"      // 顶部
	BannerPositionBottom  = "bottom"   // 底部
	BannerPositionSidebar = "sidebar"  // 侧边栏
	BannerPositionPopup   = "popup"    // 弹窗
	BannerPositionLogin   = "login"    // 登录页
	BannerPositionRecharge = "recharge" // 充值页
)

// 主题类型常量
const (
	ThemeTypeDefault  = "default"  // 默认主题
	ThemeTypeFestival = "festival" // 节日主题
	ThemeTypeCustom   = "custom"   // 自定义主题
)

// 节日类型常量
const (
	FestivalTypeNewYear        = "new_year"        // 元旦
	FestivalTypeSpringFestival = "spring_festival" // 春节
	FestivalTypeLantern        = "lantern"         // 元宵节
	FestivalTypeDragonBoat     = "dragon_boat"     // 端午节
	FestivalTypeMidAutumn      = "mid_autumn"      // 中秋节
	FestivalTypeNationalDay    = "national_day"    // 国庆节
	FestivalTypeChristmas      = "christmas"       // 圣诞节
)

// 奖励状态常量
const (
	RewardStatusPending = "pending" // 待发放
	RewardStatusGranted = "granted" // 已发放
	RewardStatusExpired = "expired" // 已过期
)

// 报表类别常量
const (
	ReportCategoryGeneral  = "general"  // 通用
	ReportCategoryFinance  = "finance"  // 财务
	ReportCategoryCustomer = "customer" // 客户
	ReportCategoryOrder    = "order"    // 订单
)

// Activity 活动模型
type Activity struct {
	ID        int64          `gorm:"primaryKey" json:"id"`
	TenantID  int64          `gorm:"not null;index" json:"tenant_id"`

	// 基本信息
	Name            string         `gorm:"size:200;not null" json:"name"`
	Description     string         `gorm:"type:text" json:"description"`
	ActivityType    string         `gorm:"size:50;not null;default:promotion" json:"activity_type"`

	// 展示配置
	DisplayPosition string         `gorm:"size:50;default:recharge" json:"display_position"`
	ImageURL        string         `gorm:"size:500" json:"image_url"`
	LinkURL         string         `gorm:"size:500" json:"link_url"`
	Content         datatypes.JSON `gorm:"type:jsonb" json:"content"`

	// 时间控制
	StartTime time.Time `gorm:"not null" json:"start_time"`
	EndTime   time.Time `gorm:"not null" json:"end_time"`

	// 参与规则
	Rules           datatypes.JSON `gorm:"type:jsonb" json:"rules"`
	MaxParticipants int            `gorm:"default:0" json:"max_participants"`

	// 状态
	IsActive  bool `gorm:"default:true" json:"is_active"`
	SortOrder int  `gorm:"default:0" json:"sort_order"`

	// 统计
	ViewCount        int `gorm:"default:0" json:"view_count"`
	ParticipantCount int `gorm:"default:0" json:"participant_count"`

	// 审计
	CreatedBy int64      `json:"created_by"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}

func (Activity) TableName() string {
	return "activities"
}

// ActivityParticipant 活动参与记录模型
type ActivityParticipant struct {
	ID         int64 `gorm:"primaryKey" json:"id"`
	TenantID   int64 `gorm:"not null;index" json:"tenant_id"`
	ActivityID int64 `gorm:"not null;index" json:"activity_id"`
	CustomerID int64 `gorm:"not null;index" json:"customer_id"`

	// 参与信息
	ParticipatedAt    time.Time      `gorm:"default:CURRENT_TIMESTAMP" json:"participated_at"`
	ParticipationData datatypes.JSON `gorm:"type:jsonb" json:"participation_data"`

	// 奖励信息
	RewardType      string     `gorm:"size:50" json:"reward_type"`
	RewardAmount    float64    `gorm:"type:decimal(15,2);default:0" json:"reward_amount"`
	RewardStatus    string     `gorm:"size:50;default:pending" json:"reward_status"`
	RewardGrantedAt *time.Time `json:"reward_granted_at,omitempty"`

	// 备注
	Remark string `gorm:"type:text" json:"remark"`

	// 关联
	Activity *Activity `gorm:"foreignKey:ActivityID" json:"activity,omitempty"`
	Customer *Customer `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
}

func (ActivityParticipant) TableName() string {
	return "activity_participants"
}

// Banner 横幅广告模型
type Banner struct {
	ID       int64 `gorm:"primaryKey" json:"id"`
	TenantID int64 `gorm:"not null;index" json:"tenant_id"`

	// 基本信息
	Title       string `gorm:"size:200;not null" json:"title"`
	Description string `gorm:"type:text" json:"description"`

	// 展示内容
	ImageURL   string `gorm:"size:500;not null" json:"image_url"`
	LinkURL    string `gorm:"size:500" json:"link_url"`
	LinkTarget string `gorm:"size:20;default:_blank" json:"link_target"`

	// 展示位置
	Position string `gorm:"size:50;not null;default:top" json:"position"`

	// 时间控制
	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`

	// 状态
	IsActive  bool `gorm:"default:true" json:"is_active"`
	SortOrder int  `gorm:"default:0" json:"sort_order"`

	// 统计
	ViewCount  int `gorm:"default:0" json:"view_count"`
	ClickCount int `gorm:"default:0" json:"click_count"`

	// 审计
	CreatedBy int64      `json:"created_by"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}

func (Banner) TableName() string {
	return "banners"
}

// Theme 主题配置模型
type Theme struct {
	ID       int64 `gorm:"primaryKey" json:"id"`
	TenantID int64 `gorm:"not null;index" json:"tenant_id"`

	// 基本信息
	Name         string `gorm:"size:100;not null" json:"name"`
	Description  string `gorm:"type:text" json:"description"`
	ThemeType    string `gorm:"size:50;default:custom" json:"theme_type"`
	FestivalType string `gorm:"size:50" json:"festival_type,omitempty"`

	// 颜色配置
	PrimaryColor    string `gorm:"size:20;default:#1890ff" json:"primary_color"`
	SecondaryColor  string `gorm:"size:20;default:#52c41a" json:"secondary_color"`
	BackgroundColor string `gorm:"size:20;default:#f0f2f5" json:"background_color"`
	TextColor       string `gorm:"size:20;default:#333333" json:"text_color"`

	// 图片资源
	LogoURL            string `gorm:"size:500" json:"logo_url"`
	FaviconURL         string `gorm:"size:500" json:"favicon_url"`
	BackgroundImageURL string `gorm:"size:500" json:"background_image_url"`
	LoginBackgroundURL string `gorm:"size:500" json:"login_background_url"`

	// 其他配置
	CustomCSS   string         `gorm:"type:text" json:"custom_css"`
	ExtraConfig datatypes.JSON `gorm:"type:jsonb" json:"extra_config"`

	// 状态
	IsDefault bool `gorm:"default:false" json:"is_default"`
	IsActive  bool `gorm:"default:true" json:"is_active"`

	// 审计
	CreatedBy int64      `json:"created_by"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}

func (Theme) TableName() string {
	return "themes"
}

// CustomReport 自定义报表定义模型
type CustomReport struct {
	ID       int64 `gorm:"primaryKey" json:"id"`
	TenantID int64 `gorm:"not null;index" json:"tenant_id"`

	// 基本信息
	Name        string `gorm:"size:200;not null" json:"name"`
	Description string `gorm:"type:text" json:"description"`
	Category    string `gorm:"size:50;default:general" json:"category"`

	// 数据源配置
	BaseTable    string         `gorm:"size:100;not null" json:"base_table"`
	SelectFields datatypes.JSON `gorm:"type:jsonb;not null" json:"select_fields"`
	JoinConfig   datatypes.JSON `gorm:"type:jsonb" json:"join_config"`

	// 查询配置
	FilterConfig   datatypes.JSON `gorm:"type:jsonb" json:"filter_config"`
	GroupByFields  datatypes.JSON `gorm:"type:jsonb" json:"group_by_fields"`
	OrderByConfig  datatypes.JSON `gorm:"type:jsonb" json:"order_by_config"`

	// 显示配置
	ColumnConfig datatypes.JSON `gorm:"type:jsonb" json:"column_config"`
	ChartConfig  datatypes.JSON `gorm:"type:jsonb" json:"chart_config"`

	// 状态
	IsPublic bool `gorm:"default:false" json:"is_public"`
	IsActive bool `gorm:"default:true" json:"is_active"`

	// 审计
	CreatedBy int64      `json:"created_by"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}

func (CustomReport) TableName() string {
	return "custom_reports"
}

// ReportExecution 报表执行记录模型
type ReportExecution struct {
	ID       int64 `gorm:"primaryKey" json:"id"`
	TenantID int64 `gorm:"not null;index" json:"tenant_id"`
	ReportID int64 `gorm:"not null;index" json:"report_id"`

	// 执行信息
	ExecutedBy int64     `gorm:"index" json:"executed_by"`
	ExecutedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"executed_at"`

	// 参数和结果
	Parameters      datatypes.JSON `gorm:"type:jsonb" json:"parameters"`
	RowCount        int            `gorm:"default:0" json:"row_count"`
	ExecutionTimeMs int            `gorm:"default:0" json:"execution_time_ms"`

	// 导出信息
	ExportFormat  string `gorm:"size:20" json:"export_format"`
	ExportFileURL string `gorm:"size:500" json:"export_file_url"`

	// 状态
	Status string `gorm:"size:50;default:completed" json:"status"`

	// 关联
	Report *CustomReport `gorm:"foreignKey:ReportID" json:"report,omitempty"`
}

func (ReportExecution) TableName() string {
	return "report_executions"
}

// SelectField 报表选择字段配置
type SelectField struct {
	Field     string `json:"field"`
	Alias     string `json:"alias"`
	Aggregate string `json:"aggregate,omitempty"` // SUM, COUNT, AVG, MAX, MIN
}

// JoinConfig 报表关联配置
type JoinConfig struct {
	Table string `json:"table"`
	On    string `json:"on"`
	Type  string `json:"type"` // LEFT, RIGHT, INNER
}

// FilterConfig 报表筛选配置
type FilterConfig struct {
	Field       string      `json:"field"`
	Operator    string      `json:"operator"` // =, !=, >, <, >=, <=, LIKE, IN, BETWEEN
	Value       interface{} `json:"value"`
	IsParameter bool        `json:"is_parameter"` // 是否为参数（运行时传入）
}

// OrderByConfig 报表排序配置
type OrderByConfig struct {
	Field     string `json:"field"`
	Direction string `json:"direction"` // ASC, DESC
}
