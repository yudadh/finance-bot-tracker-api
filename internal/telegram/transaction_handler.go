package telegram

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/yudadh/finance-bot-tracker-api/internal/domain"
	"github.com/yudadh/finance-bot-tracker-api/internal/service"
)

type transactionFinderOrCreator interface {
	CreateFromText(ctx context.Context, transaction service.CreateTransactionFromTextInput) (*service.CreateTransactionFromTextResult, error)
	FindTransactionsByUserID(
		ctx context.Context,
		userID uint64,
		startTransactionDate time.Time,
		endTransactionDate time.Time,
	) ([]service.TransactionResult, error)
}

type budgetAlert interface {
	CheckBudgetAlert(ctx context.Context, input service.BudgetStatusInput) (*service.CheckBudgetAlertResult, error)
	UpdateBudgetTimestamp(
		ctx context.Context,
		budgetID uint64,
		threshold uint8,
		now time.Time,
	) error
}

type TransactionHandler struct {
	transactionService transactionFinderOrCreator
	budgetService budgetAlert
	conversations *ConversationStore
	logger *slog.Logger
}

func NewTransactionHandler(
	transactionService transactionFinderOrCreator,
	budgetService budgetAlert,
	logger *slog.Logger,
) *TransactionHandler {
	return &TransactionHandler{
		transactionService: transactionService,
		budgetService: budgetService,
		logger: logger,
	}
}

func (h *TransactionHandler) HandleCreate(
	ctx context.Context,
	chatID int64,
	text string,
	user service.FindOrCreateUserResult,
	reply Reply,
) {
	loc, err := time.LoadLocation(user.Timezone)
	if err != nil {
		reply.Text(ctx, chatID, InternalErrorMessage())
		return
	}

	transactionInput := service.CreateTransactionFromTextInput{
		UserID:   user.ID,
		Text:     text,
		Now:      time.Now(),
		Timezone: loc,
		Currency: "IDR",
	}

	result, err := h.transactionService.CreateFromText(ctx, transactionInput)
	if err != nil {
		reply.Text(ctx, chatID, parseErrorMessage(CmdTypeTransaction))
		return
	}

	reply.Text(ctx, chatID, transactionSavedMessage(result.Transaction, result.CategoryName))

	if result.Transaction.Type != domain.TransactionTypeExpense {
		return
	}

	for _, periodType := range []domain.BudgetPeriodType{
		domain.BudgetPeriodTypeWeekly,
		domain.BudgetPeriodTypeMonthly,
	} {
		result, err := h.budgetService.CheckBudgetAlert(ctx, service.BudgetStatusInput{
			UserID: user.ID,
			PeriodType: periodType,
			Location: loc,
		})
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				h.logger.Error("failed to check budget alert", "error", err)
			}
			continue
		}
		if result.ShouldAlert {
			if err := reply.Text(ctx, chatID, budgetAlertMessage(result, periodType)); err == nil {
				if err = h.budgetService.UpdateBudgetTimestamp(ctx, result.BudgetID, result.Threshold, time.Now()); err != nil {
					h.logger.Error("failed to update budget alert timestamp", "error", err)
				}
			}
		}
	}
}