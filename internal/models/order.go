package models

import "time"

// Order 订单表
type Order struct {
	TenantModel
	CustomerID     int64      `gorm:"not null;index" json:"customer_id"`
	OrderNo        string     `gorm:"size:100;unique;not null" json:"order_no"`
	OrderType      string     `gorm:"size:20;not null;index" json:"order_type"` // recharge, settlement, transfer
	Amount         float64    `gorm:"type:decimal(15,2);not null" json:"amount"`
	Status         string     `gorm:"size:50;not null;index" json:"status"`
	PaymentMethod  string     `gorm:"size:50" json:"payment_method"`
	PaymentChannel string     `gorm:"size:100" json:"payment_channel"`
	TransactionID  string     `gorm:"size:255" json:"transaction_id"`
	PaymentTime    *time.Time `json:"payment_time"`
	VerifiedTime   *time.Time `json:"verified_time"`
	VerifiedBy     *int64     `gorm:"index" json:"verified_by"`
	Notes          string     `gorm:"type:text" json:"notes"`
	Metadata       JSONB      `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedBy      *int64     `gorm:"index" json:"created_by"`

	// 关联
	Customer   *Customer            `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	Verifier   *User                `gorm:"foreignKey:VerifiedBy" json:"verifier,omitempty"`
	Creator    *User                `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	StatusLogs []OrderStatusLog     `gorm:"foreignKey:OrderID" json:"status_logs,omitempty"`
	Transactions []AccountTransaction `gorm:"foreignKey:OrderID" json:"transactions,omitempty"`
}

// OrderStatusLog 订单状态变更日志
type OrderStatusLog struct {
	ID         int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID    int64      `gorm:"not null;index" json:"order_id"`
	FromStatus string     `gorm:"size:50" json:"from_status"`
	ToStatus   string     `gorm:"size:50;not null" json:"to_status"`
	OperatorID *int64     `gorm:"index" json:"operator_id"`
	Notes      string     `gorm:"type:text" json:"notes"`
	CreatedAt  time.Time  `gorm:"autoCreateTime" json:"created_at"`

	// 关联
	Order    *Order `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	Operator *User  `gorm:"foreignKey:OperatorID" json:"operator,omitempty"`
}

// 订单状态常量
const (
	// 充值订单状态
	OrderStatusPending   = "pending"   // 待支付
	OrderStatusPaid      = "paid"      // 已支付
	OrderStatusVerified  = "verified"  // 财务已确认
	OrderStatusCompleted = "completed" // 已完成
	OrderStatusFailed    = "failed"    // 失败
	OrderStatusCancelled = "cancelled" // 已取消

	// 转账订单状态
	OrderStatusProcessing = "processing" // 处理中

	// 订单类型
	OrderTypeRecharge   = "recharge"   // 充值订单
	OrderTypeSettlement = "settlement" // 结算订单
	OrderTypeTransfer   = "transfer"   // 转账订单
)

func (Order) TableName() string {
	return "orders"
}

func (OrderStatusLog) TableName() string {
	return "order_status_logs"
}
