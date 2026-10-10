package telegram

import (
	"context"
	"time"

	"github.com/go-telegram/bot/models"
)

type CallbackHandler struct {
	conversationHandler *ConversationHandler
	conversations *ConversationStore
}

func NewCallbackHandler(
	conversationHandler *ConversationHandler,
	conversations *ConversationStore,
) *CallbackHandler {
	return &CallbackHandler{
		conversationHandler: conversationHandler,
		conversations: conversations,
	}
}

func (h *CallbackHandler) Handle(
	ctx context.Context,
	callback *models.CallbackQuery,
	reply Reply,
) {
	if callback == nil || callback.Data == "" {
		return
	}

	chatID := callback.Message.Message.Chat.ID
	state, ok := h.conversations.Get(chatID)
	if !ok {
		return
	}

	if state.ExpiredAt.Before(time.Now()) {
		h.conversations.Delete(chatID)
		reply.Text(ctx, chatID, timeoutConversationMessage())
		return
	}

	h.conversationHandler.HandleCallback(ctx, chatID, state, reply, callback)
	return
}