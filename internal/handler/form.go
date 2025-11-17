package handler

import (
	"strconv"

	"github.com/anttna7/cjadmin/internal/middleware"
	"github.com/anttna7/cjadmin/internal/service"
	"github.com/anttna7/cjadmin/internal/utils"
	"github.com/gin-gonic/gin"
)

type FormHandler struct {
	formService *service.FormService
}

func NewFormHandler() *FormHandler {
	return &FormHandler{
		formService: service.NewFormService(),
	}
}

// Create 创建自定义表单
// 任务 10.4.1: 创建表单API
func (h *FormHandler) Create(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	var req service.CreateFormRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.CodeInvalidParams, "参数错误: "+err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	form, err := h.formService.Create(*tenantID, &req, userID)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "表单创建成功", form)
}

// Update 更新自定义表单
// 任务 10.4.2: 更新表单API
func (h *FormHandler) Update(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	formID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, utils.CodeInvalidParams, "表单ID无效")
		return
	}

	var req service.UpdateFormRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.CodeInvalidParams, "参数错误: "+err.Error())
		return
	}

	if err := h.formService.Update(*tenantID, formID, &req); err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "表单更新成功", nil)
}

// Delete 删除自定义表单
// 任务 10.4.3: 删除表单API
func (h *FormHandler) Delete(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	formID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, utils.CodeInvalidParams, "表单ID无效")
		return
	}

	if err := h.formService.Delete(*tenantID, formID); err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "表单删除成功", nil)
}

// Get 获取表单定义
// 任务 10.4.4: 获取表单定义API
func (h *FormHandler) Get(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	formID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, utils.CodeInvalidParams, "表单ID无效")
		return
	}

	form, err := h.formService.Get(*tenantID, formID)
	if err != nil {
		utils.Error(c, utils.CodeNotFound, err.Error())
		return
	}

	utils.Success(c, form)
}

// GetByCode 根据表单代码获取表单
func (h *FormHandler) GetByCode(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	formCode := c.Param("code")
	if formCode == "" {
		utils.Error(c, utils.CodeInvalidParams, "表单代码不能为空")
		return
	}

	form, err := h.formService.GetByCode(*tenantID, formCode)
	if err != nil {
		utils.Error(c, utils.CodeNotFound, err.Error())
		return
	}

	utils.Success(c, form)
}

// List 获取表单列表
// 任务 10.4.5: 获取表单列表API
func (h *FormHandler) List(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	var req service.ListFormRequest
	req.Page = 1
	req.Size = 20

	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Error(c, utils.CodeInvalidParams, "参数错误: "+err.Error())
		return
	}

	forms, total, err := h.formService.List(*tenantID, &req)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.PageSuccess(c, forms, total, req.Page, req.Size)
}

// SubmitData 提交表单数据
// 任务 10.4.6: 提交表单数据API
func (h *FormHandler) SubmitData(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	customerID, err := strconv.ParseInt(c.Param("customer_id"), 10, 64)
	if err != nil {
		utils.Error(c, utils.CodeInvalidParams, "客户ID无效")
		return
	}

	var req service.SubmitFormDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.CodeInvalidParams, "参数错误: "+err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	formData, err := h.formService.SubmitData(*tenantID, customerID, &req, userID)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "表单提交成功", formData)
}

// GetFormData 获取单条表单数据
func (h *FormHandler) GetFormData(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	dataID, err := strconv.ParseInt(c.Param("data_id"), 10, 64)
	if err != nil {
		utils.Error(c, utils.CodeInvalidParams, "数据ID无效")
		return
	}

	formData, err := h.formService.GetFormData(*tenantID, dataID)
	if err != nil {
		utils.Error(c, utils.CodeNotFound, err.Error())
		return
	}

	utils.Success(c, formData)
}

// QueryData 查询表单数据
// 任务 10.4.7: 查询表单数据API
func (h *FormHandler) QueryData(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	var req service.QueryFormDataRequest
	req.Page = 1
	req.Size = 20

	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Error(c, utils.CodeInvalidParams, "参数错误: "+err.Error())
		return
	}

	formDataList, total, err := h.formService.QueryData(*tenantID, &req)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.PageSuccess(c, formDataList, total, req.Page, req.Size)
}

// GetByCustomer 获取客户的表单数据
func (h *FormHandler) GetByCustomer(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	customerID, err := strconv.ParseInt(c.Param("customer_id"), 10, 64)
	if err != nil {
		utils.Error(c, utils.CodeInvalidParams, "客户ID无效")
		return
	}

	formDataList, err := h.formService.GetByCustomer(*tenantID, customerID)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.Success(c, formDataList)
}

// GetByForm 获取表单的所有提交数据
func (h *FormHandler) GetByForm(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	formID, err := strconv.ParseInt(c.Param("form_id"), 10, 64)
	if err != nil {
		utils.Error(c, utils.CodeInvalidParams, "表单ID无效")
		return
	}

	formDataList, err := h.formService.GetByForm(*tenantID, formID)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.Success(c, formDataList)
}

// ExportData 导出表单数据
// 任务 10.4.8: 导出表单数据API
func (h *FormHandler) ExportData(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	formID, err := strconv.ParseInt(c.Param("form_id"), 10, 64)
	if err != nil {
		utils.Error(c, utils.CodeInvalidParams, "表单ID无效")
		return
	}

	formDataList, err := h.formService.ExportData(*tenantID, formID)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	// TODO: 实现Excel导出功能（Stage 15）
	// 目前先返回JSON格式数据
	utils.Success(c, formDataList)
}

// GetStatistics 获取表单统计信息
func (h *FormHandler) GetStatistics(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	stats, err := h.formService.GetStatistics(*tenantID)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.Success(c, stats)
}
