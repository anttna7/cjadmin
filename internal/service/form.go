package service

import (
	"encoding/json"
	"errors"

	"github.com/anttna7/cjadmin/internal/database"
	"github.com/anttna7/cjadmin/internal/models"
	"gorm.io/gorm"
)

type FormService struct{}

func NewFormService() *FormService {
	return &FormService{}
}

type CreateFormRequest struct {
	FormName    string                 `json:"form_name" binding:"required"`
	FormCode    string                 `json:"form_code" binding:"required"`
	Description string                 `json:"description"`
	Schema      map[string]interface{} `json:"schema" binding:"required"`
	UISchema    map[string]interface{} `json:"ui_schema"`
	Category    string                 `json:"category"`
}

type UpdateFormRequest struct {
	FormName    string                 `json:"form_name"`
	Description string                 `json:"description"`
	Schema      map[string]interface{} `json:"schema"`
	UISchema    map[string]interface{} `json:"ui_schema"`
	Category    string                 `json:"category"`
	IsActive    *bool                  `json:"is_active"`
}

type ListFormRequest struct {
	Page     int    `form:"page" binding:"min=1"`
	Size     int    `form:"size" binding:"min=1,max=100"`
	Category string `form:"category"`
	Keyword  string `form:"keyword"`
	IsActive *bool  `form:"is_active"`
}

type SubmitFormDataRequest struct {
	FormID int64                  `json:"form_id" binding:"required"`
	Data   map[string]interface{} `json:"data" binding:"required"`
}

type QueryFormDataRequest struct {
	Page   int    `form:"page" binding:"min=1"`
	Size   int    `form:"size" binding:"min=1,max=100"`
	FormID *int64 `form:"form_id"`
}

// Create 创建自定义表单
// 任务 10.2.1: 创建表单
func (s *FormService) Create(tenantID int64, req *CreateFormRequest, createdBy int64) (*models.CustomForm, error) {
	// 检查表单代码是否已存在
	var count int64
	database.DB.Model(&models.CustomForm{}).
		Where("form_code = ? AND tenant_id = ?", req.FormCode, tenantID).
		Count(&count)
	if count > 0 {
		return nil, errors.New("表单代码已存在")
	}

	// 验证Schema格式
	schemaBytes, err := json.Marshal(req.Schema)
	if err != nil {
		return nil, errors.New("Schema格式错误")
	}

	var uiSchemaBytes []byte
	if req.UISchema != nil {
		uiSchemaBytes, err = json.Marshal(req.UISchema)
		if err != nil {
			return nil, errors.New("UISchema格式错误")
		}
	}

	form := &models.CustomForm{
		TenantModel: models.TenantModel{
			TenantID: tenantID,
		},
		FormName:    req.FormName,
		FormCode:    req.FormCode,
		Description: req.Description,
		Schema:      schemaBytes,
		UISchema:    uiSchemaBytes,
		Category:    req.Category,
		IsActive:    true,
		CreatedBy:   &createdBy,
	}

	if err := database.DB.Create(form).Error; err != nil {
		return nil, err
	}

	return form, nil
}

// Update 更新自定义表单
// 任务 10.2.2: 更新表单
func (s *FormService) Update(tenantID, formID int64, req *UpdateFormRequest) error {
	var form models.CustomForm
	if err := database.DB.Where("id = ? AND tenant_id = ?", formID, tenantID).
		First(&form).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("表单不存在")
		}
		return err
	}

	updates := map[string]interface{}{}

	if req.FormName != "" {
		updates["form_name"] = req.FormName
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Schema != nil {
		schemaBytes, err := json.Marshal(req.Schema)
		if err != nil {
			return errors.New("Schema格式错误")
		}
		updates["schema"] = schemaBytes
	}
	if req.UISchema != nil {
		uiSchemaBytes, err := json.Marshal(req.UISchema)
		if err != nil {
			return errors.New("UISchema格式错误")
		}
		updates["ui_schema"] = uiSchemaBytes
	}
	if req.Category != "" {
		updates["category"] = req.Category
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	return database.DB.Model(&form).Updates(updates).Error
}

// Delete 删除自定义表单（软删除）
// 任务 10.2.3: 删除表单
func (s *FormService) Delete(tenantID, formID int64) error {
	var form models.CustomForm
	if err := database.DB.Where("id = ? AND tenant_id = ?", formID, tenantID).
		First(&form).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("表单不存在")
		}
		return err
	}

	// 检查是否有关联的表单数据
	var dataCount int64
	database.DB.Model(&models.CustomFormData{}).
		Where("form_id = ?", formID).
		Count(&dataCount)
	if dataCount > 0 {
		return errors.New("该表单有关联数据，无法删除，请改为禁用")
	}

	return database.DB.Delete(&form).Error
}

// Get 获取表单定义
// 任务 10.2.4: 获取表单定义
func (s *FormService) Get(tenantID, formID int64) (*models.CustomForm, error) {
	var form models.CustomForm
	if err := database.DB.Where("id = ? AND tenant_id = ?", formID, tenantID).
		First(&form).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("表单不存在")
		}
		return nil, err
	}
	return &form, nil
}

// GetByCode 根据表单代码获取表单
func (s *FormService) GetByCode(tenantID int64, formCode string) (*models.CustomForm, error) {
	var form models.CustomForm
	if err := database.DB.Where("form_code = ? AND tenant_id = ? AND is_active = ?", formCode, tenantID, true).
		First(&form).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("表单不存在或已禁用")
		}
		return nil, err
	}
	return &form, nil
}

// List 获取表单列表
// 任务 10.2.5: 获取表单列表
func (s *FormService) List(tenantID int64, req *ListFormRequest) ([]models.CustomForm, int64, error) {
	var forms []models.CustomForm
	var total int64

	query := database.DB.Model(&models.CustomForm{}).Where("tenant_id = ?", tenantID)

	// 过滤条件
	if req.Category != "" {
		query = query.Where("category = ?", req.Category)
	}
	if req.IsActive != nil {
		query = query.Where("is_active = ?", *req.IsActive)
	}
	if req.Keyword != "" {
		query = query.Where("form_name LIKE ? OR form_code LIKE ? OR description LIKE ?",
			"%"+req.Keyword+"%", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}

	// 统计总数
	query.Count(&total)

	// 分页查询
	offset := (req.Page - 1) * req.Size
	if err := query.Offset(offset).Limit(req.Size).
		Order("created_at DESC").
		Find(&forms).Error; err != nil {
		return nil, 0, err
	}

	return forms, total, nil
}

// SubmitData 提交表单数据
// 任务 10.3.1: 提交表单数据
func (s *FormService) SubmitData(tenantID, customerID int64, req *SubmitFormDataRequest, submittedBy int64) (*models.CustomFormData, error) {
	// 验证表单是否存在且启用
	var form models.CustomForm
	if err := database.DB.Where("id = ? AND tenant_id = ? AND is_active = ?", req.FormID, tenantID, true).
		First(&form).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("表单不存在或已禁用")
		}
		return nil, err
	}

	// 验证客户是否存在
	var customer models.Customer
	if err := database.DB.Where("id = ? AND tenant_id = ?", customerID, tenantID).
		First(&customer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("客户不存在")
		}
		return nil, err
	}

	// 将数据转换为JSONB
	dataBytes, err := json.Marshal(req.Data)
	if err != nil {
		return nil, errors.New("数据格式错误")
	}

	formData := &models.CustomFormData{
		TenantModel: models.TenantModel{
			TenantID: tenantID,
		},
		FormID:      req.FormID,
		CustomerID:  customerID,
		Data:        dataBytes,
		SubmittedBy: &submittedBy,
	}

	if err := database.DB.Create(formData).Error; err != nil {
		return nil, err
	}

	// 加载关联数据
	database.DB.Preload("Form").Preload("Customer").First(formData, formData.ID)

	return formData, nil
}

// GetFormData 获取单条表单数据
func (s *FormService) GetFormData(tenantID, dataID int64) (*models.CustomFormData, error) {
	var formData models.CustomFormData
	if err := database.DB.Where("id = ? AND tenant_id = ?", dataID, tenantID).
		Preload("Form").
		Preload("Customer").
		First(&formData).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("表单数据不存在")
		}
		return nil, err
	}
	return &formData, nil
}

// QueryData 查询表单数据
// 任务 10.3.2: 查询表单数据
func (s *FormService) QueryData(tenantID int64, req *QueryFormDataRequest) ([]models.CustomFormData, int64, error) {
	var formDataList []models.CustomFormData
	var total int64

	query := database.DB.Model(&models.CustomFormData{}).Where("tenant_id = ?", tenantID)

	// 过滤条件
	if req.FormID != nil {
		query = query.Where("form_id = ?", *req.FormID)
	}

	// 统计总数
	query.Count(&total)

	// 分页查询
	offset := (req.Page - 1) * req.Size
	if err := query.Offset(offset).Limit(req.Size).
		Preload("Form").
		Preload("Customer").
		Order("created_at DESC").
		Find(&formDataList).Error; err != nil {
		return nil, 0, err
	}

	return formDataList, total, nil
}

// GetByCustomer 获取客户的表单数据
func (s *FormService) GetByCustomer(tenantID, customerID int64) ([]models.CustomFormData, error) {
	var formDataList []models.CustomFormData
	if err := database.DB.Where("tenant_id = ? AND customer_id = ?", tenantID, customerID).
		Preload("Form").
		Order("created_at DESC").
		Find(&formDataList).Error; err != nil {
		return nil, err
	}
	return formDataList, nil
}

// GetByForm 获取表单的所有提交数据
func (s *FormService) GetByForm(tenantID, formID int64) ([]models.CustomFormData, error) {
	var formDataList []models.CustomFormData
	if err := database.DB.Where("tenant_id = ? AND form_id = ?", tenantID, formID).
		Preload("Customer").
		Order("created_at DESC").
		Find(&formDataList).Error; err != nil {
		return nil, err
	}
	return formDataList, nil
}

// ExportData 导出表单数据
// 任务 10.3.3: 导出表单数据
func (s *FormService) ExportData(tenantID, formID int64) ([]models.CustomFormData, error) {
	// 验证表单是否存在
	var form models.CustomForm
	if err := database.DB.Where("id = ? AND tenant_id = ?", formID, tenantID).
		First(&form).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("表单不存在")
		}
		return nil, err
	}

	// 获取所有表单数据
	var formDataList []models.CustomFormData
	if err := database.DB.Where("tenant_id = ? AND form_id = ?", tenantID, formID).
		Preload("Customer").
		Order("created_at ASC").
		Find(&formDataList).Error; err != nil {
		return nil, err
	}

	return formDataList, nil
}

// GetStatistics 获取表单统计信息
func (s *FormService) GetStatistics(tenantID int64) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 表单总数
	var totalForms int64
	database.DB.Model(&models.CustomForm{}).
		Where("tenant_id = ?", tenantID).
		Count(&totalForms)
	stats["total_forms"] = totalForms

	// 启用的表单数
	var activeForms int64
	database.DB.Model(&models.CustomForm{}).
		Where("tenant_id = ? AND is_active = ?", tenantID, true).
		Count(&activeForms)
	stats["active_forms"] = activeForms

	// 表单数据总数
	var totalSubmissions int64
	database.DB.Model(&models.CustomFormData{}).
		Where("tenant_id = ?", tenantID).
		Count(&totalSubmissions)
	stats["total_submissions"] = totalSubmissions

	// 按分类统计表单数
	type CategoryStat struct {
		Category string `json:"category"`
		Count    int64  `json:"count"`
	}
	var categorystats []CategoryStat
	database.DB.Model(&models.CustomForm{}).
		Select("category, COUNT(*) as count").
		Where("tenant_id = ?", tenantID).
		Group("category").
		Scan(&categorystats)
	stats["by_category"] = categorystats

	return stats, nil
}
