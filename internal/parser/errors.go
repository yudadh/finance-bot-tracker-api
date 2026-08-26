package parser

import "errors"

var (
	ErrEmptyText = errors.New("empty text")
	ErrAmountNotFound = errors.New("amount not found")
	ErrInvalidAmount = errors.New("invalid amount")
	ErrInvalidBudgetText = errors.New("invalid budget text")
)