package model

import "time"

// PublishTask 发布任务
type PublishTask struct {
	Base
	OrgID          uint64      `gorm:"column:org_id;index"           json:"org_id"`
	RawID          uint64      `gorm:"column:raw_id"                 json:"raw_id"`
	AccountID      uint64      `gorm:"column:account_id;index"       json:"account_id"`
	TargetPlatform string      `gorm:"column:target_platform;size:20" json:"target_platform"`
	FinalTitle     string      `gorm:"column:final_title;size:512"   json:"final_title"`
	FinalText      string      `gorm:"column:final_text;type:longtext" json:"final_text"`
	FinalImages    StringSlice `gorm:"column:final_images;type:json" json:"final_images"`
	FinalTags      StringSlice `gorm:"column:final_tags;type:json"   json:"final_tags"`
	Status         string      `gorm:"column:status;size:20;default:draft" json:"status"`
	ReviewedBy     uint64      `gorm:"column:reviewed_by"            json:"reviewed_by"`
	ReviewedAt     *time.Time  `gorm:"column:reviewed_at"            json:"reviewed_at"`
	ScheduledAt    *time.Time  `gorm:"column:scheduled_at"           json:"scheduled_at"`
	PublishedAt    *time.Time  `gorm:"column:published_at"           json:"published_at"`
	PlatformPostID string      `gorm:"column:platform_post_id;size:256" json:"platform_post_id"`
	FailureReason  string      `gorm:"column:failure_reason;type:text" json:"failure_reason,omitempty"`
	RetryCount     int8        `gorm:"column:retry_count"            json:"retry_count"`
	DeletedAt      *time.Time  `gorm:"column:deleted_at;index"       json:"-"`
}

func (PublishTask) TableName() string { return "cms_publish_tasks" }

// PublishStatus 常量
const (
	PublishStatusDraft     = "draft"
	PublishStatusReviewing = "reviewing"
	PublishStatusApproved  = "approved"
	PublishStatusScheduled = "scheduled"
	PublishStatusPublished = "published"
	PublishStatusFailed    = "failed"
)
