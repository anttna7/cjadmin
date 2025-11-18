-- 第21阶段：登录安全增强
-- 包含：短信验证码、第三方OAuth绑定、黑白名单

-- 短信验证码表
CREATE TABLE IF NOT EXISTS sms_codes (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT REFERENCES tenants(id),

    -- 验证码信息
    phone VARCHAR(20) NOT NULL,
    code VARCHAR(10) NOT NULL,
    code_type VARCHAR(50) NOT NULL DEFAULT 'login', -- login, register, reset_password, bindphone

    -- 状态
    is_used BOOLEAN DEFAULT false,
    used_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,

    -- IP信息
    ip_address VARCHAR(50),
    user_agent TEXT,

    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_sms_codes_phone ON sms_codes(phone);
CREATE INDEX idx_sms_codes_code ON sms_codes(code);
CREATE INDEX idx_sms_codes_expires ON sms_codes(expires_at);
CREATE INDEX idx_sms_codes_tenant ON sms_codes(tenant_id);

-- 第三方OAuth绑定表
CREATE TABLE IF NOT EXISTS oauth_bindings (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES tenants(id),
    user_id BIGINT NOT NULL REFERENCES users(id),

    -- OAuth提供商信息
    provider VARCHAR(50) NOT NULL, -- feishu, wechat, dingtalk, github
    provider_user_id VARCHAR(200) NOT NULL, -- 第三方平台用户ID
    provider_username VARCHAR(200), -- 第三方平台用户名
    provider_avatar VARCHAR(500), -- 头像URL
    provider_email VARCHAR(200),

    -- Token信息（加密存储）
    access_token TEXT,
    refresh_token TEXT,
    token_expires_at TIMESTAMP WITH TIME ZONE,

    -- 额外信息
    extra_data JSONB,

    -- 状态
    is_active BOOLEAN DEFAULT true,

    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,

    -- 唯一约束：同一租户下，同一提供商的用户ID只能绑定一次
    CONSTRAINT uk_oauth_binding UNIQUE(tenant_id, provider, provider_user_id)
);

CREATE INDEX idx_oauth_bindings_user ON oauth_bindings(user_id);
CREATE INDEX idx_oauth_bindings_provider ON oauth_bindings(provider);
CREATE INDEX idx_oauth_bindings_tenant ON oauth_bindings(tenant_id);

-- 登录黑名单表
CREATE TABLE IF NOT EXISTS login_blacklist (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT REFERENCES tenants(id), -- NULL表示全局规则

    -- 规则类型
    rule_type VARCHAR(50) NOT NULL, -- ip, ip_range, user, device, phone
    rule_value VARCHAR(500) NOT NULL, -- IP地址、用户ID、设备指纹等

    -- 规则描述
    reason VARCHAR(500),

    -- 有效期
    expires_at TIMESTAMP WITH TIME ZONE, -- NULL表示永久

    -- 操作信息
    created_by BIGINT REFERENCES users(id),

    -- 状态
    is_active BOOLEAN DEFAULT true,

    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_login_blacklist_type ON login_blacklist(rule_type);
CREATE INDEX idx_login_blacklist_value ON login_blacklist(rule_value);
CREATE INDEX idx_login_blacklist_tenant ON login_blacklist(tenant_id);
CREATE INDEX idx_login_blacklist_expires ON login_blacklist(expires_at);

-- 登录白名单表
CREATE TABLE IF NOT EXISTS login_whitelist (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT REFERENCES tenants(id), -- NULL表示全局规则

    -- 规则类型
    rule_type VARCHAR(50) NOT NULL, -- ip, ip_range, user, device
    rule_value VARCHAR(500) NOT NULL,

    -- 规则描述
    description VARCHAR(500),

    -- 操作信息
    created_by BIGINT REFERENCES users(id),

    -- 状态
    is_active BOOLEAN DEFAULT true,

    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_login_whitelist_type ON login_whitelist(rule_type);
CREATE INDEX idx_login_whitelist_value ON login_whitelist(rule_value);
CREATE INDEX idx_login_whitelist_tenant ON login_whitelist(tenant_id);

-- 登录尝试记录表（用于限制登录频率）
CREATE TABLE IF NOT EXISTS login_attempts (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT REFERENCES tenants(id),

    -- 尝试信息
    identifier VARCHAR(200) NOT NULL, -- 用户名、手机号或IP
    identifier_type VARCHAR(50) NOT NULL, -- username, phone, ip

    -- 结果
    is_success BOOLEAN DEFAULT false,
    failure_reason VARCHAR(200),

    -- 客户端信息
    ip_address VARCHAR(50),
    user_agent TEXT,
    device_fingerprint VARCHAR(100),

    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_login_attempts_identifier ON login_attempts(identifier);
CREATE INDEX idx_login_attempts_ip ON login_attempts(ip_address);
CREATE INDEX idx_login_attempts_created ON login_attempts(created_at);
CREATE INDEX idx_login_attempts_tenant ON login_attempts(tenant_id);

-- OAuth配置表（存储租户的OAuth应用配置）
CREATE TABLE IF NOT EXISTS oauth_configs (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES tenants(id),

    -- 提供商信息
    provider VARCHAR(50) NOT NULL, -- feishu, wechat, dingtalk

    -- 应用配置
    app_id VARCHAR(200) NOT NULL,
    app_secret TEXT NOT NULL, -- 加密存储

    -- 回调地址
    redirect_uri VARCHAR(500),

    -- 额外配置
    extra_config JSONB,

    -- 状态
    is_active BOOLEAN DEFAULT true,

    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,

    -- 唯一约束
    CONSTRAINT uk_oauth_config UNIQUE(tenant_id, provider)
);

CREATE INDEX idx_oauth_configs_tenant ON oauth_configs(tenant_id);
CREATE INDEX idx_oauth_configs_provider ON oauth_configs(provider);

-- 扩展用户表，添加手机号字段（如果不存在）
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                   WHERE table_name = 'users' AND column_name = 'phone') THEN
        ALTER TABLE users ADD COLUMN phone VARCHAR(20);
        CREATE INDEX idx_users_phone ON users(phone);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                   WHERE table_name = 'users' AND column_name = 'login_method') THEN
        ALTER TABLE users ADD COLUMN login_method VARCHAR(50) DEFAULT 'password'; -- password, phone, oauth
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                   WHERE table_name = 'users' AND column_name = 'last_login_ip') THEN
        ALTER TABLE users ADD COLUMN last_login_ip VARCHAR(50);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                   WHERE table_name = 'users' AND column_name = 'last_login_at') THEN
        ALTER TABLE users ADD COLUMN last_login_at TIMESTAMP WITH TIME ZONE;
    END IF;
END $$;

-- 添加安全相关权限
INSERT INTO permissions (name, code, description, module, created_at, updated_at) VALUES
    ('查看黑名单', 'security.blacklist.view', '查看登录黑名单列表', 'security', NOW(), NOW()),
    ('管理黑名单', 'security.blacklist.manage', '添加、编辑、删除黑名单规则', 'security', NOW(), NOW()),
    ('查看白名单', 'security.whitelist.view', '查看登录白名单列表', 'security', NOW(), NOW()),
    ('管理白名单', 'security.whitelist.manage', '添加、编辑、删除白名单规则', 'security', NOW(), NOW()),
    ('查看OAuth绑定', 'oauth.bindings.view', '查看用户OAuth绑定信息', 'security', NOW(), NOW()),
    ('管理OAuth绑定', 'oauth.bindings.manage', '管理用户OAuth绑定', 'security', NOW(), NOW()),
    ('查看OAuth配置', 'oauth.config.view', '查看租户OAuth应用配置', 'security', NOW(), NOW()),
    ('管理OAuth配置', 'oauth.config.manage', '配置租户OAuth应用', 'security', NOW(), NOW()),
    ('查看登录日志', 'security.loginlog.view', '查看登录尝试记录', 'security', NOW(), NOW())
ON CONFLICT (code) DO NOTHING;

-- 添加注释
COMMENT ON TABLE sms_codes IS '短信验证码表';
COMMENT ON TABLE oauth_bindings IS '第三方OAuth账号绑定表';
COMMENT ON TABLE login_blacklist IS '登录黑名单表';
COMMENT ON TABLE login_whitelist IS '登录白名单表';
COMMENT ON TABLE login_attempts IS '登录尝试记录表';
COMMENT ON TABLE oauth_configs IS 'OAuth应用配置表';
