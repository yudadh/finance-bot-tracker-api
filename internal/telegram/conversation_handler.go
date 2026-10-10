package telegram

import (
	"context"

	"github.com/go-telegram/bot/models"
)

type ConversationFlow interface {
	Start(
		ctx context.Context,
		chatID int64,
		conversationState *ConversationState,
		reply Reply,
	)

	HandleMessage(
		ctx context.Context,
		chatID int64,
		conversationState *ConversationState,
		reply Reply,
		text string,
	)

	HandleCallback(
		ctx context.Context,
		chatID int64,
		conversationState *ConversationState,
		reply Reply,
		callbackData string,
	)

	Back(
		ctx context.Context,
		chatID int64,
		conversationState *ConversationState,
		reply Reply,
		callbackData string,
	)
}

type ConversationHandler struct {
	flow map[ConversationType]ConversationFlow
	conversations *ConversationStore
}

func NewConversationHandler(
	editTransactionFlow *EditTransactionFlow,
	deleteTransactionFlow *DeleteTransactionFlow,
	conversations *ConversationStore,
) *ConversationHandler {
	return &ConversationHandler{
		flow: map[ConversationType]ConversationFlow{
			ConversationEditTransaction: editTransactionFlow,
			ConversationDeleteTransaction: deleteTransactionFlow,
		},
		conversations: conversations,
	}
}

func (h *ConversationHandler) Start(
	ctx context.Context,
	chatID int64,
	state *ConversationState,
	reply Reply,
	message *models.Message,
) {
	if message.Text == "" {
		return
	}

	flow, ok := h.flow[state.Type]
	if !ok {
		reply.Text(ctx, chatID, InternalErrorMessage())
		return
	}

	flow.Start(ctx, chatID, state, reply)
}

func (h *ConversationHandler) HandleMessage(
	ctx context.Context,
	chatID int64,
	state *ConversationState,
	reply Reply,
	message *models.Message,
) {
	if message.Text == "" {
		return
	}

	flow, ok := h.flow[state.Type]
	if !ok {
		reply.Text(ctx, chatID, InternalErrorMessage())
		return
	}

	flow.HandleMessage(ctx, chatID, state, reply, message.Text)
}

func (h *ConversationHandler) HandleCallback(
	ctx context.Context,
	chatID int64,
	state *ConversationState,
	reply Reply,
	callback *models.CallbackQuery,
) {
	callbackData := callback.Data
	if callbackData == "cancel" {
		h.conversations.Delete(chatID)
		reply.Text(ctx, chatID, cancelCommandMessage(state.Type))
		return
	}

	flow, ok := h.flow[state.Type]
	if !ok {
		h.conversations.Delete(chatID)
	}

	if callbackData == "back" {
		flow.Back(ctx, chatID, state, reply, callbackData)
		return
	}

	flow.HandleCallback(ctx, chatID, state, reply, callback.Data)
}

