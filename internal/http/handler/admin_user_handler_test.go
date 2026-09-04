package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/yudadh/finance-bot-tracker-api/internal/domain"
	"github.com/yudadh/finance-bot-tracker-api/internal/service"
)

type fakeUserAdminService struct {
	loginToken    string
	loginErr      error
	loginCalls    int
	loginEmail    string
	loginPassword string

	registerResult *service.RegisterAdminUserResult
	registerErr    error
	registerCalls  int
	registerInput  service.RegisterAdminUserInput

	findByIDResult *service.FindByIDResult
	findByIDErr    error
	findByIDCalls  int
	findByIDValue  uint64
}

func (f *fakeUserAdminService) Register(_ context.Context, input service.RegisterAdminUserInput) (*service.RegisterAdminUserResult, error) {
	f.registerCalls++
	f.registerInput = input
	return f.registerResult, f.registerErr
}

func (f *fakeUserAdminService) Login(_ context.Context, email string, password string) (string, error) {
	f.loginCalls++
	f.loginEmail = email
	f.loginPassword = password
	return f.loginToken, f.loginErr
}

func (f *fakeUserAdminService) FindByID(_ context.Context, id uint64) (*service.FindByIDResult, error) {
	f.findByIDCalls++
	f.findByIDValue = id
	return f.findByIDResult, f.findByIDErr
}

func TestLogin_InvalidCredentialsReturnsUnauthorized(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{
			name: "unknown email",
			body: `{"email":"unknown@example.com","password":"password"}`,
		},
		{
			name: "wrong password",
			body: `{"email":"admin@example.com","password":"wrong-password"}`,
		},
	}

	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewUserAdminHandler(
		&fakeUserAdminService{loginErr: domain.ErrInvalidCredentials},
		logger,
	)

	router := gin.New()
	router.POST("/login", Handle(logger, handler.Login))

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(
				http.MethodPost,
				"/login",
				strings.NewReader(test.body),
			)
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			if response.Code != http.StatusUnauthorized {
				t.Fatalf("expected status 401, got %d", response.Code)
			}

			var body ErrorResponse
			if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
				t.Fatalf("decode response: %v", err)
			}

			if !body.Error || body.Message != "invalid email or password" {
				t.Fatalf("unexpected response: %+v", body)
			}
		})
	}

}

func TestLogin_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	token := "jwt-example-token"
	adminService := &fakeUserAdminService{loginToken: token}
	handler := NewUserAdminHandler(
		adminService,
		logger,
	)

	router := gin.New()
	router.POST("/login", Handle(logger, handler.Login))

	request := httptest.NewRequest(
		http.MethodPost,
		"/login",
		strings.NewReader(`{"email":"admin@example.com","password":"password"}`),
	)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, response.Code)
	}

	var body Response[LoginResponse]

	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body, %v", err)
	}

	if body.Data.Token == "" {
		t.Fatalf("expected token to not be empty")
	}

	if body.Data.Token != token {
		t.Fatalf("expected token %s, got %s", token, body.Data.Token)
	}

	if adminService.loginCalls != 1 {
		t.Fatalf("expected login service to be called once, got %d", adminService.loginCalls)
	}
	if adminService.loginEmail != "admin@example.com" || adminService.loginPassword != "password" {
		t.Fatalf("unexpected login input: email=%q password=%q", adminService.loginEmail, adminService.loginPassword)
	}
}

func TestRequestValidation_MalformedFormat(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{
			name: "missing email field",
			body: `{"password":"password"}`,
		},
		{
			name: "missing password field",
			body: `{"email":"admin@example.com"}`,
		},
		{
			name: "empty email",
			body: `{"email":"","password":"password"}`,
		},
		{
			name: "empty password",
			body: `{"email":"admin@example.com","password":""}`,
		},
		{
			name: "invalid email format",
			body: `{"email":"adminexample.com","password":""}`,
		},
		{
			name: "password less than 6 characters",
			body: `{"email":"admin@example.com","password":"passw"}`,
		},
	}

	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewUserAdminHandler(&fakeUserAdminService{
		loginErr: domain.ErrInvalidCredentials,
	}, logger)

	router := gin.New()
	router.POST("/login", Handle(logger, handler.Login))

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(
				http.MethodPost,
				"/login",
				strings.NewReader(test.body),
			)

			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			if response.Code != http.StatusUnprocessableEntity {
				t.Fatalf("expected status code 422, got %d", response.Code)
			}
		})
	}
}

func TestServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewUserAdminHandler(
		&fakeUserAdminService{
			loginErr: errors.New("database connection failed"),
		},
		logger,
	)

	router := gin.New()
	router.POST("/login", Handle(logger, handler.Login))

	request := httptest.NewRequest(
		http.MethodPost,
		"/login",
		strings.NewReader(`{"email":"admin@example.com","password":"password"}`),
	)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("expected status code %d, got %d", http.StatusInternalServerError, response.Code)
	}
}

func TestRegister_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	adminService := &fakeUserAdminService{
		registerResult: &service.RegisterAdminUserResult{
			ID:     42,
			Email:  "admin@example.com",
			Name:   "Admin",
			Status: domain.AdminUserStatusActive,
		},
	}
	handler := NewUserAdminHandler(adminService, logger)

	router := gin.New()
	router.POST("/register", Handle(logger, handler.Register))

	request := httptest.NewRequest(
		http.MethodPost,
		"/register",
		strings.NewReader(`{"email":"admin@example.com","password":"password","name":"Admin"}`),
	)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status code %d, got %d", http.StatusCreated, response.Code)
	}

	var body Response[RegisterAdminUserResponse]
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Error || body.Message != "user admin created successfully" {
		t.Fatalf("unexpected response envelope: %+v", body)
	}
	if body.Data.ID != 42 || body.Data.Email != "admin@example.com" || body.Data.Name != "Admin" || body.Data.Status != string(domain.AdminUserStatusActive) {
		t.Fatalf("unexpected response data: %+v", body.Data)
	}
	if adminService.registerCalls != 1 {
		t.Fatalf("expected register service to be called once, got %d", adminService.registerCalls)
	}
	if adminService.registerInput != (service.RegisterAdminUserInput{
		Email:    "admin@example.com",
		Password: "password",
		Name:     "Admin",
	}) {
		t.Fatalf("unexpected register input: %+v", adminService.registerInput)
	}
}

func TestRegister_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewUserAdminHandler(&fakeUserAdminService{
		registerErr: domain.ErrAlreadyExists,
	}, logger)

	router := gin.New()
	router.POST("/register", Handle(logger, handler.Register))

	request := httptest.NewRequest(
		http.MethodPost,
		"/register",
		strings.NewReader(`{"email":"admin@example.com","password":"password","name":"Admin"}`),
	)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf("expected status code %d, got %d", http.StatusConflict, response.Code)
	}
}

func TestGetMe_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	adminService := &fakeUserAdminService{
		findByIDResult: &service.FindByIDResult{
			ID:     42,
			Email:  "admin@example.com",
			Name:   "Admin",
			Status: domain.AdminUserStatusActive,
		},
	}
	handler := NewUserAdminHandler(adminService, logger)

	router := gin.New()
	router.GET("/me", Handle(logger, func(c *gin.Context) error {
		c.Set("admin_user_id", uint64(42))
		return handler.GetMe(c)
	}))

	request := httptest.NewRequest(http.MethodGet, "/me", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, response.Code)
	}

	var body Response[GetMeResponse]
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Error || body.Message != "user admin successfully fetched" {
		t.Fatalf("unexpected response envelope: %+v", body)
	}
	if body.Data.ID != 42 || body.Data.Email != "admin@example.com" || body.Data.Name != "Admin" || body.Data.Status != string(domain.AdminUserStatusActive) {
		t.Fatalf("unexpected response data: %+v", body.Data)
	}
	if adminService.findByIDCalls != 1 || adminService.findByIDValue != 42 {
		t.Fatalf("unexpected FindByID call: calls=%d id=%d", adminService.findByIDCalls, adminService.findByIDValue)
	}
}

func TestGetMe_AuthenticationContextErrors(t *testing.T) {
	tests := []struct {
		name  string
		value any
	}{
		{name: "missing context", value: nil},
		{name: "wrong context type", value: "42"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			adminService := &fakeUserAdminService{}
			handler := NewUserAdminHandler(adminService, logger)
			router := gin.New()
			router.GET("/me", Handle(logger, func(c *gin.Context) error {
				if test.value != nil {
					c.Set("admin_user_id", test.value)
				}
				return handler.GetMe(c)
			}))

			request := httptest.NewRequest(http.MethodGet, "/me", nil)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			if response.Code != http.StatusInternalServerError {
				t.Fatalf("expected status code %d, got %d", http.StatusInternalServerError, response.Code)
			}
			if adminService.findByIDCalls != 0 {
				t.Fatalf("expected FindByID not to be called, got %d calls", adminService.findByIDCalls)
			}
		})
	}
}

func TestGetMe_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewUserAdminHandler(&fakeUserAdminService{
		findByIDErr: domain.ErrNotFound,
	}, logger)

	router := gin.New()
	router.GET("/me", Handle(logger, func(c *gin.Context) error {
		c.Set("admin_user_id", uint64(42))
		return handler.GetMe(c)
	}))

	request := httptest.NewRequest(http.MethodGet, "/me", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status code %d, got %d", http.StatusNotFound, response.Code)
	}
}
