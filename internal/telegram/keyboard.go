package telegram

import (
	"fmt"

	"github.com/go-telegram/bot/models"
	"github.com/yudadh/finance-bot-tracker-api/internal/parser"
	"github.com/yudadh/finance-bot-tracker-api/internal/service"
)

func createTransactionKeyboardMarkup(rows []service.TransactionResult, flowType ConversationType) models.InlineKeyboardMarkup {
	var inlineKeyboard [][]models.InlineKeyboardButton

	var prefix string
	switch flowType {
	case ConversationEditTransaction:
		prefix = "edit"

	case ConversationDeleteTransaction:
		prefix = "delete"
	}

	for _, row := range rows {
		inlineKeyboard = append(inlineKeyboard, []models.InlineKeyboardButton{
			{
				Text:         fmt.Sprintf("%s - %d", row.Description, row.Amount),
				CallbackData: fmt.Sprintf("%s:%d", prefix, row.ID),
			},
		})
	}

	inlineKeyboard = append(inlineKeyboard, []models.InlineKeyboardButton{
		{
			Text:         "Kembali",
			CallbackData: "back",
		},
		{
			Text:         "Batal",
			CallbackData: "cancel",
		},
	})

	return models.InlineKeyboardMarkup{InlineKeyboard: inlineKeyboard}
}

func selectFieldToEditKeyboardMarkup() models.InlineKeyboardMarkup {
	inlineKeyboard := [][]models.InlineKeyboardButton{
		{
			{
				Text:         "Deskripsi",
				CallbackData: fmt.Sprintf("edit:%s", parser.EditFieldDescription),
			},
		},
		{
			{
				Text:         "Jumlah Nominal",
				CallbackData: fmt.Sprintf("edit:%s", parser.EditFieldAmount),
			},
		},
		{
			{
				Text:         "Deskripsi & Jumlah Nominal",
				CallbackData: fmt.Sprintf("edit:%s", parser.EditFieldBoth),
			},
		},
		{
			{
				Text:         "Kembali",
				CallbackData: "back",
			},
			{
				Text:         "Batal",
				CallbackData: "cancel",
			},
		},
	}

	return models.InlineKeyboardMarkup{InlineKeyboard: inlineKeyboard}
}

func confirmationKeyboardMarkup() models.InlineKeyboardMarkup {
	inlineKeyboard := [][]models.InlineKeyboardButton{
		{
			{
				Text:         "Kembali",
				CallbackData: "back",
			},
		},
		{
			{
				Text:         "Simpan",
				CallbackData: "save",
			},
		},
		{
			{
				Text:         "Batal",
				CallbackData: "cancel",
			},
		},
	}

	return models.InlineKeyboardMarkup{InlineKeyboard: inlineKeyboard}
}