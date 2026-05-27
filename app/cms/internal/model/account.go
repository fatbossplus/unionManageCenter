package model

import "time"

// PlatformAccount 平台绑定账号（凭证 AES-256-GCM 加密存储）
type PlatformAccount struct {
	Base
	OrgID       uint64     `gorm:"column:org_id;index"         json:"org_id"`
	Platform    string     `gorm:"column:platform;size:20"     json:"platform"`
	AccountName string     `gorm:"column:account_name;size:128" json:"account_name"`
	AccountUID  string     `gorm:"column:account_uid;size:128" json:"account_uid"`
	// 凭证密文，永不在 JSON 中暴露
	CredCipher  string     `gorm:"column:cred_cipher;type:text" json:"-"`
	CredIV      string     `gorm:"column:cred_iv;size:64"      json:"-"`
	CredVersion int        `gorm:"column:cred_version;default:1" json:"cred_version"`
	Status      int8       `gorm:"column:status;default:1"     json:"status"`
	LastUsedAt  *time.Time `gorm:"column:last_used_at"         json:"last_used_at"`
	ExpiresAt   *time.Time `gorm:"column:expires_at"           json:"expires_at"`
	Remark      string     `gorm:"column:remark;size:256"      json:"remark"`
	DeletedAt   *time.Time `gorm:"column:deleted_at;index"     json:"-"`
}

func (PlatformAccount) TableName() string { return "cms_platform_accounts" }

// CredentialAudit 凭证访问审计（只追加，不删除）
type CredentialAudit struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	AccountID  uint64    `gorm:"column:account_id;index"  json:"account_id"`
	OperatorID uint64    `gorm:"column:operator_id"       json:"operator_id"`
	Action     string    `gorm:"column:action;size:32"    json:"action"`
	Reason     string    `gorm:"column:reason;size:128"   json:"reason"`
	IP         string    `gorm:"column:ip;size:64"        json:"ip"`
	UserAgent  string    `gorm:"column:user_agent;size:256" json:"user_agent"`
	CreatedAt  time.Time `gorm:"column:created_at"        json:"created_at"`
}

func (CredentialAudit) TableName() string { return "cms_credential_audit" }

// Platform 常量
const (
	PlatformWechat   = "wechat"
	PlatformRednote  = "rednote"
	PlatformDouyin   = "douyin"
	PlatformCSdn     = "csdn"
)

// AccountStatus 常量
const (
	AccountStatusDisabled    int8 = 0
	AccountStatusNormal      int8 = 1
	AccountStatusCredInvalid int8 = 2
	AccountStatusBanned      int8 = 3
)
