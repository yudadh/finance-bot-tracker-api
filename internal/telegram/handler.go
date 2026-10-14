package telegram

import (
	"context"
	"log/slog"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type budgetService interface {
	budgetFinderAndCreator
	budgetAlert
}

type transactionService interface {
	transactionFinderOrCreator
	editTransactionService
	deleteTransactionService
}

type BotHandler struct {
	router *UpdateRouter
}

func NewBotHandler(
	transactionService transactionService,
	userService userFinderOrCreator,
	budgetService budgetService,
	conversations *ConversationStore,
	logger *slog.Logger,
) *BotHandler {
	reply := NewReplyFactory(logger)

	budgetHandler := NewBudgetHandler(budgetService, logger)
	transactionHandler := NewTransactionHandler(transactionService, budgetService, logger)

	editTransactionFlow := NewEditTransactionFlow(transactionService, conversations, logger)
	deleteTransactionFlow := NewDeleteTransactionFlow(transactionService, conversations)
	conversationHandler := NewConversationHandler(editTransactionFlow, deleteTransactionFlow, conversations)
	commandHandler := NewCommandHandler(budgetHandler, conversationHandler, conversations, logger)
	callbackHandler := NewCallbackHandler(conversationHandler, conversations)

	messageHandler := NewMessageHandler(
		userService, 
		transactionHandler, 
		commandHandler,
		conversationHandler, 
		conversations, 
		logger,
	)

	return &BotHandler{
		router: NewUpdateRouter(
			messageHandler,
			callbackHandler,
			conversations,
			reply,
			logger,
		),
	}
}

func (h *BotHandler) HandleUpdate(
	ctx context.Context,
	tgBot *bot.Bot,
	update *models.Update,
) {
	h.router.Handle(ctx, tgBot, update)
}
