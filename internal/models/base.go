package models

import (
	"time"
	"gorm.io/gorm"
)

// BaseModel 基础模型
type BaseModel struct {
	ID        int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// TenantModel 租户基础模型
type TenantModel struct {
	BaseModel
	TenantID int64 `gorm:"index;not null" json:"tenant_id"`
}
