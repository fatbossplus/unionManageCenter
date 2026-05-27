package model

import "time"

// ScrapeTask 采集任务
type ScrapeTask struct {
	Base
	OrgID          uint64     `gorm:"column:org_id;index"           json:"org_id"`
	TaskName       string     `gorm:"column:task_name;size:128"     json:"task_name"`
	Platform       string     `gorm:"column:platform;size:20"       json:"platform"`
	TaskType       string     `gorm:"column:task_type;size:20"      json:"task_type"`
	TargetParam    string     `gorm:"column:target_param;size:512"  json:"target_param"`
	TargetPlatform string     `gorm:"column:target_platform;size:20" json:"target_platform"`
	AccountID      uint64     `gorm:"column:account_id"             json:"account_id"`
	CronExpr       string     `gorm:"column:cron_expr;size:64"      json:"cron_expr"`
	FetchLimit     int        `gorm:"column:fetch_limit;default:5"  json:"fetch_limit"`
	Status         int8       `gorm:"column:status;default:1"       json:"status"`
	LastRunAt      *time.Time `gorm:"column:last_run_at"            json:"last_run_at"`
	NextRunAt      *time.Time `gorm:"column:next_run_at"            json:"next_run_at"`
	LastError      string     `gorm:"column:last_error;size:512"    json:"last_error"`
	DeletedAt      *time.Time `gorm:"column:deleted_at;index"       json:"-"`
}

func (ScrapeTask) TableName() string { return "cms_scrape_tasks" }

// TaskType 常量
const (
	TaskTypeSearchTitle  = "search_title"   // 按标题搜索
	TaskTypeFollowAuthor = "follow_author"  // 关注作者
)

// TaskStatus 常量
const (
	TaskStatusDisabled int8 = 0
	TaskStatusEnabled  int8 = 1
)
