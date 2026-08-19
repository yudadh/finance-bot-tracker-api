package parser

import (
	"testing"
	"time"

	"github.com/yudadh/finance-bot-tracker-api/internal/domain"
)

func TestParseAmount(t *testing.T) {
	text := "beli ayam 1 ekor 50.000 mantap"
	amount, description, err := ParseAmountAndDescription(text, "IDR")
	
	if amount != 50000 {
		t.Fatalf("expected 0 %d", amount)
	}

	if description != "beli ayam 1 ekor mantap" {
		t.Fatalf("expected empty string %s", description)
	}

	if err != nil {
		t.Fatal(err)
	}

}

func TestRuleBasedParser_ParseExpense(t *testing.T) {
	location := time.FixedZone("WITA", 8*60*60)
	now := time.Date(2026, 8, 19, 15, 0, 0, 0, location)

	p := NewRuleBasedParser()

	intent, err := p.Parse(ParseInput{
		Text: "beli milo 10000",
		Now: now,
		Timezone: location,
		Currency: "IDR",
		Categories: []CategoryRule{
			{
				Name: "Food", 
				Type: domain.TransactionTypeExpense, 
				Keywords: []string{"makan", "kopi", "milo"},
			},
		},
	})

	if err != nil {
		t.Fatal(err)
	}

	if intent.Amount != 10000 {
		t.Fatalf("expected amount 10000, got %d", intent.Amount)
	}

	if intent.Type != domain.TransactionTypeExpense {
		t.Fatalf("expected transaction type expense, got %s", intent.Type)
	}

	if intent.CategoryName != "Food" {
		t.Fatalf("expected category name Food, got %s", intent.CategoryName)
	}
}

func TestRuleBasedParser_ParseIncomeWithRb(t *testing.T) {
	p := NewRuleBasedParser()

	intent, err := p.Parse(ParseInput{
		Text:     "bonus proyek 750rb",
		Now:      time.Now(),
		Currency: "IDR",
	})

	if err != nil {
		t.Fatal(err)
	}

	if intent.Type != domain.TransactionTypeIncome {
		t.Fatalf("expected income, got %s", intent.Type)
	}

	if intent.Amount != 750000 {
		t.Fatalf("expected amount 750000, got %d", intent.Amount)
	}
}