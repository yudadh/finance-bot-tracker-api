package parser

import (
	"strings"

	"github.com/yudadh/finance-bot-tracker-api/internal/domain"
)

func matchCategory(
	text string,
	transactionType domain.TransactionType,
	categories []CategoryRule,
) string {
	for _, category := range categories {
		if category.Type != transactionType {
			continue
		}

		for _, keyword := range category.Keywords {
			if strings.Contains(text, strings.ToLower(keyword)) {
				return category.Name
			}
		}
	}
	
	return ""
}