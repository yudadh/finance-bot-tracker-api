package domain

import "time"

type BudgetPeriodType string

const (
	BudgetPeriodTypeWeekly  BudgetPeriodType = "weekly"
	BudgetPeriodTypeMonthly BudgetPeriodType = "monthly"
)

type Budget struct {
	ID             uint64           `gorm:"column:id;primaryKey"`
	UserID         uint64           `gorm:"column:user_id"`
	PeriodType     BudgetPeriodType `gorm:"column:period_type"`
	PeriodStart    time.Time        `gorm:"column:period_start"`
	Amount         uint64           `gorm:"column:amount"`
	Currency       string           `gorm:"column:currency"`
	Alert80SentAt  *time.Time       `gorm:"column:alert_80_sent_at"`
	Alert100SentAt *time.Time       `gorm:"column:alert_100_sent_at"`
	CreatedAt      time.Time        `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt      time.Time        `gorm:"column:updated_at;autoCreateTime;autoUpdateTime"`
	User           User             `gorm:"foreignKey:UserID;references:ID"`
}

func (b *Budget) TableName() string {
	return "budgets"
}
