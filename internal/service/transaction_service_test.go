package service

import (
	"context"
	"testing"
	"time"

	"github.com/yudadh/finance-bot-tracker-api/internal/domain"
	"github.com/yudadh/finance-bot-tracker-api/internal/parser"
	"gorm.io/datatypes"
)

type fakeTransactionRepo struct {
	created *domain.Transaction
}

func (f *fakeTransactionRepo) Create(ctx context.Context, transaction *domain.Transaction) error {
	f.created = transaction
	return nil
}

type fakeCategoryRepo struct {
	categories []domain.Category
}

func (f *fakeCategoryRepo) FindAll(ctx context.Context) ([]domain.Category, error) {
	return f.categories, nil
}

type fakeParserAttemptRepo struct {
	created *domain.ParserAttempt
}

func (f *fakeParserAttemptRepo) Create(ctx context.Context, parserAttempt *domain.ParserAttempt) error {
	f.created = parserAttempt
	return nil
}

type fakeParser struct {
	intent *parser.TransactionIntent
	err error
}

func (f *fakeParser) Parse(input parser.ParseInput) (*parser.TransactionIntent, error) {
	return f.intent, f.err
}

func TestTransactionService_CreateFromText_Success(t *testing.T) {

	ctx := context.Background()
	now := time.Date(2026, 8, 20, 0, 0, 0, 0, time.UTC)

	transactionRepo := &fakeTransactionRepo{}
	categories := []domain.Category{
		{
			ID:       1,
			Name:     "Food",
			Type:     domain.TransactionTypeExpense,
			Keywords: datatypes.JSONSlice[string]{"makan", "kopi"},
		},
	}
	categoryRepo := &fakeCategoryRepo{categories: categories}
	parserAttemptRepo := &fakeParserAttemptRepo{}

	parser := &fakeParser{intent: &parser.TransactionIntent{
		Type:            domain.TransactionTypeExpense,
		Amount:          10000,
		Currency:        "IDR",
		Description:     "makan siang",
		CategoryName:    "Food",
		TransactionDate: now,
		Confidence:      0.9,
	}}

	service := NewTransactionService(transactionRepo, categoryRepo, parserAttemptRepo, parser, nil)

	result, err := service.CreateFromText(ctx, CreateTransactionFromTextInput{
		UserID:   10,
		Text:     "makan siang 10000",
		Now:      now,
		Currency: "IDR",
	})

	if err != nil {
		t.Fatal(err)
	}

	if result.Transaction.Amount != 10000 {
		t.Fatalf("expected amount 10000, got %d", result.Transaction.Amount)
	}

	if transactionRepo.created == nil {
		t.Fatal("expected transaction to be created")
	}

	if parserAttemptRepo.created == nil {
		t.Fatal("expected parser attempt to be created")
	}

}

func TestTransactionService_CreateFromText_ParseErrorDoesNotCreateTransaction(t *testing.T) {
	transactionRepo := &fakeTransactionRepo{}

	service := NewTransactionService(
		transactionRepo,
		&fakeCategoryRepo{},
		&fakeParserAttemptRepo{},
		&fakeParser{err: parser.ErrAmountNotFound},
		nil,
	)

	_, err := service.CreateFromText(context.Background(), CreateTransactionFromTextInput{
		UserID: 1,
		Text:   "makan siang",
	})

	if err == nil {
		t.Fatal("expected error")
	}

	if transactionRepo.created != nil {
		t.Fatal("expected transaction not to be created")
	}
}

func TestTransactionService_CreateFromText_ParseErrorReturnSuccessFalse(t *testing.T) {
	transactionRepo := &fakeTransactionRepo{}
	parserAttemptRepo := &fakeParserAttemptRepo{}

	service := NewTransactionService(
		transactionRepo,
		&fakeCategoryRepo{},
		parserAttemptRepo,
		&fakeParser{err: parser.ErrAmountNotFound},
		nil,
	)

	_, err := service.CreateFromText(context.Background(), CreateTransactionFromTextInput{
		UserID: 1,
		Text:   "gak tau",
	})

	if err == nil {
		t.Fatal("expected error")
	}

	if parserAttemptRepo.created == nil {
		t.Fatal("expected parser attempt to be created")
	}

	if parserAttemptRepo.created.Success {
		t.Fatalf("expected parser attempt success not to be true, got %t", parserAttemptRepo.created.Success)
	}

	if *parserAttemptRepo.created.ErrorMessage != parser.ErrAmountNotFound.Error() {
		t.Fatalf(
			"expected error message %q, got %q",
			parser.ErrAmountNotFound.Error(),
			*parserAttemptRepo.created.ErrorMessage,
		)
	}
}

func TestTransactionService_CreateFromText_NoMatchingCategory(t *testing.T) {
	now := time.Date(2026, 8, 20, 0, 0, 0, 0, time.UTC)
	transactionRepo := &fakeTransactionRepo{}
	
	categories := []domain.Category{
		{
			ID:       1,
			Name:     "Food",
			Type:     domain.TransactionTypeExpense,
			Keywords: datatypes.JSONSlice[string]{"makan", "kopi"},
		},
	}
	categoryRepo := &fakeCategoryRepo{categories: categories}
	parserAttemptRepo := &fakeParserAttemptRepo{}
	parser := &fakeParser{intent: &parser.TransactionIntent{
		Type:            domain.TransactionTypeExpense,
		Amount:          10000,
		Currency:        "IDR",
		Description:     "makan siang",
		CategoryName: "",
		TransactionDate: now,
		Confidence:      0.9,
	}}

	service := NewTransactionService(
		transactionRepo,
		categoryRepo,
		parserAttemptRepo,
		parser,
		nil,
	)

	_, err := service.CreateFromText(context.Background(), CreateTransactionFromTextInput{
		UserID: 1,
		Text:   "gak tau",
	})

	if err != nil {
		t.Fatal(err)
	}

	if transactionRepo.created == nil {
		t.Fatal("expected created not nil")
	}

	if transactionRepo.created.CategoryID != nil {
		t.Fatalf("expected categoryID nil, but got %d", *transactionRepo.created.CategoryID)
	}
}