package models

import "time"

// CustomerAccount 客户账户表
type CustomerAccount struct {
	TenantModel
	CustomerID  int64   `gorm:"not null;index" json:"customer_id"`
	AccountType string  `gorm:"size:20;not null" json:"account_type"` // fund:资金账户, consumption:消耗账户
	Balance     float64 `gorm:"type:decimal(15,2);default:0" json:"balance"`
	CashBalance float64 `gorm:"type:decimal(15,2);default:0" json:"cash_balance"`

	// 关联
	Customer     *Customer             `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	Transactions []AccountTransaction  `gorm:"foreignKey:AccountID" json:"transactions,omitempty"`
}

// AccountTransaction 账户交易明细表
type AccountTransaction struct {
	ID                int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID          int64     `gorm:"not null;index" json:"tenant_id"`
	AccountID         int64     `gorm:"not null;index" json:"account_id"`
	OrderID           *int64    `gorm:"index" json:"order_id"`
	TransactionType   string    `gorm:"size:50;not null" json:"transaction_type"` // recharge, transfer, consume, refund
	Amount            float64   `gorm:"type:decimal(15,2);not null" json:"amount"`
	BalanceBefore     float64   `gorm:"type:decimal(15,2)" json:"balance_before"`
	BalanceAfter      float64   `gorm:"type:decimal(15,2)" json:"balance_after"`
	CashBalanceBefore float64   `gorm:"type:decimal(15,2)" json:"cash_balance_before"`
	CashBalanceAfter  float64   `gorm:"type:decimal(15,2)" json:"cash_balance_after"`
	Notes             string    `gorm:"type:text" json:"notes"`
	CreatedAt         time.Time `gorm:"autoCreateTime" json:"created_at"`

	// 关联
	Account *CustomerAccount `gorm:"foreignKey:AccountID" json:"account,omitempty"`
	Order   *Order           `gorm:"foreignKey:OrderID" json:"order,omitempty"`
}

// 账户类型常量
const (
	AccountTypeFund        = "fund"        // 资金账户
	AccountTypeConsumption = "consumption" // 消耗账户
)

// 交易类型常量
const (
	TransactionTypeRecharge = "recharge" // 充值
	TransactionTypeTransfer = "transfer" // 转账
	TransactionTypeConsume  = "consume"  // 消耗
	TransactionTypeRefund   = "refund"   // 退款
)

func (CustomerAccount) TableName() string {
	return "customer_accounts"
}

func (AccountTransaction) TableName() string {
	return "account_transactions"
}
