package service

import (
	"errors"
	"fmt"

	"github.com/anttna7/cjadmin/internal/database"
	"github.com/anttna7/cjadmin/internal/models"
	"gorm.io/gorm"
)

type TenantService struct{}

func NewTenantService() *TenantService {
	return &TenantService{}
}

type CreateTenantRequest struct {
	Name     string                 `json:"name" binding:"required"`
	Code     string                 `json:"code" binding:"required"`
	Settings map[string]interface{} `json:"settings"`
}

type UpdateTenantRequest struct {
	Name     string                 `json:"name"`
	Settings map[string]interface{} `json:"settings"`
	Status   string                 `json:"status"`
}

type ListTenantRequest struct {
	Page   int    `form:"page" binding:"min=1"`
	Size   int    `form:"size" binding:"min=1,max=100"`
	Name   string `form:"name"`
	Status string `form:"status"`
}

// Create 创建租户
// 任务 4.1.1: 创建租户
func (s *TenantService) Create(req *CreateTenantRequest, createdBy int64) (*models.Tenant, error) {
	// 检查租户代码是否已存在
	var count int64
	database.DB.Model(&models.Tenant{}).Where("code = ?", req.Code).Count(&count)
	if count > 0 {
		return nil, errors.New("租户代码已存在")
	}

	tenant := &models.Tenant{
		Name:     req.Name,
		Code:     req.Code,
		Status:   "active",
		Settings: req.Settings,
	}

	// 使用事务创建租户及相关数据
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		// 创建租户
		if err := tx.Create(tenant).Error; err != nil {
			return err
		}

		// 为租户创建默认角色
		defaultRole := &models.Role{
			TenantID:       &tenant.ID,
			Name:           "租户管理员",
			Code:           "tenant_admin",
			Description:    "租户默认管理员角色",
			IsPlatformRole: false,
			CreatedBy:      &createdBy,
		}
		if err := tx.Create(defaultRole).Error; err != nil {
			return err
		}

		// 为默认角色分配基础权限
		var permissions []models.Permission
		tx.Where("resource IN ?", []string{"customer", "order", "finance"}).Find(&permissions)

		for _, perm := range permissions {
			rolePermission := &models.RolePermission{
				RoleID:       defaultRole.ID,
				PermissionID: perm.ID,
			}
			tx.Create(rolePermission)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return tenant, nil
}

// Update 更新租户信息
// 任务 4.1.2: 更新租户信息
func (s *TenantService) Update(tenantID int64, req *UpdateTenantRequest) error {
	var tenant models.Tenant
	if err := database.DB.First(&tenant, tenantID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("租户不存在")
		}
		return err
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Settings != nil {
		updates["settings"] = req.Settings
	}
	if req.Status != "" {
		// 验证状态值
		if req.Status != "active" && req.Status != "suspended" && req.Status != "inactive" {
			return errors.New("无效的状态值")
		}
		updates["status"] = req.Status
	}

	return database.DB.Model(&tenant).Updates(updates).Error
}

// Delete 删除租户（软删除）
// 任务 4.1.3: 删除租户（软删除）
func (s *TenantService) Delete(tenantID int64) error {
	// 检查租户是否存在
	var tenant models.Tenant
	if err := database.DB.First(&tenant, tenantID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("租户不存在")
		}
		return err
	}

	// 检查租户下是否还有用户
	var userCount int64
	database.DB.Model(&models.User{}).Where("tenant_id = ?", tenantID).Count(&userCount)
	if userCount > 0 {
		return fmt.Errorf("租户下还有 %d 个用户，无法删除", userCount)
	}

	// 检查租户下是否还有客户
	var customerCount int64
	database.DB.Model(&models.Customer{}).Where("tenant_id = ?", tenantID).Count(&customerCount)
	if customerCount > 0 {
		return fmt.Errorf("租户下还有 %d 个客户，无法删除", customerCount)
	}

	// 软删除租户
	return database.DB.Delete(&tenant).Error
}

// Get 获取租户详情
// 任务 4.1.4: 获取租户详情
func (s *TenantService) Get(tenantID int64) (*models.Tenant, error) {
	var tenant models.Tenant
	if err := database.DB.First(&tenant, tenantID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("租户不存在")
		}
		return nil, err
	}
	return &tenant, nil
}

// List 获取租户列表
// 任务 4.1.5: 获取租户列表
func (s *TenantService) List(req *ListTenantRequest) ([]models.Tenant, int64, error) {
	var tenants []models.Tenant
	var total int64

	query := database.DB.Model(&models.Tenant{})

	// 过滤条件
	if req.Name != "" {
		query = query.Where("name LIKE ?", "%"+req.Name+"%")
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	// 统计总数
	query.Count(&total)

	// 分页查询
	offset := (req.Page - 1) * req.Size
	if err := query.Offset(offset).Limit(req.Size).
		Order("created_at DESC").
		Find(&tenants).Error; err != nil {
		return nil, 0, err
	}

	return tenants, total, nil
}

// UpdateStatus 修改租户状态
// 任务 4.1.6: 修改租户状态（激活/暂停）
func (s *TenantService) UpdateStatus(tenantID int64, status string) error {
	// 验证状态值
	validStatuses := []string{"active", "suspended", "inactive"}
	isValid := false
	for _, s := range validStatuses {
		if status == s {
			isValid = true
			break
		}
	}
	if !isValid {
		return errors.New("无效的状态值，必须是 active、suspended 或 inactive")
	}

	var tenant models.Tenant
	if err := database.DB.First(&tenant, tenantID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("租户不存在")
		}
		return err
	}

	return database.DB.Model(&tenant).Update("status", status).Error
}

// GetStatistics 获取租户统计信息
func (s *TenantService) GetStatistics(tenantID int64) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 用户数量
	var userCount int64
	database.DB.Model(&models.User{}).Where("tenant_id = ?", tenantID).Count(&userCount)
	stats["user_count"] = userCount

	// 客户数量
	var customerCount int64
	database.DB.Model(&models.Customer{}).Where("tenant_id = ?", tenantID).Count(&customerCount)
	stats["customer_count"] = customerCount

	// 订单数量
	var orderCount int64
	database.DB.Model(&models.Order{}).Where("tenant_id = ?", tenantID).Count(&orderCount)
	stats["order_count"] = orderCount

	// 总充值金额
	var totalAmount float64
	database.DB.Model(&models.Order{}).
		Where("tenant_id = ? AND order_type = ? AND status = ?", tenantID, "recharge", "completed").
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalAmount)
	stats["total_recharge"] = totalAmount

	return stats, nil
}
