package service

import (
	"errors"
	"fmt"
	"strings"

	"github.com/anttna7/cjadmin/internal/database"
	"github.com/anttna7/cjadmin/internal/models"
	"gorm.io/gorm"
)

type DepartmentService struct{}

func NewDepartmentService() *DepartmentService {
	return &DepartmentService{}
}

type CreateDepartmentRequest struct {
	ParentID *int64 `json:"parent_id"`
	Name     string `json:"name" binding:"required"`
	Code     string `json:"code"`
}

type UpdateDepartmentRequest struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

type DepartmentTreeNode struct {
	ID       int64                 `json:"id"`
	Name     string                `json:"name"`
	Code     string                `json:"code"`
	Level    int                   `json:"level"`
	ParentID *int64                `json:"parent_id"`
	Children []*DepartmentTreeNode `json:"children,omitempty"`
}

// Create 创建部门
// 任务 4.3.1: 创建部门
func (s *DepartmentService) Create(tenantID int64, req *CreateDepartmentRequest) (*models.Department, error) {
	// 检查部门代码是否已存在
	if req.Code != "" {
		var count int64
		database.DB.Model(&models.Department{}).
			Where("tenant_id = ? AND code = ?", tenantID, req.Code).
			Count(&count)
		if count > 0 {
			return nil, errors.New("部门代码已存在")
		}
	}

	var level int = 1
	var path string = "/"

	// 如果有父部门，获取父部门信息
	if req.ParentID != nil {
		var parentDept models.Department
		if err := database.DB.Where("id = ? AND tenant_id = ?", *req.ParentID, tenantID).
			First(&parentDept).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("父部门不存在")
			}
			return nil, err
		}
		level = parentDept.Level + 1
		path = fmt.Sprintf("%s%d/", parentDept.Path, parentDept.ID)
	}

	department := &models.Department{
		TenantModel: models.TenantModel{TenantID: tenantID},
		ParentID:    req.ParentID,
		Name:        req.Name,
		Code:        req.Code,
		Level:       level,
		Path:        path,
	}

	if err := database.DB.Create(department).Error; err != nil {
		return nil, err
	}

	// 更新path包含自己的ID
	department.Path = fmt.Sprintf("%s%d/", path, department.ID)
	database.DB.Save(department)

	return department, nil
}

// Update 更新部门
// 任务 4.3.2: 更新部门
func (s *DepartmentService) Update(tenantID, departmentID int64, req *UpdateDepartmentRequest) error {
	var department models.Department
	if err := database.DB.Where("id = ? AND tenant_id = ?", departmentID, tenantID).
		First(&department).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("部门不存在")
		}
		return err
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Code != "" {
		// 检查代码是否重复
		var count int64
		database.DB.Model(&models.Department{}).
			Where("tenant_id = ? AND code = ? AND id != ?", tenantID, req.Code, departmentID).
			Count(&count)
		if count > 0 {
			return errors.New("部门代码已存在")
		}
		updates["code"] = req.Code
	}

	return database.DB.Model(&department).Updates(updates).Error
}

// Delete 删除部门
// 任务 4.3.3: 删除部门
func (s *DepartmentService) Delete(tenantID, departmentID int64) error {
	var department models.Department
	if err := database.DB.Where("id = ? AND tenant_id = ?", departmentID, tenantID).
		First(&department).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("部门不存在")
		}
		return err
	}

	// 检查是否有子部门
	var childCount int64
	database.DB.Model(&models.Department{}).
		Where("tenant_id = ? AND parent_id = ?", tenantID, departmentID).
		Count(&childCount)
	if childCount > 0 {
		return fmt.Errorf("部门下还有 %d 个子部门，无法删除", childCount)
	}

	// 检查是否有用户
	var userCount int64
	database.DB.Model(&models.User{}).
		Where("tenant_id = ? AND department_id = ?", tenantID, departmentID).
		Count(&userCount)
	if userCount > 0 {
		return fmt.Errorf("部门下还有 %d 个用户，无法删除", userCount)
	}

	return database.DB.Delete(&department).Error
}

// GetTree 获取部门树形结构
// 任务 4.3.4: 获取部门树形结构
func (s *DepartmentService) GetTree(tenantID int64) ([]*DepartmentTreeNode, error) {
	var departments []models.Department
	if err := database.DB.Where("tenant_id = ?", tenantID).
		Order("level, id").
		Find(&departments).Error; err != nil {
		return nil, err
	}

	// 构建树形结构
	nodeMap := make(map[int64]*DepartmentTreeNode)
	var roots []*DepartmentTreeNode

	// 第一遍：创建所有节点
	for _, dept := range departments {
		node := &DepartmentTreeNode{
			ID:       dept.ID,
			Name:     dept.Name,
			Code:     dept.Code,
			Level:    dept.Level,
			ParentID: dept.ParentID,
			Children: []*DepartmentTreeNode{},
		}
		nodeMap[dept.ID] = node
	}

	// 第二遍：建立父子关系
	for _, dept := range departments {
		node := nodeMap[dept.ID]
		if dept.ParentID == nil {
			roots = append(roots, node)
		} else {
			if parent, ok := nodeMap[*dept.ParentID]; ok {
				parent.Children = append(parent.Children, node)
			}
		}
	}

	return roots, nil
}

// GetList 获取部门列表（扁平结构）
func (s *DepartmentService) GetList(tenantID int64) ([]models.Department, error) {
	var departments []models.Department
	if err := database.DB.Where("tenant_id = ?", tenantID).
		Order("level, id").
		Find(&departments).Error; err != nil {
		return nil, err
	}
	return departments, nil
}

// SetPermissions 设置部门权限
// 任务 4.3.5: 设置部门权限
func (s *DepartmentService) SetPermissions(tenantID, departmentID int64, permissionIDs []int64) error {
	// 验证部门是否存在
	var department models.Department
	if err := database.DB.Where("id = ? AND tenant_id = ?", departmentID, tenantID).
		First(&department).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("部门不存在")
		}
		return err
	}

	// 验证所有权限是否存在
	var existingPerms []models.Permission
	database.DB.Where("id IN ?", permissionIDs).Find(&existingPerms)
	if len(existingPerms) != len(permissionIDs) {
		return errors.New("部分权限不存在")
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		// 删除现有的部门权限
		if err := tx.Where("department_id = ?", departmentID).
			Delete(&models.DepartmentPermission{}).Error; err != nil {
			return err
		}

		// 添加新的权限
		for _, permID := range permissionIDs {
			deptPerm := &models.DepartmentPermission{
				DepartmentID: departmentID,
				PermissionID: permID,
			}
			if err := tx.Create(deptPerm).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// GetPermissions 获取部门权限列表
func (s *DepartmentService) GetPermissions(tenantID, departmentID int64) ([]models.Permission, error) {
	// 验证部门是否存在
	var department models.Department
	if err := database.DB.Where("id = ? AND tenant_id = ?", departmentID, tenantID).
		First(&department).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("部门不存在")
		}
		return nil, err
	}

	var permissions []models.Permission
	query := `
		SELECT p.* FROM permissions p
		INNER JOIN department_permissions dp ON p.id = dp.permission_id
		WHERE dp.department_id = ?
	`
	if err := database.DB.Raw(query, departmentID).Scan(&permissions).Error; err != nil {
		return nil, err
	}

	return permissions, nil
}

// GetInheritedPermissions 获取部门继承的所有权限（包括父部门）
func (s *DepartmentService) GetInheritedPermissions(tenantID, departmentID int64) ([]models.Permission, error) {
	// 获取部门路径上的所有部门ID
	var department models.Department
	if err := database.DB.Where("id = ? AND tenant_id = ?", departmentID, tenantID).
		First(&department).Error; err != nil {
		return nil, err
	}

	// 解析path获取所有祖先部门ID
	var departmentIDs []int64
	pathParts := strings.Split(strings.Trim(department.Path, "/"), "/")
	for _, part := range pathParts {
		if part != "" {
			var id int64
			fmt.Sscanf(part, "%d", &id)
			departmentIDs = append(departmentIDs, id)
		}
	}
	departmentIDs = append(departmentIDs, departmentID)

	// 获取所有相关部门的权限（去重）
	var permissions []models.Permission
	query := `
		SELECT DISTINCT p.* FROM permissions p
		INNER JOIN department_permissions dp ON p.id = dp.permission_id
		WHERE dp.department_id IN (?)
	`
	if err := database.DB.Raw(query, departmentIDs).Scan(&permissions).Error; err != nil {
		return nil, err
	}

	return permissions, nil
}
