package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/anttna7/cjadmin/internal/database"
	"github.com/anttna7/cjadmin/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrderService struct{}

func NewOrderService() *OrderService {
	return &OrderService{}
}

type CreateRechargeOrderRequest struct {
	CustomerID     int64   `json:"customer_id" binding:"required"`
	Amount         float64 `json:"amount" binding:"required,gt=0"`
	PaymentMethod  string  `json:"payment_method"`
	PaymentChannel string  `json:"payment_channel"`
}

// CreateRechargeOrder 创建充值订单
func (s *OrderService) CreateRechargeOrder(tenantID int64, req *CreateRechargeOrderRequest, createdBy int64) (*models.Order, error) {
	// 验证客户是否存在
	var customer models.Customer
	if err := database.DB.Where("id = ? AND tenant_id = ?", req.CustomerID, tenantID).
		First(&customer).Error; err != nil {
		return nil, errors.New("客户不存在")
	}

	// 生成订单号
	orderNo := fmt.Sprintf("RCH%s%s", time.Now().Format("20060102"), uuid.New().String()[:8])

	order := &models.Order{
		TenantModel: models.TenantModel{TenantID: tenantID},
		CustomerID:  req.CustomerID,
		OrderNo:     orderNo,
		OrderType:   models.OrderTypeRecharge,
		Amount:      req.Amount,
		Status:      models.OrderStatusPending,
		PaymentMethod: req.PaymentMethod,
		PaymentChannel: req.PaymentChannel,
		CreatedBy:   &createdBy,
	}

	if err := database.DB.Create(order).Error; err != nil {
		return nil, err
	}

	// 记录状态变更
	s.logStatusChange(order.ID, "", models.OrderStatusPending, &createdBy, "创建充值订单")

	return order, nil
}

// PaymentCallback 支付回调（模拟支付完成）
func (s *OrderService) PaymentCallback(orderNo, transactionID string) error {
	var order models.Order
	if err := database.DB.Where("order_no = ?", orderNo).First(&order).Error; err != nil {
		return err
	}

	if order.Status != models.OrderStatusPending {
		return errors.New("订单状态不正确")
	}

	// 更新订单状态为已支付
	now := time.Now()
	updates := map[string]interface{}{
		"status":         models.OrderStatusPaid,
		"transaction_id": transactionID,
		"payment_time":   &now,
	}

	if err := database.DB.Model(&order).Updates(updates).Error; err != nil {
		return err
	}

	// 记录状态变更
	s.logStatusChange(order.ID, models.OrderStatusPending, models.OrderStatusPaid, nil, "支付完成")

	return nil
}

// VerifyRechargeOrder 财务审核充值订单
func (s *OrderService) VerifyRechargeOrder(tenantID, orderID, verifiedBy int64, approved bool, notes string) error {
	var order models.Order
	if err := database.DB.Where("id = ? AND tenant_id = ?", orderID, tenantID).
		First(&order).Error; err != nil {
		return err
	}

	if order.OrderType != models.OrderTypeRecharge {
		return errors.New("订单类型错误")
	}

	if order.Status != models.OrderStatusPaid {
		return errors.New("订单状态不正确，必须是已支付状态")
	}

	if !approved {
		// 审核不通过，订单失败
		order.Status = models.OrderStatusFailed
		order.VerifiedBy = &verifiedBy
		now := time.Now()
		order.VerifiedTime = &now
		order.Notes = notes

		if err := database.DB.Save(&order).Error; err != nil {
			return err
		}

		s.logStatusChange(order.ID, models.OrderStatusPaid, models.OrderStatusFailed, &verifiedBy, notes)
		return nil
	}

	// 开始数据库事务
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// 更新订单状态为已审核
		now := time.Now()
		order.Status = models.OrderStatusVerified
		order.VerifiedBy = &verifiedBy
		order.VerifiedTime = &now

		if err := tx.Save(&order).Error; err != nil {
			return err
		}

		// 记录状态变更
		s.logStatusChange(order.ID, models.OrderStatusPaid, models.OrderStatusVerified, &verifiedBy, "财务审核通过")

		// 更新客户资金账户余额
		var fundAccount models.CustomerAccount
		if err := tx.Where("customer_id = ? AND account_type = ?", order.CustomerID, models.AccountTypeFund).
			First(&fundAccount).Error; err != nil {
			return err
		}

		// 记录余额变更前的值
		balanceBefore := fundAccount.Balance
		cashBalanceBefore := fundAccount.CashBalance

		// 更新余额（充值金额计入现金余额和总余额）
		fundAccount.Balance += order.Amount
		fundAccount.CashBalance += order.Amount

		if err := tx.Save(&fundAccount).Error; err != nil {
			return err
		}

		// 记录交易明细
		transaction := &models.AccountTransaction{
			TenantID:          order.TenantID,
			AccountID:         fundAccount.ID,
			OrderID:           &order.ID,
			TransactionType:   models.TransactionTypeRecharge,
			Amount:            order.Amount,
			BalanceBefore:     balanceBefore,
			BalanceAfter:      fundAccount.Balance,
			CashBalanceBefore: cashBalanceBefore,
			CashBalanceAfter:  fundAccount.CashBalance,
			Notes:             fmt.Sprintf("充值订单: %s", order.OrderNo),
		}

		if err := tx.Create(transaction).Error; err != nil {
			return err
		}

		// 更新订单状态为已完成
		order.Status = models.OrderStatusCompleted
		if err := tx.Save(&order).Error; err != nil {
			return err
		}

		s.logStatusChange(order.ID, models.OrderStatusVerified, models.OrderStatusCompleted, &verifiedBy, "余额入账完成")

		return nil
	})
}

// CreateTransferOrder 创建转账订单（从资金账户转到消耗账户）
func (s *OrderService) CreateTransferOrder(tenantID, customerID int64, amount float64, createdBy int64) (*models.Order, error) {
	if amount <= 0 {
		return nil, errors.New("转账金额必须大于0")
	}

	// 验证客户是否存在
	var customer models.Customer
	if err := database.DB.Where("id = ? AND tenant_id = ?", customerID, tenantID).
		First(&customer).Error; err != nil {
		return nil, errors.New("客户不存在")
	}

	// 开始事务
	var order *models.Order
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		// 获取资金账户
		var fundAccount models.CustomerAccount
		if err := tx.Where("customer_id = ? AND account_type = ?", customerID, models.AccountTypeFund).
			First(&fundAccount).Error; err != nil {
			return errors.New("资金账户不存在")
		}

		// 检查余额是否足够
		if fundAccount.Balance < amount {
			return errors.New("资金账户余额不足")
		}

		// 获取消耗账户
		var consumptionAccount models.CustomerAccount
		if err := tx.Where("customer_id = ? AND account_type = ?", customerID, models.AccountTypeConsumption).
			First(&consumptionAccount).Error; err != nil {
			return errors.New("消耗账户不存在")
		}

		// 生成订单号
		orderNo := fmt.Sprintf("TRF%s%s", time.Now().Format("20060102"), uuid.New().String()[:8])

		// 创建转账订单
		order = &models.Order{
			TenantModel: models.TenantModel{TenantID: tenantID},
			CustomerID:  customerID,
			OrderNo:     orderNo,
			OrderType:   models.OrderTypeTransfer,
			Amount:      amount,
			Status:      models.OrderStatusCompleted,
			CreatedBy:   &createdBy,
		}

		if err := tx.Create(order).Error; err != nil {
			return err
		}

		// 从资金账户扣除
		fundBalanceBefore := fundAccount.Balance
		fundCashBalanceBefore := fundAccount.CashBalance

		fundAccount.Balance -= amount
		// 优先扣除现金余额
		if fundAccount.CashBalance >= amount {
			fundAccount.CashBalance -= amount
		} else {
			fundAccount.CashBalance = 0
		}

		if err := tx.Save(&fundAccount).Error; err != nil {
			return err
		}

		// 记录资金账户交易
		fundTransaction := &models.AccountTransaction{
			TenantID:          tenantID,
			AccountID:         fundAccount.ID,
			OrderID:           &order.ID,
			TransactionType:   models.TransactionTypeTransfer,
			Amount:            -amount,
			BalanceBefore:     fundBalanceBefore,
			BalanceAfter:      fundAccount.Balance,
			CashBalanceBefore: fundCashBalanceBefore,
			CashBalanceAfter:  fundAccount.CashBalance,
			Notes:             fmt.Sprintf("转出到消耗账户: %s", orderNo),
		}
		if err := tx.Create(fundTransaction).Error; err != nil {
			return err
		}

		// 增加到消耗账户
		consBalanceBefore := consumptionAccount.Balance
		consCashBalanceBefore := consumptionAccount.CashBalance

		consumptionAccount.Balance += amount
		// 保持现金余额比例
		if fundCashBalanceBefore > 0 {
			cashRatio := fundAccount.CashBalance / fundBalanceBefore
			consumptionAccount.CashBalance += amount * cashRatio
		}

		if err := tx.Save(&consumptionAccount).Error; err != nil {
			return err
		}

		// 记录消耗账户交易
		consTransaction := &models.AccountTransaction{
			TenantID:          tenantID,
			AccountID:         consumptionAccount.ID,
			OrderID:           &order.ID,
			TransactionType:   models.TransactionTypeTransfer,
			Amount:            amount,
			BalanceBefore:     consBalanceBefore,
			BalanceAfter:      consumptionAccount.Balance,
			CashBalanceBefore: consCashBalanceBefore,
			CashBalanceAfter:  consumptionAccount.CashBalance,
			Notes:             fmt.Sprintf("从资金账户转入: %s", orderNo),
		}
		if err := tx.Create(consTransaction).Error; err != nil {
			return err
		}

		s.logStatusChange(order.ID, "", models.OrderStatusCompleted, &createdBy, "转账完成")

		return nil
	})

	if err != nil {
		return nil, err
	}

	return order, nil
}

// logStatusChange 记录订单状态变更日志
func (s *OrderService) logStatusChange(orderID int64, fromStatus, toStatus string, operatorID *int64, notes string) {
	log := &models.OrderStatusLog{
		OrderID:    orderID,
		FromStatus: fromStatus,
		ToStatus:   toStatus,
		OperatorID: operatorID,
		Notes:      notes,
	}
	database.DB.Create(log)
}

// GetOrder 获取订单详情
func (s *OrderService) GetOrder(tenantID, orderID int64) (*models.Order, error) {
	var order models.Order
	if err := database.DB.Where("id = ? AND tenant_id = ?", orderID, tenantID).
		Preload("Customer").
		Preload("StatusLogs").
		First(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

// ListOrders 获取订单列表
func (s *OrderService) ListOrders(tenantID int64, orderType, status string, customerID *int64, page, size int) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64

	query := database.DB.Model(&models.Order{}).Where("tenant_id = ?", tenantID)

	if orderType != "" {
		query = query.Where("order_type = ?", orderType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if customerID != nil {
		query = query.Where("customer_id = ?", *customerID)
	}

	query.Count(&total)

	offset := (page - 1) * size
	if err := query.Offset(offset).Limit(size).
		Preload("Customer").
		Order("created_at DESC").
		Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}
