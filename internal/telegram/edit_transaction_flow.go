package telegram

import (
	"context"
	"errors"
	"log/slog"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/yudadh/finance-bot-tracker-api/internal/parser"
	"github.com/yudadh/finance-bot-tracker-api/internal/service"
)

type editTransactionService interface {
	FindTransactionsByUserID(
		ctx context.Context,
		userID uint64,
		startTransactionDate time.Time,
		endTransactionDate time.Time,
	) ([]service.TransactionResult, error)

	FindTransactionByID(
		ctx context.Context,
		id uint64,
	) (*service.TransactionResult, error)

	ParseEditTransactionFromText(
		ctx context.Context,
		input service.ParseEditTransactionFromTextInput,
	) (*service.ParseEditTransactionFromTextResult, error)

	UpdateTransactionFromChat(
		ctx context.Context,
		input service.UpdateTransactionFromChatInput,
	) (*service.UpdateTransactionFromChatResult, error)
}

type EditTransactionFlow struct {
	transactionService editTransactionService
	conversations      *ConversationStore
	logger             *slog.Logger
}

func NewEditTransactionFlow(
	transactionService editTransactionService,
	conversations *ConversationStore,
	logger *slog.Logger,
) *EditTransactionFlow {
	return &EditTransactionFlow{
		transactionService: transactionService,
		conversations:      conversations,
		logger:             logger,
	}
}

func (f *EditTransactionFlow) Start(
	ctx context.Context,
	chatID int64,
	state *ConversationState,
	reply Reply,
) {
	f.conversations.Set(chatID, state)

	reply.Text(
		ctx,
		chatID,
		handleDateSelectionMessage(),
	)
}

func (f *EditTransactionFlow) HandleMessage(
	ctx context.Context,
	chatID int64,
	state *ConversationState,
	reply Reply,
	text string,
) {
	switch state.Step {
	case StepSelectDate:
		f.handleDateSelection(ctx, chatID, reply, state, text)
	case StepInputValue:
		f.handleInputValue(ctx, chatID, reply, state, text)
	}

}

func (f *EditTransactionFlow) HandleCallback(
	ctx context.Context,
	chatID int64,
	state *ConversationState,
	reply Reply,
	callbackData string,
) {
	switch state.Step {
	case StepSelectTransaction:
		f.handleTransactionSelected(ctx, chatID, reply, state, callbackData)
	case StepSelectField:
		f.handleFieldSelection(ctx, chatID, reply, state, callbackData)
	case StepConfirm:
		f.handleConfirmEditTransaction(ctx, chatID, reply, state, callbackData)
	}
}

func (f *EditTransactionFlow) Back(
	ctx context.Context,
	chatID int64,
	state *ConversationState,
	reply Reply,
	callbackData string,
) {
	editState := state.EditState
	if editState == nil {
		reply.Text(ctx, chatID, InternalErrorMessage())
		return
	}

	switch state.Step {
	case StepSelectDate:
		f.conversations.Delete(chatID)
		reply.Text(ctx, chatID, cancelCommandMessage(state.Type))
	case StepSelectTransaction:
		editState.TransactionID = 0
		editState.Field = ""
		editState.NewValue = nil
		state.Step = StepSelectDate
		reply.Text(
			ctx,
			chatID,
			handleDateSelectionMessage(),
		)
	case StepSelectField:
		f.handleDateSelection(ctx, chatID, reply, state, editState.DateText)
	case StepInputValue:
		state.Step = StepSelectField
		reply.Inline(ctx, chatID, "Pilih bagian yang ingin di edit", selectFieldToEditKeyboardMarkup())
	case StepConfirm:
		state.Step = StepInputValue
		reply.Text(ctx, chatID, editTransactionInputMessage(editState.Field))
	}
}

func (f *EditTransactionFlow) handleDateSelection(
	ctx context.Context,
	chatID int64,
	reply Reply,
	state *ConversationState,
	text string,
) {
	editState := state.EditState
	if editState == nil {
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

	state.Step = StepSelectTransaction
	state.ExpiredAt = state.ExpiredAt.Add(2 * time.Minute)
	editState.DateText = text

	reply.Inline(
		ctx,
		chatID,
		"Pilih salah satu transaksi",
		createTransactionKeyboardMarkup(result, state.Type),
	)
}

func (f *EditTransactionFlow) handleTransactionSelected(
	ctx context.Context,
	chatID int64,
	reply Reply,
	state *ConversationState,
	callbackData string,
) {
	transactionId, err := strconv.ParseUint(strings.Split(callbackData, ":")[1], 10, 64)
	if err != nil || transactionId == 0 {
		reply.Text(ctx, chatID, InternalErrorMessage())
		return
	}

	_, err = f.transactionService.FindTransactionByID(ctx, transactionId)
	if err != nil {
		reply.Text(ctx, chatID, InternalErrorMessage())
		return
	}

	state.EditState.TransactionID = transactionId
	state.Step = StepSelectField
	state.ExpiredAt = state.ExpiredAt.Add(2 * time.Minute)

	reply.Inline(ctx, chatID, "Pilih bagian yang ingin di edit", selectFieldToEditKeyboardMarkup())
}

func (f *EditTransactionFlow) handleFieldSelection(
	ctx context.Context,
	chatID int64,
	reply Reply,
	state *ConversationState,
	callbackData string,
) {
	prefix, field, ok := strings.Cut(callbackData, ":")
	if !ok || prefix != "edit" || field == "" {
		reply.Text(ctx, chatID, InternalErrorMessage())
	}

	editState := state.EditState
	if editState == nil {
		reply.Text(ctx, chatID, InternalErrorMessage())
		return
	}

	fieldType := parser.EditField(field)
	authorizedField := []parser.EditField{
		parser.EditFieldDescription,
		parser.EditFieldAmount,
		parser.EditFieldBoth,
	}
	if !slices.Contains(authorizedField, fieldType) {
		reply.Text(ctx, chatID, InternalErrorMessage())
		return
	}

	editState.Field = string(fieldType)
	state.Step = StepInputValue
	state.ExpiredAt = state.ExpiredAt.Add(2 * time.Minute)

	reply.Text(ctx, chatID, editTransactionInputMessage(field))
}

func (f *EditTransactionFlow) handleInputValue(
	ctx context.Context,
	chatID int64,
	reply Reply,
	state *ConversationState,
	text string,
) {
	editState := state.EditState
	if editState == nil {
		reply.Text(ctx, chatID, InternalErrorMessage())
		return
	}

	input := service.ParseEditTransactionFromTextInput{
		TransactionID: editState.TransactionID,
		UserID:        state.UserID,
		Text:          text,
		Currency:      "IDR",
		FieldType:     editState.Field,
	}

	intent, err := f.transactionService.ParseEditTransactionFromText(ctx, input)
	if err != nil {
		switch {
		case errors.Is(err, parser.ErrAmountNotFound),
			errors.Is(err, parser.ErrInvalidAmount):
			reply.Text(ctx, chatID, editAmountErrorMessage())
			return
		}

		f.logger.ErrorContext(ctx, "failed to parse edit transaction from text",
			"chat_id", chatID,
			"user_id", state.UserID,
			"transaction_id", editState.TransactionID,
			"field", editState.Field,
			"error", err,
		)
		reply.Text(ctx, chatID, InternalErrorMessage())
		return
	}

	editState.NewValue = intent.EditTransactionIntent
	state.Step = StepConfirm
	state.ExpiredAt = state.ExpiredAt.Add(2 * time.Minute)

	reply.Inline(ctx, chatID, editTransactionConfirmationMessage(editState.Field, intent.EditTransactionIntent), confirmationKeyboardMarkup())
}

func (f *EditTransactionFlow) handleConfirmEditTransaction(
	ctx context.Context,
	chatID int64,
	reply Reply,
	state *ConversationState,
	callbackData string,
) {
	if callbackData != "save" {
		return
	}

	editState := state.EditState
	if editState == nil {
		reply.Text(ctx, chatID, InternalErrorMessage())
		return
	}

	_, err := f.transactionService.UpdateTransactionFromChat(ctx, service.UpdateTransactionFromChatInput{
		TransactionID: editState.TransactionID,
		UserID:        state.UserID,
		Value:         editState.NewValue,
		FieldType:     editState.Field,
	})

	if err != nil {
		reply.Text(ctx, chatID, InternalErrorMessage())
		return
	}

	f.conversations.Delete(chatID)
	reply.Text(ctx, chatID, editTransactionSuccessMessage())
}
