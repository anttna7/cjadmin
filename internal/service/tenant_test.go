package service

import (
	"testing"
	"time"

	"cjadmin/internal/test"

	"github.com/stretchr/testify/assert"
)

func TestTenantService_CreateTenant(t *testing.T) {
	testDB := test.SetupTestDB(t)
	defer testDB.TearDown(t)

	tenantService := NewTenantService(testDB.DB)

	tests := []struct {
		name        string
		req         *CreateTenantRequest
		wantErr     bool
		errContains string
	}{
		{
			name: "successful tenant creation",
			req: &CreateTenantRequest{
				TenantName:    "Test Company",
				ContactPerson: "John Doe",
				ContactPhone:  "13800138000",
				ContactEmail:  "contact@testcompany.com",
			},
			wantErr: false,
		},
		{
			name: "duplicate tenant name",
			req: &CreateTenantRequest{
				TenantName:    "Test Company",
				ContactPerson: "Jane Doe",
				ContactPhone:  "13800138001",
				ContactEmail:  "jane@testcompany.com",
			},
			wantErr:     true,
			errContains: "租户名称已存在",
		},
		{
			name: "missing required fields",
			req: &CreateTenantRequest{
				TenantName: "Incomplete Tenant",
			},
			wantErr:     true,
			errContains: "联系人不能为空",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tenant, err := tenantService.CreateTenant(tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, tenant)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, tenant)
				assert.Equal(t, tt.req.TenantName, tenant.TenantName)
				assert.Equal(t, "active", tenant.Status)
				assert.NotZero(t, tenant.ExpireAt)
			}
		})
	}
}

func TestTenantService_UpdateTenant(t *testing.T) {
	testDB := test.SetupTestDB(t)
	defer testDB.TearDown(t)

	tenantService := NewTenantService(testDB.DB)
	tenant := testDB.CreateTestTenant(t, "test-tenant")

	tests := []struct {
		name     string
		tenantID int64
		req      *UpdateTenantRequest
		wantErr  bool
	}{
		{
			name:     "successful update",
			tenantID: tenant.ID,
			req: &UpdateTenantRequest{
				TenantName:    strPtr("Updated Tenant Name"),
				ContactPerson: strPtr("Updated Contact"),
			},
			wantErr: false,
		},
		{
			name:     "tenant not found",
			tenantID: 99999,
			req: &UpdateTenantRequest{
				TenantName: strPtr("Non-existent"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updatedTenant, err := tenantService.UpdateTenant(tt.tenantID, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, updatedTenant)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, updatedTenant)
				if tt.req.TenantName != nil {
					assert.Equal(t, *tt.req.TenantName, updatedTenant.TenantName)
				}
				if tt.req.ContactPerson != nil {
					assert.Equal(t, *tt.req.ContactPerson, updatedTenant.ContactPerson)
				}
			}
		})
	}
}

func TestTenantService_GetTenant(t *testing.T) {
	testDB := test.SetupTestDB(t)
	defer testDB.TearDown(t)

	tenantService := NewTenantService(testDB.DB)
	tenant := testDB.CreateTestTenant(t, "test-tenant")

	tests := []struct {
		name     string
		tenantID int64
		wantErr  bool
	}{
		{
			name:     "get existing tenant",
			tenantID: tenant.ID,
			wantErr:  false,
		},
		{
			name:     "get non-existent tenant",
			tenantID: 99999,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			foundTenant, err := tenantService.GetTenant(tt.tenantID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, foundTenant)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, foundTenant)
				assert.Equal(t, tenant.ID, foundTenant.ID)
				assert.Equal(t, tenant.TenantName, foundTenant.TenantName)
			}
		})
	}
}

func TestTenantService_ListTenants(t *testing.T) {
	testDB := test.SetupTestDB(t)
	defer testDB.TearDown(t)

	tenantService := NewTenantService(testDB.DB)

	// 创建多个测试租户
	testDB.CreateTestTenant(t, "tenant-1")
	testDB.CreateTestTenant(t, "tenant-2")
	testDB.CreateTestTenant(t, "tenant-3")

	tests := []struct {
		name      string
		page      int
		pageSize  int
		status    string
		wantCount int
		wantErr   bool
	}{
		{
			name:      "list all tenants",
			page:      1,
			pageSize:  10,
			status:    "",
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:      "list with pagination",
			page:      1,
			pageSize:  2,
			status:    "",
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "filter by status",
			page:      1,
			pageSize:  10,
			status:    "active",
			wantCount: 3,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tenants, total, err := tenantService.ListTenants(tt.page, tt.pageSize, tt.status, "")

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, int64(3), total)
				assert.Len(t, tenants, tt.wantCount)
			}
		})
	}
}

func TestTenantService_DisableTenant(t *testing.T) {
	testDB := test.SetupTestDB(t)
	defer testDB.TearDown(t)

	tenantService := NewTenantService(testDB.DB)
	tenant := testDB.CreateTestTenant(t, "test-tenant")

	// 禁用租户
	err := tenantService.DisableTenant(tenant.ID)
	assert.NoError(t, err)

	// 验证租户状态
	disabledTenant, err := tenantService.GetTenant(tenant.ID)
	assert.NoError(t, err)
	assert.Equal(t, "disabled", disabledTenant.Status)
}

func TestTenantService_EnableTenant(t *testing.T) {
	testDB := test.SetupTestDB(t)
	defer testDB.TearDown(t)

	tenantService := NewTenantService(testDB.DB)
	tenant := testDB.CreateTestTenant(t, "test-tenant")

	// 先禁用
	err := tenantService.DisableTenant(tenant.ID)
	assert.NoError(t, err)

	// 再启用
	err = tenantService.EnableTenant(tenant.ID)
	assert.NoError(t, err)

	// 验证租户状态
	enabledTenant, err := tenantService.GetTenant(tenant.ID)
	assert.NoError(t, err)
	assert.Equal(t, "active", enabledTenant.Status)
}

func TestTenantService_ExtendExpiry(t *testing.T) {
	testDB := test.SetupTestDB(t)
	defer testDB.TearDown(t)

	tenantService := NewTenantService(testDB.DB)
	tenant := testDB.CreateTestTenant(t, "test-tenant")

	originalExpiry := tenant.ExpireAt

	// 延长1年
	err := tenantService.ExtendExpiry(tenant.ID, 365)
	assert.NoError(t, err)

	// 验证过期时间
	updatedTenant, err := tenantService.GetTenant(tenant.ID)
	assert.NoError(t, err)
	assert.True(t, updatedTenant.ExpireAt.After(originalExpiry))

	// 计算差值应该约为365天
	duration := updatedTenant.ExpireAt.Sub(originalExpiry)
	expectedDuration := 365 * 24 * time.Hour
	assert.InDelta(t, expectedDuration, duration, float64(time.Hour))
}

func TestTenantService_DeleteTenant(t *testing.T) {
	testDB := test.SetupTestDB(t)
	defer testDB.TearDown(t)

	tenantService := NewTenantService(testDB.DB)
	tenant := testDB.CreateTestTenant(t, "test-tenant")

	// 删除租户
	err := tenantService.DeleteTenant(tenant.ID)
	assert.NoError(t, err)

	// 验证租户已被软删除
	_, err = tenantService.GetTenant(tenant.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "租户不存在")
}
