package parser

import (
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
		Type: transactionType,
		Amount: amount,
		Currency: input.Currency,
		CategoryName: categoryName,
		Description: description,
		TransactionDate: parseSimpleDate(text, input.Now, input.Timezone),
		Confidence: confidence,
	}, nil
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