package model

import "time"

// Base 通用字段，所有模型继承
type Base struct {
	ID        uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// PageReq 分页请求
type PageReq struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

func (p *PageReq) Normalize() {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.PageSize <= 0 || p.PageSize > 100 {
		p.PageSize = 20
	}
}

func (p *PageReq) Offset() int { return (p.Page - 1) * p.PageSize }
