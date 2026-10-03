package telegram

import "time"

type ConversationType string

const (
	ConversationEditTransaction   ConversationType = "edit_transaction"
	ConversationDeleteTransaction ConversationType = "delete_transaction"
)

type ConversationStep string

const (
	StepSelectTransaction ConversationStep = "select_transactions"
	StepSelectField       ConversationStep = "select_field"
	StepInputValue        ConversationStep = "input_value"
	StepConfirm           ConversationStep = "confirm"
	StepCancel            ConversationStep = "cancel"
)

type ConversationState struct {
	Type ConversationType
	Step ConversationStep

	TransactionID uint64
	Field         string
	NewValue      string

	ExpiredAt time.Time
}
