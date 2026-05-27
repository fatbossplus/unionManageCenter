package model

import "time"

type User struct {
	Base
	Username    string     `gorm:"uniqueIndex;size:64"   json:"username"`
	Password    string     `gorm:"size:128"              json:"-"`
	Email       *string    `gorm:"uniqueIndex;size:128"  json:"email"`
	Phone       string     `gorm:"size:20"               json:"phone"`
	RealName    string     `gorm:"size:64"               json:"real_name"`
	Avatar      string     `gorm:"size:256"              json:"avatar"`
	Status      int8       `gorm:"default:1"             json:"status"`
	CertStatus  int8       `gorm:"default:0"             json:"cert_status"`
	Source      string     `gorm:"size:32;default:'web'" json:"source"`
	LastLoginAt *time.Time `                             json:"last_login_at"`
	LastLoginIP string     `gorm:"size:64"               json:"last_login_ip"`
	Roles       []Role     `gorm:"many2many:user_roles;" json:"roles,omitempty"`
}

func (User) TableName() string { return "users" }

type Role struct {
	Base
	Name        string       `gorm:"size:64"            json:"name"`
	Code        string       `gorm:"uniqueIndex;size:64" json:"code"`
	Description string       `gorm:"size:256"           json:"description"`
	Status      int8         `gorm:"default:1"          json:"status"`
	Sort        int          `gorm:"default:0"          json:"sort"`
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
}

func (Role) TableName() string { return "roles" }

type Permission struct {
	Base
	ParentID  uint64       `gorm:"default:0"   json:"parent_id"`
	Name      string       `gorm:"size:64"     json:"name"`
	Code      string       `gorm:"size:128"    json:"code"`
	Type      int8         `gorm:"default:1"   json:"type"`
	Path      string       `gorm:"size:256"    json:"path"`
	Method    string       `gorm:"size:16"     json:"method"`
	Icon      string       `gorm:"size:64"     json:"icon"`
	Component string       `gorm:"size:128"    json:"component"`
	Sort      int          `gorm:"default:0"   json:"sort"`
	Visible   int8         `gorm:"default:1"   json:"visible"`
	Status    int8         `gorm:"default:1"   json:"status"`
	Children  []Permission `gorm:"-"           json:"children,omitempty"`
}

func (Permission) TableName() string { return "permissions" }

type UserRole struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement"`
	UserID    uint64    `gorm:"index"`
	RoleID    uint64    `gorm:"index"`
	CreatedAt time.Time
}

func (UserRole) TableName() string { return "user_roles" }

type RolePermission struct {
	RoleID       uint64 `gorm:"primaryKey;column:role_id"`
	PermissionID uint64 `gorm:"primaryKey;column:permission_id"`
}

func (RolePermission) TableName() string { return "role_permissions" }
