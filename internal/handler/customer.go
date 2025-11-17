package handler

import (
	"strconv"

	"github.com/anttna7/cjadmin/internal/middleware"
	"github.com/anttna7/cjadmin/internal/service"
	"github.com/anttna7/cjadmin/internal/utils"
	"github.com/gin-gonic/gin"
)

type CustomerHandler struct {
	customerService *service.CustomerService
}

func NewCustomerHandler() *CustomerHandler {
	return &CustomerHandler{
		customerService: service.NewCustomerService(),
	}
}

// Create 创建客户
func (h *CustomerHandler) Create(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	var req service.CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.CodeInvalidParams, "参数错误: "+err.Error())
		return
	}

	customer, err := h.customerService.Create(*tenantID, &req)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "客户创建成功", customer)
}

// Update 更新客户
func (h *CustomerHandler) Update(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	customerID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, utils.CodeInvalidParams, "客户ID无效")
		return
	}

	var req service.UpdateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.CodeInvalidParams, "参数错误: "+err.Error())
		return
	}

	if err := h.customerService.Update(*tenantID, customerID, &req); err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "客户更新成功", nil)
}

// Delete 删除客户
func (h *CustomerHandler) Delete(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	customerID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, utils.CodeInvalidParams, "客户ID无效")
		return
	}

	if err := h.customerService.Delete(*tenantID, customerID); err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "客户删除成功", nil)
}

// Get 获取客户详情
func (h *CustomerHandler) Get(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	customerID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, utils.CodeInvalidParams, "客户ID无效")
		return
	}

	customer, err := h.customerService.Get(*tenantID, customerID)
	if err != nil {
		utils.Error(c, utils.CodeNotFound, err.Error())
		return
	}

	utils.Success(c, customer)
}

// List 获取客户列表
func (h *CustomerHandler) List(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	var req service.ListCustomerRequest
	req.Page = 1
	req.Size = 20

	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Error(c, utils.CodeInvalidParams, "参数错误: "+err.Error())
		return
	}

	customers, total, err := h.customerService.List(*tenantID, &req)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.PageSuccess(c, customers, total, req.Page, req.Size)
}

// GetAccounts 获取客户账户信息
func (h *CustomerHandler) GetAccounts(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	customerID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, utils.CodeInvalidParams, "客户ID无效")
		return
	}

	accounts, err := h.customerService.GetCustomerAccounts(*tenantID, customerID)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.Success(c, accounts)
}

// ExportCustomers 导出客户数据到Excel
// 任务 15.4.1: 导出客户数据API
func (h *CustomerHandler) ExportCustomers(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	// 获取所有客户（不分页）
	var req service.ListCustomerRequest
	req.Page = 1
	req.Size = 10000 // 最多导出1万条

	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Error(c, utils.CodeInvalidParams, "参数错误: "+err.Error())
		return
	}

	customers, _, err := h.customerService.List(*tenantID, &req)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	// 导出到Excel
	exporter := utils.NewExcelExporter()
	excelData, err := exporter.ExportCustomers(customers)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	// 设置响应头
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=customers.xlsx")
	c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", excelData)
}

// GetImportTemplate 获取客户导入模板
// 任务 15.4.2: 获取导入模板API
func (h *CustomerHandler) GetImportTemplate(c *gin.Context) {
	templateData, err := utils.GetExampleCustomersTemplate()
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	// 设置响应头
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=customer_template.xlsx")
	c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", templateData)
}

// ImportCustomers 从Excel导入客户数据
// 任务 15.4.3: 导入客户数据API
func (h *CustomerHandler) ImportCustomers(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		utils.Error(c, utils.CodeInvalidParams, "请上传Excel文件")
		return
	}

	// 检查文件大小（最大10MB）
	if file.Size > 10*1024*1024 {
		utils.Error(c, utils.CodeInvalidParams, "文件大小不能超过10MB")
		return
	}

	// 检查文件类型
	if file.Header.Get("Content-Type") != "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" &&
		file.Header.Get("Content-Type") != "application/vnd.ms-excel" {
		// 也检查文件扩展名
		if len(file.Filename) < 5 || (file.Filename[len(file.Filename)-5:] != ".xlsx" && file.Filename[len(file.Filename)-4:] != ".xls") {
			utils.Error(c, utils.CodeInvalidParams, "只支持Excel文件格式(.xlsx, .xls)")
			return
		}
	}

	// 读取文件内容
	fileContent, err := file.Open()
	if err != nil {
		utils.Error(c, utils.CodeInternalError, "读取文件失败: "+err.Error())
		return
	}
	defer fileContent.Close()

	fileData := make([]byte, file.Size)
	_, err = fileContent.Read(fileData)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, "读取文件失败: "+err.Error())
		return
	}

	// 解析Excel数据
	importer := utils.NewExcelImporter()
	customerData, err := importer.ImportCustomers(fileData)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, "解析Excel失败: "+err.Error())
		return
	}

	// 验证数据
	validationErrors, err := importer.ValidateCustomerImportData(customerData)
	if err != nil {
		utils.Error(c, utils.CodeInvalidParams, gin.H{
			"message": "数据验证失败",
			"errors":  validationErrors,
		})
		return
	}

	// 转换为CreateCustomerRequest
	var requests []service.CreateCustomerRequest
	for _, data := range customerData {
		req := service.CreateCustomerRequest{
			CustomerName:  data.CustomerName,
			ContactPerson: data.ContactPerson,
			ContactPhone:  data.ContactPhone,
			ContactEmail:  data.ContactEmail,
			Address:       data.Address,
			CustomerType:  data.CustomerType,
		}
		requests = append(requests, req)
	}

	// 批量导入
	userID := middleware.GetUserID(c)
	importLog, err := h.customerService.ImportCustomers(*tenantID, file.Filename, requests, userID)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "导入完成", gin.H{
		"import_log":    importLog,
		"total_count":   importLog.TotalCount,
		"success_count": importLog.SuccessCount,
		"fail_count":    importLog.FailCount,
	})
}
