package service

import (
	"errors"
	"time"

	"github.com/anttna7/cjadmin/internal/database"
	"github.com/anttna7/cjadmin/internal/models"
	"gorm.io/gorm"
)

type ContractService struct{}

func NewContractService() *ContractService {
	return &ContractService{}
}

type CreateContractRequest struct {
	CustomerID     int64      `json:"customer_id" binding:"required"`
	ContractNo     string     `json:"contract_no" binding:"required"`
	ContractName   string     `json:"contract_name"`
	ContractAmount float64    `json:"contract_amount"`
	StartDate      *time.Time `json:"start_date"`
	EndDate        *time.Time `json:"end_date"`
	FileURL        string     `json:"file_url"`
	Notes          string     `json:"notes"`
}

type UpdateContractRequest struct {
	ContractName   string     `json:"contract_name"`
	ContractAmount float64    `json:"contract_amount"`
	StartDate      *time.Time `json:"start_date"`
	EndDate        *time.Time `json:"end_date"`
	FileURL        string     `json:"file_url"`
	Status         string     `json:"status"`
	Notes          string     `json:"notes"`
}

type ListContractRequest struct {
	Page       int    `form:"page" binding:"min=1"`
	Size       int    `form:"size" binding:"min=1,max=100"`
	CustomerID *int64 `form:"customer_id"`
	Status     string `form:"status"`
	Keyword    string `form:"keyword"`
}

// Create 创建合同
// 任务 6.2.1: 创建合同
func (s *ContractService) Create(tenantID int64, req *CreateContractRequest) (*models.Contract, error) {
	// 验证客户是否存在
	var customer models.Customer
	if err := database.DB.Where("id = ? AND tenant_id = ?", req.CustomerID, tenantID).
		First(&customer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("客户不存在")
		}
		return nil, err
	}

	// 检查合同编号是否重复
	var count int64
	database.DB.Model(&models.Contract{}).
		Where("tenant_id = ? AND contract_no = ?", tenantID, req.ContractNo).
		Count(&count)
	if count > 0 {
		return nil, errors.New("合同编号已存在")
	}

	contract := &models.Contract{
		TenantModel: models.TenantModel{
			TenantID: tenantID,
		},
		CustomerID:     req.CustomerID,
		ContractNo:     req.ContractNo,
		ContractName:   req.ContractName,
		ContractAmount: req.ContractAmount,
		StartDate:      req.StartDate,
		EndDate:        req.EndDate,
		FileURL:        req.FileURL,
		Status:         "active",
		Notes:          req.Notes,
	}

	if err := database.DB.Create(contract).Error; err != nil {
		return nil, err
	}

	// 加载客户信息
	database.DB.Preload("Customer").First(contract, contract.ID)

	return contract, nil
}

// Update 更新合同
// 任务 6.2.2: 更新合同
func (s *ContractService) Update(tenantID, contractID int64, req *UpdateContractRequest) error {
	var contract models.Contract
	if err := database.DB.Where("id = ? AND tenant_id = ?", contractID, tenantID).
		First(&contract).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("合同不存在")
		}
		return err
	}

	updates := map[string]interface{}{}
	if req.ContractName != "" {
		updates["contract_name"] = req.ContractName
	}
	if req.ContractAmount > 0 {
		updates["contract_amount"] = req.ContractAmount
	}
	if req.StartDate != nil {
		updates["start_date"] = req.StartDate
	}
	if req.EndDate != nil {
		updates["end_date"] = req.EndDate
	}
	if req.FileURL != "" {
		updates["file_url"] = req.FileURL
	}
	if req.Status != "" {
		// 验证状态值
		validStatuses := []string{"active", "expired", "cancelled"}
		isValid := false
		for _, s := range validStatuses {
			if req.Status == s {
				isValid = true
				break
			}
		}
		if !isValid {
			return errors.New("无效的状态值")
		}
		updates["status"] = req.Status
	}
	if req.Notes != "" {
		updates["notes"] = req.Notes
	}

	return database.DB.Model(&contract).Updates(updates).Error
}

// Delete 删除合同（软删除）
// 任务 6.2.3: 删除合同
func (s *ContractService) Delete(tenantID, contractID int64) error {
	result := database.DB.Where("id = ? AND tenant_id = ?", contractID, tenantID).
		Delete(&models.Contract{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("合同不存在")
	}

	return nil
}

// Get 获取合同详情
func (s *ContractService) Get(tenantID, contractID int64) (*models.Contract, error) {
	var contract models.Contract
	if err := database.DB.Where("id = ? AND tenant_id = ?", contractID, tenantID).
		Preload("Customer").
		First(&contract).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("合同不存在")
		}
		return nil, err
	}
	return &contract, nil
}

// List 获取合同列表
// 任务 6.2.4: 获取合同列表
func (s *ContractService) List(tenantID int64, req *ListContractRequest) ([]models.Contract, int64, error) {
	var contracts []models.Contract
	var total int64

	query := database.DB.Model(&models.Contract{}).Where("tenant_id = ?", tenantID)

	// 过滤条件
	if req.CustomerID != nil {
		query = query.Where("customer_id = ?", *req.CustomerID)
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.Keyword != "" {
		query = query.Where("contract_no LIKE ? OR contract_name LIKE ?",
			"%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}

	// 统计总数
	query.Count(&total)

	// 分页查询
	offset := (req.Page - 1) * req.Size
	if err := query.Offset(offset).Limit(req.Size).
		Preload("Customer").
		Order("created_at DESC").
		Find(&contracts).Error; err != nil {
		return nil, 0, err
	}

	return contracts, total, nil
}

// GetByCustomer 获取客户的所有合同
func (s *ContractService) GetByCustomer(tenantID, customerID int64) ([]models.Contract, error) {
	var contracts []models.Contract
	if err := database.DB.Where("tenant_id = ? AND customer_id = ?", tenantID, customerID).
		Order("created_at DESC").
		Find(&contracts).Error; err != nil {
		return nil, err
	}
	return contracts, nil
}

// UpdateStatus 更新合同状态
func (s *ContractService) UpdateStatus(tenantID, contractID int64, status string) error {
	// 验证状态值
	validStatuses := []string{"active", "expired", "cancelled"}
	isValid := false
	for _, s := range validStatuses {
		if status == s {
			isValid = true
			break
		}
	}
	if !isValid {
		return errors.New("无效的状态值：必须是 active、expired 或 cancelled")
	}

	var contract models.Contract
	if err := database.DB.Where("id = ? AND tenant_id = ?", contractID, tenantID).
		First(&contract).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("合同不存在")
		}
		return err
	}

	return database.DB.Model(&contract).Update("status", status).Error
}

// GetStatistics 获取合同统计信息
func (s *ContractService) GetStatistics(tenantID int64, customerID *int64) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	query := database.DB.Model(&models.Contract{}).Where("tenant_id = ?", tenantID)
	if customerID != nil {
		query = query.Where("customer_id = ?", *customerID)
	}

	// 合同总数
	var totalCount int64
	query.Count(&totalCount)
	stats["total_count"] = totalCount

	// 有效合同数
	var activeCount int64
	query.Where("status = ?", "active").Count(&activeCount)
	stats["active_count"] = activeCount

	// 已过期合同数
	var expiredCount int64
	query.Where("status = ?", "expired").Count(&expiredCount)
	stats["expired_count"] = expiredCount

	// 合同总金额
	var totalAmount float64
	database.DB.Model(&models.Contract{}).
		Where("tenant_id = ? AND status = ?", tenantID, "active").
		Select("COALESCE(SUM(contract_amount), 0)").
		Scan(&totalAmount)
	stats["total_amount"] = totalAmount

	return stats, nil
}
