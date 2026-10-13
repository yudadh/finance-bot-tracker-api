package telegram

import (
	"time"

	"github.com/yudadh/finance-bot-tracker-api/internal/parser"
)

type ConversationType string

const (
	ConversationEditTransaction   ConversationType = "edit_transaction"
	ConversationDeleteTransaction ConversationType = "delete_transaction"
)

type ConversationStep string

const (
	StepSelectDate        ConversationStep = "select_date"
	StepSelectTransaction ConversationStep = "select_transactions"
	StepSelectField       ConversationStep = "select_field"
	StepInputValue        ConversationStep = "input_value"
	StepConfirm           ConversationStep = "confirm"
	StepCancel            ConversationStep = "cancel"
)

type EditStateValue struct {
	TransactionID uint64
	DateText      string
	Field         string
	NewValue      *parser.EditTransactionIntent
}

type DeleteStateValue struct {
	TransactionID uint64
}

type ConversationState struct {
	Type ConversationType
	Step ConversationStep

	UserID      uint64
	TelegramID  int64
	Text        string
	EditState   *EditStateValue
	DeleteState *DeleteStateValue

	ExpiredAt time.Time
}
