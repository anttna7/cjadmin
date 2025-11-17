package service

import (
	"testing"
	"time"

	"cjadmin/internal/models"
	"cjadmin/internal/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthService_Login(t *testing.T) {
	// 设置测试数据库
	testDB := test.SetupTestDB(t)
	defer testDB.TearDown(t)

	// 创建认证服务
	authService := NewAuthService(testDB.DB)

	// 创建测试租户
	tenant := testDB.CreateTestTenant(t, "test-tenant")

	// 创建测试用户（密码为 "password123"）
	hashedPassword := "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"
	user := &models.User{
		TenantID: tenant.ID,
		Username: "testuser",
		Password: hashedPassword,
		RealName: "Test User",
		Email:    "test@example.com",
		Status:   "active",
	}
	require.NoError(t, testDB.DB.Create(user).Error)

	tests := []struct {
		name        string
		username    string
		password    string
		wantErr     bool
		errContains string
	}{
		{
			name:     "successful login",
			username: "testuser",
			password: "password123",
			wantErr:  false,
		},
		{
			name:        "wrong password",
			username:    "testuser",
			password:    "wrongpassword",
			wantErr:     true,
			errContains: "密码错误",
		},
		{
			name:        "user not found",
			username:    "nonexistent",
			password:    "password123",
			wantErr:     true,
			errContains: "用户不存在",
		},
		{
			name:        "empty username",
			username:    "",
			password:    "password123",
			wantErr:     true,
			errContains: "用户不存在",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, loginUser, err := authService.Login(tt.username, tt.password)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Empty(t, token)
				assert.Nil(t, loginUser)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
				assert.NotNil(t, loginUser)
				assert.Equal(t, tt.username, loginUser.Username)
			}
		})
	}
}

func TestAuthService_CreateUser(t *testing.T) {
	testDB := test.SetupTestDB(t)
	defer testDB.TearDown(t)

	authService := NewAuthService(testDB.DB)
	tenant := testDB.CreateTestTenant(t, "test-tenant")
	dept := testDB.CreateTestDepartment(t, tenant.ID, "IT部门", nil)

	tests := []struct {
		name        string
		req         *CreateUserRequest
		wantErr     bool
		errContains string
	}{
		{
			name: "successful user creation",
			req: &CreateUserRequest{
				Username:     "newuser",
				Password:     "password123",
				RealName:     "New User",
				Email:        "newuser@example.com",
				Phone:        "13800138000",
				DepartmentID: &dept.ID,
			},
			wantErr: false,
		},
		{
			name: "duplicate username",
			req: &CreateUserRequest{
				Username: "newuser",
				Password: "password123",
				RealName: "Duplicate User",
				Email:    "duplicate@example.com",
			},
			wantErr:     true,
			errContains: "用户名已存在",
		},
		{
			name: "weak password",
			req: &CreateUserRequest{
				Username: "weakpass",
				Password: "123",
				RealName: "Weak User",
				Email:    "weak@example.com",
			},
			wantErr:     true,
			errContains: "密码长度不能少于",
		},
		{
			name: "invalid email",
			req: &CreateUserRequest{
				Username: "invalidemail",
				Password: "password123",
				RealName: "Invalid Email User",
				Email:    "invalid-email",
			},
			wantErr:     true,
			errContains: "邮箱格式不正确",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := authService.CreateUser(tenant.ID, tt.req, 0)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.req.Username, user.Username)
				assert.Equal(t, tt.req.RealName, user.RealName)
				assert.NotEqual(t, tt.req.Password, user.Password) // Password should be hashed
			}
		})
	}
}

func TestAuthService_UpdateUser(t *testing.T) {
	testDB := test.SetupTestDB(t)
	defer testDB.TearDown(t)

	authService := NewAuthService(testDB.DB)
	tenant := testDB.CreateTestTenant(t, "test-tenant")
	user := testDB.CreateTestUser(t, tenant.ID, "testuser", "")

	tests := []struct {
		name    string
		userID  int64
		req     *UpdateUserRequest
		wantErr bool
	}{
		{
			name:   "successful update",
			userID: user.ID,
			req: &UpdateUserRequest{
				RealName: strPtr("Updated Name"),
				Email:    strPtr("updated@example.com"),
			},
			wantErr: false,
		},
		{
			name:   "user not found",
			userID: 99999,
			req: &UpdateUserRequest{
				RealName: strPtr("Updated Name"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updatedUser, err := authService.UpdateUser(tenant.ID, tt.userID, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, updatedUser)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, updatedUser)
				if tt.req.RealName != nil {
					assert.Equal(t, *tt.req.RealName, updatedUser.RealName)
				}
				if tt.req.Email != nil {
					assert.Equal(t, *tt.req.Email, updatedUser.Email)
				}
			}
		})
	}
}

func TestAuthService_ChangePassword(t *testing.T) {
	testDB := test.SetupTestDB(t)
	defer testDB.TearDown(t)

	authService := NewAuthService(testDB.DB)
	tenant := testDB.CreateTestTenant(t, "test-tenant")

	// 创建测试用户（密码为 "password123"）
	hashedPassword := "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"
	user := &models.User{
		TenantID: tenant.ID,
		Username: "testuser",
		Password: hashedPassword,
		RealName: "Test User",
		Status:   "active",
	}
	require.NoError(t, testDB.DB.Create(user).Error)

	tests := []struct {
		name        string
		oldPassword string
		newPassword string
		wantErr     bool
		errContains string
	}{
		{
			name:        "successful password change",
			oldPassword: "password123",
			newPassword: "newpassword123",
			wantErr:     false,
		},
		{
			name:        "wrong old password",
			oldPassword: "wrongpassword",
			newPassword: "newpassword123",
			wantErr:     true,
			errContains: "原密码错误",
		},
		{
			name:        "weak new password",
			oldPassword: "password123",
			newPassword: "123",
			wantErr:     true,
			errContains: "密码长度不能少于",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := authService.ChangePassword(tenant.ID, user.ID, tt.oldPassword, tt.newPassword)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)

				// 验证新密码可以登录
				_, _, err := authService.Login(user.Username, tt.newPassword)
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthService_DisableUser(t *testing.T) {
	testDB := test.SetupTestDB(t)
	defer testDB.TearDown(t)

	authService := NewAuthService(testDB.DB)
	tenant := testDB.CreateTestTenant(t, "test-tenant")
	user := testDB.CreateTestUser(t, tenant.ID, "testuser", "")

	// 禁用用户
	err := authService.DisableUser(tenant.ID, user.ID)
	assert.NoError(t, err)

	// 验证用户状态
	var disabledUser models.User
	err = testDB.DB.First(&disabledUser, user.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "disabled", disabledUser.Status)

	// 验证禁用的用户无法登录
	_, _, err = authService.Login(user.Username, "password123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "用户已被禁用")
}

func TestAuthService_ValidateToken(t *testing.T) {
	testDB := test.SetupTestDB(t)
	defer testDB.TearDown(t)

	authService := NewAuthService(testDB.DB)
	tenant := testDB.CreateTestTenant(t, "test-tenant")
	user := testDB.CreateTestUser(t, tenant.ID, "testuser", "")

	// 生成有效的token
	token, err := authService.GenerateToken(user)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// 验证有效token
	claims, err := authService.ValidateToken(token)
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, user.Username, claims.Username)

	// 验证无效token
	_, err = authService.ValidateToken("invalid.token.string")
	assert.Error(t, err)
}

// Helper function
func strPtr(s string) *string {
	return &s
}
