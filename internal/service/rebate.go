package service

import (
	"errors"
	"fmt"

	"cjadmin/internal/models"
	"gorm.io/gorm"
)

// RebateService 返点服务
type RebateService struct {
	db *gorm.DB
}

// NewRebateService 创建返点服务
func NewRebateService(db *gorm.DB) *RebateService {
	return &RebateService{db: db}
}

// RebateResult 返点计算结果
type RebateResult struct {
	PaymentType  string  `json:"payment_type"`  // public 或 private
	Amount       float64 `json:"amount"`        // 充值金额
	RebateRate   float64 `json:"rebate_rate"`   // 返点比例(%)
	TotalRebate  float64 `json:"total_rebate"`  // 返点总金额
	CashRate     float64 `json:"cash_rate"`     // 现金比例(%)
	GiftRate     float64 `json:"gift_rate"`     // 赠款比例(%)
	CashAmount   float64 `json:"cash_amount"`   // 现金金额
	GiftAmount   float64 `json:"gift_amount"`   // 赠款金额
	FinalBalance float64 `json:"final_balance"` // 最终到账余额
}

// CalculateRebate 计算返点
func (s *RebateService) CalculateRebate(
	customerID int64,
	amount float64,
	paymentType string, // "public" 或 "private"
) (*RebateResult, error) {
	// 验证参数
	if amount <= 0 {
		return nil, errors.New("充值金额必须大于0")
	}

	if paymentType != models.PaymentAccountTypePublic && paymentType != models.PaymentAccountTypePrivate {
		return nil, fmt.Errorf("无效的支付类型: %s", paymentType)
	}

	// 获取客户信息
	var customer models.Customer
	if err := s.db.First(&customer, customerID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("客户不存在")
		}
		return nil, fmt.Errorf("查询客户失败: %w", err)
	}

	// 根据支付类型获取返点配置
	var rebateRate, cashRate, giftRate float64
	if paymentType == models.PaymentAccountTypePublic {
		rebateRate = customer.PublicRebateRate
		cashRate = customer.PublicCashRate
		giftRate = customer.PublicGiftRate
	} else {
		rebateRate = customer.PrivateRebateRate
		cashRate = customer.PrivateCashRate
		giftRate = customer.PrivateGiftRate
	}

	// 验证返点配置
	if err := s.validateRebateConfig(rebateRate, cashRate, giftRate); err != nil {
		return nil, err
	}

	// 计算返点金额
	totalRebate := amount * rebateRate / 100
	cashAmount := totalRebate * cashRate / 100
	giftAmount := totalRebate * giftRate / 100

	// 计算最终到账余额（充值金额 + 返点现金 + 返点赠款）
	finalBalance := amount + cashAmount + giftAmount

	result := &RebateResult{
		PaymentType:  paymentType,
		Amount:       amount,
		RebateRate:   rebateRate,
		TotalRebate:  totalRebate,
		CashRate:     cashRate,
		GiftRate:     giftRate,
		CashAmount:   cashAmount,
		GiftAmount:   giftAmount,
		FinalBalance: finalBalance,
	}

	return result, nil
}

// validateRebateConfig 验证返点配置
func (s *RebateService) validateRebateConfig(rebateRate, cashRate, giftRate float64) error {
	// 返点比例不能为负
	if rebateRate < 0 {
		return errors.New("返点比例不能为负数")
	}

	// 返点比例不能超过100%
	if rebateRate > 100 {
		return errors.New("返点比例不能超过100%")
	}

	// 现金比例不能为负
	if cashRate < 0 {
		return errors.New("现金比例不能为负数")
	}

	// 赠款比例不能为负
	if giftRate < 0 {
		return errors.New("赠款比例不能为负数")
	}

	// 现金比例 + 赠款比例应该等于100%
	total := cashRate + giftRate
	if total < 99.99 || total > 100.01 { // 允许浮点误差
		return fmt.Errorf("现金比例(%.2f%%) + 赠款比例(%.2f%%) 必须等于100%%", cashRate, giftRate)
	}

	return nil
}

// GetCustomerRebateConfig 获取客户的返点配置
func (s *RebateService) GetCustomerRebateConfig(customerID int64) (map[string]interface{}, error) {
	var customer models.Customer
	if err := s.db.First(&customer, customerID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("客户不存在")
		}
		return nil, fmt.Errorf("查询客户失败: %w", err)
	}

	return map[string]interface{}{
		"customer_id":   customer.ID,
		"customer_name": customer.CustomerName,
		"public_rebate": map[string]float64{
			"rebate_rate": customer.PublicRebateRate,
			"cash_rate":   customer.PublicCashRate,
			"gift_rate":   customer.PublicGiftRate,
		},
		"private_rebate": map[string]float64{
			"rebate_rate": customer.PrivateRebateRate,
			"cash_rate":   customer.PrivateCashRate,
			"gift_rate":   customer.PrivateGiftRate,
		},
	}, nil
}

// UpdateCustomerRebateConfig 更新客户的返点配置
func (s *RebateService) UpdateCustomerRebateConfig(
	customerID int64,
	paymentType string,
	rebateRate, cashRate, giftRate float64,
) error {
	// 验证参数
	if paymentType != models.PaymentAccountTypePublic && paymentType != models.PaymentAccountTypePrivate {
		return fmt.Errorf("无效的支付类型: %s", paymentType)
	}

	// 验证返点配置
	if err := s.validateRebateConfig(rebateRate, cashRate, giftRate); err != nil {
		return err
	}

	// 检查客户是否存在
	var count int64
	if err := s.db.Model(&models.Customer{}).Where("id = ?", customerID).Count(&count).Error; err != nil {
		return fmt.Errorf("查询客户失败: %w", err)
	}
	if count == 0 {
		return errors.New("客户不存在")
	}

	// 更新返点配置
	updates := make(map[string]interface{})
	if paymentType == models.PaymentAccountTypePublic {
		updates["public_rebate_rate"] = rebateRate
		updates["public_cash_rate"] = cashRate
		updates["public_gift_rate"] = giftRate
	} else {
		updates["private_rebate_rate"] = rebateRate
		updates["private_cash_rate"] = cashRate
		updates["private_gift_rate"] = giftRate
	}

	return s.db.Model(&models.Customer{}).
		Where("id = ?", customerID).
		Updates(updates).Error
}

// BatchUpdateRebateConfig 批量更新客户返点配置
func (s *RebateService) BatchUpdateRebateConfig(
	customerIDs []int64,
	paymentType string,
	rebateRate, cashRate, giftRate float64,
) error {
	// 验证参数
	if len(customerIDs) == 0 {
		return errors.New("客户列表不能为空")
	}

	if paymentType != models.PaymentAccountTypePublic && paymentType != models.PaymentAccountTypePrivate {
		return fmt.Errorf("无效的支付类型: %s", paymentType)
	}

	// 验证返点配置
	if err := s.validateRebateConfig(rebateRate, cashRate, giftRate); err != nil {
		return err
	}

	// 更新返点配置
	updates := make(map[string]interface{})
	if paymentType == models.PaymentAccountTypePublic {
		updates["public_rebate_rate"] = rebateRate
		updates["public_cash_rate"] = cashRate
		updates["public_gift_rate"] = giftRate
	} else {
		updates["private_rebate_rate"] = rebateRate
		updates["private_cash_rate"] = cashRate
		updates["private_gift_rate"] = giftRate
	}

	return s.db.Model(&models.Customer{}).
		Where("id IN ?", customerIDs).
		Updates(updates).Error
}

// PreviewRebate 预览返点（不保存到数据库）
func (s *RebateService) PreviewRebate(
	amount float64,
	rebateRate, cashRate, giftRate float64,
	paymentType string,
) (*RebateResult, error) {
	// 验证参数
	if amount <= 0 {
		return nil, errors.New("充值金额必须大于0")
	}

	if paymentType != models.PaymentAccountTypePublic && paymentType != models.PaymentAccountTypePrivate {
		return nil, fmt.Errorf("无效的支付类型: %s", paymentType)
	}

	// 验证返点配置
	if err := s.validateRebateConfig(rebateRate, cashRate, giftRate); err != nil {
		return nil, err
	}

	// 计算返点金额
	totalRebate := amount * rebateRate / 100
	cashAmount := totalRebate * cashRate / 100
	giftAmount := totalRebate * giftRate / 100
	finalBalance := amount + cashAmount + giftAmount

	result := &RebateResult{
		PaymentType:  paymentType,
		Amount:       amount,
		RebateRate:   rebateRate,
		TotalRebate:  totalRebate,
		CashRate:     cashRate,
		GiftRate:     giftRate,
		CashAmount:   cashAmount,
		GiftAmount:   giftAmount,
		FinalBalance: finalBalance,
	}

	return result, nil
}

// GetRebateStatistics 获取返点统计
func (s *RebateService) GetRebateStatistics(
	tenantID int64,
	startDate, endDate string,
) (map[string]interface{}, error) {
	var result struct {
		TotalOrders      int64
		TotalAmount      float64
		TotalRebate      float64
		TotalCashRebate  float64
		TotalGiftRebate  float64
		PublicOrders     int64
		PublicAmount     float64
		PublicRebate     float64
		PrivateOrders    int64
		PrivateAmount    float64
		PrivateRebate    float64
	}

	// 构建基础查询
	query := s.db.Model(&models.Order{}).
		Where("tenant_id = ?", tenantID).
		Where("order_type = ?", models.OrderTypeRecharge).
		Where("status = ?", models.OrderStatusCompleted)

	if startDate != "" {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("created_at <= ?", endDate)
	}

	// 总计统计
	if err := query.
		Select("COUNT(*) as total_orders, "+
			"COALESCE(SUM(amount), 0) as total_amount, "+
			"COALESCE(SUM(rebate_amount), 0) as total_rebate, "+
			"COALESCE(SUM(cash_amount), 0) as total_cash_rebate, "+
			"COALESCE(SUM(gift_amount), 0) as total_gift_rebate").
		Scan(&result).Error; err != nil {
		return nil, err
	}

	// 对公统计
	queryPublic := query.Where("payment_type = ?", models.PaymentAccountTypePublic)
	if err := queryPublic.
		Select("COUNT(*) as public_orders, "+
			"COALESCE(SUM(amount), 0) as public_amount, "+
			"COALESCE(SUM(rebate_amount), 0) as public_rebate").
		Scan(&result).Error; err != nil {
		return nil, err
	}

	// 对私统计
	queryPrivate := query.Where("payment_type = ?", models.PaymentAccountTypePrivate)
	if err := queryPrivate.
		Select("COUNT(*) as private_orders, "+
			"COALESCE(SUM(amount), 0) as private_amount, "+
			"COALESCE(SUM(rebate_amount), 0) as private_rebate").
		Scan(&result).Error; err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total": map[string]interface{}{
			"orders":       result.TotalOrders,
			"amount":       result.TotalAmount,
			"rebate":       result.TotalRebate,
			"cash_rebate":  result.TotalCashRebate,
			"gift_rebate":  result.TotalGiftRebate,
		},
		"public": map[string]interface{}{
			"orders": result.PublicOrders,
			"amount": result.PublicAmount,
			"rebate": result.PublicRebate,
		},
		"private": map[string]interface{}{
			"orders": result.PrivateOrders,
			"amount": result.PrivateAmount,
			"rebate": result.PrivateRebate,
		},
	}, nil
}
