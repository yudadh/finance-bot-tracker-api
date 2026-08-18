package domain

import (
	"time"

	"gorm.io/gorm"
)

type TransactionType string

const (
	TransactionTypeIncome  TransactionType = "income"
	TransactionTypeExpense TransactionType = "expense"
)

type TransactionSource string

const (
	TransactionSourceTelegram TransactionSource = "telegram"
	TransactionSourceAdmin    TransactionSource = "admin"
)

type Transaction struct {
	ID               uint64            `gorm:"column:id;primaryKey"`
	UserID           uint64            `gorm:"column:user_id"`
	CategoryID       *uint64           `gorm:"column:category_id"`
	Type             TransactionType   `gorm:"column:type"`
	Amount           int64             `gorm:"column:amount"`
	Currency         string            `gorm:"column:currency"`
	Description      string            `gorm:"column:description"`
	TransactionDate  time.Time         `gorm:"column:transaction_date"`
	Source           TransactionSource `gorm:"column:source"`
	RawText          *string           `gorm:"column:raw_text"`
	ParserConfidence *float64          `gorm:"column:parser_confidence"`
	CreatedAt        time.Time         `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt        time.Time         `gorm:"column:updated_at;autoCreateTime;autoUpdateTime"`
	DeletedAt        gorm.DeletedAt    `gorm:"column:deleted_at;index"`
	User             User              `gorm:"foreignKey:UserID;references:ID"`
	Category         *Category         `gorm:"foreignKey:CategoryID;references:ID"`
}

func (t *Transaction) TableName() string {
	return "transactions"
}
