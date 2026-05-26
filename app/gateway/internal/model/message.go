package model

import "time"

type Message struct {
	ID        uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint64     `gorm:"index;default:0"          json:"user_id"`
	Title     string     `gorm:"size:128"                 json:"title"`
	Content   string     `gorm:"type:text"                json:"content"`
	Type      string     `gorm:"size:32;default:'system'" json:"type"`
	IsRead    int8       `gorm:"default:0"                json:"is_read"`
	RefID     uint64     `gorm:"default:0"                json:"ref_id"`
	RefType   string     `gorm:"size:32"                  json:"ref_type"`
	CreatedAt time.Time  `                                json:"created_at"`
	ReadAt    *time.Time `                                json:"read_at"`
}

func (Message) TableName() string { return "messages" }
