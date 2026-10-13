package parser

import (
	"errors"
	"strings"

	"github.com/yudadh/finance-bot-tracker-api/internal/domain"
)

type RuleBasedParser struct{}

func NewRuleBasedParser() *RuleBasedParser {
	return &RuleBasedParser{}
}

func (p *RuleBasedParser) Parse(input ParseInput) (*TransactionIntent, error) {
	text := strings.TrimSpace(strings.ToLower(input.Text))

	if text == "" {
		return nil, ErrEmptyText
	}

	amount, description, err := ParseAmountAndDescription(text, input.Currency)
	if err != nil {
		return nil, err
	}

	transactionType := detectTransactionType(description)
	categoryName := matchCategory(description, transactionType, input.Categories)

	confidence := 0.80
	if categoryName != "" {
		confidence = 0.90
	}

	return &TransactionIntent{
		Type:            transactionType,
		Amount:          amount,
		Currency:        input.Currency,
		CategoryName:    categoryName,
		Description:     description,
		TransactionDate: parseSimpleDate(text, input.Now, input.Timezone),
		Confidence:      confidence,
	}, nil
}

func (p *RuleBasedParser) ParseEdit(input ParseEditInput) (*EditTransactionIntent, error) {
	text := strings.TrimSpace(strings.ToLower(input.Text))
	if text == "" {
		return nil, ErrEmptyText
	}

	switch input.FieldType {
	case EditFieldDescription:
		transactionType := detectTransactionType(text)
		categoryName := matchCategory(text, transactionType, input.Categories)

		return &EditTransactionIntent{
			Type:         &transactionType,
			Currency:     input.Currency,
			CategoryName: &categoryName,
			Description:  &text,
		}, nil

	case EditFieldAmount:
		amount, err := ParseAmount(text, input.Currency)
		if err != nil {
			return nil, err
		}

		return &EditTransactionIntent{
			Amount:   &amount,
			Currency: input.Currency,
		}, nil

	case EditFieldBoth:
		amount, description, err := ParseAmountAndDescription(text, input.Currency)
		if err != nil {
			return nil, err
		}

		transactionType := detectTransactionType(description)
		categoryName := matchCategory(description, transactionType, input.Categories)

		confidence := 0.80
		if categoryName != "" {
			confidence = 0.90
		}

		return &EditTransactionIntent{
			Type:         &transactionType,
			Amount:       &amount,
			Currency:     input.Currency,
			CategoryName: &categoryName,
			Description:  &description,
			Confidence:   &confidence,
		}, nil

	default:
		return nil, errors.New("unknown edit field type")

	}
}

func detectTransactionType(text string) domain.TransactionType {
	incomeKeywords := []string{
		"gaji",
		"bonus",
		"freelance",
		"dibayar",
		"bayaran",
		"pendapatan",
		"income",
	}

	for _, keyword := range incomeKeywords {
		if strings.Contains(text, keyword) {
			return domain.TransactionTypeIncome
		}
	}

	return domain.TransactionTypeExpense
}
