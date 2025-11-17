package middleware

import (
	"github.com/anttna7/cjadmin/internal/database"
	"github.com/anttna7/cjadmin/internal/models"
	"github.com/anttna7/cjadmin/internal/utils"
	"github.com/gin-gonic/gin"
)

// RequirePermission 权限验证中间件
func RequirePermission(permissionCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 平台管理员拥有所有权限
		if IsPlatformAdmin(c) {
			c.Next()
			return
		}

		userID := GetUserID(c)
		if userID == 0 {
			utils.ErrorWithStatus(c, 401, utils.CodeUnauthorized, "未登录")
			c.Abort()
			return
		}

		// 检查用户是否拥有该权限
		hasPermission := checkUserPermission(userID, permissionCode)
		if !hasPermission {
			utils.ErrorWithStatus(c, 403, utils.CodeForbidden, "没有权限")
			c.Abort()
			return
		}

		c.Next()
	}
}

// checkUserPermission 检查用户是否拥有指定权限
func checkUserPermission(userID int64, permissionCode string) bool {
	var count int64

	// 通过用户角色检查权限
	query := `
		SELECT COUNT(*) FROM permissions p
		INNER JOIN role_permissions rp ON p.id = rp.permission_id
		INNER JOIN user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = ? AND p.code = ?
	`
	database.DB.Raw(query, userID, permissionCode).Scan(&count)

	if count > 0 {
		return true
	}

	// 通过部门权限检查（部门权限继承）
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return false
	}

	if user.DepartmentID != nil {
		// 获取部门路径上的所有部门ID（包括父部门）
		departmentIDs := getDepartmentHierarchy(*user.DepartmentID)

		query = `
			SELECT COUNT(*) FROM permissions p
			INNER JOIN department_permissions dp ON p.id = dp.permission_id
			WHERE dp.department_id IN (?) AND p.code = ?
		`
		database.DB.Raw(query, departmentIDs, permissionCode).Scan(&count)

		if count > 0 {
			return true
		}
	}

	return false
}

// getDepartmentHierarchy 获取部门及其所有父部门的ID列表
func getDepartmentHierarchy(departmentID int64) []int64 {
	var ids []int64
	ids = append(ids, departmentID)

	var dept models.Department
	currentID := departmentID

	for {
		if err := database.DB.First(&dept, currentID).Error; err != nil {
			break
		}

		if dept.ParentID == nil {
			break
		}

		ids = append(ids, *dept.ParentID)
		currentID = *dept.ParentID
	}

	return ids
}

// GetUserPermissions 获取用户的所有权限
func GetUserPermissions(userID int64) ([]string, error) {
	var permissions []string

	// 通过角色获取权限
	query := `
		SELECT DISTINCT p.code FROM permissions p
		INNER JOIN role_permissions rp ON p.id = rp.permission_id
		INNER JOIN user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = ?
	`
	if err := database.DB.Raw(query, userID).Scan(&permissions).Error; err != nil {
		return nil, err
	}

	// 通过部门获取权限
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return permissions, nil
	}

	if user.DepartmentID != nil {
		departmentIDs := getDepartmentHierarchy(*user.DepartmentID)
		var deptPermissions []string

		query = `
			SELECT DISTINCT p.code FROM permissions p
			INNER JOIN department_permissions dp ON p.id = dp.permission_id
			WHERE dp.department_id IN (?)
		`
		database.DB.Raw(query, departmentIDs).Scan(&deptPermissions)

		// 合并权限
		permMap := make(map[string]bool)
		for _, p := range permissions {
			permMap[p] = true
		}
		for _, p := range deptPermissions {
			permMap[p] = true
		}

		permissions = make([]string, 0, len(permMap))
		for p := range permMap {
			permissions = append(permissions, p)
		}
	}

	return permissions, nil
}
