package models

// User 用户表
type User struct {
	BaseModel
	TenantID         *int64 `gorm:"index" json:"tenant_id"`
	DepartmentID     *int64 `gorm:"index" json:"department_id"`
	Username         string `gorm:"size:100;unique;not null" json:"username"`
	PasswordHash     string `gorm:"size:255;not null" json:"-"`
	RealName         string `gorm:"size:100" json:"real_name"`
	Email            string `gorm:"size:255" json:"email"`
	Phone            string `gorm:"size:50" json:"phone"`
	Status           string `gorm:"size:20;default:active" json:"status"`
	IsPlatformAdmin  bool   `gorm:"default:false" json:"is_platform_admin"`

	// 关联
	Tenant     *Tenant     `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
	Department *Department `gorm:"foreignKey:DepartmentID" json:"department,omitempty"`
	Roles      []Role      `gorm:"many2many:user_roles" json:"roles,omitempty"`
}

// Role 角色表
type Role struct {
	BaseModel
	TenantID       *int64 `gorm:"index" json:"tenant_id"`
	Name           string `gorm:"size:100;not null" json:"name"`
	Code           string `gorm:"size:100;not null" json:"code"`
	Description    string `gorm:"type:text" json:"description"`
	IsPlatformRole bool   `gorm:"default:false" json:"is_platform_role"`
	CreatedBy      *int64 `gorm:"index" json:"created_by"`

	// 关联
	Tenant      *Tenant      `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
	Permissions []Permission `gorm:"many2many:role_permissions" json:"permissions,omitempty"`
	Creator     *User        `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

// Permission 权限表
type Permission struct {
	ID          int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Code        string `gorm:"size:100;unique;not null" json:"code"`
	Name        string `gorm:"size:255;not null" json:"name"`
	Resource    string `gorm:"size:100" json:"resource"` // customer, order, finance等
	Action      string `gorm:"size:50" json:"action"`    // create, read, update, delete, import, export
	Description string `gorm:"type:text" json:"description"`
	CreatedAt   int64  `gorm:"autoCreateTime" json:"created_at"`
}

// UserRole 用户角色关联表
type UserRole struct {
	ID        int64 `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    int64 `gorm:"not null;index" json:"user_id"`
	RoleID    int64 `gorm:"not null;index" json:"role_id"`
	CreatedAt int64 `gorm:"autoCreateTime" json:"created_at"`
}

// RolePermission 角色权限关联表
type RolePermission struct {
	ID           int64 `gorm:"primaryKey;autoIncrement" json:"id"`
	RoleID       int64 `gorm:"not null;index" json:"role_id"`
	PermissionID int64 `gorm:"not null;index" json:"permission_id"`
	CreatedAt    int64 `gorm:"autoCreateTime" json:"created_at"`
}

// DepartmentPermission 部门权限表
type DepartmentPermission struct {
	ID           int64 `gorm:"primaryKey;autoIncrement" json:"id"`
	DepartmentID int64 `gorm:"not null;index" json:"department_id"`
	PermissionID int64 `gorm:"not null;index" json:"permission_id"`
	CreatedAt    int64 `gorm:"autoCreateTime" json:"created_at"`
}

func (User) TableName() string {
	return "users"
}

func (Role) TableName() string {
	return "roles"
}

func (Permission) TableName() string {
	return "permissions"
}

func (UserRole) TableName() string {
	return "user_roles"
}

func (RolePermission) TableName() string {
	return "role_permissions"
}

func (DepartmentPermission) TableName() string {
	return "department_permissions"
}
