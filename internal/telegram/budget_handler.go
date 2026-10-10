package telegram

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/yudadh/finance-bot-tracker-api/internal/domain"
	"github.com/yudadh/finance-bot-tracker-api/internal/parser"
	"github.com/yudadh/finance-bot-tracker-api/internal/service"
)

type budgetFinderAndCreator interface {
	SetBudget(ctx context.Context, budgetInput service.CreateBudgetInput) (*service.CreateBudgetResult, error)
	GetBudgetStatus(ctx context.Context, user service.BudgetStatusInput) (*service.BudgetStatus, error)
}

type BudgetHandler struct {
	budgetService budgetFinderAndCreator
	logger *slog.Logger
}

func NewBudgetHandler(
	budgetService budgetFinderAndCreator,
	logger *slog.Logger,
) *BudgetHandler {
	return &BudgetHandler{
		budgetService: budgetService,
		logger: logger,
	}
}

func (h *BudgetHandler) handleBudgetCommand(
	ctx context.Context,
	chatID int64,
	reply Reply,
	user service.FindOrCreateUserResult,
) {
	loc, err := time.LoadLocation(user.Timezone)
	if err != nil {
		reply.Text(ctx, chatID, InternalErrorMessage())
		return
	}

	budgetStatus, err := h.budgetService.GetBudgetStatus(ctx, service.BudgetStatusInput{
		UserID:     user.ID,
		PeriodType: domain.BudgetPeriodTypeMonthly,
		Location:   loc,
	})

	if err != nil {
		reply.Text(ctx, chatID, InternalErrorMessage())
		return
	}

	reply.Text(ctx, chatID, budgetStatusMessage(budgetStatus))
}

func (h *BudgetHandler) handleSetBudgetCommand(
	ctx context.Context,
	chatID int64,
	text string,
	reply Reply,
	user service.FindOrCreateUserResult,
) {
	res, err := parser.ParseBudget(text)

	if err != nil {
		if errors.Is(err, parser.ErrInvalidBudgetText) {
			reply.Text(ctx, chatID, parseErrorMessage(CmdTypeBudget))
			return
		}
		if errors.Is(err, parser.ErrAmountMustBePositive) {
			reply.Text(ctx, chatID, parseErrorAmountMessage(CmdTypeBudget))
			return
		}
		reply.Text(ctx, chatID, InternalErrorMessage())
		return
	}

	loc, err := time.LoadLocation(user.Timezone)
	if err != nil {
		reply.Text(ctx, chatID, InternalErrorMessage())
		return
	}

	budget, err := h.budgetService.SetBudget(ctx, service.CreateBudgetInput{
		UserID:     user.ID,
		PeriodType: res.PeriodType,
		Now:        time.Now(),
		Amount:     uint64(res.Amount),
		Currency:   "IDR",
		Loc:        loc,
	})

	if err != nil {
		if errors.Is(err, domain.ErrAlreadyExists) {
			reply.Text(ctx, chatID, budgetAlreadyExistsMessage())
			return
		}
		reply.Text(ctx, chatID, InternalErrorMessage())
		return
	}

	reply.Text(ctx, chatID, budgetCreatedMessage(budget))
}