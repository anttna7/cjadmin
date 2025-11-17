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

type CustomerService struct{}

func NewCustomerService() *CustomerService {
	return &CustomerService{}
}

type CreateCustomerRequest struct {
	CustomerCode  string                 `json:"customer_code"`
	CustomerName  string                 `json:"customer_name" binding:"required"`
	CustomerType  string                 `json:"customer_type"`
	ContactPerson string                 `json:"contact_person"`
	Phone         string                 `json:"phone"`
	Email         string                 `json:"email"`
	Address       string                 `json:"address"`
	Industry      string                 `json:"industry"`
	Source        string                 `json:"source"`
	AssignedTo    *int64                 `json:"assigned_to"`
	CustomFields  map[string]interface{} `json:"custom_fields"`
}

type UpdateCustomerRequest struct {
	CustomerName  string                 `json:"customer_name"`
	CustomerType  string                 `json:"customer_type"`
	ContactPerson string                 `json:"contact_person"`
	Phone         string                 `json:"phone"`
	Email         string                 `json:"email"`
	Address       string                 `json:"address"`
	Industry      string                 `json:"industry"`
	Source        string                 `json:"source"`
	AssignedTo    *int64                 `json:"assigned_to"`
	Status        string                 `json:"status"`
	CustomFields  map[string]interface{} `json:"custom_fields"`
}

type ListCustomerRequest struct {
	Page         int    `form:"page" binding:"min=1"`
	Size         int    `form:"size" binding:"min=1,max=100"`
	CustomerName string `form:"customer_name"`
	Status       string `form:"status"`
	AssignedTo   *int64 `form:"assigned_to"`
}

// Create 创建客户
func (s *CustomerService) Create(tenantID int64, req *CreateCustomerRequest) (*models.Customer, error) {
	// 生成客户编号
	if req.CustomerCode == "" {
		req.CustomerCode = fmt.Sprintf("CUS%s", uuid.New().String()[:8])
	}

	// 检查客户编号是否重复
	var count int64
	database.DB.Model(&models.Customer{}).
		Where("tenant_id = ? AND customer_code = ?", tenantID, req.CustomerCode).
		Count(&count)
	if count > 0 {
		return nil, errors.New("客户编号已存在")
	}

	customer := &models.Customer{
		TenantModel: models.TenantModel{
			TenantID: tenantID,
		},
		CustomerCode:  req.CustomerCode,
		CustomerName:  req.CustomerName,
		CustomerType:  req.CustomerType,
		ContactPerson: req.ContactPerson,
		Phone:         req.Phone,
		Email:         req.Email,
		Address:       req.Address,
		Industry:      req.Industry,
		Source:        req.Source,
		AssignedTo:    req.AssignedTo,
		Status:        "active",
		CustomFields:  req.CustomFields,
	}

	if err := database.DB.Create(customer).Error; err != nil {
		return nil, err
	}

	// 创建客户账户（资金账户和消耗账户）
	accounts := []models.CustomerAccount{
		{
			TenantModel: models.TenantModel{TenantID: tenantID},
			CustomerID:  customer.ID,
			AccountType: models.AccountTypeFund,
			Balance:     0,
			CashBalance: 0,
		},
		{
			TenantModel: models.TenantModel{TenantID: tenantID},
			CustomerID:  customer.ID,
			AccountType: models.AccountTypeConsumption,
			Balance:     0,
			CashBalance: 0,
		},
	}

	for _, account := range accounts {
		database.DB.Create(&account)
	}

	return customer, nil
}

// Update 更新客户
func (s *CustomerService) Update(tenantID, customerID int64, req *UpdateCustomerRequest) error {
	var customer models.Customer
	if err := database.DB.Where("id = ? AND tenant_id = ?", customerID, tenantID).
		First(&customer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("客户不存在")
		}
		return err
	}

	updates := map[string]interface{}{}
	if req.CustomerName != "" {
		updates["customer_name"] = req.CustomerName
	}
	if req.CustomerType != "" {
		updates["customer_type"] = req.CustomerType
	}
	if req.ContactPerson != "" {
		updates["contact_person"] = req.ContactPerson
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Address != "" {
		updates["address"] = req.Address
	}
	if req.Industry != "" {
		updates["industry"] = req.Industry
	}
	if req.Source != "" {
		updates["source"] = req.Source
	}
	if req.AssignedTo != nil {
		updates["assigned_to"] = req.AssignedTo
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.CustomFields != nil {
		updates["custom_fields"] = req.CustomFields
	}

	return database.DB.Model(&customer).Updates(updates).Error
}

// Delete 删除客户（软删除）
func (s *CustomerService) Delete(tenantID, customerID int64) error {
	result := database.DB.Where("id = ? AND tenant_id = ?", customerID, tenantID).
		Delete(&models.Customer{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("客户不存在")
	}
	return nil
}

// Get 获取客户详情
func (s *CustomerService) Get(tenantID, customerID int64) (*models.Customer, error) {
	var customer models.Customer
	if err := database.DB.Where("id = ? AND tenant_id = ?", customerID, tenantID).
		Preload("AssignedUser").
		Preload("Accounts").
		First(&customer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("客户不存在")
		}
		return nil, err
	}
	return &customer, nil
}

// List 获取客户列表
func (s *CustomerService) List(tenantID int64, req *ListCustomerRequest) ([]models.Customer, int64, error) {
	var customers []models.Customer
	var total int64

	query := database.DB.Model(&models.Customer{}).Where("tenant_id = ?", tenantID)

	// 过滤条件
	if req.CustomerName != "" {
		query = query.Where("customer_name LIKE ?", "%"+req.CustomerName+"%")
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.AssignedTo != nil {
		query = query.Where("assigned_to = ?", req.AssignedTo)
	}

	// 统计总数
	query.Count(&total)

	// 分页查询
	offset := (req.Page - 1) * req.Size
	if err := query.Offset(offset).Limit(req.Size).
		Preload("AssignedUser").
		Order("created_at DESC").
		Find(&customers).Error; err != nil {
		return nil, 0, err
	}

	return customers, total, nil
}

// ImportCustomers 批量导入客户
func (s *CustomerService) ImportCustomers(tenantID int64, data []CreateCustomerRequest, importedBy int64) (*models.ImportLog, error) {
	importLog := &models.ImportLog{
		TenantModel: models.TenantModel{TenantID: tenantID},
		ImportType:  "customer",
		TotalRows:   len(data),
		SuccessRows: 0,
		FailedRows:  0,
		Status:      "processing",
		ImportedBy:  &importedBy,
	}

	database.DB.Create(importLog)

	errorDetails := make(map[string]interface{})

	for i, req := range data {
		_, err := s.Create(tenantID, &req)
		if err != nil {
			importLog.FailedRows++
			errorDetails[fmt.Sprintf("row_%d", i+1)] = err.Error()
		} else {
			importLog.SuccessRows++
		}
	}

	if importLog.FailedRows > 0 {
		importLog.ErrorDetails = errorDetails
	}

	importLog.Status = "completed"
	database.DB.Save(importLog)

	return importLog, nil
}

// ExportCustomers 导出客户数据
func (s *CustomerService) ExportCustomers(tenantID int64, filters map[string]interface{}) ([]models.Customer, error) {
	var customers []models.Customer

	query := database.DB.Where("tenant_id = ?", tenantID)

	// 应用过滤条件
	for key, value := range filters {
		query = query.Where(key+" = ?", value)
	}

	if err := query.Preload("AssignedUser").Find(&customers).Error; err != nil {
		return nil, err
	}

	return customers, nil
}

// GetCustomerAccounts 获取客户账户信息
func (s *CustomerService) GetCustomerAccounts(tenantID, customerID int64) ([]models.CustomerAccount, error) {
	var accounts []models.CustomerAccount

	if err := database.DB.Where("tenant_id = ? AND customer_id = ?", tenantID, customerID).
		Find(&accounts).Error; err != nil {
		return nil, err
	}

	return accounts, nil
}
