package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yudadh/finance-bot-tracker-api/internal/domain"
	"github.com/yudadh/finance-bot-tracker-api/internal/service"
)

type TransactionService interface {
	FindTransactionsByUserID(
		ctx context.Context,
		userID uint64,
		startTransactionDate time.Time,
		endTransactionDate time.Time,
	) ([]service.TransactionResult, error)
}

type TransactionHandler struct {
	transactionService TransactionService
}

func NewTransactionHandler(transactionService TransactionService) *TransactionHandler {
	return &TransactionHandler{
		transactionService: transactionService,
	}
}

func parseDateToUTC(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}

	date, err := time.ParseInLocation(
		time.DateOnly,
		value,
		time.UTC,
	)
	if err != nil {
		return nil, err
	}

	return &date, nil
}

func parsePositiveID(value string) (uint64, error) {
	result, err := strconv.ParseUint(value, 10, 64)
	if err != nil || result == 0 {
		return 0, domain.ErrInvalidID
	}

	return result, nil
}

func (h *TransactionHandler) GetTransactionsByUserPeriod(c *gin.Context) error {
	userID, err := parsePositiveID(c.Param("id"))
	if err != nil {
		return err
	}

	var query GetTransactionsByUserPeriodQuery

	if err := c.ShouldBindQuery(&query); err != nil {
		return err
	}

	startDate, err := parseDateToUTC(query.StartDate)
	if err != nil {
		return err
	}

	endDate, err := parseDateToUTC(query.EndDate)
	if err != nil {
		return err
	}

	result, err := h.transactionService.FindTransactionsByUserID(
		c.Request.Context(),
		userID,
		*startDate,
		*endDate,
	)

	if err != nil {
		return err
	}

	data := make([]GetTransactionsByUserPeriodResponse, 0, len(result))
	for _, res := range result {
		userTransaction := GetTransactionsByUserPeriodResponse{
			ID:              res.ID,
			UserID:          res.UserID,
			Type:            string(res.Type),
			Amount:          res.Amount,
			Currency:        res.Currency,
			Description:     res.Description,
			TransactionDate: res.TransactionDate,
			Source:          string(res.Source),
		}

		if res.Category != nil {
			userTransaction.Category = &categoryTransactionResponse{
				ID:   res.Category.ID,
				Name: res.Category.Name,
			}
		}

		data = append(data, userTransaction)
	}

	ResponseSuccess(c, http.StatusOK, "user transactions fetched successfully", data)
	return nil
}
