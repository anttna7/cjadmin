-- 充值与支付管理系统数据库迁移
-- Migration: 010_add_payment_and_credit_management
-- Date: 2025-11-17
-- Description: 添加收款账户管理、返点政策、授信管理等功能

-- =============================================================================
-- 第一部分：扩展 customers 表
-- =============================================================================

-- 添加账户和代理商信息
ALTER TABLE customers ADD COLUMN IF NOT EXISTS account_id VARCHAR(100);
ALTER TABLE customers ADD COLUMN IF NOT EXISTS agent_name VARCHAR(200);
ALTER TABLE customers ADD COLUMN IF NOT EXISTS agent_id VARCHAR(100);
ALTER TABLE customers ADD COLUMN IF NOT EXISTS channel VARCHAR(100);

-- 添加报备和行业信息
ALTER TABLE customers ADD COLUMN IF NOT EXISTS report_tags TEXT[];
ALTER TABLE customers ADD COLUMN IF NOT EXISTS report_industry VARCHAR(200);
ALTER TABLE customers ADD COLUMN IF NOT EXISTS industry_level1 VARCHAR(200);
ALTER TABLE customers ADD COLUMN IF NOT EXISTS industry_level2 VARCHAR(200);

-- 添加返点配置（对公充值）
ALTER TABLE customers ADD COLUMN IF NOT EXISTS public_rebate_rate DECIMAL(5,2) DEFAULT 0;
ALTER TABLE customers ADD COLUMN IF NOT EXISTS public_cash_rate DECIMAL(5,2) DEFAULT 100;
ALTER TABLE customers ADD COLUMN IF NOT EXISTS public_gift_rate DECIMAL(5,2) DEFAULT 0;

-- 添加返点配置（对私充值）
ALTER TABLE customers ADD COLUMN IF NOT EXISTS private_rebate_rate DECIMAL(5,2) DEFAULT 0;
ALTER TABLE customers ADD COLUMN IF NOT EXISTS private_cash_rate DECIMAL(5,2) DEFAULT 100;
ALTER TABLE customers ADD COLUMN IF NOT EXISTS private_gift_rate DECIMAL(5,2) DEFAULT 0;

-- 添加业务信息
ALTER TABLE customers ADD COLUMN IF NOT EXISTS business_platform VARCHAR(200);
ALTER TABLE customers ADD COLUMN IF NOT EXISTS contract_period VARCHAR(100);

-- 添加索引
CREATE INDEX IF NOT EXISTS idx_customers_account_id ON customers(account_id);
CREATE INDEX IF NOT EXISTS idx_customers_agent_id ON customers(agent_id);
CREATE INDEX IF NOT EXISTS idx_customers_industry ON customers(industry_level1, industry_level2);

-- 添加注释
COMMENT ON COLUMN customers.account_id IS '客户账户ID';
COMMENT ON COLUMN customers.agent_name IS '代理商名称';
COMMENT ON COLUMN customers.agent_id IS '代理商ID';
COMMENT ON COLUMN customers.channel IS '客户渠道';
COMMENT ON COLUMN customers.report_tags IS '报备标签数组';
COMMENT ON COLUMN customers.report_industry IS '报备行业类别';
COMMENT ON COLUMN customers.industry_level1 IS '一级行业';
COMMENT ON COLUMN customers.industry_level2 IS '二级行业';
COMMENT ON COLUMN customers.public_rebate_rate IS '对公充值返点比例(%)';
COMMENT ON COLUMN customers.public_cash_rate IS '对公返点现金比例(%)';
COMMENT ON COLUMN customers.public_gift_rate IS '对公返点赠款比例(%)';
COMMENT ON COLUMN customers.private_rebate_rate IS '对私充值返点比例(%)';
COMMENT ON COLUMN customers.private_cash_rate IS '对私返点现金比例(%)';
COMMENT ON COLUMN customers.private_gift_rate IS '对私返点赠款比例(%)';
COMMENT ON COLUMN customers.business_platform IS '业务平台';
COMMENT ON COLUMN customers.contract_period IS '合同期限';

-- =============================================================================
-- 第二部分：扩展 customer_accounts 表
-- =============================================================================

-- 添加赠款余额和授信余额
ALTER TABLE customer_accounts ADD COLUMN IF NOT EXISTS gift_balance DECIMAL(15,2) DEFAULT 0;
ALTER TABLE customer_accounts ADD COLUMN IF NOT EXISTS credit_balance DECIMAL(15,2) DEFAULT 0;
ALTER TABLE customer_accounts ADD COLUMN IF NOT EXISTS credit_limit DECIMAL(15,2) DEFAULT 0;

-- 添加注释
COMMENT ON COLUMN customer_accounts.gift_balance IS '赠款余额（返点赠送的金额）';
COMMENT ON COLUMN customer_accounts.credit_balance IS '授信余额（已使用的授信额度）';
COMMENT ON COLUMN customer_accounts.credit_limit IS '授信额度上限';

-- =============================================================================
-- 第三部分：创建 payment_accounts 表（收款账户管理）
-- =============================================================================

CREATE TABLE IF NOT EXISTS payment_accounts (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,

    -- 基本信息
    account_name VARCHAR(200) NOT NULL,
    account_type VARCHAR(50) NOT NULL,  -- public（对公）, private（对私）
    payment_method VARCHAR(50) NOT NULL, -- bank_card（银行卡）, alipay（支付宝）, wechat（微信）, api（API接口）

    -- 账户信息
    account_number VARCHAR(200),
    account_holder VARCHAR(200),
    bank_name VARCHAR(200),
    bank_branch VARCHAR(200),

    -- API接口信息（当 payment_method = 'api' 时使用）
    api_config_id BIGINT,  -- 关联 payment_api_configs 表
    api_type VARCHAR(50),  -- BOC（中国银行）, ICBC（工商银行）, ALIPAY（支付宝）等
    api_merchant_id VARCHAR(200),
    api_app_id VARCHAR(200),
    api_secret_key TEXT,  -- 加密存储
    api_public_key TEXT,
    api_endpoint VARCHAR(500),

    -- 轮询策略配置
    status VARCHAR(20) DEFAULT 'active',  -- active（启用）, inactive（停用）, maintenance（维护中）
    is_auto_rotate BOOLEAN DEFAULT true,  -- 是否参与自动轮询
    rotate_strategy VARCHAR(50) DEFAULT 'weight',  -- weight（权重）, round_robin（轮流）, balance（余额）, frequency（频率）, random（随机）
    rotate_weight INT DEFAULT 1,  -- 轮询权重（1-100）
    rotate_priority INT DEFAULT 0,  -- 优先级（数字越大优先级越高）

    -- 余额和限额
    current_balance DECIMAL(15,2) DEFAULT 0,
    daily_limit DECIMAL(15,2),  -- 单日限额
    single_limit DECIMAL(15,2), -- 单笔限额
    monthly_limit DECIMAL(15,2), -- 月限额

    -- 使用统计
    use_frequency INT DEFAULT 0,  -- 使用次数
    total_amount DECIMAL(15,2) DEFAULT 0,  -- 累计收款金额
    total_count INT DEFAULT 0,  -- 累计收款次数
    today_amount DECIMAL(15,2) DEFAULT 0,  -- 今日收款金额
    today_count INT DEFAULT 0,  -- 今日收款次数
    this_month_amount DECIMAL(15,2) DEFAULT 0,  -- 本月收款金额
    this_month_count INT DEFAULT 0,  -- 本月收款次数
    last_used_at TIMESTAMP,  -- 最后使用时间

    -- 附加信息
    description TEXT,
    remark TEXT,

    -- 时间戳
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,

    -- 外键约束
    CONSTRAINT fk_payment_accounts_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_payment_accounts_tenant ON payment_accounts(tenant_id);
CREATE INDEX IF NOT EXISTS idx_payment_accounts_type ON payment_accounts(account_type);
CREATE INDEX IF NOT EXISTS idx_payment_accounts_status ON payment_accounts(status);
CREATE INDEX IF NOT EXISTS idx_payment_accounts_method ON payment_accounts(payment_method);
CREATE INDEX IF NOT EXISTS idx_payment_accounts_rotate ON payment_accounts(is_auto_rotate, status);
CREATE INDEX IF NOT EXISTS idx_payment_accounts_deleted ON payment_accounts(deleted_at);

-- 添加注释
COMMENT ON TABLE payment_accounts IS '收款账户表';
COMMENT ON COLUMN payment_accounts.account_type IS '账户类型：public-对公, private-对私';
COMMENT ON COLUMN payment_accounts.payment_method IS '支付方式：bank_card-银行卡, alipay-支付宝, wechat-微信, api-API接口';
COMMENT ON COLUMN payment_accounts.rotate_strategy IS '轮询策略：weight-权重, round_robin-轮流, balance-余额, frequency-频率, random-随机';

-- =============================================================================
-- 第四部分：创建 payment_api_configs 表（支付接口配置）
-- =============================================================================

CREATE TABLE IF NOT EXISTS payment_api_configs (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,

    -- 提供商信息
    provider_name VARCHAR(200) NOT NULL,  -- 中国银行, 工商银行, 支付宝等
    provider_code VARCHAR(100) NOT NULL,  -- BOC, ICBC, ALIPAY等
    provider_type VARCHAR(50) NOT NULL,   -- bank（银行）, payment_platform（支付平台）, acquirer（收单机构）

    -- API配置
    api_version VARCHAR(50),
    api_endpoint VARCHAR(500) NOT NULL,
    api_sandbox_endpoint VARCHAR(500),
    is_sandbox BOOLEAN DEFAULT false,

    -- 认证信息
    merchant_id VARCHAR(200),
    app_id VARCHAR(200),
    app_secret TEXT,  -- 加密存储
    public_key TEXT,
    private_key TEXT, -- 加密存储
    cert_path VARCHAR(500),

    -- 扩展配置（JSON格式，用于存储特定提供商的额外参数）
    extra_config JSONB,

    -- 功能支持
    supports_payment BOOLEAN DEFAULT true,
    supports_query BOOLEAN DEFAULT true,
    supports_refund BOOLEAN DEFAULT false,
    supports_transfer BOOLEAN DEFAULT false,

    -- 状态
    status VARCHAR(20) DEFAULT 'active',  -- active, inactive

    -- 附加信息
    description TEXT,
    remark TEXT,

    -- 时间戳
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,

    -- 外键约束
    CONSTRAINT fk_payment_api_configs_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_payment_api_configs_tenant ON payment_api_configs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_payment_api_configs_provider ON payment_api_configs(provider_code);
CREATE INDEX IF NOT EXISTS idx_payment_api_configs_status ON payment_api_configs(status);
CREATE INDEX IF NOT EXISTS idx_payment_api_configs_deleted ON payment_api_configs(deleted_at);
CREATE INDEX IF NOT EXISTS idx_payment_api_configs_extra ON payment_api_configs USING GIN (extra_config);

-- 添加注释
COMMENT ON TABLE payment_api_configs IS '支付接口配置表';
COMMENT ON COLUMN payment_api_configs.provider_type IS '提供商类型：bank-银行, payment_platform-支付平台, acquirer-收单机构';
COMMENT ON COLUMN payment_api_configs.extra_config IS '扩展配置JSON，用于存储不同提供商的特定参数';

-- 添加 payment_accounts 表的外键约束（现在 payment_api_configs 表已创建）
ALTER TABLE payment_accounts ADD CONSTRAINT IF NOT EXISTS fk_payment_accounts_api_config
    FOREIGN KEY (api_config_id) REFERENCES payment_api_configs(id) ON DELETE SET NULL;

-- =============================================================================
-- 第五部分：创建 credit_records 表（授信记录）
-- =============================================================================

CREATE TABLE IF NOT EXISTS credit_records (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    customer_id BIGINT NOT NULL,

    -- 授信信息
    record_type VARCHAR(50) NOT NULL,  -- grant（授信）, adjust（调整）, repay（还款）, consume（消费）
    credit_amount DECIMAL(15,2) NOT NULL,  -- 授信/调整/还款金额
    before_credit DECIMAL(15,2),  -- 操作前授信余额
    after_credit DECIMAL(15,2),   -- 操作后授信余额
    before_limit DECIMAL(15,2),   -- 操作前授信额度
    after_limit DECIMAL(15,2),    -- 操作后授信额度

    -- 申请信息
    apply_reason TEXT,
    applicant_id BIGINT,  -- 申请人
    apply_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- 审批信息
    approval_status VARCHAR(50) DEFAULT 'pending',  -- pending（待审批）, approved（已批准）, rejected（已拒绝）, cancelled（已取消）
    approver_id BIGINT,  -- 实际审批人
    assigned_approver_id BIGINT,  -- 指定审批人（超级管理员可指定）
    auto_assigned BOOLEAN DEFAULT false,  -- 是否自动分配审批人
    approval_time TIMESTAMP,
    approval_remark TEXT,

    -- 还款信息（record_type = 'repay' 时使用）
    repay_amount DECIMAL(15,2),
    repay_method VARCHAR(50),  -- cash（现金）, transfer（转账）, deduction（扣款）
    repay_order_id BIGINT,  -- 关联的充值订单ID

    -- 附件
    attachments JSONB,  -- 附件列表 JSON

    -- 附加信息
    remark TEXT,

    -- 时间戳
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,

    -- 外键约束
    CONSTRAINT fk_credit_records_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    CONSTRAINT fk_credit_records_customer FOREIGN KEY (customer_id) REFERENCES customers(id) ON DELETE CASCADE,
    CONSTRAINT fk_credit_records_applicant FOREIGN KEY (applicant_id) REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT fk_credit_records_approver FOREIGN KEY (approver_id) REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT fk_credit_records_assigned_approver FOREIGN KEY (assigned_approver_id) REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT fk_credit_records_repay_order FOREIGN KEY (repay_order_id) REFERENCES orders(id) ON DELETE SET NULL
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_credit_records_tenant ON credit_records(tenant_id);
CREATE INDEX IF NOT EXISTS idx_credit_records_customer ON credit_records(customer_id);
CREATE INDEX IF NOT EXISTS idx_credit_records_type ON credit_records(record_type);
CREATE INDEX IF NOT EXISTS idx_credit_records_status ON credit_records(approval_status);
CREATE INDEX IF NOT EXISTS idx_credit_records_approver ON credit_records(approver_id);
CREATE INDEX IF NOT EXISTS idx_credit_records_assigned ON credit_records(assigned_approver_id);
CREATE INDEX IF NOT EXISTS idx_credit_records_applicant ON credit_records(applicant_id);
CREATE INDEX IF NOT EXISTS idx_credit_records_created ON credit_records(created_at);
CREATE INDEX IF NOT EXISTS idx_credit_records_deleted ON credit_records(deleted_at);

-- 添加注释
COMMENT ON TABLE credit_records IS '授信记录表';
COMMENT ON COLUMN credit_records.record_type IS '记录类型：grant-授信, adjust-调整, repay-还款, consume-消费';
COMMENT ON COLUMN credit_records.approval_status IS '审批状态：pending-待审批, approved-已批准, rejected-已拒绝, cancelled-已取消';

-- =============================================================================
-- 第六部分：扩展 orders 表
-- =============================================================================

-- 添加返点和支付账户相关字段
ALTER TABLE orders ADD COLUMN IF NOT EXISTS payment_account_id BIGINT;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS payment_type VARCHAR(50);  -- public（对公）, private（对私）
ALTER TABLE orders ADD COLUMN IF NOT EXISTS rebate_rate DECIMAL(5,2);
ALTER TABLE orders ADD COLUMN IF NOT EXISTS rebate_amount DECIMAL(15,2);
ALTER TABLE orders ADD COLUMN IF NOT EXISTS cash_amount DECIMAL(15,2);
ALTER TABLE orders ADD COLUMN IF NOT EXISTS gift_amount DECIMAL(15,2);

-- 添加外键约束
ALTER TABLE orders ADD CONSTRAINT IF NOT EXISTS fk_orders_payment_account
    FOREIGN KEY (payment_account_id) REFERENCES payment_accounts(id) ON DELETE SET NULL;

-- 添加索引
CREATE INDEX IF NOT EXISTS idx_orders_payment_account ON orders(payment_account_id);
CREATE INDEX IF NOT EXISTS idx_orders_payment_type ON orders(payment_type);

-- 添加注释
COMMENT ON COLUMN orders.payment_account_id IS '收款账户ID';
COMMENT ON COLUMN orders.payment_type IS '支付类型：public-对公, private-对私';
COMMENT ON COLUMN orders.rebate_rate IS '返点比例(%)';
COMMENT ON COLUMN orders.rebate_amount IS '返点总金额';
COMMENT ON COLUMN orders.cash_amount IS '返点现金金额';
COMMENT ON COLUMN orders.gift_amount IS '返点赠款金额';

-- =============================================================================
-- 第七部分：创建 import_logs 表（如果不存在）
-- =============================================================================

CREATE TABLE IF NOT EXISTS import_logs (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,

    -- 导入信息
    import_type VARCHAR(50) NOT NULL,  -- customer, order, payment_account等
    file_name VARCHAR(255),

    -- 导入统计
    total_rows INT DEFAULT 0,
    success_rows INT DEFAULT 0,
    failed_rows INT DEFAULT 0,

    -- 错误详情
    error_details JSONB,

    -- 状态
    status VARCHAR(20) DEFAULT 'processing',  -- processing, completed, failed

    -- 导入人
    imported_by BIGINT,

    -- 时间戳
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,

    -- 外键约束
    CONSTRAINT fk_import_logs_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    CONSTRAINT fk_import_logs_importer FOREIGN KEY (imported_by) REFERENCES users(id) ON DELETE SET NULL
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_import_logs_tenant ON import_logs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_import_logs_type ON import_logs(import_type);
CREATE INDEX IF NOT EXISTS idx_import_logs_status ON import_logs(status);
CREATE INDEX IF NOT EXISTS idx_import_logs_importer ON import_logs(imported_by);
CREATE INDEX IF NOT EXISTS idx_import_logs_created ON import_logs(created_at);

-- 添加注释
COMMENT ON TABLE import_logs IS '导入日志表';
COMMENT ON COLUMN import_logs.import_type IS '导入类型：customer-客户, order-订单, payment_account-收款账户等';

-- =============================================================================
-- 完成
-- =============================================================================
