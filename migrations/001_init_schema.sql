-- 客户管理+收款系统数据库初始化脚本
-- PostgreSQL 18+

-- 1. 租户表
CREATE TABLE IF NOT EXISTS tenants (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(100) UNIQUE NOT NULL,
    status VARCHAR(20) DEFAULT 'active',
    settings JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_tenants_status ON tenants(status) WHERE deleted_at IS NULL;

-- 2. 部门表
CREATE TABLE IF NOT EXISTS departments (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES tenants(id),
    parent_id BIGINT REFERENCES departments(id),
    name VARCHAR(255) NOT NULL,
    code VARCHAR(100),
    level INT DEFAULT 1,
    path VARCHAR(500),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    UNIQUE(tenant_id, code)
);

CREATE INDEX idx_departments_tenant ON departments(tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_departments_parent ON departments(parent_id) WHERE deleted_at IS NULL;

-- 3. 用户表
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT REFERENCES tenants(id),
    department_id BIGINT REFERENCES departments(id),
    username VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    real_name VARCHAR(100),
    email VARCHAR(255),
    phone VARCHAR(50),
    status VARCHAR(20) DEFAULT 'active',
    is_platform_admin BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_users_tenant ON users(tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_department ON users(department_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_status ON users(status) WHERE deleted_at IS NULL;

-- 4. 角色表
CREATE TABLE IF NOT EXISTS roles (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT REFERENCES tenants(id),
    name VARCHAR(100) NOT NULL,
    code VARCHAR(100) NOT NULL,
    description TEXT,
    is_platform_role BOOLEAN DEFAULT FALSE,
    created_by BIGINT REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    UNIQUE(tenant_id, code)
);

CREATE INDEX idx_roles_tenant ON roles(tenant_id) WHERE deleted_at IS NULL;

-- 5. 权限表
CREATE TABLE IF NOT EXISTS permissions (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    resource VARCHAR(100),
    action VARCHAR(50),
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 6. 角色权限关联表
CREATE TABLE IF NOT EXISTS role_permissions (
    id BIGSERIAL PRIMARY KEY,
    role_id BIGINT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id BIGINT NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(role_id, permission_id)
);

CREATE INDEX idx_role_permissions_role ON role_permissions(role_id);
CREATE INDEX idx_role_permissions_perm ON role_permissions(permission_id);

-- 7. 用户角色关联表
CREATE TABLE IF NOT EXISTS user_roles (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id BIGINT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, role_id)
);

CREATE INDEX idx_user_roles_user ON user_roles(user_id);
CREATE INDEX idx_user_roles_role ON user_roles(role_id);

-- 8. 部门权限表
CREATE TABLE IF NOT EXISTS department_permissions (
    id BIGSERIAL PRIMARY KEY,
    department_id BIGINT NOT NULL REFERENCES departments(id) ON DELETE CASCADE,
    permission_id BIGINT NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(department_id, permission_id)
);

CREATE INDEX idx_dept_permissions_dept ON department_permissions(department_id);
CREATE INDEX idx_dept_permissions_perm ON department_permissions(permission_id);

-- 9. 客户信息表
CREATE TABLE IF NOT EXISTS customers (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES tenants(id),
    customer_code VARCHAR(100) NOT NULL,
    customer_name VARCHAR(255) NOT NULL,
    customer_type VARCHAR(50),
    contact_person VARCHAR(100),
    phone VARCHAR(50),
    email VARCHAR(255),
    address TEXT,
    industry VARCHAR(100),
    source VARCHAR(100),
    assigned_to BIGINT REFERENCES users(id),
    status VARCHAR(20) DEFAULT 'active',
    custom_fields JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    UNIQUE(tenant_id, customer_code)
);

CREATE INDEX idx_customers_tenant ON customers(tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_customers_assigned ON customers(assigned_to) WHERE deleted_at IS NULL;
CREATE INDEX idx_customers_status ON customers(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_customers_name ON customers(customer_name) WHERE deleted_at IS NULL;

-- 10. 合同归档表
CREATE TABLE IF NOT EXISTS contracts (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES tenants(id),
    customer_id BIGINT NOT NULL REFERENCES customers(id),
    contract_no VARCHAR(100) NOT NULL,
    contract_name VARCHAR(255),
    contract_amount DECIMAL(15, 2),
    start_date DATE,
    end_date DATE,
    file_url VARCHAR(500),
    status VARCHAR(20) DEFAULT 'active',
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    UNIQUE(tenant_id, contract_no)
);

CREATE INDEX idx_contracts_tenant ON contracts(tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_contracts_customer ON contracts(customer_id) WHERE deleted_at IS NULL;

-- 11. 订单表
CREATE TABLE IF NOT EXISTS orders (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES tenants(id),
    customer_id BIGINT NOT NULL REFERENCES customers(id),
    order_no VARCHAR(100) UNIQUE NOT NULL,
    order_type VARCHAR(20) NOT NULL,
    amount DECIMAL(15, 2) NOT NULL,
    status VARCHAR(50) NOT NULL,
    payment_method VARCHAR(50),
    payment_channel VARCHAR(100),
    transaction_id VARCHAR(255),
    payment_time TIMESTAMP,
    verified_time TIMESTAMP,
    verified_by BIGINT REFERENCES users(id),
    notes TEXT,
    metadata JSONB,
    created_by BIGINT REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_orders_tenant ON orders(tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_orders_customer ON orders(customer_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_orders_type_status ON orders(order_type, status) WHERE deleted_at IS NULL;
CREATE INDEX idx_orders_created_at ON orders(created_at DESC) WHERE deleted_at IS NULL;

-- 12. 订单状态变更日志
CREATE TABLE IF NOT EXISTS order_status_logs (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL REFERENCES orders(id),
    from_status VARCHAR(50),
    to_status VARCHAR(50) NOT NULL,
    operator_id BIGINT REFERENCES users(id),
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_order_logs_order ON order_status_logs(order_id);

-- 13. 客户账户表
CREATE TABLE IF NOT EXISTS customer_accounts (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES tenants(id),
    customer_id BIGINT NOT NULL REFERENCES customers(id),
    account_type VARCHAR(20) NOT NULL,
    balance DECIMAL(15, 2) DEFAULT 0,
    cash_balance DECIMAL(15, 2) DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(customer_id, account_type)
);

CREATE INDEX idx_accounts_tenant ON customer_accounts(tenant_id);
CREATE INDEX idx_accounts_customer ON customer_accounts(customer_id);

-- 14. 账户交易明细表
CREATE TABLE IF NOT EXISTS account_transactions (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES tenants(id),
    account_id BIGINT NOT NULL REFERENCES customer_accounts(id),
    order_id BIGINT REFERENCES orders(id),
    transaction_type VARCHAR(50) NOT NULL,
    amount DECIMAL(15, 2) NOT NULL,
    balance_before DECIMAL(15, 2),
    balance_after DECIMAL(15, 2),
    cash_balance_before DECIMAL(15, 2),
    cash_balance_after DECIMAL(15, 2),
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_transactions_tenant ON account_transactions(tenant_id);
CREATE INDEX idx_transactions_account ON account_transactions(account_id);
CREATE INDEX idx_transactions_order ON account_transactions(order_id);
CREATE INDEX idx_transactions_created ON account_transactions(created_at DESC);

-- 15. 发票管理表
CREATE TABLE IF NOT EXISTS invoices (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES tenants(id),
    customer_id BIGINT NOT NULL REFERENCES customers(id),
    order_id BIGINT REFERENCES orders(id),
    invoice_no VARCHAR(100) UNIQUE NOT NULL,
    invoice_type VARCHAR(20),
    invoice_title VARCHAR(255),
    tax_no VARCHAR(100),
    amount DECIMAL(15, 2) NOT NULL,
    tax_amount DECIMAL(15, 2),
    status VARCHAR(20) DEFAULT 'pending',
    issue_date DATE,
    file_url VARCHAR(500),
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_invoices_tenant ON invoices(tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_invoices_customer ON invoices(customer_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_invoices_order ON invoices(order_id) WHERE deleted_at IS NULL;

-- 16. 自定义表单定义
CREATE TABLE IF NOT EXISTS custom_forms (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT REFERENCES tenants(id),
    form_name VARCHAR(255) NOT NULL,
    form_code VARCHAR(100) NOT NULL,
    form_type VARCHAR(50),
    form_schema JSONB NOT NULL,
    status VARCHAR(20) DEFAULT 'active',
    created_by BIGINT REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    UNIQUE(tenant_id, form_code)
);

CREATE INDEX idx_forms_tenant ON custom_forms(tenant_id) WHERE deleted_at IS NULL;

-- 17. 自定义表单数据
CREATE TABLE IF NOT EXISTS custom_form_data (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES tenants(id),
    form_id BIGINT NOT NULL REFERENCES custom_forms(id),
    entity_type VARCHAR(50),
    entity_id BIGINT,
    form_data JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_form_data_tenant ON custom_form_data(tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_form_data_form ON custom_form_data(form_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_form_data_entity ON custom_form_data(entity_type, entity_id) WHERE deleted_at IS NULL;

-- 18. 系统设置表
CREATE TABLE IF NOT EXISTS system_settings (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT REFERENCES tenants(id),
    category VARCHAR(50) NOT NULL,
    key VARCHAR(100) NOT NULL,
    value TEXT,
    value_type VARCHAR(20) DEFAULT 'string',
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, category, key)
);

CREATE INDEX idx_settings_tenant_category ON system_settings(tenant_id, category);

-- 19. 审计日志表
CREATE TABLE IF NOT EXISTS audit_logs (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT REFERENCES tenants(id),
    user_id BIGINT REFERENCES users(id),
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50),
    resource_id BIGINT,
    ip_address VARCHAR(50),
    user_agent TEXT,
    request_data JSONB,
    response_data JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_audit_tenant ON audit_logs(tenant_id);
CREATE INDEX idx_audit_user ON audit_logs(user_id);
CREATE INDEX idx_audit_created ON audit_logs(created_at DESC);

-- 20. 导入日志表
CREATE TABLE IF NOT EXISTS import_logs (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES tenants(id),
    import_type VARCHAR(50) NOT NULL,
    file_name VARCHAR(255),
    total_rows INT,
    success_rows INT,
    failed_rows INT,
    error_details JSONB,
    status VARCHAR(20),
    imported_by BIGINT REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_import_logs_tenant ON import_logs(tenant_id) WHERE deleted_at IS NULL;

-- 插入基础权限数据
INSERT INTO permissions (code, name, resource, action, description) VALUES
-- 租户管理
('tenant.create', '创建租户', 'tenant', 'create', '创建新租户'),
('tenant.read', '查看租户', 'tenant', 'read', '查看租户信息'),
('tenant.update', '更新租户', 'tenant', 'update', '更新租户信息'),
('tenant.delete', '删除租户', 'tenant', 'delete', '删除租户'),

-- 角色管理
('role.create', '创建角色', 'role', 'create', '创建新角色'),
('role.read', '查看角色', 'role', 'read', '查看角色信息'),
('role.update', '更新角色', 'role', 'update', '更新角色信息'),
('role.delete', '删除角色', 'role', 'delete', '删除角色'),

-- 用户管理
('user.create', '创建用户', 'user', 'create', '创建新用户'),
('user.read', '查看用户', 'user', 'read', '查看用户信息'),
('user.update', '更新用户', 'user', 'update', '更新用户信息'),
('user.delete', '删除用户', 'user', 'delete', '删除用户'),

-- 客户管理
('customer.create', '创建客户', 'customer', 'create', '创建新客户'),
('customer.read', '查看客户', 'customer', 'read', '查看客户信息'),
('customer.update', '更新客户', 'customer', 'update', '更新客户信息'),
('customer.delete', '删除客户', 'customer', 'delete', '删除客户'),
('customer.import', '导入客户', 'customer', 'import', '批量导入客户'),
('customer.export', '导出客户', 'customer', 'export', '导出客户数据'),

-- 订单管理
('order.create', '创建订单', 'order', 'create', '创建新订单'),
('order.read', '查看订单', 'order', 'read', '查看订单信息'),
('order.update', '更新订单', 'order', 'update', '更新订单信息'),
('order.delete', '删除订单', 'order', 'delete', '删除订单'),
('order.import', '导入订单', 'order', 'import', '批量导入订单'),
('order.export', '导出订单', 'order', 'export', '导出订单数据'),

-- 财务管理
('finance.verify', '财务审核', 'finance', 'verify', '审核充值订单'),
('finance.invoice.create', '创建发票', 'invoice', 'create', '创建发票'),
('finance.invoice.read', '查看发票', 'invoice', 'read', '查看发票信息'),
('finance.invoice.update', '更新发票', 'invoice', 'update', '更新发票信息'),
('finance.invoice.delete', '删除发票', 'invoice', 'delete', '删除发票'),

-- 表单管理
('form.create', '创建表单', 'form', 'create', '创建自定义表单'),
('form.read', '查看表单', 'form', 'read', '查看表单'),
('form.update', '更新表单', 'form', 'update', '更新表单'),
('form.delete', '删除表单', 'form', 'delete', '删除表单'),

-- 系统设置
('system.setting', '系统设置', 'system', 'setting', '管理系统设置')
ON CONFLICT (code) DO NOTHING;

-- 创建默认的超级管理员角色
INSERT INTO roles (tenant_id, name, code, description, is_platform_role)
VALUES (NULL, '超级管理员', 'super_admin', '平台超级管理员，拥有所有权限', TRUE)
ON CONFLICT DO NOTHING;

-- 为超级管理员角色分配所有权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT
    (SELECT id FROM roles WHERE code = 'super_admin' LIMIT 1),
    id
FROM permissions
ON CONFLICT DO NOTHING;

-- 创建触发器，自动更新 updated_at 字段
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 为需要的表添加触发器
DO $$
DECLARE
    t text;
BEGIN
    FOREACH t IN ARRAY ARRAY['tenants', 'departments', 'users', 'roles', 'customers', 'contracts', 'orders', 'customer_accounts', 'invoices', 'custom_forms', 'custom_form_data', 'system_settings', 'import_logs']
    LOOP
        EXECUTE format('
            DROP TRIGGER IF EXISTS update_%I_updated_at ON %I;
            CREATE TRIGGER update_%I_updated_at
                BEFORE UPDATE ON %I
                FOR EACH ROW
                EXECUTE FUNCTION update_updated_at_column();
        ', t, t, t, t);
    END LOOP;
END$$;
