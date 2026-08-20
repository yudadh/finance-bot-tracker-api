package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/yudadh/finance-bot-tracker-api/internal/domain"
	"github.com/yudadh/finance-bot-tracker-api/internal/parser"
)

type CreateTransactionFromTextInput struct {
	UserID   uint64
	Text     string
	Now      time.Time
	Timezone *time.Location
	Currency string
}

type CreateTransactionFromTextResult struct {
	Transaction *domain.Transaction
	CategoryName string
}

type transactionRepository interface {
	Create(ctx context.Context, transaction *domain.Transaction) error
}

type categoryRepository interface {
	FindAll(ctx context.Context) ([]domain.Category, error)
}

type parserAttemptRepository interface {
	Create(ctx context.Context, parserAttempt *domain.ParserAttempt) error
}

type TransactionService struct {
	transactionRepo transactionRepository
	categoryRepo categoryRepository
	parserAttemptRepo parserAttemptRepository
	parser parser.TransactionParser
	logger *slog.Logger
}

func NewTransactionService(
	transactionRepo transactionRepository,
	categoryRepo categoryRepository,
	parserAttemptRepo parserAttemptRepository,
	parser parser.TransactionParser,
	logger *slog.Logger,
) *TransactionService {
	return &TransactionService{
		transactionRepo: transactionRepo,
		categoryRepo: categoryRepo,
		parserAttemptRepo: parserAttemptRepo,
		parser: parser,
		logger: logger,
	}
}

func (s *TransactionService) CreateFromText(
	ctx context.Context,
	input CreateTransactionFromTextInput,
) (*CreateTransactionFromTextResult, error) {
	categories, err := s.categoryRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	categoryRules := buildCategoryRules(categories)

	
	intent, err := s.parser.Parse(parser.ParseInput{
		Text: input.Text,
		Now: input.Now,
		Timezone: input.Timezone,
		Currency: input.Currency,
		Categories: categoryRules,
	})

	parserAttempt := domain.ParserAttempt{
		UserID: input.UserID,
		RawText: input.Text,
		ParserType: domain.ParserTypeRuleBased,
		Success: err == nil,
	}

	if err != nil {
		message := err.Error()
		parserAttempt.ErrorMessage = &message

		if errParser := s.parserAttemptRepo.Create(ctx, &parserAttempt); errParser != nil && s.logger != nil {
			s.logger.Error("failed to create parserAttempt", "error", errParser.Error())
		}

		return nil, err
	}

	parserAttempt.Confidence = &intent.Confidence

	if err = s.parserAttemptRepo.Create(ctx, &parserAttempt); err != nil && s.logger != nil {
		s.logger.Error("failed to create parserAttempt", "error", err.Error())
	}

	categoryID := findCategoryID(intent.CategoryName, intent.Type, categories)

	transaction := domain.Transaction{
		UserID: input.UserID,
		CategoryID: categoryID,
		Type: intent.Type,
		Amount: intent.Amount,
		Currency: intent.Currency,
		Description: intent.Description,
		TransactionDate: intent.TransactionDate,
		Source: domain.TransactionSourceTelegram,
		RawText: &input.Text,
		ParserConfidence: &intent.Confidence,
	}
	
	if err := s.transactionRepo.Create(ctx, &transaction); err != nil {
		return nil, err
	}

	return &CreateTransactionFromTextResult{
		Transaction: &transaction,
		CategoryName: intent.CategoryName,
	}, nil

}

func buildCategoryRules(categories []domain.Category) []parser.CategoryRule {
	var categoryRules []parser.CategoryRule
	for _, category := range categories {
		categoryRule := parser.CategoryRule{
			Type: category.Type,
			Name: category.Name,
			Keywords: []string(category.Keywords),
		}
		categoryRules = append(categoryRules, categoryRule)
	}

	return categoryRules
}

func findCategoryID(
	categoryName string, 
	categoryType domain.TransactionType,
	categories []domain.Category,
) *uint64 {
	for _, category := range categories {
		if category.Name == categoryName && category.Type == categoryType {
			return &category.ID
		}
	}

	return nil
}