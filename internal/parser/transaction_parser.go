package parser

import (
	"time"

	"github.com/yudadh/finance-bot-tracker-api/internal/domain"
)

type TransactionIntent struct {
	Type            domain.TransactionType
	Amount          int64
	Currency        string
	CategoryName    string
	Description     string
	TransactionDate time.Time
	Confidence      float64
}

type EditTransactionIntent struct {
	Type         *domain.TransactionType
	Amount       *int64
	Currency     string
	CategoryName *string
	Description  *string
	Confidence   *float64
}

type TransactionParser interface {
	Parse(input ParseInput) (*TransactionIntent, error)
	ParseEdit(input ParseEditInput) (*EditTransactionIntent, error)
}

type ParseInput struct {
	Text       string
	Now        time.Time
	Timezone   *time.Location
	Currency   string
	Categories []CategoryRule
}

type EditField string

const (
	EditFieldDescription EditField = "description"
	EditFieldAmount      EditField = "amount"
	EditFieldBoth        EditField = "description_amount"
)

type ParseEditInput struct {
	Text      string
	FieldType EditField
	Currency  string
	Categories []CategoryRule
}

type CategoryRule struct {
	Name     string
	Type     domain.TransactionType
	Keywords []string
}
