package handler

import "time"

type GetTransactionsByUserPeriodQuery struct {
	StartDate string `form:"start_date" binding:"required,datetime=2006-01-02"`
	EndDate   string `form:"end_date" binding:"required,datetime=2006-01-02"`
}

type categoryTransactionResponse struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

type GetTransactionsByUserPeriodResponse struct {
	ID              uint64                       `json:"id"`
	UserID          uint64                       `json:"user_id"`
	Type            string                       `json:"type"`
	Amount          int64                        `json:"amount"`
	Currency        string                       `json:"currency"`
	Description     string                       `json:"description"`
	TransactionDate time.Time                    `json:"transaction_date"`
	Source          string                       `json:"source"`
	Category        *categoryTransactionResponse `json:"category"`
}
