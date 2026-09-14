package service

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/yudadh/finance-bot-tracker-api/internal/domain"
	"github.com/yudadh/finance-bot-tracker-api/internal/parser"
	"gorm.io/datatypes"
)

type fakeTransactionRepo struct {
	created *domain.Transaction

	findByUserIDWithCategoryCalls  int
	userID                         uint64
	startTransactionDate           time.Time
	endTransactionDate             time.Time
	findByUserIDWithCategoryResult []domain.Transaction
	findByUserIDWithCategoryErr    error
}

func (f *fakeTransactionRepo) Create(ctx context.Context, transaction *domain.Transaction) error {
	f.created = transaction
	return nil
}

func (f *fakeTransactionRepo) FindByUserIDWithCategory(
	ctx context.Context,
	userID uint64,
	startTransactionDate time.Time,
	endTransactionDate time.Time,
) ([]domain.Transaction, error) {
	f.findByUserIDWithCategoryCalls++
	f.userID = userID
	f.startTransactionDate = startTransactionDate
	f.endTransactionDate = endTransactionDate
	return f.findByUserIDWithCategoryResult, f.findByUserIDWithCategoryErr
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

type fakeUserRepo struct {
	calls  int
	userID uint64
	result *domain.User
	err    error
}

func (f *fakeUserRepo) FindByID(ctx context.Context, id uint64) (*domain.User, error) {
	f.calls++
	f.userID = id
	return f.result, f.err
}

type fakeParser struct {
	intent *parser.TransactionIntent
	err    error
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
	userRepo := &fakeUserRepo{}

	parser := &fakeParser{intent: &parser.TransactionIntent{
		Type:            domain.TransactionTypeExpense,
		Amount:          10000,
		Currency:        "IDR",
		Description:     "makan siang",
		CategoryName:    "Food",
		TransactionDate: now,
		Confidence:      0.9,
	}}

	service := NewTransactionService(
		transactionRepo,
		categoryRepo,
		parserAttemptRepo,
		userRepo,
		parser,
		nil,
	)

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
		&fakeUserRepo{},
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
		&fakeUserRepo{},
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
		CategoryName:    "",
		TransactionDate: now,
		Confidence:      0.9,
	}}

	service := NewTransactionService(
		transactionRepo,
		categoryRepo,
		parserAttemptRepo,
		&fakeUserRepo{},
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

func TestFindTransactionsByUserID_RepositoryError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	service := NewTransactionService(
		&fakeTransactionRepo{findByUserIDWithCategoryErr: errors.New("database connection failed")},
		&fakeCategoryRepo{},
		&fakeParserAttemptRepo{},
		&fakeUserRepo{},
		&fakeParser{},
		logger,
	)
	date := time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC)
	_, err := service.FindTransactionsByUserID(context.Background(), 1, date, date)

	if err == nil {
		t.Fatal("expected error not nil")
	}
}

func TestFindTransactionsByUserID_SuccessMapsAllFieldsAndDates(t *testing.T) {
	startDate := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC)
	transactionDate := time.Date(2026, 9, 5, 0, 0, 0, 0, time.UTC)
	categoryID := uint64(7)

	userRepo := &fakeUserRepo{result: &domain.User{ID: 42}}
	transactionRepo := &fakeTransactionRepo{
		findByUserIDWithCategoryResult: []domain.Transaction{
			{
				ID:              100,
				UserID:          42,
				CategoryID:      &categoryID,
				Type:            domain.TransactionTypeExpense,
				Amount:          12500,
				Currency:        "IDR",
				Description:     "lunch",
				TransactionDate: transactionDate,
				Source:          domain.TransactionSourceTelegram,
				Category:        &domain.Category{ID: categoryID, Name: "Food"},
			},
			{
				ID:              101,
				UserID:          42,
				Type:            domain.TransactionTypeIncome,
				Amount:          500000,
				Currency:        "IDR",
				Description:     "salary",
				TransactionDate: endDate,
				Source:          domain.TransactionSourceAdmin,
			},
		},
	}

	svc := NewTransactionService(
		transactionRepo,
		&fakeCategoryRepo{},
		&fakeParserAttemptRepo{},
		userRepo,
		&fakeParser{},
		nil,
	)

	result, err := svc.FindTransactionsByUserID(context.Background(), 42, startDate, endDate)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if userRepo.calls != 1 || userRepo.userID != 42 {
		t.Fatalf("unexpected user lookup: calls=%d userID=%d", userRepo.calls, userRepo.userID)
	}
	if transactionRepo.findByUserIDWithCategoryCalls != 1 {
		t.Fatalf("expected one transaction lookup, got %d", transactionRepo.findByUserIDWithCategoryCalls)
	}
	if transactionRepo.userID != 42 || !transactionRepo.startTransactionDate.Equal(startDate) || !transactionRepo.endTransactionDate.Equal(endDate) {
		t.Fatalf("unexpected repository input: user=%d start=%v end=%v", transactionRepo.userID, transactionRepo.startTransactionDate, transactionRepo.endTransactionDate)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 transactions, got %d", len(result))
	}

	got := result[0]
	if got.ID != 100 || got.UserID != 42 || got.Type != domain.TransactionTypeExpense || got.Amount != 12500 || got.Currency != "IDR" || got.Description != "lunch" || !got.TransactionDate.Equal(transactionDate) || got.Source != domain.TransactionSourceTelegram {
		t.Fatalf("unexpected mapped transaction: %+v", got)
	}
	if got.Category == nil || got.Category.ID != categoryID || got.Category.Name != "Food" {
		t.Fatalf("unexpected mapped category: %+v", got.Category)
	}
	if result[1].Category != nil {
		t.Fatalf("expected nil category, got %+v", result[1].Category)
	}
}

func TestFindTransactionsByUserID_InvalidDateRangeDoesNotQueryRepositories(t *testing.T) {
	userRepo := &fakeUserRepo{result: &domain.User{ID: 42}}
	transactionRepo := &fakeTransactionRepo{}
	svc := NewTransactionService(
		transactionRepo,
		&fakeCategoryRepo{},
		&fakeParserAttemptRepo{},
		userRepo,
		&fakeParser{},
		nil,
	)

	startDate := time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC)
	_, err := svc.FindTransactionsByUserID(context.Background(), 42, startDate, endDate)

	if !errors.Is(err, domain.ErrInvalidDateRange) {
		t.Fatalf("expected invalid date range, got %v", err)
	}
	if userRepo.calls != 0 || transactionRepo.findByUserIDWithCategoryCalls != 0 {
		t.Fatalf("expected no repository calls, got user=%d transactions=%d", userRepo.calls, transactionRepo.findByUserIDWithCategoryCalls)
	}
}

func TestFindTransactionsByUserID_UserNotFound(t *testing.T) {
	userRepo := &fakeUserRepo{err: domain.ErrNotFound}
	transactionRepo := &fakeTransactionRepo{}
	svc := NewTransactionService(
		transactionRepo,
		&fakeCategoryRepo{},
		&fakeParserAttemptRepo{},
		userRepo,
		&fakeParser{},
		nil,
	)

	date := time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC)
	_, err := svc.FindTransactionsByUserID(context.Background(), 42, date, date)

	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}
	var notFound *domain.NotFoundError
	if !errors.As(err, &notFound) || notFound.Resource != domain.ResourceUser {
		t.Fatalf("expected user not found error, got %v", err)
	}
	if transactionRepo.findByUserIDWithCategoryCalls != 0 {
		t.Fatal("expected transaction repository not to be called")
	}
}

func TestFindTransactionsByUserID_UserRepositoryError(t *testing.T) {
	userErr := errors.New("database connection failed")
	userRepo := &fakeUserRepo{err: userErr}
	transactionRepo := &fakeTransactionRepo{}
	svc := NewTransactionService(
		transactionRepo,
		&fakeCategoryRepo{},
		&fakeParserAttemptRepo{},
		userRepo,
		&fakeParser{},
		nil,
	)

	date := time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC)
	_, err := svc.FindTransactionsByUserID(context.Background(), 42, date, date)

	if !errors.Is(err, userErr) {
		t.Fatalf("expected wrapped user error, got %v", err)
	}
	if transactionRepo.findByUserIDWithCategoryCalls != 0 {
		t.Fatal("expected transaction repository not to be called")
	}
}

func TestFindTransactionsByUserID_EmptyResultIsNonNil(t *testing.T) {
	userRepo := &fakeUserRepo{result: &domain.User{ID: 42}}
	svc := NewTransactionService(
		&fakeTransactionRepo{findByUserIDWithCategoryResult: []domain.Transaction{}},
		&fakeCategoryRepo{},
		&fakeParserAttemptRepo{},
		userRepo,
		&fakeParser{},
		nil,
	)

	date := time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC)
	result, err := svc.FindTransactionsByUserID(context.Background(), 42, date, date)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil empty result")
	}
	if len(result) != 0 {
		t.Fatalf("expected empty result, got %d items", len(result))
	}
}
