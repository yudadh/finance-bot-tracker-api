package parser

import (
	"time"

	"github.com/yudadh/finance-bot-tracker-api/internal/domain"
)

type TransactionIntent struct {
	Type            domain.TransactionType
	Amount          uint64
	Currency        string
	CategoryName    string
	TransactionDate time.Time
	Confidence      float64
}

type TransactionParser interface {
	Parser(input ParseInput) (*TransactionIntent, error)
}

type ParseInput struct {
	Text string
	Now time.Time
	Timezone *time.Location
	Currency string
	Categories []CategoryRule
}

type CategoryRule struct {
	Name string
	Type domain.TransactionType
	Keywords []string
}