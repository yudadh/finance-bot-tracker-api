package telegram

import (
	"context"
	"log/slog"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type UpdateRouter struct {
	messages      *MessageHandler
	callbacks     *CallbackHandler
	conversations *ConversationStore
	replies       *ReplyFactory
	logger        *slog.Logger
}

func NewUpdateRouter(
	messageHandler *MessageHandler,
	callbackHandler *CallbackHandler,
	conversations *ConversationStore,
	reply *ReplyFactory,
	logger *slog.Logger,
) *UpdateRouter {
	return &UpdateRouter{
		messages:      messageHandler,
		callbacks:     callbackHandler,
		conversations: conversations,
		replies:       reply,
		logger:        logger,
	}
}

func (r *UpdateRouter) Handle(
	ctx context.Context,
	tgBot *bot.Bot,
	update *models.Update,
) {

	reply := r.replies.For(tgBot)

	if update.CallbackQuery != nil {
		r.callbacks.Handle(ctx, update.CallbackQuery, reply)
		return
	}

	if update.Message != nil {
		r.messages.Handle(ctx, update.Message, reply)
	}
}
