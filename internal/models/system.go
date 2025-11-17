package models

import "time"

// SystemSetting 系统设置表
type SystemSetting struct {
	ID          int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID    *int64     `gorm:"index" json:"tenant_id"` // NULL表示平台级设置
	Category    string     `gorm:"size:50;not null" json:"category"` // security, notification, login
	Key         string     `gorm:"size:100;not null" json:"key"`
	Value       string     `gorm:"type:text" json:"value"`
	ValueType   string     `gorm:"size:20;default:string" json:"value_type"` // string, number, boolean, json
	Description string     `gorm:"type:text" json:"description"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updated_at"`

	// 关联
	Tenant *Tenant `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
}

// AuditLog 审计日志表
type AuditLog struct {
	ID            int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID      *int64    `gorm:"index" json:"tenant_id"`
	UserID        *int64    `gorm:"index" json:"user_id"`
	Action        string    `gorm:"size:100;not null" json:"action"`
	ResourceType  string    `gorm:"size:50" json:"resource_type"`
	ResourceID    *int64    `json:"resource_id"`
	IPAddress     string    `gorm:"size:50" json:"ip_address"`
	UserAgent     string    `gorm:"type:text" json:"user_agent"`
	RequestData   JSONB     `gorm:"type:jsonb" json:"request_data,omitempty"`
	ResponseData  JSONB     `gorm:"type:jsonb" json:"response_data,omitempty"`
	ChangedFields JSONB     `gorm:"type:jsonb" json:"changed_fields,omitempty"` // 字段级变更追踪：{"field_name": {"before": "old_value", "after": "new_value"}}
	CreatedAt     time.Time `gorm:"autoCreateTime;index:idx_audit_created" json:"created_at"`

	// 关联
	Tenant *Tenant `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
	User   *User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// FieldChange 字段变更记录结构
type FieldChange struct {
	Before interface{} `json:"before"`
	After  interface{} `json:"after"`
}

// ImportLog 导入日志表
type ImportLog struct {
	TenantModel
	ImportType   string `gorm:"size:50;not null" json:"import_type"` // customer, order等
	FileName     string `gorm:"size:255" json:"file_name"`
	TotalRows    int    `json:"total_rows"`
	SuccessRows  int    `json:"success_rows"`
	FailedRows   int    `json:"failed_rows"`
	ErrorDetails JSONB  `gorm:"type:jsonb" json:"error_details,omitempty"`
	Status       string `gorm:"size:20" json:"status"` // processing, completed, failed
	ImportedBy   *int64 `gorm:"index" json:"imported_by"`

	// 关联
	Importer *User `gorm:"foreignKey:ImportedBy" json:"importer,omitempty"`
}

func (SystemSetting) TableName() string {
	return "system_settings"
}

func (AuditLog) TableName() string {
	return "audit_logs"
}

func (ImportLog) TableName() string {
	return "import_logs"
}
