package service

import (
	"github.com/anttna7/cjadmin/internal/database"
	"github.com/anttna7/cjadmin/internal/models"
)

type PermissionService struct{}

func NewPermissionService() *PermissionService {
	return &PermissionService{}
}

// GetAll 获取所有权限列表
// 任务 5.2.1: 获取所有权限列表
func (s *PermissionService) GetAll() ([]models.Permission, error) {
	var permissions []models.Permission
	if err := database.DB.Order("resource, action").Find(&permissions).Error; err != nil {
		return nil, err
	}
	return permissions, nil
}

// GetByResource 按资源类型获取权限
func (s *PermissionService) GetByResource(resource string) ([]models.Permission, error) {
	var permissions []models.Permission
	if err := database.DB.Where("resource = ?", resource).
		Order("action").
		Find(&permissions).Error; err != nil {
		return nil, err
	}
	return permissions, nil
}

// GetGrouped 获取按资源分组的权限列表
func (s *PermissionService) GetGrouped() (map[string][]models.Permission, error) {
	var permissions []models.Permission
	if err := database.DB.Order("resource, action").Find(&permissions).Error; err != nil {
		return nil, err
	}

	// 按资源分组
	grouped := make(map[string][]models.Permission)
	for _, perm := range permissions {
		grouped[perm.Resource] = append(grouped[perm.Resource], perm)
	}

	return grouped, nil
}
