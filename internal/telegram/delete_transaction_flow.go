package telegram

import (
	"context"
)

type DeleteTransactionFlow struct {
	conversations *ConversationStore
}

func NewDeleteTransactionFlow(
	conversations *ConversationStore,
) *DeleteTransactionFlow {
	return &DeleteTransactionFlow{
		conversations: conversations,
	}
}

func (f *DeleteTransactionFlow) Start(
	ctx context.Context,
	chatID int64,
	state *ConversationState,
	reply Reply,
) {
	f.conversations.Set(chatID, state)
	
	reply.Text(
		ctx, 
		chatID, 
		"Masukkan tanggal transaksi dari data yang akan di hapus dengan format (DD-MM-YYYY):",
	)
}

func (f *DeleteTransactionFlow) HandleMessage(
	ctx context.Context,
	chatID int64,
	state *ConversationState,
	reply Reply,
	text string,
) {
	switch state.Step {
	case StepSelectDate:
		f.handleSelectDate(ctx, chatID, reply, text)
	}
}

func (f *DeleteTransactionFlow) HandleCallback(
	ctx context.Context,
	chatID int64,
	state *ConversationState,
	reply Reply,
	callbackData string,
) {
	switch state.Step {
	case StepSelectTransaction:
		f.handleTransactionSelected(ctx, chatID, reply, callbackData)
	case StepConfirm:
		f.handleConfirmDeleteTransaction(ctx, chatID, reply, callbackData)
	}
}

func (f *DeleteTransactionFlow) Back(
	ctx context.Context,
	chatID int64,
	state *ConversationState,
	reply Reply,
	callbackData string,
) {
	switch state.Step {
	case StepSelectDate:
		f.conversations.Delete(chatID)
		reply.Text(ctx, chatID, cancelCommandMessage(state.Type))
	case StepSelectTransaction:
		state.Step = StepSelectDate
	case StepConfirm:
		state.Step = StepInputValue
	}
}

func (f *DeleteTransactionFlow) handleSelectDate(
	ctx context.Context,
	chatID int64,
	reply Reply,
	text string,
) {

}

func (f *DeleteTransactionFlow) handleTransactionSelected(
	ctx context.Context,
	chatID int64,
	reply Reply,
	callbackData string,
) {

}

func (f *DeleteTransactionFlow) handleConfirmDeleteTransaction(
	ctx context.Context,
	chatID int64,
	reply Reply,
	callbackData string,
) {

}

