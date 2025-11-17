package handler

import (
	"strconv"

	"github.com/anttna7/cjadmin/internal/middleware"
	"github.com/anttna7/cjadmin/internal/service"
	"github.com/anttna7/cjadmin/internal/utils"
	"github.com/gin-gonic/gin"
)

type DepartmentHandler struct {
	deptService *service.DepartmentService
}

func NewDepartmentHandler() *DepartmentHandler {
	return &DepartmentHandler{
		deptService: service.NewDepartmentService(),
	}
}

// Create 创建部门
func (h *DepartmentHandler) Create(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	var req service.CreateDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.CodeInvalidParams, "参数错误: "+err.Error())
		return
	}

	department, err := h.deptService.Create(*tenantID, &req)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "部门创建成功", department)
}

// Update 更新部门
func (h *DepartmentHandler) Update(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	departmentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, utils.CodeInvalidParams, "部门ID无效")
		return
	}

	var req service.UpdateDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.CodeInvalidParams, "参数错误: "+err.Error())
		return
	}

	if err := h.deptService.Update(*tenantID, departmentID, &req); err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "部门更新成功", nil)
}

// Delete 删除部门
func (h *DepartmentHandler) Delete(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	departmentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, utils.CodeInvalidParams, "部门ID无效")
		return
	}

	if err := h.deptService.Delete(*tenantID, departmentID); err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "部门删除成功", nil)
}

// GetTree 获取部门树形结构
func (h *DepartmentHandler) GetTree(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	tree, err := h.deptService.GetTree(*tenantID)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.Success(c, tree)
}

// GetList 获取部门列表
func (h *DepartmentHandler) GetList(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	departments, err := h.deptService.GetList(*tenantID)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.Success(c, departments)
}

// SetPermissions 设置部门权限
func (h *DepartmentHandler) SetPermissions(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	departmentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, utils.CodeInvalidParams, "部门ID无效")
		return
	}

	type PermissionsRequest struct {
		PermissionIDs []int64 `json:"permission_ids" binding:"required"`
	}

	var req PermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.CodeInvalidParams, "参数错误: "+err.Error())
		return
	}

	if err := h.deptService.SetPermissions(*tenantID, departmentID, req.PermissionIDs); err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "权限设置成功", nil)
}

// GetPermissions 获取部门权限列表
func (h *DepartmentHandler) GetPermissions(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	departmentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, utils.CodeInvalidParams, "部门ID无效")
		return
	}

	permissions, err := h.deptService.GetPermissions(*tenantID, departmentID)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.Success(c, permissions)
}

// GetInheritedPermissions 获取部门继承的所有权限
func (h *DepartmentHandler) GetInheritedPermissions(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	departmentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, utils.CodeInvalidParams, "部门ID无效")
		return
	}

	permissions, err := h.deptService.GetInheritedPermissions(*tenantID, departmentID)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.Success(c, permissions)
}
