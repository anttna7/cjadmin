package models

import "time"

// Customer 客户信息表
type Customer struct {
	TenantModel
	CustomerCode  string `gorm:"size:100;not null" json:"customer_code"`
	CustomerName  string `gorm:"size:255;not null" json:"customer_name"`
	CustomerType  string `gorm:"size:50" json:"customer_type"` // 企业/个人
	ContactPerson string `gorm:"size:100" json:"contact_person"`
	Phone         string `gorm:"size:50" json:"phone"`
	Email         string `gorm:"size:255" json:"email"`
	Address       string `gorm:"type:text" json:"address"`
	Industry      string `gorm:"size:100" json:"industry"`
	Source        string `gorm:"size:100" json:"source"` // 客户来源
	AssignedTo    *int64 `gorm:"index" json:"assigned_to"`
	Status        string `gorm:"size:20;default:active" json:"status"`
	CustomFields  JSONB  `gorm:"type:jsonb" json:"custom_fields,omitempty"`

	// 账户和代理商信息（第19阶段新增）
	AccountID string `gorm:"size:100;index" json:"account_id,omitempty"`
	AgentName string `gorm:"size:200" json:"agent_name,omitempty"`
	AgentID   string `gorm:"size:100;index" json:"agent_id,omitempty"`
	Channel   string `gorm:"size:100" json:"channel,omitempty"`

	// 报备和行业信息（第19阶段新增）
	ReportTags      []string `gorm:"type:text[]" json:"report_tags,omitempty"`
	ReportIndustry  string   `gorm:"size:200" json:"report_industry,omitempty"`
	IndustryLevel1  string   `gorm:"size:200" json:"industry_level1,omitempty"`
	IndustryLevel2  string   `gorm:"size:200" json:"industry_level2,omitempty"`

	// 返点配置 - 对公充值（第19阶段新增）
	PublicRebateRate float64 `gorm:"type:decimal(5,2);default:0" json:"public_rebate_rate"`
	PublicCashRate   float64 `gorm:"type:decimal(5,2);default:100" json:"public_cash_rate"`
	PublicGiftRate   float64 `gorm:"type:decimal(5,2);default:0" json:"public_gift_rate"`

	// 返点配置 - 对私充值（第19阶段新增）
	PrivateRebateRate float64 `gorm:"type:decimal(5,2);default:0" json:"private_rebate_rate"`
	PrivateCashRate   float64 `gorm:"type:decimal(5,2);default:100" json:"private_cash_rate"`
	PrivateGiftRate   float64 `gorm:"type:decimal(5,2);default:0" json:"private_gift_rate"`

	// 业务信息（第19阶段新增）
	BusinessPlatform string `gorm:"size:200" json:"business_platform,omitempty"`
	ContractPeriod   string `gorm:"size:100" json:"contract_period,omitempty"`

	// 关联
	AssignedUser *User      `gorm:"foreignKey:AssignedTo" json:"assigned_user,omitempty"`
	Contracts    []Contract `gorm:"foreignKey:CustomerID" json:"contracts,omitempty"`
	Accounts     []CustomerAccount `gorm:"foreignKey:CustomerID" json:"accounts,omitempty"`
}

// Contract 合同归档表
type Contract struct {
	TenantModel
	CustomerID     int64      `gorm:"not null;index" json:"customer_id"`
	ContractNo     string     `gorm:"size:100;not null" json:"contract_no"`
	ContractName   string     `gorm:"size:255" json:"contract_name"`
	ContractAmount float64    `gorm:"type:decimal(15,2)" json:"contract_amount"`
	StartDate      *time.Time `gorm:"type:date" json:"start_date"`
	EndDate        *time.Time `gorm:"type:date" json:"end_date"`
	FileURL        string     `gorm:"size:500" json:"file_url"`
	Status         string     `gorm:"size:20;default:active" json:"status"`
	Notes          string     `gorm:"type:text" json:"notes"`

	// 关联
	Customer *Customer `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
}

func (Customer) TableName() string {
	return "customers"
}

func (Contract) TableName() string {
	return "contracts"
}
