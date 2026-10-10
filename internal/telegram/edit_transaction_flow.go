package telegram

import (
	"context"
	"fmt"
	"time"

	"github.com/go-telegram/bot/models"
	"github.com/yudadh/finance-bot-tracker-api/internal/service"
)

type transactionFinder interface {
	FindTransactionsByUserID(
		ctx context.Context,
		userID uint64,
		startTransactionDate time.Time,
		endTransactionDate time.Time,
	) ([]service.TransactionResult, error)
}

type EditTransactionFlow struct {
	transactionService transactionFinder
	conversations *ConversationStore
}

func NewEditTransactionFlow(
	transactionService transactionFinder,
	conversations *ConversationStore,
) *EditTransactionFlow {
	return &EditTransactionFlow{
		transactionService: transactionService,
		conversations: conversations,
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
		"Masukkan tanggal transaksi dari data yang akan di edit dengan format (DD-MM-YYYY):",
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
		f.handleInputValue(ctx, chatID, reply, text)
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
		f.handleTransactionSelected(ctx, chatID, reply, callbackData)
	case StepSelectField:
		f.handleFieldSelection(ctx, chatID, reply, callbackData)
	case StepConfirm:
		f.handleConfirmEditTransaction(ctx, chatID, reply, callbackData)
	}
}

func (f *EditTransactionFlow) Back(
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
		reply.Text(
			ctx, 
			chatID, 
			"Masukkan tanggal transaksi dari data yang akan di edit dengan format (DD-MM-YYYY):",
		)
	case StepSelectField:
		state.Step = StepSelectTransaction
	case StepInputValue:
		state.Step = StepSelectField
	case StepConfirm:
		state.Step = StepInputValue
	}
}

func (f *EditTransactionFlow) handleDateSelection(
	ctx context.Context,
	chatID int64,
	reply Reply,
	state *ConversationState,
	text string,
) {
	date, err := time.Parse("02-01-2006", text)
	if err != nil {
		reply.Text(ctx, chatID, "format tanggal tidak valid, silahkan coba lagi")
		return
	}

	result, err := f.transactionService.FindTransactionsByUserID(ctx, state.UserID, date, date.AddDate(0, 0, 1))
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
	state.ExpiredAt = state.ExpiredAt.Add(2 *time.Minute)

	reply.Inline(
		ctx, 
		chatID, 
		"Pilih salah satu transaksi", 
		createTransactionKeyboardMarkup(result),
	)
}

func (f *EditTransactionFlow) handleTransactionSelected(
	ctx context.Context,
	chatID int64,
	reply Reply,
	callbackData string,
) {
	
}

func (f *EditTransactionFlow) handleFieldSelection(
	ctx context.Context,
	chatID int64,
	reply Reply,
	callbackData string,
) {
	
}

func (f *EditTransactionFlow) handleInputValue(
	ctx context.Context,
	chatID int64,
	reply Reply,
	text string,
) {
	
}

func (f *EditTransactionFlow) handleConfirmEditTransaction(
	ctx context.Context,
	chatID int64,
	reply Reply,
	callbackData string,
) {

}

func createTransactionKeyboardMarkup(rows []service.TransactionResult) models.InlineKeyboardMarkup {
	var inlineKeyboard [][]models.InlineKeyboardButton
	
	for _, row := range rows {
		inlineKeyboard = append(inlineKeyboard, []models.InlineKeyboardButton{
			{
				Text: fmt.Sprintf("%s %d - %s", row.Description, row.Amount, row.TransactionDate.Format("02-01-2006")),
				CallbackData: fmt.Sprintf("edit:%d", row.ID),
			},
		})
	}

	inlineKeyboard = append(inlineKeyboard, []models.InlineKeyboardButton{
		{
			Text: "Kembali",
			CallbackData: "back",
		},
		{
			Text: "Batal",
			CallbackData: "cancel",
		},
	})

	return models.InlineKeyboardMarkup{InlineKeyboard: inlineKeyboard}
}