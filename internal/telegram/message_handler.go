package telegram

import (
	"context"
	"log/slog"
	"time"

	"github.com/go-telegram/bot/models"
	"github.com/yudadh/finance-bot-tracker-api/internal/service"
)

type userFinderOrCreator interface {
	FindOrCreate(ctx context.Context, userInput *service.FindOrCreateUserInput) (*service.FindOrCreateUserResult, error)
}

type MessageHandler struct {
	userService userFinderOrCreator
	transactionHandler *TransactionHandler
	commandHandler *CommandHandler
	conversationHandler *ConversationHandler
	conversations *ConversationStore
	logger *slog.Logger
}

func NewMessageHandler(
	userService userFinderOrCreator,
	transactionHandler *TransactionHandler,
	commandHandler *CommandHandler,
	conversationHandler *ConversationHandler,
	conversations *ConversationStore,
	logger *slog.Logger,
) *MessageHandler {
	return &MessageHandler{
		userService: userService,
		transactionHandler: transactionHandler,
		commandHandler: commandHandler,
		conversationHandler: conversationHandler,
		conversations: conversations,
		logger: logger,
	}
}

func (h *MessageHandler) Handle(
	ctx context.Context,
	message *models.Message,
	reply Reply,
) {
	text := message.Text
	if text == "" {
		return
	}

	chatID := message.Chat.ID

	state, ok := h.conversations.Get(chatID)
	if ok {
		if state.ExpiredAt.Before(time.Now()) {
			h.conversations.Delete(chatID)
			reply.Text(ctx, chatID, timeoutConversationMessage())
			return
		}
		
		h.conversationHandler.HandleMessage(ctx, chatID, state, reply, message)
		return
	}

	user, err := h.userService.FindOrCreate(ctx, &service.FindOrCreateUserInput{
		TelegramID: message.From.ID,
		TelegramUsername: message.From.Username,
		FirstName: message.From.FirstName,
		LastName: message.From.LastName,
		LanguageCode: message.From.LanguageCode,
	})

	if err != nil {
		h.logger.ErrorContext(ctx, "error fetch or create user", "error:", err)
		reply.Text(ctx, chatID, InternalErrorMessage())
		return
	}

	if h.commandHandler.HandleCommand(ctx, chatID, user, message, reply) {
		return
	}

	h.transactionHandler.HandleCreate(ctx, chatID, text, *user, reply)
	return

}