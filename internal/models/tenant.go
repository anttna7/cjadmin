package models

import (
	"database/sql/driver"
	"encoding/json"
)

// Tenant 租户表
type Tenant struct {
	BaseModel
	Name     string         `gorm:"size:255;not null" json:"name"`
	Code     string         `gorm:"size:100;unique;not null" json:"code"`
	Status   string         `gorm:"size:20;default:active" json:"status"` // active, suspended, inactive
	Settings JSONB          `gorm:"type:jsonb" json:"settings,omitempty"`
}

// Department 部门表
type Department struct {
	TenantModel
	ParentID *int64 `gorm:"index" json:"parent_id"`
	Name     string `gorm:"size:255;not null" json:"name"`
	Code     string `gorm:"size:100" json:"code"`
	Level    int    `gorm:"default:1" json:"level"`
	Path     string `gorm:"size:500" json:"path"` // 部门路径，如 /1/2/3/

	// 关联
	Parent   *Department  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children []Department `gorm:"foreignKey:ParentID" json:"children,omitempty"`
}

// JSONB 自定义类型，用于PostgreSQL的JSONB字段
type JSONB map[string]interface{}

// Scan 实现 sql.Scanner 接口
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = make(map[string]interface{})
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

// Value 实现 driver.Valuer 接口
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

func (Tenant) TableName() string {
	return "tenants"
}

func (Department) TableName() string {
	return "departments"
}
