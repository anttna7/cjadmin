-- 为审计日志表添加字段级变更追踪字段
-- Migration: 009_add_changed_fields_to_audit_logs
-- Date: 2025-11-17

ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS changed_fields JSONB;

-- 为 changed_fields 字段创建 GIN 索引以提高查询性能
CREATE INDEX IF NOT EXISTS idx_audit_logs_changed_fields ON audit_logs USING GIN (changed_fields);

-- 添加注释
COMMENT ON COLUMN audit_logs.changed_fields IS '字段级变更追踪：{"field_name": {"before": "old_value", "after": "new_value"}}';
