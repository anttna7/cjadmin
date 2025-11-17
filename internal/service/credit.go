package service

import (
	"errors"
	"fmt"
	"time"

	"cjadmin/internal/models"
	"gorm.io/gorm"
)

// CreditService 授信服务
type CreditService struct {
	db *gorm.DB
}

// NewCreditService 创建授信服务
func NewCreditService(db *gorm.DB) *CreditService {
	return &CreditService{db: db}
}

// ApplyCredit 申请授信
func (s *CreditService) ApplyCredit(
	tenantID, customerID, applicantID int64,
	amount float64,
	reason string,
	assignedApproverID *int64,
) (*models.CreditRecord, error) {
	// 验证参数
	if amount <= 0 {
		return nil, errors.New("授信金额必须大于0")
	}

	// 检查客户是否存在
	var customer models.Customer
	if err := s.db.First(&customer, customerID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("客户不存在")
		}
		return nil, fmt.Errorf("查询客户失败: %w", err)
	}

	// 获取客户当前账户信息
	var account models.CustomerAccount
	if err := s.db.Where("customer_id = ? AND account_type = ?",
		customerID, models.AccountTypeFund).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("客户账户不存在")
		}
		return nil, fmt.Errorf("查询客户账户失败: %w", err)
	}

	// 创建授信记录
	record := &models.CreditRecord{
		TenantID:           tenantID,
		CustomerID:         customerID,
		RecordType:         models.CreditRecordTypeGrant,
		CreditAmount:       amount,
		BeforeCredit:       account.CreditBalance,
		BeforeLimit:        account.CreditLimit,
		AfterLimit:         account.CreditLimit + amount, // 增加授信额度
		ApplyReason:        reason,
		ApplicantID:        &applicantID,
		ApplyTime:          time.Now(),
		ApprovalStatus:     models.CreditApprovalStatusPending,
		AssignedApproverID: assignedApproverID,
		AutoAssigned:       assignedApproverID == nil,
	}

	// 如果没有指定审批人，自动分配
	if assignedApproverID == nil {
		approvers, err := s.GetApprovers(tenantID)
		if err != nil {
			return nil, fmt.Errorf("获取审批人失败: %w", err)
		}
		if len(approvers) > 0 {
			record.AssignedApproverID = &approvers[0].ID
		}
	}

	if err := s.db.Create(record).Error; err != nil {
		return nil, fmt.Errorf("创建授信记录失败: %w", err)
	}

	return record, nil
}

// ApproveCredit 审批授信
func (s *CreditService) ApproveCredit(
	tenantID, recordID, approverID int64,
	approved bool,
	remark string,
) error {
	// 开启事务
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 获取授信记录
		var record models.CreditRecord
		if err := tx.Where("id = ? AND tenant_id = ?", recordID, tenantID).
			First(&record).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("授信记录不存在")
			}
			return fmt.Errorf("查询授信记录失败: %w", err)
		}

		// 检查状态
		if record.ApprovalStatus != models.CreditApprovalStatusPending {
			return fmt.Errorf("授信记录已审批，当前状态：%s", record.ApprovalStatus)
		}

		// 更新审批信息
		now := time.Now()
		updates := map[string]interface{}{
			"approver_id":     approverID,
			"approval_time":   now,
			"approval_remark": remark,
			"updated_at":      now,
		}

		if approved {
			updates["approval_status"] = models.CreditApprovalStatusApproved

			// 如果审批通过，更新客户账户的授信额度
			var account models.CustomerAccount
			if err := tx.Where("customer_id = ? AND account_type = ?",
				record.CustomerID, models.AccountTypeFund).First(&account).Error; err != nil {
				return fmt.Errorf("查询客户账户失败: %w", err)
			}

			// 更新授信额度
			newLimit := account.CreditLimit + record.CreditAmount
			if err := tx.Model(&account).Updates(map[string]interface{}{
				"credit_limit": newLimit,
				"updated_at":   now,
			}).Error; err != nil {
				return fmt.Errorf("更新授信额度失败: %w", err)
			}

			// 更新授信记录的 after_credit 和 after_limit
			updates["after_credit"] = account.CreditBalance
			updates["after_limit"] = newLimit
		} else {
			updates["approval_status"] = models.CreditApprovalStatusRejected
		}

		// 更新授信记录
		if err := tx.Model(&record).Updates(updates).Error; err != nil {
			return fmt.Errorf("更新授信记录失败: %w", err)
		}

		return nil
	})
}

// GetApprovers 获取有授信审批权限的用户列表
func (s *CreditService) GetApprovers(tenantID int64) ([]models.User, error) {
	var users []models.User

	// 查询拥有 "credit.approve" 权限的用户
	err := s.db.Distinct("users.*").
		Table("users").
		Joins("JOIN user_roles ON users.id = user_roles.user_id").
		Joins("JOIN role_permissions ON user_roles.role_id = role_permissions.role_id").
		Joins("JOIN permissions ON role_permissions.permission_id = permissions.id").
		Where("users.tenant_id = ? OR users.tenant_id IS NULL", tenantID).
		Where("permissions.code = ?", "credit.approve").
		Where("users.status = ?", "active").
		Find(&users).Error

	return users, err
}

// CancelCredit 取消授信申请
func (s *CreditService) CancelCredit(tenantID, recordID, userID int64) error {
	var record models.CreditRecord
	if err := s.db.Where("id = ? AND tenant_id = ?", recordID, tenantID).
		First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("授信记录不存在")
		}
		return fmt.Errorf("查询授信记录失败: %w", err)
	}

	// 只有待审批状态才能取消
	if record.ApprovalStatus != models.CreditApprovalStatusPending {
		return fmt.Errorf("只有待审批状态才能取消，当前状态：%s", record.ApprovalStatus)
	}

	// 只有申请人才能取消
	if record.ApplicantID == nil || *record.ApplicantID != userID {
		return errors.New("只有申请人才能取消授信申请")
	}

	updates := map[string]interface{}{
		"approval_status": models.CreditApprovalStatusCancelled,
		"updated_at":      time.Now(),
	}

	return s.db.Model(&record).Updates(updates).Error
}

// AdjustCredit 调整授信额度
func (s *CreditService) AdjustCredit(
	tenantID, customerID, applicantID int64,
	adjustAmount float64,
	reason string,
) (*models.CreditRecord, error) {
	// 验证参数
	if adjustAmount == 0 {
		return nil, errors.New("调整金额不能为0")
	}

	return s.db.Transaction(func(tx *gorm.DB) (*models.CreditRecord, error) {
		// 获取客户账户
		var account models.CustomerAccount
		if err := tx.Where("customer_id = ? AND account_type = ?",
			customerID, models.AccountTypeFund).First(&account).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("客户账户不存在")
			}
			return nil, fmt.Errorf("查询客户账户失败: %w", err)
		}

		// 检查调整后的额度是否为负
		newLimit := account.CreditLimit + adjustAmount
		if newLimit < 0 {
			return nil, errors.New("调整后的授信额度不能为负数")
		}

		// 创建调整记录
		record := &models.CreditRecord{
			TenantID:       tenantID,
			CustomerID:     customerID,
			RecordType:     models.CreditRecordTypeAdjust,
			CreditAmount:   adjustAmount,
			BeforeCredit:   account.CreditBalance,
			AfterCredit:    account.CreditBalance,
			BeforeLimit:    account.CreditLimit,
			AfterLimit:     newLimit,
			ApplyReason:    reason,
			ApplicantID:    &applicantID,
			ApplyTime:      time.Now(),
			ApprovalStatus: models.CreditApprovalStatusApproved, // 调整直接通过
			ApproverID:     &applicantID,
			ApprovalTime:   func() *time.Time { t := time.Now(); return &t }(),
		}

		if err := tx.Create(record).Error; err != nil {
			return nil, fmt.Errorf("创建调整记录失败: %w", err)
		}

		// 更新账户授信额度
		if err := tx.Model(&account).Updates(map[string]interface{}{
			"credit_limit": newLimit,
			"updated_at":   time.Now(),
		}).Error; err != nil {
			return nil, fmt.Errorf("更新授信额度失败: %w", err)
		}

		return record, nil
	})
}

// RepayCredit 还款
func (s *CreditService) RepayCredit(
	tenantID, customerID int64,
	repayAmount float64,
	repayMethod string,
	orderID *int64,
) (*models.CreditRecord, error) {
	// 验证参数
	if repayAmount <= 0 {
		return nil, errors.New("还款金额必须大于0")
	}

	return s.db.Transaction(func(tx *gorm.DB) (*models.CreditRecord, error) {
		// 获取客户账户
		var account models.CustomerAccount
		if err := tx.Where("customer_id = ? AND account_type = ?",
			customerID, models.AccountTypeFund).First(&account).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("客户账户不存在")
			}
			return nil, fmt.Errorf("查询客户账户失败: %w", err)
		}

		// 检查还款金额是否超过授信余额
		if repayAmount > account.CreditBalance {
			return nil, fmt.Errorf("还款金额(%.2f)超过授信余额(%.2f)", repayAmount, account.CreditBalance)
		}

		// 创建还款记录
		record := &models.CreditRecord{
			TenantID:       tenantID,
			CustomerID:     customerID,
			RecordType:     models.CreditRecordTypeRepay,
			CreditAmount:   repayAmount,
			BeforeCredit:   account.CreditBalance,
			AfterCredit:    account.CreditBalance - repayAmount,
			BeforeLimit:    account.CreditLimit,
			AfterLimit:     account.CreditLimit,
			ApplyTime:      time.Now(),
			ApprovalStatus: models.CreditApprovalStatusApproved,
			RepayAmount:    repayAmount,
			RepayMethod:    repayMethod,
			RepayOrderID:   orderID,
		}

		if err := tx.Create(record).Error; err != nil {
			return nil, fmt.Errorf("创建还款记录失败: %w", err)
		}

		// 更新账户授信余额
		if err := tx.Model(&account).Updates(map[string]interface{}{
			"credit_balance": gorm.Expr("credit_balance - ?", repayAmount),
			"updated_at":     time.Now(),
		}).Error; err != nil {
			return nil, fmt.Errorf("更新授信余额失败: %w", err)
		}

		return record, nil
	})
}

// GetCreditRecords 获取授信记录列表
func (s *CreditService) GetCreditRecords(
	tenantID int64,
	customerID *int64,
	recordType string,
	status string,
	page, pageSize int,
) ([]models.CreditRecord, int64, error) {
	var records []models.CreditRecord
	var total int64

	query := s.db.Model(&models.CreditRecord{}).
		Where("tenant_id = ?", tenantID).
		Where("deleted_at IS NULL")

	if customerID != nil {
		query = query.Where("customer_id = ?", *customerID)
	}

	if recordType != "" {
		query = query.Where("record_type = ?", recordType)
	}

	if status != "" {
		query = query.Where("approval_status = ?", status)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.
		Preload("Customer").
		Preload("Applicant").
		Preload("Approver").
		Preload("AssignedApprover").
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&records).Error; err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

// GetCreditRecord 获取授信记录详情
func (s *CreditService) GetCreditRecord(tenantID, recordID int64) (*models.CreditRecord, error) {
	var record models.CreditRecord
	if err := s.db.
		Preload("Customer").
		Preload("Applicant").
		Preload("Approver").
		Preload("AssignedApprover").
		Preload("RepayOrder").
		Where("id = ? AND tenant_id = ?", recordID, tenantID).
		First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("授信记录不存在")
		}
		return nil, fmt.Errorf("查询授信记录失败: %w", err)
	}

	return &record, nil
}

// GetCustomerCreditInfo 获取客户授信信息
func (s *CreditService) GetCustomerCreditInfo(customerID int64) (map[string]interface{}, error) {
	var account models.CustomerAccount
	if err := s.db.Where("customer_id = ? AND account_type = ?",
		customerID, models.AccountTypeFund).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("客户账户不存在")
		}
		return nil, fmt.Errorf("查询客户账户失败: %w", err)
	}

	// 计算可用授信额度
	availableCredit := account.CreditLimit - account.CreditBalance

	return map[string]interface{}{
		"customer_id":      customerID,
		"credit_limit":     account.CreditLimit,
		"credit_balance":   account.CreditBalance,
		"available_credit": availableCredit,
		"usage_rate":       func() float64 {
			if account.CreditLimit > 0 {
				return account.CreditBalance / account.CreditLimit * 100
			}
			return 0
		}(),
	}, nil
}

// GetCreditStatistics 获取授信统计
func (s *CreditService) GetCreditStatistics(tenantID int64) (map[string]interface{}, error) {
	var result struct {
		TotalCustomers  int64
		TotalLimit      float64
		TotalBalance    float64
		AvailableCredit float64
		PendingCount    int64
		ApprovedCount   int64
		RejectedCount   int64
	}

	// 统计客户数和授信总额
	if err := s.db.Model(&models.CustomerAccount{}).
		Select("COUNT(DISTINCT customer_id) as total_customers, "+
			"COALESCE(SUM(credit_limit), 0) as total_limit, "+
			"COALESCE(SUM(credit_balance), 0) as total_balance").
		Where("tenant_id = ? AND account_type = ?", tenantID, models.AccountTypeFund).
		Scan(&result).Error; err != nil {
		return nil, err
	}

	result.AvailableCredit = result.TotalLimit - result.TotalBalance

	// 统计审批状态
	var statusCounts []struct {
		Status string
		Count  int64
	}
	if err := s.db.Model(&models.CreditRecord{}).
		Select("approval_status as status, COUNT(*) as count").
		Where("tenant_id = ?", tenantID).
		Where("deleted_at IS NULL").
		Group("approval_status").
		Scan(&statusCounts).Error; err != nil {
		return nil, err
	}

	for _, sc := range statusCounts {
		switch sc.Status {
		case models.CreditApprovalStatusPending:
			result.PendingCount = sc.Count
		case models.CreditApprovalStatusApproved:
			result.ApprovedCount = sc.Count
		case models.CreditApprovalStatusRejected:
			result.RejectedCount = sc.Count
		}
	}

	return map[string]interface{}{
		"total_customers":  result.TotalCustomers,
		"total_limit":      result.TotalLimit,
		"total_balance":    result.TotalBalance,
		"available_credit": result.AvailableCredit,
		"usage_rate": func() float64 {
			if result.TotalLimit > 0 {
				return result.TotalBalance / result.TotalLimit * 100
			}
			return 0
		}(),
		"pending_count":  result.PendingCount,
		"approved_count": result.ApprovedCount,
		"rejected_count": result.RejectedCount,
	}, nil
}
