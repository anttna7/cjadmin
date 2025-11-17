package service

import (
	"testing"

	"cjadmin/internal/test"

	"github.com/stretchr/testify/assert"
)

func TestCustomerService_CreateCustomer(t *testing.T) {
	testDB := test.SetupTestDB(t)
	defer testDB.TearDown(t)

	customerService := NewCustomerService(testDB.DB)
	tenant := testDB.CreateTestTenant(t, "test-tenant")
	user := testDB.CreateTestUser(t, tenant.ID, "testuser", "")

	tests := []struct {
		name        string
		req         *CreateCustomerRequest
		wantErr     bool
		errContains string
	}{
		{
			name: "successful customer creation",
			req: &CreateCustomerRequest{
				CustomerName: "Test Customer Inc.",
				CustomerType: "enterprise",
				Industry:     "IT",
				Source:       "website",
				Level:        "A",
				Country:      "China",
				Province:     "Beijing",
				City:         "Beijing",
			},
			wantErr: false,
		},
		{
			name: "duplicate customer name",
			req: &CreateCustomerRequest{
				CustomerName: "Test Customer Inc.",
				CustomerType: "enterprise",
				Industry:     "IT",
			},
			wantErr:     true,
			errContains: "客户名称已存在",
		},
		{
			name: "missing required fields",
			req: &CreateCustomerRequest{
				CustomerType: "enterprise",
			},
			wantErr:     true,
			errContains: "客户名称不能为空",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			customer, err := customerService.CreateCustomer(tenant.ID, tt.req, user.ID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, customer)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, customer)
				assert.Equal(t, tt.req.CustomerName, customer.CustomerName)
				assert.Equal(t, "active", customer.Status)
				assert.Equal(t, user.ID, customer.CreatedBy)
			}
		})
	}
}

func TestCustomerService_UpdateCustomer(t *testing.T) {
	testDB := test.SetupTestDB(t)
	defer testDB.TearDown(t)

	customerService := NewCustomerService(testDB.DB)
	tenant := testDB.CreateTestTenant(t, "test-tenant")
	customer := testDB.CreateTestCustomer(t, tenant.ID, "Original Customer")

	tests := []struct {
		name       string
		customerID int64
		req        *UpdateCustomerRequest
		wantErr    bool
	}{
		{
			name:       "successful update",
			customerID: customer.ID,
			req: &UpdateCustomerRequest{
				CustomerName: strPtr("Updated Customer Name"),
				Industry:     strPtr("Finance"),
				Level:        strPtr("B"),
			},
			wantErr: false,
		},
		{
			name:       "customer not found",
			customerID: 99999,
			req: &UpdateCustomerRequest{
				CustomerName: strPtr("Non-existent"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updatedCustomer, err := customerService.UpdateCustomer(tenant.ID, tt.customerID, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, updatedCustomer)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, updatedCustomer)
				if tt.req.CustomerName != nil {
					assert.Equal(t, *tt.req.CustomerName, updatedCustomer.CustomerName)
				}
				if tt.req.Industry != nil {
					assert.Equal(t, *tt.req.Industry, updatedCustomer.Industry)
				}
				if tt.req.Level != nil {
					assert.Equal(t, *tt.req.Level, updatedCustomer.Level)
				}
			}
		})
	}
}

func TestCustomerService_GetCustomer(t *testing.T) {
	testDB := test.SetupTestDB(t)
	defer testDB.TearDown(t)

	customerService := NewCustomerService(testDB.DB)
	tenant := testDB.CreateTestTenant(t, "test-tenant")
	customer := testDB.CreateTestCustomer(t, tenant.ID, "Test Customer")

	tests := []struct {
		name       string
		customerID int64
		wantErr    bool
	}{
		{
			name:       "get existing customer",
			customerID: customer.ID,
			wantErr:    false,
		},
		{
			name:       "get non-existent customer",
			customerID: 99999,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			foundCustomer, err := customerService.GetCustomer(tenant.ID, tt.customerID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, foundCustomer)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, foundCustomer)
				assert.Equal(t, customer.ID, foundCustomer.ID)
				assert.Equal(t, customer.CustomerName, foundCustomer.CustomerName)
			}
		})
	}
}

func TestCustomerService_ListCustomers(t *testing.T) {
	testDB := test.SetupTestDB(t)
	defer testDB.TearDown(t)

	customerService := NewCustomerService(testDB.DB)
	tenant := testDB.CreateTestTenant(t, "test-tenant")
	user := testDB.CreateTestUser(t, tenant.ID, "testuser", "")

	// 创建多个测试客户
	testDB.CreateTestCustomer(t, tenant.ID, "Customer A")
	testDB.CreateTestCustomer(t, tenant.ID, "Customer B")
	testDB.CreateTestCustomer(t, tenant.ID, "Customer C")

	tests := []struct {
		name      string
		page      int
		pageSize  int
		level     string
		wantCount int
		wantErr   bool
	}{
		{
			name:      "list all customers",
			page:      1,
			pageSize:  10,
			level:     "",
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:      "list with pagination",
			page:      1,
			pageSize:  2,
			level:     "",
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "filter by level",
			page:      1,
			pageSize:  10,
			level:     "A",
			wantCount: 3, // All test customers are level A
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			customers, total, err := customerService.ListCustomers(tenant.ID, tt.page, tt.pageSize, "", tt.level, "", "")

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, int64(3), total)
				assert.Len(t, customers, tt.wantCount)
			}
		})
	}
}

func TestCustomerService_SearchCustomers(t *testing.T) {
	testDB := test.SetupTestDB(t)
	defer testDB.TearDown(t)

	customerService := NewCustomerService(testDB.DB)
	tenant := testDB.CreateTestTenant(t, "test-tenant")

	// 创建测试客户
	testDB.CreateTestCustomer(t, tenant.ID, "Apple Inc.")
	testDB.CreateTestCustomer(t, tenant.ID, "Microsoft Corp.")
	testDB.CreateTestCustomer(t, tenant.ID, "Google LLC")

	tests := []struct {
		name      string
		keyword   string
		wantCount int
	}{
		{
			name:      "search by partial name",
			keyword:   "Apple",
			wantCount: 1,
		},
		{
			name:      "search by common term",
			keyword:   "Inc",
			wantCount: 1,
		},
		{
			name:      "no results",
			keyword:   "NonExistent",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			customers, total, err := customerService.SearchCustomers(tenant.ID, tt.keyword, 1, 10)

			assert.NoError(t, err)
			assert.Equal(t, int64(tt.wantCount), total)
			assert.Len(t, customers, tt.wantCount)
		})
	}
}

func TestCustomerService_DeleteCustomer(t *testing.T) {
	testDB := test.SetupTestDB(t)
	defer testDB.TearDown(t)

	customerService := NewCustomerService(testDB.DB)
	tenant := testDB.CreateTestTenant(t, "test-tenant")
	customer := testDB.CreateTestCustomer(t, tenant.ID, "Test Customer")

	// 删除客户
	err := customerService.DeleteCustomer(tenant.ID, customer.ID)
	assert.NoError(t, err)

	// 验证客户已被软删除
	_, err = customerService.GetCustomer(tenant.ID, customer.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "客户不存在")
}

func TestCustomerService_GetStatistics(t *testing.T) {
	testDB := test.SetupTestDB(t)
	defer testDB.TearDown(t)

	customerService := NewCustomerService(testDB.DB)
	tenant := testDB.CreateTestTenant(t, "test-tenant")

	// 创建不同级别的客户
	testDB.CreateTestCustomer(t, tenant.ID, "Customer A Level")
	testDB.CreateTestCustomer(t, tenant.ID, "Customer A Level 2")

	// 创建B级客户
	customerB := testDB.CreateTestCustomer(t, tenant.ID, "Customer B Level")
	testDB.DB.Model(customerB).Update("level", "B")

	// 获取统计数据
	stats, err := customerService.GetStatistics(tenant.ID)
	assert.NoError(t, err)
	assert.NotNil(t, stats)

	// 验证总数
	assert.Equal(t, int64(3), stats["total_customers"])

	// 验证等级分布
	levelDistribution := stats["level_distribution"].(map[string]int64)
	assert.Equal(t, int64(2), levelDistribution["A"])
	assert.Equal(t, int64(1), levelDistribution["B"])
}

func TestCustomerService_BatchUpdateLevel(t *testing.T) {
	testDB := test.SetupTestDB(t)
	defer testDB.TearDown(t)

	customerService := NewCustomerService(testDB.DB)
	tenant := testDB.CreateTestTenant(t, "test-tenant")

	// 创建多个客户
	customer1 := testDB.CreateTestCustomer(t, tenant.ID, "Customer 1")
	customer2 := testDB.CreateTestCustomer(t, tenant.ID, "Customer 2")
	customer3 := testDB.CreateTestCustomer(t, tenant.ID, "Customer 3")

	customerIDs := []int64{customer1.ID, customer2.ID, customer3.ID}

	// 批量更新等级
	err := customerService.BatchUpdateLevel(tenant.ID, customerIDs, "S")
	assert.NoError(t, err)

	// 验证所有客户都已更新
	for _, id := range customerIDs {
		customer, err := customerService.GetCustomer(tenant.ID, id)
		assert.NoError(t, err)
		assert.Equal(t, "S", customer.Level)
	}
}
