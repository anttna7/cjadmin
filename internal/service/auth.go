package service

import (
	"errors"
	"fmt"

	"github.com/anttna7/cjadmin/internal/config"
	"github.com/anttna7/cjadmin/internal/database"
	"github.com/anttna7/cjadmin/internal/models"
	"github.com/anttna7/cjadmin/internal/utils"
	"gorm.io/gorm"
)

type AuthService struct {
	cfg *config.Config
}

func NewAuthService(cfg *config.Config) *AuthService {
	return &AuthService{cfg: cfg}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string      `json:"token"`
	User  *UserInfo   `json:"user"`
}

type UserInfo struct {
	ID              int64   `json:"id"`
	Username        string  `json:"username"`
	RealName        string  `json:"real_name"`
	Email           string  `json:"email"`
	TenantID        *int64  `json:"tenant_id"`
	IsPlatformAdmin bool    `json:"is_platform_admin"`
	Permissions     []string `json:"permissions"`
}

// Login 用户登录
func (s *AuthService) Login(req *LoginRequest) (*LoginResponse, error) {
	var user models.User

	// 查询用户
	if err := database.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户名或密码错误")
		}
		return nil, err
	}

	// 验证密码
	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		return nil, errors.New("用户名或密码错误")
	}

	// 检查用户状态
	if user.Status != "active" {
		return nil, fmt.Errorf("用户状态异常: %s", user.Status)
	}

	// 生成token
	token, err := utils.GenerateToken(
		user.ID,
		user.Username,
		user.TenantID,
		user.IsPlatformAdmin,
		s.cfg.JWT.ExpireHours,
	)
	if err != nil {
		return nil, err
	}

	// 获取用户权限
	permissions := []string{}
	if user.IsPlatformAdmin {
		// 平台管理员拥有所有权限
		database.DB.Model(&models.Permission{}).Pluck("code", &permissions)
	} else {
		// 获取普通用户的权限
		query := `
			SELECT DISTINCT p.code FROM permissions p
			INNER JOIN role_permissions rp ON p.id = rp.permission_id
			INNER JOIN user_roles ur ON rp.role_id = ur.role_id
			WHERE ur.user_id = ?
		`
		database.DB.Raw(query, user.ID).Scan(&permissions)
	}

	userInfo := &UserInfo{
		ID:              user.ID,
		Username:        user.Username,
		RealName:        user.RealName,
		Email:           user.Email,
		TenantID:        user.TenantID,
		IsPlatformAdmin: user.IsPlatformAdmin,
		Permissions:     permissions,
	}

	return &LoginResponse{
		Token: token,
		User:  userInfo,
	}, nil
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6"`
	RealName string `json:"real_name"`
	Email    string `json:"email" binding:"email"`
	Phone    string `json:"phone"`
	TenantID *int64 `json:"tenant_id"`
}

// Register 用户注册（仅限平台管理员创建用户）
func (s *AuthService) Register(req *RegisterRequest) (*models.User, error) {
	// 检查用户名是否已存在
	var count int64
	database.DB.Model(&models.User{}).Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		return nil, errors.New("用户名已存在")
	}

	// 加密密码
	passwordHash, err := utils.HashPassword(req.Password, s.cfg.Security.BcryptCost)
	if err != nil {
		return nil, err
	}

	// 创建用户
	user := &models.User{
		Username:     req.Username,
		PasswordHash: passwordHash,
		RealName:     req.RealName,
		Email:        req.Email,
		Phone:        req.Phone,
		TenantID:     req.TenantID,
		Status:       "active",
	}

	if err := database.DB.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

// CreateFirstAdmin 创建第一个平台管理员
func (s *AuthService) CreateFirstAdmin(username, password string) error {
	// 检查是否已存在管理员
	var count int64
	database.DB.Model(&models.User{}).Where("is_platform_admin = ?", true).Count(&count)
	if count > 0 {
		return errors.New("平台管理员已存在")
	}

	// 加密密码
	passwordHash, err := utils.HashPassword(password, s.cfg.Security.BcryptCost)
	if err != nil {
		return err
	}

	// 创建管理员
	user := &models.User{
		Username:        username,
		PasswordHash:    passwordHash,
		RealName:        "超级管理员",
		Status:          "active",
		IsPlatformAdmin: true,
	}

	if err := database.DB.Create(user).Error; err != nil {
		return err
	}

	// 分配超级管理员角色
	var superAdminRole models.Role
	if err := database.DB.Where("code = ?", "super_admin").First(&superAdminRole).Error; err == nil {
		userRole := &models.UserRole{
			UserID: user.ID,
			RoleID: superAdminRole.ID,
		}
		database.DB.Create(userRole)
	}

	return nil
}
