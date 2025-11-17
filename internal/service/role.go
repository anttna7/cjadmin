package service

import (
	"errors"
	"fmt"

	"github.com/anttna7/cjadmin/internal/database"
	"github.com/anttna7/cjadmin/internal/models"
	"gorm.io/gorm"
)

type RoleService struct{}

func NewRoleService() *RoleService {
	return &RoleService{}
}

type CreateRoleRequest struct {
	Name        string  `json:"name" binding:"required"`
	Code        string  `json:"code" binding:"required"`
	Description string  `json:"description"`
	TenantID    *int64  `json:"tenant_id"`
	PermissionIDs []int64 `json:"permission_ids"`
}

type UpdateRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ListRoleRequest struct {
	Page     int    `form:"page" binding:"min=1"`
	Size     int    `form:"size" binding:"min=1,max=100"`
	Name     string `form:"name"`
	TenantID *int64 `form:"tenant_id"`
}

// CreatePlatformRole 创建平台角色（仅平台管理员可调用）
// 任务 5.1.1: 创建平台角色
func (s *RoleService) CreatePlatformRole(req *CreateRoleRequest, createdBy int64) (*models.Role, error) {
	// 检查角色代码是否已存在
	var count int64
	database.DB.Model(&models.Role{}).
		Where("code = ? AND tenant_id IS NULL", req.Code).
		Count(&count)
	if count > 0 {
		return nil, errors.New("平台角色代码已存在")
	}

	role := &models.Role{
		TenantID:       nil,
		Name:           req.Name,
		Code:           req.Code,
		Description:    req.Description,
		IsPlatformRole: true,
		CreatedBy:      &createdBy,
	}

	return s.createRoleWithPermissions(role, req.PermissionIDs)
}

// CreateTenantRole 创建租户角色
// 任务 5.1.2: 创建租户角色
func (s *RoleService) CreateTenantRole(tenantID int64, req *CreateRoleRequest, createdBy int64) (*models.Role, error) {
	// 检查租户角色代码是否已存在
	var count int64
	database.DB.Model(&models.Role{}).
		Where("code = ? AND tenant_id = ?", req.Code, tenantID).
		Count(&count)
	if count > 0 {
		return nil, errors.New("租户角色代码已存在")
	}

	// 验证权限不超过创建者权限
	creatorPermissions, err := s.getUserPermissions(createdBy)
	if err != nil {
		return nil, err
	}

	if !s.validatePermissions(req.PermissionIDs, creatorPermissions) {
		return nil, errors.New("不能分配超出自己权限范围的权限")
	}

	role := &models.Role{
		TenantID:       &tenantID,
		Name:           req.Name,
		Code:           req.Code,
		Description:    req.Description,
		IsPlatformRole: false,
		CreatedBy:      &createdBy,
	}

	return s.createRoleWithPermissions(role, req.PermissionIDs)
}

// createRoleWithPermissions 创建角色并分配权限（内部方法）
func (s *RoleService) createRoleWithPermissions(role *models.Role, permissionIDs []int64) (*models.Role, error) {
	return role, database.DB.Transaction(func(tx *gorm.DB) error {
		// 创建角色
		if err := tx.Create(role).Error; err != nil {
			return err
		}

		// 分配权限
		if len(permissionIDs) > 0 {
			for _, permID := range permissionIDs {
				rolePermission := &models.RolePermission{
					RoleID:       role.ID,
					PermissionID: permID,
				}
				if err := tx.Create(rolePermission).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}

// Update 更新角色
// 任务 5.1.3: 更新角色
func (s *RoleService) Update(roleID int64, req *UpdateRoleRequest) error {
	var role models.Role
	if err := database.DB.First(&role, roleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("角色不存在")
		}
		return err
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}

	return database.DB.Model(&role).Updates(updates).Error
}

// Delete 删除角色
// 任务 5.1.4: 删除角色
func (s *RoleService) Delete(roleID int64) error {
	var role models.Role
	if err := database.DB.First(&role, roleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("角色不存在")
		}
		return err
	}

	// 检查是否有用户在使用此角色
	var userCount int64
	database.DB.Model(&models.UserRole{}).Where("role_id = ?", roleID).Count(&userCount)
	if userCount > 0 {
		return fmt.Errorf("角色下还有 %d 个用户，无法删除", userCount)
	}

	// 使用事务删除角色及其权限关联
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// 删除角色权限关联
		if err := tx.Where("role_id = ?", roleID).Delete(&models.RolePermission{}).Error; err != nil {
			return err
		}

		// 删除角色
		return tx.Delete(&role).Error
	})
}

// GetList 获取角色列表
// 任务 5.1.5: 获取角色列表
func (s *RoleService) GetList(req *ListRoleRequest, isPlatformAdmin bool, userTenantID *int64) ([]models.Role, int64, error) {
	var roles []models.Role
	var total int64

	query := database.DB.Model(&models.Role{})

	// 权限过滤
	if isPlatformAdmin {
		// 平台管理员可以查看所有角色
		if req.TenantID != nil {
			query = query.Where("tenant_id = ?", *req.TenantID)
		}
	} else {
		// 租户用户只能查看自己租户的角色
		if userTenantID != nil {
			query = query.Where("tenant_id = ?", *userTenantID)
		} else {
			return nil, 0, errors.New("用户未关联到任何租户")
		}
	}

	// 名称过滤
	if req.Name != "" {
		query = query.Where("name LIKE ?", "%"+req.Name+"%")
	}

	// 统计总数
	query.Count(&total)

	// 分页查询
	offset := (req.Page - 1) * req.Size
	if err := query.Offset(offset).Limit(req.Size).
		Preload("Permissions").
		Order("created_at DESC").
		Find(&roles).Error; err != nil {
		return nil, 0, err
	}

	return roles, total, nil
}

// Get 获取角色详情
func (s *RoleService) Get(roleID int64) (*models.Role, error) {
	var role models.Role
	if err := database.DB.Preload("Permissions").First(&role, roleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("角色不存在")
		}
		return nil, err
	}
	return &role, nil
}

// SetPermissions 为角色设置权限
// 任务 5.2.2: 为角色分配权限
func (s *RoleService) SetPermissions(roleID int64, permissionIDs []int64, operatorID int64) error {
	// 验证角色是否存在
	var role models.Role
	if err := database.DB.First(&role, roleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("角色不存在")
		}
		return err
	}

	// 如果是租户角色，验证操作者的权限
	if role.TenantID != nil {
		operatorPermissions, err := s.getUserPermissions(operatorID)
		if err != nil {
			return err
		}

		if !s.validatePermissions(permissionIDs, operatorPermissions) {
			return errors.New("不能分配超出自己权限范围的权限")
		}
	}

	// 验证所有权限是否存在
	var existingPerms []models.Permission
	database.DB.Where("id IN ?", permissionIDs).Find(&existingPerms)
	if len(existingPerms) != len(permissionIDs) {
		return errors.New("部分权限不存在")
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		// 删除现有的角色权限
		if err := tx.Where("role_id = ?", roleID).
			Delete(&models.RolePermission{}).Error; err != nil {
			return err
		}

		// 添加新的权限
		for _, permID := range permissionIDs {
			rolePerm := &models.RolePermission{
				RoleID:       roleID,
				PermissionID: permID,
			}
			if err := tx.Create(rolePerm).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// RemovePermission 移除角色权限
// 任务 5.2.3: 移除角色权限
func (s *RoleService) RemovePermission(roleID, permissionID int64) error {
	result := database.DB.Where("role_id = ? AND permission_id = ?", roleID, permissionID).
		Delete(&models.RolePermission{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("权限关联不存在")
	}

	return nil
}

// GetPermissions 获取角色的权限列表
// 任务 5.2.4: 获取角色的权限列表
func (s *RoleService) GetPermissions(roleID int64) ([]models.Permission, error) {
	var role models.Role
	if err := database.DB.First(&role, roleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("角色不存在")
		}
		return nil, err
	}

	var permissions []models.Permission
	query := `
		SELECT p.* FROM permissions p
		INNER JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = ?
	`
	if err := database.DB.Raw(query, roleID).Scan(&permissions).Error; err != nil {
		return nil, err
	}

	return permissions, nil
}

// AssignRoleToUser 为用户分配角色
// 任务 5.3.1: 为用户分配角色
func (s *RoleService) AssignRoleToUser(userID, roleID int64) error {
	// 验证用户是否存在
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户不存在")
		}
		return err
	}

	// 验证角色是否存在
	var role models.Role
	if err := database.DB.First(&role, roleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("角色不存在")
		}
		return err
	}

	// 验证租户匹配
	if role.TenantID != nil && user.TenantID != nil {
		if *role.TenantID != *user.TenantID {
			return errors.New("用户和角色不属于同一租户")
		}
	}

	// 检查是否已经分配
	var count int64
	database.DB.Model(&models.UserRole{}).
		Where("user_id = ? AND role_id = ?", userID, roleID).
		Count(&count)
	if count > 0 {
		return errors.New("用户已拥有该角色")
	}

	userRole := &models.UserRole{
		UserID: userID,
		RoleID: roleID,
	}

	return database.DB.Create(userRole).Error
}

// RemoveRoleFromUser 移除用户角色
// 任务 5.3.2: 移除用户角色
func (s *RoleService) RemoveRoleFromUser(userID, roleID int64) error {
	result := database.DB.Where("user_id = ? AND role_id = ?", userID, roleID).
		Delete(&models.UserRole{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("用户角色关联不存在")
	}

	return nil
}

// GetUserRoles 获取用户的角色列表
// 任务 5.3.3: 获取用户的角色列表
func (s *RoleService) GetUserRoles(userID int64) ([]models.Role, error) {
	var roles []models.Role
	query := `
		SELECT r.* FROM roles r
		INNER JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = ?
	`
	if err := database.DB.Raw(query, userID).Scan(&roles).Error; err != nil {
		return nil, err
	}

	return roles, nil
}

// BatchAssignRoles 批量为用户分配角色
// 任务 5.3.4: 批量分配角色
func (s *RoleService) BatchAssignRoles(userID int64, roleIDs []int64) error {
	// 验证用户是否存在
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户不存在")
		}
		return err
	}

	// 验证所有角色是否存在
	var roles []models.Role
	database.DB.Where("id IN ?", roleIDs).Find(&roles)
	if len(roles) != len(roleIDs) {
		return errors.New("部分角色不存在")
	}

	// 验证租户匹配
	for _, role := range roles {
		if role.TenantID != nil && user.TenantID != nil {
			if *role.TenantID != *user.TenantID {
				return fmt.Errorf("角色 %s 与用户不属于同一租户", role.Name)
			}
		}
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		// 删除现有的用户角色
		if err := tx.Where("user_id = ?", userID).
			Delete(&models.UserRole{}).Error; err != nil {
			return err
		}

		// 添加新的角色
		for _, roleID := range roleIDs {
			userRole := &models.UserRole{
				UserID: userID,
				RoleID: roleID,
			}
			if err := tx.Create(userRole).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// getUserPermissions 获取用户的权限ID列表（内部辅助方法）
func (s *RoleService) getUserPermissions(userID int64) ([]int64, error) {
	var permissionIDs []int64
	query := `
		SELECT DISTINCT p.id FROM permissions p
		INNER JOIN role_permissions rp ON p.id = rp.permission_id
		INNER JOIN user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = ?
	`
	if err := database.DB.Raw(query, userID).Scan(&permissionIDs).Error; err != nil {
		return nil, err
	}
	return permissionIDs, nil
}

// validatePermissions 验证权限是否在允许范围内（内部辅助方法）
// 任务 5.1.6: 角色权限验证（不超过创建者权限）
func (s *RoleService) validatePermissions(requestedPermissions, allowedPermissions []int64) bool {
	allowedMap := make(map[int64]bool)
	for _, id := range allowedPermissions {
		allowedMap[id] = true
	}

	for _, id := range requestedPermissions {
		if !allowedMap[id] {
			return false
		}
	}

	return true
}
