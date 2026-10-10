package telegram

import (
	"context"
	"log/slog"
	"time"

	"github.com/go-telegram/bot/models"
	"github.com/yudadh/finance-bot-tracker-api/internal/service"
)

type CommandHandler struct {
	budgetHandler       *BudgetHandler
	conversationHandler *ConversationHandler
	conversations       *ConversationStore
	logger              *slog.Logger
}

func NewCommandHandler(
	budgetHandler *BudgetHandler,
	conversations *ConversationStore,
	logger *slog.Logger,
) *CommandHandler {
	return &CommandHandler{
		budgetHandler: budgetHandler,
		conversations: conversations,
		logger:        logger,
	}
}

func (h *CommandHandler) HandleCommand(
	ctx context.Context,
	chatID int64,
	user *service.FindOrCreateUserResult,
	message *models.Message,
	reply Reply,
) bool {
	text := message.Text

	switch DetectCommand(text) {
	case CommandStart:
		// h.SendTimezoneSelection(ctx, b, chatID)
		reply.Text(ctx, chatID, welcomeMessage())
		return true

	case CommandHelp:
		reply.Text(ctx, chatID, helpMessage())
		return true

	case CommandEdit:
		h.conversationHandler.Start(
			ctx, 
			chatID, 
			&ConversationState{
				Type:       ConversationEditTransaction,
				Step:       StepSelectDate,
				UserID:     user.ID,
				TelegramID: message.From.ID,
				ExpiredAt:  time.Now().Add(5 * time.Minute),
			},
			reply,
			message,
		)
		return true

	case CommandBudget:
		h.budgetHandler.handleBudgetCommand(ctx, chatID, reply, *user)
		return true

	case CommandSetBudget:
		h.budgetHandler.handleSetBudgetCommand(ctx, chatID, text, reply, *user)
		return true
	}

	return false
}
