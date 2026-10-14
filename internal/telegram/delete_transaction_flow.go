package telegram

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/yudadh/finance-bot-tracker-api/internal/service"
)

type deleteTransactionService interface {
	FindTransactionsByUserID(
		ctx context.Context,
		userID uint64,
		startTransactionDate time.Time,
		endTransactionDate time.Time,
	) ([]service.TransactionResult, error)

	HardDeleteByID(
		ctx context.Context,
		id uint64,
		userID uint64,
	) (*service.DeleteTransactionResult, error)
}

type DeleteTransactionFlow struct {
	transactionService deleteTransactionService
	conversations *ConversationStore
}

func NewDeleteTransactionFlow(
	transactionService deleteTransactionService,
	conversations *ConversationStore,
) *DeleteTransactionFlow {
	return &DeleteTransactionFlow{
		transactionService: transactionService,
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
		f.handleDateSelection(ctx, chatID, reply, state, text)
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
		f.handleTransactionSelected(ctx, chatID, reply, state, callbackData)
	case StepConfirm:
		f.handleConfirmDeleteTransaction(ctx, chatID, reply, state, callbackData)
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

func (f *DeleteTransactionFlow) handleDateSelection(
	ctx context.Context,
	chatID int64,
	reply Reply,
	state *ConversationState,
	text string,
) {
	deleteState := state.DeleteState
	if deleteState == nil {
		reply.Text(ctx, chatID, InternalErrorMessage())
		return
	}

	date, err := time.Parse("02-01-2006", text)
	if err != nil {
		reply.Text(ctx, chatID, "format tanggal tidak valid, silahkan coba lagi dengan format (DD-MM-YYYY)")
		return
	}

	result, err := f.transactionService.FindTransactionsByUserID(ctx, state.UserID, date, date)
	if err != nil {
		reply.Text(ctx, chatID, InternalErrorMessage())
		return
	}

	state, ok := f.conversations.Get(chatID)
	if !ok {
		reply.Text(ctx, chatID, cancelCommandMessage(state.Type))
		return
	}

	deleteState.DateText = text
	state.Step = StepSelectTransaction
	state.ExpiredAt = state.ExpiredAt.Add(2 * time.Minute)

	reply.Inline(
		ctx,
		chatID,
		"Pilih salah satu transaksi",
		createTransactionKeyboardMarkup(result, state.Type),
	)
}

func (f *DeleteTransactionFlow) handleTransactionSelected(
	ctx context.Context,
	chatID int64,
	reply Reply,
	state *ConversationState,
	callbackData string,
) {
	prefix, stringID, ok := strings.Cut(callbackData, ":")
	if !ok || prefix != "delete" || stringID == "" {
		reply.Text(ctx, chatID, InternalErrorMessage())
		return
	}

	deleteState := state.DeleteState
	if deleteState == nil {
		reply.Text(ctx, chatID, InternalErrorMessage())
		return
	}

	id, err := strconv.ParseUint(stringID, 10, 64)
	if err != nil {
		reply.Text(ctx, chatID, InternalErrorMessage())
		return
	}

	deleteState.TransactionID = id
	state.Step = StepConfirm
	state.ExpiredAt = state.ExpiredAt.Add(2 *time.Minute)

	reply.Inline(ctx, chatID, "Anda yakin ingin menghapus transaksi?", confirmationKeyboardMarkup())
}

func (f *DeleteTransactionFlow) handleConfirmDeleteTransaction(
	ctx context.Context,
	chatID int64,
	reply Reply,
	state *ConversationState,
	callbackData string,
) {
	if callbackData != "save" {
		return
	}

	deleteState := state.DeleteState
	if deleteState == nil {
		reply.Text(ctx, chatID, InternalErrorMessage())
		return
	}

	_, err := f.transactionService.HardDeleteByID(ctx, deleteState.TransactionID, state.UserID)
	if err != nil {
		reply.Text(ctx, chatID, InternalErrorMessage())
		return
	}

	reply.Text(ctx, chatID, deleteTransactionSuccessMessage())
}

