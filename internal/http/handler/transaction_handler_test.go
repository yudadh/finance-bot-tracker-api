package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yudadh/finance-bot-tracker-api/internal/domain"
	"github.com/yudadh/finance-bot-tracker-api/internal/service"
)

type fakeTransactionService struct {
	calls      int
	userID     uint64
	startDate  time.Time
	endDate    time.Time
	requestCtx context.Context
	result     []service.TransactionResult
	err        error
}

func (f *fakeTransactionService) FindTransactionsByUserID(
	ctx context.Context,
	userID uint64,
	startDate time.Time,
	endDate time.Time,
) ([]service.TransactionResult, error) {
	f.calls++
	f.requestCtx = ctx
	f.userID = userID
	f.startDate = startDate
	f.endDate = endDate
	return f.result, f.err
}

func newTransactionHandlerRouter(transactionService *fakeTransactionService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	router := gin.New()
	h := NewTransactionHandler(transactionService)
	router.GET("/transactions/users/:id", Handle(logger, h.GetTransactionsByUserPeriod))
	return router
}

type transactionResponseEnvelope struct {
	Error   bool                                  `json:"error"`
	Message string                                `json:"message"`
	Data    []GetTransactionsByUserPeriodResponse `json:"data"`
}

func TestGetTransactionsByUserPeriod_SuccessMapsRequestAndResponse(t *testing.T) {
	requestContextKey := struct{}{}
	requestContextValue := "request-value"
	requestContext := context.WithValue(context.Background(), requestContextKey, requestContextValue)
	transactionDate := time.Date(2026, 9, 5, 0, 0, 0, 0, time.UTC)
	category := &service.TransactionCategory{ID: 7, Name: "Food"}

	transactionService := &fakeTransactionService{
		result: []service.TransactionResult{
			{
				ID:              100,
				UserID:          42,
				Type:            domain.TransactionTypeExpense,
				Amount:          12500,
				Currency:        "IDR",
				Description:     "lunch",
				TransactionDate: transactionDate,
				Source:          domain.TransactionSourceTelegram,
				Category:        category,
			},
			{
				ID:              101,
				UserID:          42,
				Type:            domain.TransactionTypeIncome,
				Amount:          500000,
				Currency:        "IDR",
				Description:     "salary",
				TransactionDate: transactionDate,
				Source:          domain.TransactionSourceAdmin,
			},
		},
	}
	router := newTransactionHandlerRouter(transactionService)

	request := httptest.NewRequest(
		http.MethodGet,
		"/transactions/users/42?start_date=2026-09-01&end_date=2026-09-10",
		nil,
	).WithContext(requestContext)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", response.Code, response.Body.String())
	}
	var body transactionResponseEnvelope
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Error || body.Message != "user transactions fetched successfully" {
		t.Fatalf("unexpected response envelope: %+v", body)
	}
	if len(body.Data) != 2 {
		t.Fatalf("expected 2 transactions, got %d", len(body.Data))
	}
	if transactionService.calls != 1 || transactionService.userID != 42 {
		t.Fatalf("unexpected service call: calls=%d userID=%d", transactionService.calls, transactionService.userID)
	}
	expectedStart := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	expectedEnd := time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC)
	if !transactionService.startDate.Equal(expectedStart) || !transactionService.endDate.Equal(expectedEnd) {
		t.Fatalf("unexpected service dates: start=%v end=%v", transactionService.startDate, transactionService.endDate)
	}
	if transactionService.requestCtx.Value(requestContextKey) != requestContextValue {
		t.Fatal("expected request context to be passed to service")
	}

	got := body.Data[0]
	if got.ID != 100 || got.UserID != 42 || got.Type != "expense" || got.Amount != 12500 || got.Currency != "IDR" || got.Description != "lunch" || got.Source != "telegram" {
		t.Fatalf("unexpected mapped response: %+v", got)
	}
	if !got.TransactionDate.Equal(transactionDate) {
		t.Fatalf("expected transaction date %v, got %v", transactionDate, got.TransactionDate)
	}
	if got.Category == nil || got.Category.ID != 7 || got.Category.Name != "Food" {
		t.Fatalf("unexpected category response: %+v", got.Category)
	}
	if body.Data[1].Category != nil {
		t.Fatalf("expected uncategorized transaction to have nil category, got %+v", body.Data[1].Category)
	}
}

func TestGetTransactionsByUserPeriod_RequiredDates(t *testing.T) {
	tests := []struct {
		name  string
		query string
	}{
		{name: "missing start date", query: "?end_date=2026-09-10"},
		{name: "missing end date", query: "?start_date=2026-09-01"},
		{name: "missing both dates", query: ""},
		{name: "invalid start date", query: "?start_date=not-a-date&end_date=2026-09-10"},
		{name: "invalid end date", query: "?start_date=2026-09-01&end_date=not-a-date"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			transactionService := &fakeTransactionService{}
			router := newTransactionHandlerRouter(transactionService)
			request := httptest.NewRequest(http.MethodGet, "/transactions/users/42"+test.query, nil)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			if response.Code != http.StatusUnprocessableEntity {
				t.Fatalf("expected status 422, got %d: %s", response.Code, response.Body.String())
			}
			if transactionService.calls != 0 {
				t.Fatal("expected service not to be called")
			}
		})
	}
}

func TestGetTransactionsByUserPeriod_InvalidUserIDReturnsBadRequest(t *testing.T) {
	for _, userID := range []string{"abc", "0", "-1"} {
		t.Run(userID, func(t *testing.T) {
			transactionService := &fakeTransactionService{}
			router := newTransactionHandlerRouter(transactionService)
			request := httptest.NewRequest(
				http.MethodGet,
				"/transactions/users/"+userID+"?start_date=2026-09-01&end_date=2026-09-10",
				nil,
			)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			if response.Code != http.StatusBadRequest {
				t.Fatalf("expected status 400, got %d: %s", response.Code, response.Body.String())
			}
			if transactionService.calls != 0 {
				t.Fatal("expected service not to be called")
			}
		})
	}
}

func TestGetTransactionsByUserPeriod_ServiceErrors(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		statusCode int
	}{
		{
			name:       "invalid date range",
			err:        domain.ErrInvalidDateRange,
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "user not found",
			err:        &domain.NotFoundError{Resource: domain.ResourceUser},
			statusCode: http.StatusNotFound,
		},
		{
			name:       "unexpected error",
			err:        errors.New("database unavailable"),
			statusCode: http.StatusInternalServerError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			transactionService := &fakeTransactionService{err: test.err}
			router := newTransactionHandlerRouter(transactionService)
			request := httptest.NewRequest(
				http.MethodGet,
				"/transactions/users/42?start_date=2026-09-01&end_date=2026-09-10",
				nil,
			)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			if response.Code != test.statusCode {
				t.Fatalf("expected status %d, got %d: %s", test.statusCode, response.Code, response.Body.String())
			}
			if transactionService.calls != 1 {
				t.Fatalf("expected service to be called once, got %d", transactionService.calls)
			}
		})
	}
}

func TestGetTransactionsByUserPeriod_EmptyResultReturnsEmptyArray(t *testing.T) {
	transactionService := &fakeTransactionService{result: []service.TransactionResult{}}
	router := newTransactionHandlerRouter(transactionService)
	request := httptest.NewRequest(
		http.MethodGet,
		"/transactions/users/42?start_date=2026-09-01&end_date=2026-09-10",
		nil,
	)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	var body struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !bytes.Equal(bytes.TrimSpace(body.Data), []byte("[]")) {
		t.Fatalf("expected data to be [], got %s", body.Data)
	}
}
