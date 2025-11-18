-- 第20阶段：活动营销与数据分析
-- 创建时间：2025-11-18

-- =====================================================
-- 1. 活动管理表
-- =====================================================
CREATE TABLE IF NOT EXISTS activities (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES tenants(id),

    -- 基本信息
    name VARCHAR(200) NOT NULL,
    description TEXT,
    activity_type VARCHAR(50) NOT NULL DEFAULT 'promotion',  -- promotion, discount, gift, lottery

    -- 展示配置
    display_position VARCHAR(50) DEFAULT 'recharge',  -- login, recharge, dashboard, popup
    image_url VARCHAR(500),
    link_url VARCHAR(500),
    content JSONB,  -- 活动详细内容配置

    -- 时间控制
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,

    -- 参与规则
    rules JSONB,  -- 参与条件、奖励规则等
    max_participants INT DEFAULT 0,  -- 0表示不限制

    -- 状态
    is_active BOOLEAN DEFAULT true,
    sort_order INT DEFAULT 0,

    -- 统计
    view_count INT DEFAULT 0,
    participant_count INT DEFAULT 0,

    -- 审计
    created_by BIGINT REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_activities_tenant ON activities(tenant_id);
CREATE INDEX idx_activities_type ON activities(activity_type);
CREATE INDEX idx_activities_position ON activities(display_position);
CREATE INDEX idx_activities_time ON activities(start_time, end_time);
CREATE INDEX idx_activities_active ON activities(is_active) WHERE deleted_at IS NULL;

-- =====================================================
-- 2. 活动参与记录表
-- =====================================================
CREATE TABLE IF NOT EXISTS activity_participants (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES tenants(id),
    activity_id BIGINT NOT NULL REFERENCES activities(id),
    customer_id BIGINT NOT NULL REFERENCES customers(id),

    -- 参与信息
    participated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    participation_data JSONB,  -- 参与时提交的数据

    -- 奖励信息
    reward_type VARCHAR(50),  -- cash, gift, coupon, points
    reward_amount DECIMAL(15,2) DEFAULT 0,
    reward_status VARCHAR(50) DEFAULT 'pending',  -- pending, granted, expired
    reward_granted_at TIMESTAMP WITH TIME ZONE,

    -- 备注
    remark TEXT,

    UNIQUE(activity_id, customer_id)
);

CREATE INDEX idx_activity_participants_tenant ON activity_participants(tenant_id);
CREATE INDEX idx_activity_participants_activity ON activity_participants(activity_id);
CREATE INDEX idx_activity_participants_customer ON activity_participants(customer_id);

-- =====================================================
-- 3. 横幅广告表
-- =====================================================
CREATE TABLE IF NOT EXISTS banners (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES tenants(id),

    -- 基本信息
    title VARCHAR(200) NOT NULL,
    description TEXT,

    -- 展示内容
    image_url VARCHAR(500) NOT NULL,
    link_url VARCHAR(500),
    link_target VARCHAR(20) DEFAULT '_blank',  -- _blank, _self

    -- 展示位置
    position VARCHAR(50) NOT NULL DEFAULT 'top',  -- top, bottom, sidebar, popup, login, recharge

    -- 时间控制
    start_time TIMESTAMP WITH TIME ZONE,
    end_time TIMESTAMP WITH TIME ZONE,

    -- 状态
    is_active BOOLEAN DEFAULT true,
    sort_order INT DEFAULT 0,

    -- 统计
    view_count INT DEFAULT 0,
    click_count INT DEFAULT 0,

    -- 审计
    created_by BIGINT REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_banners_tenant ON banners(tenant_id);
CREATE INDEX idx_banners_position ON banners(position);
CREATE INDEX idx_banners_time ON banners(start_time, end_time);
CREATE INDEX idx_banners_active ON banners(is_active) WHERE deleted_at IS NULL;

-- =====================================================
-- 4. 主题配置表
-- =====================================================
CREATE TABLE IF NOT EXISTS themes (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES tenants(id),

    -- 基本信息
    name VARCHAR(100) NOT NULL,
    description TEXT,
    theme_type VARCHAR(50) DEFAULT 'custom',  -- default, festival, custom
    festival_type VARCHAR(50),  -- new_year, spring_festival, christmas, etc.

    -- 颜色配置
    primary_color VARCHAR(20) DEFAULT '#1890ff',
    secondary_color VARCHAR(20) DEFAULT '#52c41a',
    background_color VARCHAR(20) DEFAULT '#f0f2f5',
    text_color VARCHAR(20) DEFAULT '#333333',

    -- 图片资源
    logo_url VARCHAR(500),
    favicon_url VARCHAR(500),
    background_image_url VARCHAR(500),
    login_background_url VARCHAR(500),

    -- 其他配置
    custom_css TEXT,
    extra_config JSONB,

    -- 状态
    is_default BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,

    -- 审计
    created_by BIGINT REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_themes_tenant ON themes(tenant_id);
CREATE INDEX idx_themes_type ON themes(theme_type);
CREATE INDEX idx_themes_default ON themes(is_default) WHERE deleted_at IS NULL;

-- =====================================================
-- 5. 自定义报表定义表
-- =====================================================
CREATE TABLE IF NOT EXISTS custom_reports (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES tenants(id),

    -- 基本信息
    name VARCHAR(200) NOT NULL,
    description TEXT,
    category VARCHAR(50) DEFAULT 'general',  -- general, finance, customer, order

    -- 数据源配置
    base_table VARCHAR(100) NOT NULL,  -- customers, orders, customer_accounts, etc.
    select_fields JSONB NOT NULL,  -- [{field: "customer_name", alias: "客户名称", aggregate: null}]
    join_config JSONB,  -- [{table: "customer_accounts", on: "customers.id = customer_accounts.customer_id", type: "LEFT"}]

    -- 查询配置
    filter_config JSONB,  -- [{field: "status", operator: "=", value: null, is_parameter: true}]
    group_by_fields JSONB,  -- ["customer_id", "order_type"]
    order_by_config JSONB,  -- [{field: "created_at", direction: "DESC"}]

    -- 显示配置
    column_config JSONB,  -- 列宽、格式化等
    chart_config JSONB,  -- 图表类型、配置

    -- 状态
    is_public BOOLEAN DEFAULT false,  -- 是否公开给其他用户
    is_active BOOLEAN DEFAULT true,

    -- 审计
    created_by BIGINT REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_custom_reports_tenant ON custom_reports(tenant_id);
CREATE INDEX idx_custom_reports_category ON custom_reports(category);
CREATE INDEX idx_custom_reports_creator ON custom_reports(created_by);

-- =====================================================
-- 6. 报表执行记录表
-- =====================================================
CREATE TABLE IF NOT EXISTS report_executions (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES tenants(id),
    report_id BIGINT NOT NULL REFERENCES custom_reports(id),

    -- 执行信息
    executed_by BIGINT REFERENCES users(id),
    executed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    -- 参数和结果
    parameters JSONB,  -- 执行时的参数
    row_count INT DEFAULT 0,
    execution_time_ms INT DEFAULT 0,

    -- 导出信息
    export_format VARCHAR(20),  -- excel, csv
    export_file_url VARCHAR(500),

    -- 状态
    status VARCHAR(50) DEFAULT 'completed'  -- running, completed, failed
);

CREATE INDEX idx_report_executions_tenant ON report_executions(tenant_id);
CREATE INDEX idx_report_executions_report ON report_executions(report_id);
CREATE INDEX idx_report_executions_user ON report_executions(executed_by);
CREATE INDEX idx_report_executions_time ON report_executions(executed_at);

-- =====================================================
-- 7. 添加权限
-- =====================================================

-- 活动管理权限
INSERT INTO permissions (code, name, description, module, created_at, updated_at) VALUES
('activity.view', '查看活动', '查看活动列表和详情', 'marketing', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('activity.create', '创建活动', '创建新活动', 'marketing', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('activity.update', '编辑活动', '编辑活动信息', 'marketing', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('activity.delete', '删除活动', '删除活动', 'marketing', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (code) DO NOTHING;

-- 横幅管理权限
INSERT INTO permissions (code, name, description, module, created_at, updated_at) VALUES
('banner.view', '查看横幅', '查看横幅列表和详情', 'marketing', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('banner.create', '创建横幅', '创建新横幅', 'marketing', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('banner.update', '编辑横幅', '编辑横幅信息', 'marketing', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('banner.delete', '删除横幅', '删除横幅', 'marketing', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (code) DO NOTHING;

-- 主题管理权限
INSERT INTO permissions (code, name, description, module, created_at, updated_at) VALUES
('theme.view', '查看主题', '查看主题列表和详情', 'marketing', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('theme.create', '创建主题', '创建新主题', 'marketing', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('theme.update', '编辑主题', '编辑主题信息', 'marketing', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('theme.delete', '删除主题', '删除主题', 'marketing', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('theme.switch', '切换主题', '切换当前使用的主题', 'marketing', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (code) DO NOTHING;

-- 报表管理权限
INSERT INTO permissions (code, name, description, module, created_at, updated_at) VALUES
('report.view', '查看报表', '查看报表列表和详情', 'report', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('report.create', '创建报表', '创建新报表', 'report', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('report.update', '编辑报表', '编辑报表定义', 'report', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('report.delete', '删除报表', '删除报表', 'report', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('report.execute', '执行报表', '执行报表查询', 'report', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('report.export', '导出报表', '导出报表数据', 'report', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (code) DO NOTHING;

-- =====================================================
-- 8. 插入预设节日主题模板（平台级别）
-- =====================================================
-- 注意：这些是平台级别的模板，tenant_id为0或NULL表示平台模板
-- 租户可以基于这些模板创建自己的主题

COMMENT ON TABLE activities IS '活动管理表 - 存储营销活动信息，支持多种活动类型和展示位置';
COMMENT ON TABLE activity_participants IS '活动参与记录表 - 记录客户参与活动的信息和奖励';
COMMENT ON TABLE banners IS '横幅广告表 - 管理各位置的横幅广告展示';
COMMENT ON TABLE themes IS '主题配置表 - 管理系统界面主题，支持节日主题切换';
COMMENT ON TABLE custom_reports IS '自定义报表定义表 - 存储用户自定义的报表查询配置';
COMMENT ON TABLE report_executions IS '报表执行记录表 - 记录报表执行历史和导出记录';
