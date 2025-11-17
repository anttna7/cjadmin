package handler

import (
	"strconv"

	"github.com/anttna7/cjadmin/internal/middleware"
	"github.com/anttna7/cjadmin/internal/service"
	"github.com/anttna7/cjadmin/internal/utils"
	"github.com/gin-gonic/gin"
)

type RoleHandler struct {
	roleService *service.RoleService
	permService *service.PermissionService
}

func NewRoleHandler() *RoleHandler {
	return &RoleHandler{
		roleService: service.NewRoleService(),
		permService: service.NewPermissionService(),
	}
}

// CreatePlatformRole 创建平台角色
// 任务 5.4.1: 创建角色API
func (h *RoleHandler) CreatePlatformRole(c *gin.Context) {
	var req service.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.CodeInvalidParams, "参数错误: "+err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	role, err := h.roleService.CreatePlatformRole(&req, userID)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "平台角色创建成功", role)
}

// CreateTenantRole 创建租户角色
func (h *RoleHandler) CreateTenantRole(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	var req service.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.CodeInvalidParams, "参数错误: "+err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	role, err := h.roleService.CreateTenantRole(*tenantID, &req, userID)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "租户角色创建成功", role)
}

// Update 更新角色
// 任务 5.4.2: 更新角色API
func (h *RoleHandler) Update(c *gin.Context) {
	roleID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, utils.CodeInvalidParams, "角色ID无效")
		return
	}

	var req service.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.CodeInvalidParams, "参数错误: "+err.Error())
		return
	}

	if err := h.roleService.Update(roleID, &req); err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "角色更新成功", nil)
}

// Delete 删除角色
// 任务 5.4.3: 删除角色API
func (h *RoleHandler) Delete(c *gin.Context) {
	roleID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, utils.CodeInvalidParams, "角色ID无效")
		return
	}

	if err := h.roleService.Delete(roleID); err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "角色删除成功", nil)
}

// GetList 获取角色列表
// 任务 5.4.4: 获取角色列表API
func (h *RoleHandler) GetList(c *gin.Context) {
	var req service.ListRoleRequest
	req.Page = 1
	req.Size = 20

	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Error(c, utils.CodeInvalidParams, "参数错误: "+err.Error())
		return
	}

	isPlatformAdmin := middleware.IsPlatformAdmin(c)
	tenantID := middleware.GetTenantID(c)

	roles, total, err := h.roleService.GetList(&req, isPlatformAdmin, tenantID)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.PageSuccess(c, roles, total, req.Page, req.Size)
}

// Get 获取角色详情
func (h *RoleHandler) Get(c *gin.Context) {
	roleID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, utils.CodeInvalidParams, "角色ID无效")
		return
	}

	role, err := h.roleService.Get(roleID)
	if err != nil {
		utils.Error(c, utils.CodeNotFound, err.Error())
		return
	}

	utils.Success(c, role)
}

// SetPermissions 为角色设置权限
// 任务 5.4.5: 分配角色权限API
func (h *RoleHandler) SetPermissions(c *gin.Context) {
	roleID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, utils.CodeInvalidParams, "角色ID无效")
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

	userID := middleware.GetUserID(c)
	if err := h.roleService.SetPermissions(roleID, req.PermissionIDs, userID); err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "权限设置成功", nil)
}

// GetPermissions 获取角色的权限列表
func (h *RoleHandler) GetPermissions(c *gin.Context) {
	roleID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, utils.CodeInvalidParams, "角色ID无效")
		return
	}

	permissions, err := h.roleService.GetPermissions(roleID)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.Success(c, permissions)
}

// AssignRoleToUser 为用户分配角色
// 任务 5.4.7: 分配用户角色API
func (h *RoleHandler) AssignRoleToUser(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
	if err != nil {
		utils.Error(c, utils.CodeInvalidParams, "用户ID无效")
		return
	}

	type AssignRoleRequest struct {
		RoleID int64 `json:"role_id" binding:"required"`
	}

	var req AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.CodeInvalidParams, "参数错误: "+err.Error())
		return
	}

	if err := h.roleService.AssignRoleToUser(userID, req.RoleID); err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "角色分配成功", nil)
}

// RemoveRoleFromUser 移除用户角色
func (h *RoleHandler) RemoveRoleFromUser(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
	if err != nil {
		utils.Error(c, utils.CodeInvalidParams, "用户ID无效")
		return
	}

	roleID, err := strconv.ParseInt(c.Param("role_id"), 10, 64)
	if err != nil {
		utils.Error(c, utils.CodeInvalidParams, "角色ID无效")
		return
	}

	if err := h.roleService.RemoveRoleFromUser(userID, roleID); err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "角色移除成功", nil)
}

// GetUserRoles 获取用户的角色列表
func (h *RoleHandler) GetUserRoles(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
	if err != nil {
		utils.Error(c, utils.CodeInvalidParams, "用户ID无效")
		return
	}

	roles, err := h.roleService.GetUserRoles(userID)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.Success(c, roles)
}

// BatchAssignRoles 批量为用户分配角色
func (h *RoleHandler) BatchAssignRoles(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
	if err != nil {
		utils.Error(c, utils.CodeInvalidParams, "用户ID无效")
		return
	}

	type BatchAssignRequest struct {
		RoleIDs []int64 `json:"role_ids" binding:"required"`
	}

	var req BatchAssignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.CodeInvalidParams, "参数错误: "+err.Error())
		return
	}

	if err := h.roleService.BatchAssignRoles(userID, req.RoleIDs); err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "角色批量分配成功", nil)
}

// GetAllPermissions 获取所有权限列表
// 任务 5.4.6: 获取权限列表API
func (h *RoleHandler) GetAllPermissions(c *gin.Context) {
	permissions, err := h.permService.GetAll()
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.Success(c, permissions)
}

// GetPermissionsByResource 按资源获取权限
func (h *RoleHandler) GetPermissionsByResource(c *gin.Context) {
	resource := c.Query("resource")
	if resource == "" {
		utils.Error(c, utils.CodeInvalidParams, "资源类型不能为空")
		return
	}

	permissions, err := h.permService.GetByResource(resource)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.Success(c, permissions)
}

// GetGroupedPermissions 获取分组的权限列表
func (h *RoleHandler) GetGroupedPermissions(c *gin.Context) {
	permissions, err := h.permService.GetGrouped()
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.Success(c, permissions)
}
