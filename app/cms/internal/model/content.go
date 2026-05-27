package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// StringSlice JSON 序列化字符串切片（用于图片/标签列表）
type StringSlice []string

func (s StringSlice) Value() (driver.Value, error) {
	if s == nil {
		return "[]", nil
	}
	b, err := json.Marshal(s)
	return string(b), err
}

func (s *StringSlice) Scan(val any) error {
	if val == nil {
		*s = []string{}
		return nil
	}
	var bs []byte
	switch v := val.(type) {
	case []byte:
		bs = v
	case string:
		bs = []byte(v)
	default:
		return fmt.Errorf("unsupported type: %T", val)
	}
	return json.Unmarshal(bs, s)
}

// RawContent 原始内容（采集结果）
type RawContent struct {
	Base
	TaskID     uint64      `gorm:"column:task_id;index"       json:"task_id"`
	OrgID      uint64      `gorm:"column:org_id"              json:"org_id"`
	Platform   string      `gorm:"column:platform;size:20"    json:"platform"`
	SourceURL  string      `gorm:"column:source_url;size:1024" json:"source_url"`
	SourceHash string      `gorm:"column:source_hash;size:64;uniqueIndex" json:"source_hash"`
	Title      string      `gorm:"column:title;size:512"      json:"title"`
	Author     string      `gorm:"column:author;size:128"     json:"author"`
	BodyText   string      `gorm:"column:body_text;type:longtext" json:"body_text"`
	BodyImages StringSlice `gorm:"column:body_images;type:json" json:"body_images"`
	FetchedAt  time.Time   `gorm:"column:fetched_at"          json:"fetched_at"`
	ProcStatus string      `gorm:"column:proc_status;size:20;default:pending" json:"proc_status"`
	ProcError  string      `gorm:"column:proc_error;size:512" json:"proc_error"`
}

func (RawContent) TableName() string { return "cms_raw_contents" }

// ProcStatus 常量
const (
	ProcStatusPending    = "pending"
	ProcStatusProcessing = "processing"
	ProcStatusDone       = "done"
	ProcStatusFailed     = "failed"
	ProcStatusSkipped    = "skipped"
)

// ProcessRecord AI 流水线处理记录
type ProcessRecord struct {
	ID         uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	RawID      uint64          `gorm:"column:raw_id;index:idx_raw_round" json:"raw_id"`
	Round      int             `gorm:"column:round;index:idx_raw_round"  json:"round"`
	RoundName  string          `gorm:"column:round_name;size:32"  json:"round_name"`
	InputHash  string          `gorm:"column:input_hash;size:64"  json:"input_hash"`
	InputText  string          `gorm:"column:input_text;type:longtext" json:"-"`
	OutputText string          `gorm:"column:output_text;type:longtext" json:"output_text"`
	DriverUsed string          `gorm:"column:driver_used;size:64" json:"driver_used"`
	ModelUsed  string          `gorm:"column:model_used;size:64"  json:"model_used"`
	Result     string          `gorm:"column:result;size:20"      json:"result"` // pass|fail|retry
	ScoreJSON  *ScoreResult    `gorm:"column:score_json;type:json;serializer:json" json:"score_json,omitempty"`
	IssuesJSON StringSlice     `gorm:"column:issues_json;type:json" json:"issues_json,omitempty"`
	RetryCount int             `gorm:"column:retry_count"         json:"retry_count"`
	DurationMs int             `gorm:"column:duration_ms"         json:"duration_ms"`
	CreatedAt  time.Time       `gorm:"column:created_at"          json:"created_at"`
}

func (ProcessRecord) TableName() string { return "cms_process_records" }

// ScoreResult 第4轮 AI 自审评分结果
type ScoreResult struct {
	Scores   map[string]int `json:"scores"`   // 各维度评分
	Passed   bool           `json:"passed"`
	Comment  string         `json:"comment"`
}

// Round 常量
const (
	RoundClean      = 1
	RoundScan       = 2
	RoundRewrite    = 3
	RoundSelfReview = 4
	RoundFormat     = 5
)

var RoundNames = map[int]string{
	RoundClean:      "clean",
	RoundScan:       "scan",
	RoundRewrite:    "rewrite",
	RoundSelfReview: "self_review",
	RoundFormat:     "format",
}

// ProcessResult 常量
const (
	ProcResultPass    = "pass"
	ProcResultFail    = "fail"
	ProcResultRetry   = "retry"
	ProcResultPending = "pending"
)
