package service

import (
	"context"
	"errors"
	"fmt"
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
	Transaction  *domain.Transaction
	CategoryName string
}

type ParseEditTransactionFromTextInput struct {
	TransactionID uint64
	UserID        uint64
	Text          string
	Currency      string
	FieldType     string
}

type ParseEditTransactionFromTextResult struct {
	EditTransactionIntent *parser.EditTransactionIntent
}

type UpdateTransactionFromChatInput struct {
	TransactionID uint64
	UserID        uint64
	Value         *parser.EditTransactionIntent
	FieldType     string
}

type UpdateTransactionFromChatResult struct {
	ID uint64
}

type DeleteTransactionResult struct {
	ID uint64
	Description string
	Amount int64
}

type transactionRepository interface {
	Create(ctx context.Context, transaction *domain.Transaction) error
	FindByUserIDWithCategory(
		ctx context.Context,
		userID uint64,
		startTransactionDate time.Time,
		endTransactionDate time.Time,
	) ([]domain.Transaction, error)
	FindByID(
		ctx context.Context,
		id uint64,
	) (*domain.Transaction, error)
	FindByIDAndUserID(
		ctx context.Context,
		id uint64,
		userID uint64,
	) (*domain.Transaction, error)
	Update(ctx context.Context, transaction *domain.Transaction) error
	HardDelete(
		ctx context.Context,
		transaction *domain.Transaction,
	) error
}

type categoryRepository interface {
	FindAllWithoutPagination(ctx context.Context) ([]domain.Category, error)
}

type parserAttemptRepository interface {
	Create(ctx context.Context, parserAttempt *domain.ParserAttempt) error
}

type userFinder interface {
	FindByID(ctx context.Context, id uint64) (*domain.User, error)
}

type TransactionService struct {
	transactionRepo   transactionRepository
	categoryRepo      categoryRepository
	parserAttemptRepo parserAttemptRepository
	userRepo          userFinder
	parser            parser.TransactionParser
	logger            *slog.Logger
}

type TransactionCategory struct {
	ID   uint64
	Name string
}

type TransactionResult struct {
	ID              uint64
	UserID          uint64
	Type            domain.TransactionType
	Amount          int64
	Currency        string
	Description     string
	TransactionDate time.Time
	Source          domain.TransactionSource
	Category        *TransactionCategory
}

func NewTransactionService(
	transactionRepo transactionRepository,
	categoryRepo categoryRepository,
	parserAttemptRepo parserAttemptRepository,
	userRepo userFinder,
	parser parser.TransactionParser,
	logger *slog.Logger,
) *TransactionService {
	return &TransactionService{
		transactionRepo:   transactionRepo,
		categoryRepo:      categoryRepo,
		parserAttemptRepo: parserAttemptRepo,
		userRepo:          userRepo,
		parser:            parser,
		logger:            logger,
	}
}

func (s *TransactionService) CreateFromText(
	ctx context.Context,
	input CreateTransactionFromTextInput,
) (*CreateTransactionFromTextResult, error) {
	categories, err := s.categoryRepo.FindAllWithoutPagination(ctx)
	if err != nil {
		return nil, err
	}

	categoryRules := buildCategoryRules(categories)

	intent, err := s.parser.Parse(parser.ParseInput{
		Text:       input.Text,
		Now:        input.Now,
		Timezone:   input.Timezone,
		Currency:   input.Currency,
		Categories: categoryRules,
	})

	parserAttempt := domain.ParserAttempt{
		UserID:     input.UserID,
		RawText:    input.Text,
		ParserType: domain.ParserTypeRuleBased,
		Success:    err == nil,
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
	transactionDate := normalizeDateToUTC(intent.TransactionDate, input.Timezone)

	transaction := domain.Transaction{
		UserID:           input.UserID,
		CategoryID:       categoryID,
		Type:             intent.Type,
		Amount:           intent.Amount,
		Currency:         intent.Currency,
		Description:      intent.Description,
		TransactionDate:  transactionDate,
		Source:           domain.TransactionSourceTelegram,
		RawText:          &input.Text,
		ParserConfidence: &intent.Confidence,
	}

	if err := s.transactionRepo.Create(ctx, &transaction); err != nil {
		return nil, err
	}

	return &CreateTransactionFromTextResult{
		Transaction:  &transaction,
		CategoryName: intent.CategoryName,
	}, nil

}

func buildCategoryRules(categories []domain.Category) []parser.CategoryRule {
	var categoryRules []parser.CategoryRule
	for _, category := range categories {
		categoryRule := parser.CategoryRule{
			Type:     category.Type,
			Name:     category.Name,
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

func normalizeDateToUTC(t time.Time, loc *time.Location) time.Time {
	year, month, day := t.In(loc).Date()

	return time.Date(
		year,
		month,
		day,
		0, 0, 0, 0,
		time.UTC,
	)
}

func (s *TransactionService) FindTransactionsByUserID(
	ctx context.Context,
	userID uint64,
	startTransactionDate time.Time,
	endTransactionDate time.Time,
) ([]TransactionResult, error) {

	if startTransactionDate.After(endTransactionDate) {
		return nil, domain.ErrInvalidDateRange
	}

	_, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, &domain.NotFoundError{Resource: domain.ResourceUser}
		}

		return nil, fmt.Errorf(
			"load user by id: %w",
			err,
		)
	}

	result, err := s.transactionRepo.FindByUserIDWithCategory(ctx, userID, startTransactionDate, endTransactionDate)
	if err != nil {
		return nil, fmt.Errorf(
			"load transactions by user: %w",
			err,
		)
	}

	transactions := make([]TransactionResult, 0, len(result))

	for _, res := range result {
		transactionData := TransactionResult{
			ID:              res.ID,
			UserID:          res.UserID,
			Type:            res.Type,
			Amount:          res.Amount,
			Currency:        res.Currency,
			Description:     res.Description,
			TransactionDate: res.TransactionDate,
			Source:          res.Source,
		}

		if res.Category != nil {
			transactionData.Category = &TransactionCategory{
				ID:   res.Category.ID,
				Name: res.Category.Name,
			}
		}

		transactions = append(transactions, transactionData)
	}

	return transactions, nil
}

func (s *TransactionService) FindTransactionByID(
	ctx context.Context,
	id uint64,
) (*TransactionResult, error) {
	result, err := s.transactionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf(
			"load transaction by id: %w",
			err,
		)
	}

	return &TransactionResult{
		ID:              result.ID,
		UserID:          result.UserID,
		Type:            result.Type,
		Amount:          result.Amount,
		Currency:        result.Currency,
		Description:     result.Description,
		TransactionDate: result.TransactionDate,
		Source:          result.Source,
	}, nil
}

func (s *TransactionService) FindTransactionByIDAndUserID(
	ctx context.Context,
	id uint64,
	userID uint64,
) (*TransactionResult, error) {
	result, err := s.transactionRepo.FindByIDAndUserID(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf(
			"load transaction by id and user_id: %w",
			err,
		)
	}

	return &TransactionResult{
		ID:              result.ID,
		UserID:          result.UserID,
		Type:            result.Type,
		Amount:          result.Amount,
		Currency:        result.Currency,
		Description:     result.Description,
		TransactionDate: result.TransactionDate,
		Source:          result.Source,
	}, nil
}

func (s *TransactionService) ParseEditTransactionFromText(
	ctx context.Context,
	input ParseEditTransactionFromTextInput,
) (*ParseEditTransactionFromTextResult, error) {
	fieldType := parser.EditField(input.FieldType)

	var categories []domain.Category
	var categoryRules []parser.CategoryRule

	if fieldType == parser.EditFieldDescription || fieldType == parser.EditFieldBoth {
		results, err := s.categoryRepo.FindAllWithoutPagination(ctx)
		if err != nil {
			return nil, fmt.Errorf(
				"load categories: %w",
				err,
			)
		}

		categories = results
		categoryRules = buildCategoryRules(categories)
	}

	intent, err := s.parser.ParseEdit(parser.ParseEditInput{
		Text:       input.Text,
		FieldType:  fieldType,
		Currency:   input.Currency,
		Categories: categoryRules,
	})

	parserAttempt := &domain.ParserAttempt{
		UserID:     input.UserID,
		RawText:    input.Text,
		ParserType: domain.ParserTypeRuleBased,
		Success:    err == nil,
	}

	if err != nil {
		if fieldType == parser.EditFieldBoth {
			errMessage := err.Error()
			parserAttempt.ErrorMessage = &errMessage
			if errParserAttempt := s.parserAttemptRepo.Create(ctx, parserAttempt); errParserAttempt != nil && s.logger != nil {
				s.logger.ErrorContext(ctx, "failed to create parserAttempt", "error", errParserAttempt)
			}
		}

		return nil, fmt.Errorf(
			"parse edit transaction: %w",
			err,
		)
	}

	if fieldType == parser.EditFieldBoth {
		parserAttempt.Confidence = intent.Confidence
		if errParserAttempt := s.parserAttemptRepo.Create(ctx, parserAttempt); errParserAttempt != nil && s.logger != nil {
			s.logger.ErrorContext(ctx, "failed to create parserAttempt", "error", errParserAttempt)
		}
	}

	return &ParseEditTransactionFromTextResult{
		EditTransactionIntent: intent,
	}, nil
}

func (s *TransactionService) UpdateTransactionFromChat(
	ctx context.Context,
	input UpdateTransactionFromChatInput,
) (*UpdateTransactionFromChatResult, error) {
	fieldType := parser.EditField(input.FieldType)
	var categories []domain.Category

	if fieldType == parser.EditFieldDescription || fieldType == parser.EditFieldBoth {
		results, err := s.categoryRepo.FindAllWithoutPagination(ctx)
		if err != nil {
			return nil, fmt.Errorf(
				"load categories: %w",
				err,
			)
		}

		categories = results
	}

	result, err := s.transactionRepo.FindByID(ctx, input.TransactionID)
	if err != nil {
		return nil, fmt.Errorf(
			"load transaction by id: %w",
			err,
		)
	}

	if result.UserID != input.UserID {
		return nil, domain.ErrNotFound
	}

	intent := input.Value

	switch fieldType {
	case parser.EditFieldDescription:

		categoryID := findCategoryID(*intent.CategoryName, *intent.Type, categories)

		result.Description = *intent.Description
		result.CategoryID = categoryID
		result.Type = *intent.Type
	case parser.EditFieldAmount:
		result.Amount = *intent.Amount

	case parser.EditFieldBoth:
		categoryID := findCategoryID(*intent.CategoryName, *intent.Type, categories)

		result.Description = *intent.Description
		result.CategoryID = categoryID
		result.Type = *intent.Type
		result.Amount = *intent.Amount
	}

	result.Currency = intent.Currency

	err = s.transactionRepo.Update(ctx, result)
	if err != nil {
		return nil, fmt.Errorf(
			"update transaction: %w",
			err,
		)
	}

	return &UpdateTransactionFromChatResult{ID: result.ID}, nil
}

func (s *TransactionService) HardDeleteByID(
	ctx context.Context,
	id uint64,
	userID uint64,
) (*DeleteTransactionResult, error) {
	result, err := s.transactionRepo.FindByIDAndUserID(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf(
			"load transaction by id and user_id: %w",
			err,
		)
	}

	err = s.transactionRepo.HardDelete(ctx, result)
	if err != nil {
		return nil, fmt.Errorf(
			"hard delete transaction: %w",
			err,
		)
	}

	return &DeleteTransactionResult{
		ID: result.ID,
		Description: result.Description,
		Amount: result.Amount,
	}, nil
}