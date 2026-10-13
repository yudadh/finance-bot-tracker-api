package parser

import (
	"errors"
	"testing"
	"time"

	"github.com/yudadh/finance-bot-tracker-api/internal/domain"
)

func TestParseAmountAndDescription(t *testing.T) {
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
		Text:     "beli milo 10000",
		Now:      now,
		Timezone: location,
		Currency: "IDR",
		Categories: []CategoryRule{
			{
				Name:     "Food",
				Type:     domain.TransactionTypeExpense,
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

func TestParseAmount(t *testing.T) {
	text := "10 rb"
	amount, err := ParseAmount(text, "IDR")

	if err != nil {
		t.Fatal("expected error to be nil")
	}

	if amount != 10000 {
		t.Fatalf("expected amount %d, got %d", 10000, amount)
	}
}

func TestRuleBasedParser_ParseEdit_SuccessByFieldType(t *testing.T) {
	p := NewRuleBasedParser()
	categories := []CategoryRule{
		{
			Name:     "Food",
			Type:     domain.TransactionTypeExpense,
			Keywords: []string{"makan", "kopi"},
		},
	}

	tests := []struct {
		name        string
		fieldType   EditField
		text        string
		amount      int64
		description string
	}{
		{
			name:        "description",
			fieldType:   EditFieldDescription,
			text:        "  makan siang  ",
			description: "makan siang",
		},
		{
			name:      "amount",
			fieldType: EditFieldAmount,
			text:      "10 rb",
			amount:    10000,
		},
		{
			name:        "description and amount",
			fieldType:   EditFieldBoth,
			text:        "makan siang 10 rb",
			amount:      10000,
			description: "makan siang",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			intent, err := p.ParseEdit(ParseEditInput{
				Text:       tt.text,
				FieldType:  tt.fieldType,
				Currency:   "IDR",
				Categories: categories,
			})
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if tt.amount != 0 {
				if intent.Amount == nil || *intent.Amount != tt.amount {
					t.Fatalf("expected amount %d, got %v", tt.amount, intent.Amount)
				}
			} else if intent.Amount != nil {
				t.Fatalf("expected no amount, got %d", *intent.Amount)
			}

			if tt.description != "" {
				if intent.Description == nil || *intent.Description != tt.description {
					t.Fatalf("expected description %q, got %v", tt.description, intent.Description)
				}
			} else if intent.Description != nil {
				t.Fatalf("expected no description, got %q", *intent.Description)
			}
		})
	}
}

func TestRuleBasedParser_ParseEdit_FailureByFieldType(t *testing.T) {
	p := NewRuleBasedParser()

	tests := []struct {
		name      string
		fieldType EditField
		text      string
		err       error
	}{
		{
			name:      "description with empty text",
			fieldType: EditFieldDescription,
			err:       ErrEmptyText,
		},
		{
			name:      "amount without amount",
			fieldType: EditFieldAmount,
			text:      "not a nominal",
			err:       ErrAmountNotFound,
		},
		{
			name:      "description and amount without amount",
			fieldType: EditFieldBoth,
			text:      "makan siang",
			err:       ErrAmountNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := p.ParseEdit(ParseEditInput{
				Text:      tt.text,
				FieldType: tt.fieldType,
				Currency:  "IDR",
			})
			if !errors.Is(err, tt.err) {
				t.Fatalf("expected error %v, got %v", tt.err, err)
			}
		})
	}
}
