package parser

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/yudadh/finance-bot-tracker-api/internal/domain"
)

type ParseBudgetResult struct {
	PeriodType domain.BudgetPeriodType
	Amount int64
}

var matchBudgetAmount = regexp.MustCompile(`(?i)\b(bulanan|mingguan|weekly|monthly)\b\s+(\d{1,3}(?:\.\d{3})*|\d+)\b`)

func ParseBudget(text string) (*ParseBudgetResult, error) {
	match := matchBudgetAmount.FindStringSubmatch(text)
	
	if match == nil {
		return nil, ErrInvalidBudgetText
	}

	amountText := strings.ReplaceAll(match[2], ".", "")
	amount, err := strconv.ParseInt(amountText, 10, 64)

	if err != nil {
		return nil, err
	}

	if amount <= 0 {
		return nil, ErrAmountMustBePositive
	}

	var periodType domain.BudgetPeriodType

	switch match[1] {
	case "bulanan", "monthly":
		periodType = domain.BudgetPeriodTypeMonthly

	case "mingguan", "weekly":
		periodType = domain.BudgetPeriodTypeWeekly
	}
	
	return &ParseBudgetResult{
		PeriodType: periodType,
		Amount: amount,
	}, nil
}