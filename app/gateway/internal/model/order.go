package model

import "time"

type Order struct {
	Base
	OrderNo      string     `gorm:"uniqueIndex;size:64"  json:"order_no"`
	Type         string     `gorm:"size:32"              json:"type"`
	Status       int8       `gorm:"default:1"            json:"status"`
	PayMethod    string     `gorm:"size:32"              json:"pay_method"`
	Amount       float64    `gorm:"type:decimal(12,2)"   json:"amount"`
	UserID       uint64     `gorm:"index"                json:"user_id"`
	OrgID        uint64     `gorm:"index;default:0"      json:"org_id"`
	Remark       string     `gorm:"size:256"             json:"remark"`
	PaidAt       *time.Time `                            json:"paid_at"`
	RefundedAt   *time.Time `                            json:"refunded_at"`
	RefundReason string     `gorm:"size:256"             json:"refund_reason"`
	User         *User      `gorm:"foreignKey:UserID"    json:"user,omitempty"`
	Org          *Org       `gorm:"foreignKey:OrgID"     json:"org,omitempty"`
}

func (Order) TableName() string { return "orders" }
