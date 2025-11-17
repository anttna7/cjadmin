package service

import (
	"errors"
	"math/rand"
	"sort"
	"time"

	"cjadmin/internal/models"
	"gorm.io/gorm"
)

// PaymentRotationService 收款账户轮询服务
type PaymentRotationService struct {
	db *gorm.DB
}

// NewPaymentRotationService 创建收款账户轮询服务
func NewPaymentRotationService(db *gorm.DB) *PaymentRotationService {
	return &PaymentRotationService{db: db}
}

// SelectPaymentAccount 根据策略选择收款账户
func (s *PaymentRotationService) SelectPaymentAccount(
	tenantID int64,
	accountType string, // public 或 private
	amount float64,
	strategy string,
) (*models.PaymentAccount, error) {
	// 获取可用的收款账户列表
	accounts, err := s.getAvailableAccounts(tenantID, accountType, amount)
	if err != nil {
		return nil, err
	}

	if len(accounts) == 0 {
		return nil, errors.New("没有可用的收款账户")
	}

	// 如果只有一个账户，直接返回
	if len(accounts) == 1 {
		return &accounts[0], nil
	}

	// 根据策略选择账户
	var selected *models.PaymentAccount
	switch strategy {
	case models.RotateStrategyWeight:
		selected, err = s.selectByWeight(accounts)
	case models.RotateStrategyRoundRobin:
		selected, err = s.selectByRoundRobin(accounts)
	case models.RotateStrategyBalance:
		selected, err = s.selectByBalance(accounts)
	case models.RotateStrategyFrequency:
		selected, err = s.selectByFrequency(accounts)
	case models.RotateStrategyRandom:
		selected, err = s.selectByRandom(accounts)
	default:
		// 默认使用权重策略
		selected, err = s.selectByWeight(accounts)
	}

	if err != nil {
		return nil, err
	}

	return selected, nil
}

// getAvailableAccounts 获取可用的收款账户列表
func (s *PaymentRotationService) getAvailableAccounts(
	tenantID int64,
	accountType string,
	amount float64,
) ([]models.PaymentAccount, error) {
	var accounts []models.PaymentAccount

	query := s.db.Where("tenant_id = ?", tenantID).
		Where("account_type = ?", accountType).
		Where("status = ?", models.PaymentAccountStatusActive).
		Where("is_auto_rotate = ?", true).
		Where("deleted_at IS NULL")

	// 检查单笔限额
	if amount > 0 {
		query = query.Where("(single_limit IS NULL OR single_limit >= ?)", amount)
	}

	// 检查今日限额
	query = query.Where("(daily_limit IS NULL OR (daily_limit - today_amount) >= ?)", amount)

	// 检查月限额
	query = query.Where("(monthly_limit IS NULL OR (monthly_limit - this_month_amount) >= ?)", amount)

	// 按优先级排序
	query = query.Order("rotate_priority DESC")

	if err := query.Find(&accounts).Error; err != nil {
		return nil, err
	}

	return accounts, nil
}

// selectByWeight 权重策略：按照权重随机选择
func (s *PaymentRotationService) selectByWeight(accounts []models.PaymentAccount) (*models.PaymentAccount, error) {
	if len(accounts) == 0 {
		return nil, errors.New("账户列表为空")
	}

	// 计算总权重
	totalWeight := 0
	for _, account := range accounts {
		if account.RotateWeight > 0 {
			totalWeight += account.RotateWeight
		}
	}

	if totalWeight == 0 {
		// 如果所有权重都为0，使用随机策略
		return s.selectByRandom(accounts)
	}

	// 生成随机数
	rand.Seed(time.Now().UnixNano())
	randomWeight := rand.Intn(totalWeight)

	// 根据权重选择账户
	currentWeight := 0
	for i, account := range accounts {
		if account.RotateWeight > 0 {
			currentWeight += account.RotateWeight
			if randomWeight < currentWeight {
				return &accounts[i], nil
			}
		}
	}

	// 兜底返回第一个
	return &accounts[0], nil
}

// selectByRoundRobin 轮流策略：按照最后使用时间轮流
func (s *PaymentRotationService) selectByRoundRobin(accounts []models.PaymentAccount) (*models.PaymentAccount, error) {
	if len(accounts) == 0 {
		return nil, errors.New("账户列表为空")
	}

	// 按最后使用时间排序（最早使用的排在前面）
	sort.Slice(accounts, func(i, j int) bool {
		// 如果账户从未使用过，优先使用
		if accounts[i].LastUsedAt == nil {
			return true
		}
		if accounts[j].LastUsedAt == nil {
			return false
		}
		return accounts[i].LastUsedAt.Before(*accounts[j].LastUsedAt)
	})

	return &accounts[0], nil
}

// selectByBalance 余额策略：优先选择余额高的账户
func (s *PaymentRotationService) selectByBalance(accounts []models.PaymentAccount) (*models.PaymentAccount, error) {
	if len(accounts) == 0 {
		return nil, errors.New("账户列表为空")
	}

	// 按余额降序排序
	sort.Slice(accounts, func(i, j int) bool {
		return accounts[i].CurrentBalance > accounts[j].CurrentBalance
	})

	return &accounts[0], nil
}

// selectByFrequency 频率策略：优先选择使用次数少的账户
func (s *PaymentRotationService) selectByFrequency(accounts []models.PaymentAccount) (*models.PaymentAccount, error) {
	if len(accounts) == 0 {
		return nil, errors.New("账户列表为空")
	}

	// 按使用频率升序排序（使用次数少的优先）
	sort.Slice(accounts, func(i, j int) bool {
		return accounts[i].UseFrequency < accounts[j].UseFrequency
	})

	return &accounts[0], nil
}

// selectByRandom 随机策略：完全随机选择
func (s *PaymentRotationService) selectByRandom(accounts []models.PaymentAccount) (*models.PaymentAccount, error) {
	if len(accounts) == 0 {
		return nil, errors.New("账户列表为空")
	}

	rand.Seed(time.Now().UnixNano())
	index := rand.Intn(len(accounts))
	return &accounts[index], nil
}

// UpdateAccountStatistics 更新账户使用统计
func (s *PaymentRotationService) UpdateAccountStatistics(
	accountID int64,
	amount float64,
) error {
	now := time.Now()

	updates := map[string]interface{}{
		"use_frequency":     gorm.Expr("use_frequency + 1"),
		"total_amount":      gorm.Expr("total_amount + ?", amount),
		"total_count":       gorm.Expr("total_count + 1"),
		"today_amount":      gorm.Expr("today_amount + ?", amount),
		"today_count":       gorm.Expr("today_count + 1"),
		"this_month_amount": gorm.Expr("this_month_amount + ?", amount),
		"this_month_count":  gorm.Expr("this_month_count + 1"),
		"last_used_at":      now,
		"updated_at":        now,
	}

	return s.db.Model(&models.PaymentAccount{}).
		Where("id = ?", accountID).
		Updates(updates).Error
}

// ResetDailyStatistics 重置日统计（定时任务调用）
func (s *PaymentRotationService) ResetDailyStatistics() error {
	updates := map[string]interface{}{
		"today_amount": 0,
		"today_count":  0,
		"updated_at":   time.Now(),
	}

	return s.db.Model(&models.PaymentAccount{}).Updates(updates).Error
}

// ResetMonthlyStatistics 重置月统计（定时任务调用）
func (s *PaymentRotationService) ResetMonthlyStatistics() error {
	updates := map[string]interface{}{
		"this_month_amount": 0,
		"this_month_count":  0,
		"updated_at":        time.Now(),
	}

	return s.db.Model(&models.PaymentAccount{}).Updates(updates).Error
}

// GetAccountsByType 获取指定类型的所有账户
func (s *PaymentRotationService) GetAccountsByType(
	tenantID int64,
	accountType string,
) ([]models.PaymentAccount, error) {
	var accounts []models.PaymentAccount

	err := s.db.Where("tenant_id = ?", tenantID).
		Where("account_type = ?", accountType).
		Where("deleted_at IS NULL").
		Order("rotate_priority DESC, created_at DESC").
		Find(&accounts).Error

	return accounts, err
}

// CheckAccountAvailability 检查账户是否可用
func (s *PaymentRotationService) CheckAccountAvailability(
	account *models.PaymentAccount,
	amount float64,
) error {
	// 检查账户状态
	if account.Status != models.PaymentAccountStatusActive {
		return errors.New("账户未启用")
	}

	// 检查是否参与轮询
	if !account.IsAutoRotate {
		return errors.New("账户未开启自动轮询")
	}

	// 检查单笔限额
	if account.SingleLimit > 0 && amount > account.SingleLimit {
		return errors.New("超过单笔限额")
	}

	// 检查今日限额
	if account.DailyLimit > 0 {
		remaining := account.DailyLimit - account.TodayAmount
		if amount > remaining {
			return errors.New("超过今日剩余限额")
		}
	}

	// 检查月限额
	if account.MonthlyLimit > 0 {
		remaining := account.MonthlyLimit - account.ThisMonthAmount
		if amount > remaining {
			return errors.New("超过本月剩余限额")
		}
	}

	return nil
}

// GetAccountStatistics 获取账户统计信息
func (s *PaymentRotationService) GetAccountStatistics(
	tenantID int64,
	accountType string,
) (map[string]interface{}, error) {
	var result struct {
		TotalAccounts  int64
		ActiveAccounts int64
		TotalAmount    float64
		TotalCount     int64
		TodayAmount    float64
		TodayCount     int64
	}

	// 总账户数
	if err := s.db.Model(&models.PaymentAccount{}).
		Where("tenant_id = ? AND account_type = ? AND deleted_at IS NULL", tenantID, accountType).
		Count(&result.TotalAccounts).Error; err != nil {
		return nil, err
	}

	// 启用账户数
	if err := s.db.Model(&models.PaymentAccount{}).
		Where("tenant_id = ? AND account_type = ? AND status = ? AND deleted_at IS NULL",
			tenantID, accountType, models.PaymentAccountStatusActive).
		Count(&result.ActiveAccounts).Error; err != nil {
		return nil, err
	}

	// 统计金额和次数
	if err := s.db.Model(&models.PaymentAccount{}).
		Select("COALESCE(SUM(total_amount), 0) as total_amount, "+
			"COALESCE(SUM(total_count), 0) as total_count, "+
			"COALESCE(SUM(today_amount), 0) as today_amount, "+
			"COALESCE(SUM(today_count), 0) as today_count").
		Where("tenant_id = ? AND account_type = ? AND deleted_at IS NULL", tenantID, accountType).
		Scan(&result).Error; err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_accounts":  result.TotalAccounts,
		"active_accounts": result.ActiveAccounts,
		"total_amount":    result.TotalAmount,
		"total_count":     result.TotalCount,
		"today_amount":    result.TodayAmount,
		"today_count":     result.TodayCount,
	}, nil
}
