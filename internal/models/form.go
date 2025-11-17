package models

// CustomForm 自定义表单定义
type CustomForm struct {
	BaseModel
	TenantID   *int64 `gorm:"index" json:"tenant_id"`
	FormName   string `gorm:"size:255;not null" json:"form_name"`
	FormCode   string `gorm:"size:100;not null" json:"form_code"`
	FormType   string `gorm:"size:50" json:"form_type"` // customer, order等
	FormSchema JSONB  `gorm:"type:jsonb;not null" json:"form_schema"`
	Status     string `gorm:"size:20;default:active" json:"status"`
	CreatedBy  *int64 `gorm:"index" json:"created_by"`

	// 关联
	Tenant  *Tenant `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
	Creator *User   `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

// CustomFormData 自定义表单数据
type CustomFormData struct {
	TenantModel
	FormID     int64  `gorm:"not null;index" json:"form_id"`
	EntityType string `gorm:"size:50" json:"entity_type"` // customer, order等
	EntityID   *int64 `json:"entity_id"`
	FormData   JSONB  `gorm:"type:jsonb;not null" json:"form_data"`

	// 关联
	Form *CustomForm `gorm:"foreignKey:FormID" json:"form,omitempty"`
}

func (CustomForm) TableName() string {
	return "custom_forms"
}

func (CustomFormData) TableName() string {
	return "custom_form_data"
}
