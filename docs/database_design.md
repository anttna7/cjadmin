# 客户管理+收款系统 - 数据库设计文档

## 系统架构
- 单体应用架构
- 多租户SaaS模式（基于tenant_id隔离）
- 技术栈：Go 1.25+, PostgreSQL 18+, Gin 1.11+

## 核心表结构设计

### 1. 租户和组织架构

#### tenants (租户表)
```sql
CREATE TABLE tenants (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(100) UNIQUE NOT NULL,
    status VARCHAR(20) DEFAULT 'active', -- active, suspended, inactive
    settings JSONB, -- 租户配置
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);
```

#### departments (部门表)
```sql
CREATE TABLE departments (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES tenants(id),
    parent_id BIGINT REFERENCES departments(id),
    name VARCHAR(255) NOT NULL,
    code VARCHAR(100),
    level INT DEFAULT 1,
    path VARCHAR(500), -- 部门路径，如 /1/2/3/
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    UNIQUE(tenant_id, code)
);
CREATE INDEX idx_departments_tenant ON departments(tenant_id);
CREATE INDEX idx_departments_parent ON departments(parent_id);
```

### 2. 用户和角色权限

#### users (用户表)
```sql
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT REFERENCES tenants(id), -- NULL表示平台用户
    department_id BIGINT REFERENCES departments(id),
    username VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    real_name VARCHAR(100),
    email VARCHAR(255),
    phone VARCHAR(50),
    status VARCHAR(20) DEFAULT 'active',
    is_platform_admin BOOLEAN DEFAULT FALSE, -- 是否平台管理员
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);
CREATE INDEX idx_users_tenant ON users(tenant_id);
CREATE INDEX idx_users_department ON users(department_id);
```

#### roles (角色表)
```sql
CREATE TABLE roles (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT REFERENCES tenants(id), -- NULL表示平台角色
    name VARCHAR(100) NOT NULL,
    code VARCHAR(100) NOT NULL,
    description TEXT,
    is_platform_role BOOLEAN DEFAULT FALSE, -- 是否平台角色
    created_by BIGINT REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    UNIQUE(tenant_id, code)
);
CREATE INDEX idx_roles_tenant ON roles(tenant_id);
```

#### permissions (权限表)
```sql
CREATE TABLE permissions (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    resource VARCHAR(100), -- 资源类型：customer, order, finance等
    action VARCHAR(50), -- 操作：create, read, update, delete, import, export
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### role_permissions (角色权限关联表)
```sql
CREATE TABLE role_permissions (
    id BIGSERIAL PRIMARY KEY,
    role_id BIGINT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id BIGINT NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(role_id, permission_id)
);
CREATE INDEX idx_role_permissions_role ON role_permissions(role_id);
```

#### user_roles (用户角色关联表)
```sql
CREATE TABLE user_roles (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id BIGINT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, role_id)
);
CREATE INDEX idx_user_roles_user ON user_roles(user_id);
```

#### department_permissions (部门权限表)
```sql
CREATE TABLE department_permissions (
    id BIGSERIAL PRIMARY KEY,
    department_id BIGINT NOT NULL REFERENCES departments(id) ON DELETE CASCADE,
    permission_id BIGINT NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(department_id, permission_id)
);
CREATE INDEX idx_dept_permissions_dept ON department_permissions(department_id);
```

### 3. 客户管理

#### customers (客户信息表)
```sql
CREATE TABLE customers (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES tenants(id),
    customer_code VARCHAR(100) NOT NULL,
    customer_name VARCHAR(255) NOT NULL,
    customer_type VARCHAR(50), -- 企业/个人
    contact_person VARCHAR(100),
    phone VARCHAR(50),
    email VARCHAR(255),
    address TEXT,
    industry VARCHAR(100),
    source VARCHAR(100), -- 客户来源
    assigned_to BIGINT REFERENCES users(id), -- 负责业务员
    status VARCHAR(20) DEFAULT 'active',
    custom_fields JSONB, -- 自定义字段
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    UNIQUE(tenant_id, customer_code)
);
CREATE INDEX idx_customers_tenant ON customers(tenant_id);
CREATE INDEX idx_customers_assigned ON customers(assigned_to);
CREATE INDEX idx_customers_status ON customers(status);
```

#### contracts (合同归档表)
```sql
CREATE TABLE contracts (
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
CREATE INDEX idx_contracts_tenant ON contracts(tenant_id);
CREATE INDEX idx_contracts_customer ON contracts(customer_id);
```

### 4. 订单管理

#### orders (订单表)
```sql
CREATE TABLE orders (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES tenants(id),
    customer_id BIGINT NOT NULL REFERENCES customers(id),
    order_no VARCHAR(100) UNIQUE NOT NULL,
    order_type VARCHAR(20) NOT NULL, -- recharge:充值, settlement:结算, transfer:转账
    amount DECIMAL(15, 2) NOT NULL,
    status VARCHAR(50) NOT NULL, -- 订单状态（根据类型不同有不同状态）
    payment_method VARCHAR(50),
    payment_channel VARCHAR(100),
    transaction_id VARCHAR(255), -- 第三方交易ID
    payment_time TIMESTAMP,
    verified_time TIMESTAMP, -- 财务审核时间
    verified_by BIGINT REFERENCES users(id),
    notes TEXT,
    metadata JSONB, -- 扩展字段
    created_by BIGINT REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);
CREATE INDEX idx_orders_tenant ON orders(tenant_id);
CREATE INDEX idx_orders_customer ON orders(customer_id);
CREATE INDEX idx_orders_type_status ON orders(order_type, status);
CREATE INDEX idx_orders_created_at ON orders(created_at DESC);
```

#### order_status_logs (订单状态变更日志)
```sql
CREATE TABLE order_status_logs (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL REFERENCES orders(id),
    from_status VARCHAR(50),
    to_status VARCHAR(50) NOT NULL,
    operator_id BIGINT REFERENCES users(id),
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_order_logs_order ON order_status_logs(order_id);
```

### 5. 账户余额管理

#### customer_accounts (客户账户表)
```sql
CREATE TABLE customer_accounts (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES tenants(id),
    customer_id BIGINT NOT NULL REFERENCES customers(id),
    account_type VARCHAR(20) NOT NULL, -- fund:资金账户, consumption:消耗账户
    balance DECIMAL(15, 2) DEFAULT 0, -- 总余额
    cash_balance DECIMAL(15, 2) DEFAULT 0, -- 现金余额
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(customer_id, account_type)
);
CREATE INDEX idx_accounts_tenant ON customer_accounts(tenant_id);
CREATE INDEX idx_accounts_customer ON customer_accounts(customer_id);
```

#### account_transactions (账户交易明细表)
```sql
CREATE TABLE account_transactions (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES tenants(id),
    account_id BIGINT NOT NULL REFERENCES customer_accounts(id),
    order_id BIGINT REFERENCES orders(id),
    transaction_type VARCHAR(50) NOT NULL, -- recharge:充值, transfer:转账, consume:消耗, refund:退款
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
```

### 6. 财务管理

#### invoices (发票管理表)
```sql
CREATE TABLE invoices (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES tenants(id),
    customer_id BIGINT NOT NULL REFERENCES customers(id),
    order_id BIGINT REFERENCES orders(id),
    invoice_no VARCHAR(100) UNIQUE NOT NULL,
    invoice_type VARCHAR(20), -- 增值税专用发票、普通发票
    invoice_title VARCHAR(255),
    tax_no VARCHAR(100),
    amount DECIMAL(15, 2) NOT NULL,
    tax_amount DECIMAL(15, 2),
    status VARCHAR(20) DEFAULT 'pending', -- pending, issued, cancelled
    issue_date DATE,
    file_url VARCHAR(500),
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);
CREATE INDEX idx_invoices_tenant ON invoices(tenant_id);
CREATE INDEX idx_invoices_customer ON invoices(customer_id);
CREATE INDEX idx_invoices_order ON invoices(order_id);
```

### 7. 自定义表单

#### custom_forms (自定义表单定义)
```sql
CREATE TABLE custom_forms (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT REFERENCES tenants(id),
    form_name VARCHAR(255) NOT NULL,
    form_code VARCHAR(100) NOT NULL,
    form_type VARCHAR(50), -- customer, order, etc.
    form_schema JSONB NOT NULL, -- 表单字段定义
    status VARCHAR(20) DEFAULT 'active',
    created_by BIGINT REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    UNIQUE(tenant_id, form_code)
);
CREATE INDEX idx_forms_tenant ON custom_forms(tenant_id);
```

#### custom_form_data (自定义表单数据)
```sql
CREATE TABLE custom_form_data (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES tenants(id),
    form_id BIGINT NOT NULL REFERENCES custom_forms(id),
    entity_type VARCHAR(50), -- customer, order等
    entity_id BIGINT, -- 关联的实体ID
    form_data JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_form_data_tenant ON custom_form_data(tenant_id);
CREATE INDEX idx_form_data_form ON custom_form_data(form_id);
CREATE INDEX idx_form_data_entity ON custom_form_data(entity_type, entity_id);
```

### 8. 系统设置

#### system_settings (系统设置表)
```sql
CREATE TABLE system_settings (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT REFERENCES tenants(id), -- NULL表示平台级设置
    category VARCHAR(50) NOT NULL, -- security, notification, login
    key VARCHAR(100) NOT NULL,
    value TEXT,
    value_type VARCHAR(20) DEFAULT 'string', -- string, number, boolean, json
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, category, key)
);
CREATE INDEX idx_settings_tenant_category ON system_settings(tenant_id, category);
```

#### audit_logs (审计日志表)
```sql
CREATE TABLE audit_logs (
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
```

#### import_logs (导入日志表)
```sql
CREATE TABLE import_logs (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES tenants(id),
    import_type VARCHAR(50) NOT NULL, -- customer, order等
    file_name VARCHAR(255),
    total_rows INT,
    success_rows INT,
    failed_rows INT,
    error_details JSONB,
    status VARCHAR(20), -- processing, completed, failed
    imported_by BIGINT REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_import_logs_tenant ON import_logs(tenant_id);
```

## 订单状态流转

### 充值订单状态流转
1. `pending` - 待支付
2. `paid` - 已支付（客户完成充值）
3. `verified` - 财务已确认到账
4. `completed` - 已完成（余额已入账）
5. `failed` - 失败
6. `cancelled` - 已取消

### 转账订单状态流转
1. `pending` - 待处理
2. `processing` - 处理中
3. `completed` - 已完成
4. `failed` - 失败

### 结算订单状态流转
1. `pending` - 待结算
2. `processing` - 结算中
3. `completed` - 已完成
4. `failed` - 失败

## 权限继承逻辑

1. 部门权限继承：子部门自动继承父部门的权限
2. 用户权限 = 角色权限 ∪ 部门权限
3. 租户角色的权限不能超过创建者的权限
4. 平台角色只能由平台管理员创建和管理

## 索引优化说明

- 所有外键都已建立索引
- 租户ID在所有多租户表上都有索引（数据隔离查询优化）
- 常用查询字段（状态、时间等）都有索引
- 使用JSONB类型存储灵活配置，支持GIN索引

## 数据安全

- 软删除：使用deleted_at字段
- 审计日志：记录所有重要操作
- 多租户隔离：严格的tenant_id过滤
- 密码加密：使用bcrypt
- 敏感数据：可考虑加密存储
