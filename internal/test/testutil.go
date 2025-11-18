package test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"cjadmin/internal/config"
	"cjadmin/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestDB 测试数据库配置
type TestDB struct {
	DB *gorm.DB
}

// SetupTestDB 设置测试数据库
func SetupTestDB(t *testing.T) *TestDB {
	// 从环境变量读取测试数据库配置
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres")
	dbname := getEnv("DB_NAME", "cjadmin_test")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // 测试时关闭SQL日志
	})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// 自动迁移所有模型
	err = db.AutoMigrate(
		&models.Tenant{},
		&models.User{},
		&models.Department{},
		&models.Role{},
		&models.Permission{},
		&models.Customer{},
		&models.Contact{},
		&models.FollowUp{},
		&models.Contract{},
		&models.ContractItem{},
		&models.Order{},
		&models.OrderItem{},
		&models.Payment{},
		&models.Invoice{},
		&models.CustomForm{},
		&models.CustomFormData{},
		&models.SystemSetting{},
		&models.AuditLog{},
		&models.FileUpload{},
		// 安全相关模型
		&models.SMSCode{},
		&models.OAuthBinding{},
		&models.LoginBlacklist{},
		&models.LoginWhitelist{},
		&models.LoginAttempt{},
		&models.OAuthConfig{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return &TestDB{DB: db}
}

// TearDownTestDB 清理测试数据库
func (tdb *TestDB) TearDown(t *testing.T) {
	// 清空所有表数据
	tables := []string{
		// 安全相关表
		"oauth_configs", "login_attempts", "login_whitelist",
		"login_blacklist", "oauth_bindings", "sms_codes",
		// 其他表
		"file_uploads", "audit_logs", "system_settings",
		"custom_form_data", "custom_forms", "invoices",
		"payments", "order_items", "orders",
		"contract_items", "contracts", "follow_ups",
		"contacts", "customers", "role_permissions",
		"user_roles", "permissions", "roles",
		"users", "departments", "tenants",
	}

	for _, table := range tables {
		tdb.DB.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table))
	}
}

// CreateTestTenant 创建测试租户
func (tdb *TestDB) CreateTestTenant(t *testing.T, name string) *models.Tenant {
	tenant := &models.Tenant{
		TenantName: name,
		ContactPerson: "Test Contact",
		ContactPhone: "13800138000",
		ContactEmail: fmt.Sprintf("test@%s.com", name),
		Status: "active",
		ExpireAt: time.Now().AddDate(1, 0, 0),
	}

	if err := tdb.DB.Create(tenant).Error; err != nil {
		t.Fatalf("Failed to create test tenant: %v", err)
	}

	return tenant
}

// CreateTestUser 创建测试用户
func (tdb *TestDB) CreateTestUser(t *testing.T, tenantID int64, username string, role string) *models.User {
	user := &models.User{
		TenantID: tenantID,
		Username: username,
		Password: "$2a$10$test.hashed.password", // 预哈希的测试密码
		RealName: "Test User",
		Email: fmt.Sprintf("%s@test.com", username),
		Phone: "13800138000",
		Status: "active",
	}

	if err := tdb.DB.Create(user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// 如果指定了角色，创建角色并关联
	if role != "" {
		testRole := &models.Role{
			TenantID: tenantID,
			RoleName: role,
			Description: "Test role",
			Status: "active",
		}
		if err := tdb.DB.Create(testRole).Error; err != nil {
			t.Fatalf("Failed to create test role: %v", err)
		}

		// 关联用户和角色
		if err := tdb.DB.Exec("INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)",
			user.ID, testRole.ID).Error; err != nil {
			t.Fatalf("Failed to assign role to user: %v", err)
		}
	}

	return user
}

// CreateTestCustomer 创建测试客户
func (tdb *TestDB) CreateTestCustomer(t *testing.T, tenantID int64, name string) *models.Customer {
	customer := &models.Customer{
		TenantID: tenantID,
		CustomerName: name,
		CustomerType: "enterprise",
		Industry: "IT",
		Source: "website",
		Level: "A",
		Status: "active",
		Country: "China",
		Province: "Beijing",
		City: "Beijing",
	}

	if err := tdb.DB.Create(customer).Error; err != nil {
		t.Fatalf("Failed to create test customer: %v", err)
	}

	return customer
}

// CreateTestDepartment 创建测试部门
func (tdb *TestDB) CreateTestDepartment(t *testing.T, tenantID int64, name string, parentID *int64) *models.Department {
	dept := &models.Department{
		TenantID: tenantID,
		DeptName: name,
		ParentID: parentID,
		DeptLevel: 1,
		SortOrder: 0,
		Status: "active",
	}

	if parentID != nil {
		dept.DeptLevel = 2
	}

	if err := tdb.DB.Create(dept).Error; err != nil {
		t.Fatalf("Failed to create test department: %v", err)
	}

	return dept
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// MockConfig 创建测试用的配置
func MockConfig() *config.Config {
	return &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: 8080,
			Mode: "test",
			BaseURL: "http://localhost:8080",
		},
		Database: config.DatabaseConfig{
			Host: "localhost",
			Port: 5432,
			User: "postgres",
			Password: "postgres",
			DBName: "cjadmin_test",
			SSLMode: "disable",
		},
		JWT: config.JWTConfig{
			Secret: "test-secret-key",
			ExpireHours: 24,
		},
		Upload: config.UploadConfig{
			MaxSize: 10 * 1024 * 1024, // 10MB
			Path: "./test_uploads",
		},
		Security: config.SecurityConfig{
			BcryptCost: 4, // 使用较低的成本以加快测试速度
			PasswordMinLength: 6,
			SessionTimeout: 3600,
		},
	}
}
