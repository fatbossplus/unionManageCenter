package model

import "time"

type FinanceAccount struct {
	Base
	OrgID       uint64 `gorm:"index"     json:"org_id"`
	Type        string `gorm:"size:32"   json:"type"`
	AccountName string `gorm:"size:128"  json:"account_name"`
	AccountNo   string `gorm:"size:128"  json:"account_no"`
	BankName    string `gorm:"size:128"  json:"bank_name"`
	IsDefault   int8   `gorm:"default:0" json:"is_default"`
	Status      int8   `gorm:"default:1" json:"status"`
}

func (FinanceAccount) TableName() string { return "finance_accounts" }

type FinanceSettlement struct {
	Base
	OrgID       uint64     `gorm:"index"                json:"org_id"`
	AccountID   uint64     `gorm:"default:0"            json:"account_id"`
	Amount      float64    `gorm:"type:decimal(12,2)"   json:"amount"`
	Status      int8       `gorm:"default:1"            json:"status"`
	Period      string     `gorm:"size:32"              json:"period"`
	PeriodStart time.Time  `gorm:"type:date"            json:"period_start"`
	PeriodEnd   time.Time  `gorm:"type:date"            json:"period_end"`
	Remark      string     `gorm:"size:256"             json:"remark"`
	SettledAt   *time.Time `                            json:"settled_at"`
	OperatorID  uint64     `gorm:"default:0"            json:"operator_id"`
	Org         *Org       `gorm:"foreignKey:OrgID"     json:"org,omitempty"`
}

func (FinanceSettlement) TableName() string { return "finance_settlements" }
