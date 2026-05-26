package model

import "time"

type Org struct {
	Base
	Name         string `gorm:"size:128"     json:"name"`
	Type         string `gorm:"size:32"      json:"type"`
	Description  string `gorm:"type:text"    json:"description"`
	Logo         string `gorm:"size:256"     json:"logo"`
	Region       string `gorm:"size:64"      json:"region"`
	LeaderID     uint64 `gorm:"default:0"    json:"leader_id"`
	ContactEmail string `gorm:"size:128"     json:"contact_email"`
	ContactPhone string `gorm:"size:20"      json:"contact_phone"`
	Status       int8   `gorm:"default:2"    json:"status"`
	MemberCount  int    `gorm:"default:0"    json:"member_count"`
	Leader       *User  `gorm:"foreignKey:LeaderID" json:"leader,omitempty"`
}

func (Org) TableName() string { return "orgs" }

// OrgMember 表无 deleted_at，使用独立结构
type OrgMember struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	OrgID     uint64    `gorm:"index"             json:"org_id"`
	UserID    uint64    `gorm:"index"             json:"user_id"`
	Role      string    `gorm:"size:32"           json:"role"`
	Status    int8      `gorm:"default:1"         json:"status"`
	User      *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (OrgMember) TableName() string { return "org_members" }
