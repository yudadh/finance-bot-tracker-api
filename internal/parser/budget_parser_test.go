package parser

import (
	"errors"
	"testing"

	"github.com/yudadh/finance-bot-tracker-api/internal/domain"
)

func TestParseBudget_Success(t *testing.T) {
	tests := []struct{
		name string
		input string
		expectedPeriodType domain.BudgetPeriodType
		expectedAmount int64
	}{
		{
			name: "monthly period type",
			input: "setbudget bulanan 3000000",
			expectedPeriodType: domain.BudgetPeriodTypeMonthly,
			expectedAmount: 3000000,
		},
		{
			name: "weekly period type",
			input: "setbudget mingguan 3000000",
			expectedPeriodType: domain.BudgetPeriodTypeWeekly,
			expectedAmount: 3000000,
		},
		{
			name: "with dotted amount",
			input: "setbudget bulanan 3.000.000",
			expectedPeriodType: domain.BudgetPeriodTypeMonthly,
			expectedAmount: 3000000,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := ParseBudget(test.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
	
			if result.PeriodType != test.expectedPeriodType {
				t.Errorf("expected %s got %s", test.expectedPeriodType, result.PeriodType)
			}
	
			if result.Amount != test.expectedAmount {
				t.Errorf("expected %d got %d", test.expectedAmount, result.Amount)
			}
		})
	}
}

func TestParseBudget_InvalidText(t *testing.T) {
	tests := []struct{
		name string
		input string
	}{
		{
			name: "empty text",
			input: "",
		},
		{
			name: "missing amount",
			input: "setbudget bulanan",
		},
		{
			name: "missing period type",
			input: "setbudget 3000000",
		},
		{
			name: "unknown period type",
			input: "setbudget tahunan 3000000",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ParseBudget(test.input)
			if !errors.Is(err, ErrInvalidBudgetText) {
				t.Fatalf("expected ErrInvalidBudgetText, got %v", err)
			}
		})
	}
}