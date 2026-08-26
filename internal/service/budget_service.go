package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/yudadh/finance-bot-tracker-api/internal/domain"
)

type BudgetRepository interface {
	Create(ctx context.Context, budget *domain.Budget) error
	FindByUserPeriod(ctx context.Context, userID uint64, period time.Time, periodType domain.BudgetPeriodType) (*domain.Budget, error)
	Update(ctx context.Context, budget *domain.Budget) error
	FindByID(ctx context.Context, budgetID uint64) (*domain.Budget, error)
}

type transactionSummarizer interface {
	SumTransactionsByUserAndDateRange(
		ctx context.Context,
		userID uint64,
		transactionType domain.TransactionType,
		startDate time.Time,
		endDate time.Time,
	) (*int64, error)
}

type BudgetService struct {
	budgetRepo      BudgetRepository
	transactionRepo transactionSummarizer
	logger          *slog.Logger
}

type CreateBudgetInput struct {
	UserID     uint64
	PeriodType domain.BudgetPeriodType
	Now        time.Time
	Amount     uint64
	Currency   string
	Loc        *time.Location
}

type CreateBudgetResult struct {
	UserID      uint64
	PeriodType  domain.BudgetPeriodType
	PeriodStart time.Time
	PeriodEnd   time.Time
	Amount      uint64
	Currency    string
}

type BudgetStatusInput struct {
	UserID     uint64
	PeriodType domain.BudgetPeriodType
	Location   *time.Location
}

type BudgetStatus struct {
	TotalTransactionAmount int64
	BudgetAmount           uint64
	PeriodStart            time.Time
	PeriodEnd              time.Time
}

type CheckBudgetAlertResult struct {
	BudgetID               uint64
	TotalTransactionAmount int64
	BudgetAmount           uint64
	UsedPercentage         uint64
	Threshold              uint8
	ShouldAlert            bool
}

func NewBudgetService(
	budgetRepo BudgetRepository,
	transactionRepo transactionSummarizer,
	logger *slog.Logger,
) *BudgetService {
	return &BudgetService{
		budgetRepo:      budgetRepo,
		transactionRepo: transactionRepo,
		logger:          logger,
	}
}

func setBudgetWindow(
	now time.Time,
	loc *time.Location,
	periodType domain.BudgetPeriodType,
) (time.Time, time.Time) {
	current := now.In(loc)

	var startDate time.Time
	var endDate time.Time

	switch periodType {
	case domain.BudgetPeriodTypeMonthly:
		startDate = time.Date(current.Year(), current.Month(), 1, 0, 0, 0, 0, loc)
		endDate = startDate.AddDate(0, 1, 0)

	case domain.BudgetPeriodTypeWeekly:
		day := int(now.Weekday())

		start := current.AddDate(0, 0, -(day - 1))
		startDate = time.Date(
			start.Year(),
			start.Month(),
			start.Day(),
			0, 0, 0, 0,
			loc,
		)

		endDate = startDate.AddDate(0, 0, 7)
	}

	return startDate, endDate
}

func (s *BudgetService) SetBudget(
	ctx context.Context,
	budgetInput CreateBudgetInput,
) (*CreateBudgetResult, error) {
	startDate, endDate := setBudgetWindow(
		budgetInput.Now,
		budgetInput.Loc,
		budgetInput.PeriodType,
	)

	budget := &domain.Budget{
		UserID:      budgetInput.UserID,
		PeriodType:  budgetInput.PeriodType,
		PeriodStart: startDate,
		Amount:      budgetInput.Amount,
		Currency:    budgetInput.Currency,
	}

	err := s.budgetRepo.Create(ctx, budget)

	if err != nil {
		if errors.Is(err, domain.ErrAlreadyExists) {
			return nil, domain.ErrAlreadyExists
		}

		return nil, err
	}

	return &CreateBudgetResult{
		UserID:      budget.UserID,
		PeriodType:  budget.PeriodType,
		PeriodStart: budget.PeriodStart,
		PeriodEnd:   endDate,
		Amount:      budget.Amount,
		Currency:    budget.Currency,
	}, nil
}

func (s *BudgetService) GetBudgetStatus(ctx context.Context, user BudgetStatusInput) (*BudgetStatus, error) {
	startDate, endDate := setBudgetWindow(time.Now(), user.Location, user.PeriodType)

	budget, err := s.budgetRepo.FindByUserPeriod(
		ctx,
		user.UserID,
		startDate,
		user.PeriodType,
	)

	if err != nil {
		return nil, err
	}

	totalTransactionAmount, err := s.transactionRepo.SumTransactionsByUserAndDateRange(
		ctx,
		budget.UserID,
		domain.TransactionTypeExpense,
		startDate,
		endDate,
	)

	if err != nil {
		return nil, err
	}

	return &BudgetStatus{
		TotalTransactionAmount: *totalTransactionAmount,
		BudgetAmount:           budget.Amount,
		PeriodStart:            startDate,
		PeriodEnd:              endDate,
	}, nil
}

func (s *BudgetService) CheckBudgetAlert(
	ctx context.Context,
	input BudgetStatusInput,
) (*CheckBudgetAlertResult, error) {
	startDate, endDate := setBudgetWindow(time.Now(), input.Location, input.PeriodType)

	budget, err := s.budgetRepo.FindByUserPeriod(
		ctx,
		input.UserID,
		startDate,
		input.PeriodType,
	)

	if err != nil {
		return nil, err
	}

	if budget.Alert100SentAt != nil {
		return &CheckBudgetAlertResult{
			ShouldAlert: false,
		}, nil
	}

	totalTransactionAmount, err := s.transactionRepo.SumTransactionsByUserAndDateRange(
		ctx,
		input.UserID,
		domain.TransactionTypeExpense,
		startDate,
		endDate,
	)

	if err != nil {
		return nil, err
	}

	if budget.Amount == 0 {
		return &CheckBudgetAlertResult{
			ShouldAlert: false,
		}, nil
	}

	usedPercentage := uint64(*totalTransactionAmount) * 100 / budget.Amount

	if usedPercentage >= 100 {
		return &CheckBudgetAlertResult{
			TotalTransactionAmount: *totalTransactionAmount,
			BudgetAmount:           budget.Amount,
			UsedPercentage:         usedPercentage,
			ShouldAlert:            true,
			Threshold:              100,
		}, nil
	}

	if usedPercentage >= 80 && budget.Alert80SentAt == nil {
		return &CheckBudgetAlertResult{
			TotalTransactionAmount: *totalTransactionAmount,
			BudgetAmount:           budget.Amount,
			UsedPercentage:         usedPercentage,
			ShouldAlert:            true,
			Threshold:              80,
		}, nil
	}

	return &CheckBudgetAlertResult{
		ShouldAlert: false,
	}, nil

}

func (s *BudgetService) UpdateBudgetTimestamp(
	ctx context.Context,
	budgetID uint64,
	threshold uint8,
	now time.Time,
) error {
	budget, err := s.budgetRepo.FindByID(ctx, budgetID)
	if err != nil {
		return err
	}

	switch threshold {
	case 80:
		budget.Alert80SentAt = &now
	case 100:
		budget.Alert100SentAt = &now
	default:
		return domain.ErrInvalidInput
	}

	return s.budgetRepo.Update(ctx, budget)
}
