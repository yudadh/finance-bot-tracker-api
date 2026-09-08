package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/yudadh/finance-bot-tracker-api/internal/domain"
	"github.com/yudadh/finance-bot-tracker-api/internal/pagination"
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

type fakeUserService struct {
	findAllUsersCalls      int
	findAllusersQuery      service.FindAllUsersQuery
	findAllUsersResult     []service.FindAllUserResult
	findAllUsersPagination *pagination.Meta
	findAllUsersErr        error
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

func (f *fakeUserService) FindAll(_ context.Context, query service.FindAllUsersQuery) ([]service.FindAllUserResult, *pagination.Meta, error) {
	f.findAllUsersCalls++
	f.findAllusersQuery = query
	return f.findAllUsersResult, f.findAllUsersPagination, f.findAllUsersErr
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
		&fakeUserService{},
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
		&fakeUserService{},
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
	handler := NewUserAdminHandler(
		&fakeUserAdminService{loginErr: domain.ErrInvalidCredentials},
		&fakeUserService{},
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
		&fakeUserService{},
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
	handler := NewUserAdminHandler(
		adminService,
		&fakeUserService{},
		logger,
	)

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
	handler := NewUserAdminHandler(
		&fakeUserAdminService{
			registerErr: domain.ErrAlreadyExists,
		},
		&fakeUserService{},
		logger,
	)

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
	handler := NewUserAdminHandler(
		adminService,
		&fakeUserService{},
		logger,
	)

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
			handler := NewUserAdminHandler(
				adminService,
				&fakeUserService{},
				logger,
			)
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
	handler := NewUserAdminHandler(
		&fakeUserAdminService{findByIDErr: domain.ErrNotFound},
		&fakeUserService{},
		logger,
	)

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

func TestGetAllUsers_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	userService := &fakeUserService{findAllUsersErr: errors.New("database connection failed")}
	handler := NewUserAdminHandler(
		&fakeUserAdminService{},
		userService,
		logger,
	)
	router := gin.New()
	router.GET("/admin/users", Handle(logger, handler.GetAllUsers))

	query := url.Values{}
	query.Set("page", "1")
	query.Set("per_page", "10")

	request := httptest.NewRequest(http.MethodGet, "/admin/users?"+query.Encode(), nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("expected status code %d, got %d", http.StatusInternalServerError, response.Code)
	}
}

func TestGetAllUsers_InvalidQuery(t *testing.T) {
	tests := []struct {
		name  string
		query ListUsersQuery
	}{
		{
			name: "invalid page",
			query: ListUsersQuery{
				PerPage: 10,
				Page:    0,
			},
		},
		{
			name: "invalid per_page",
			query: ListUsersQuery{
				PerPage: 5,
				Page:    1,
			},
		},
		{
			name: "invalid status",
			query: ListUsersQuery{
				PerPage: 10,
				Page:    1,
				Status:  "invalid status",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			userService := &fakeUserService{}
			handler := NewUserAdminHandler(
				&fakeUserAdminService{},
				userService,
				logger,
			)
			router := gin.New()
			router.GET("/admin/users", Handle(logger, handler.GetAllUsers))

			query := url.Values{}
			query.Set("page", strconv.Itoa(test.query.Page))
			query.Set("per_page", strconv.Itoa(test.query.PerPage))
			query.Set("status", test.query.Status)

			request := httptest.NewRequest(http.MethodGet, "/admin/users?"+query.Encode(), nil)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			if response.Code != http.StatusUnprocessableEntity {
				t.Fatalf("expected status code %d, got %d", http.StatusUnprocessableEntity, response.Code)
			}
		})
	}
}

func isSlice(v any) bool {
	if v == nil {
		return false
	}

	return reflect.TypeOf(v).Kind() == reflect.Slice
}

func TestGetAllUsers_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	userService := &fakeUserService{
		findAllUsersResult: []service.FindAllUserResult{
			{
				ID:               1,
				TelegramID:       111111,
				TelegramUsername: "john_doe",
				FirstName:        "John",
				LastName:         "Doe",
				LanguageCode:     "id",
				Timezone:         "Asia/Makassar",
				Status:           domain.UserStatusActive,
			},
			{
				ID:               2,
				TelegramID:       222222,
				TelegramUsername: "fulan_doe",
				FirstName:        "Fulan",
				LastName:         "Doe",
				LanguageCode:     "id",
				Timezone:         "Asia/Makassar",
				Status:           domain.UserStatusInactive,
			},
		},
		findAllUsersPagination: &pagination.Meta{
			Page:       2,
			PerPage:    20,
			Total:      22,
			TotalPages: 2,
		},
	}
	handler := NewUserAdminHandler(
		&fakeUserAdminService{},
		userService,
		logger,
	)
	router := gin.New()
	router.GET("/admin/users", Handle(logger, handler.GetAllUsers))

	query := url.Values{}
	query.Set("page", "2")
	query.Set("per_page", "20")
	query.Set("status", "active")
	query.Set("search", "john")

	request := httptest.NewRequest(http.MethodGet, "/admin/users?"+query.Encode(), nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, response.Code)
	}

	var body ResponseWithMeta[[]ListUsersResponse, pagination.Meta]
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response %v", err)
	}

	if body.Error || body.Message != "users successfully fetched" {
		t.Fatalf("unexpected response envelope: %+v", body)
	}
	if !isSlice(body.Data) || len(body.Data) != 2 {
		t.Fatalf("expected two users, got %+v", body.Data)
	}
	if body.Data[0] != (ListUsersResponse{
		ID:               1,
		TelegramID:       111111,
		TelegramUsername: "john_doe",
		FirstName:        "John",
		LastName:         "Doe",
		LanguageCode:     "id",
		Timezone:         "Asia/Makassar",
		Status:           string(domain.UserStatusActive),
	}) {
		t.Fatalf("unexpected first user: %+v", body.Data[0])
	}
	if body.Data[1].Status != string(domain.UserStatusInactive) {
		t.Fatalf("expected second user status %q, got %q", domain.UserStatusInactive, body.Data[1].Status)
	}
	if body.Meta != *userService.findAllUsersPagination {
		t.Fatalf("unexpected pagination metadata: %+v", body.Meta)
	}

	if userService.findAllUsersCalls != 1 {
		t.Fatalf("expected FindAll to be called once, got %d", userService.findAllUsersCalls)
	}
	expectedQuery := service.FindAllUsersQuery{
		Param:  pagination.Params{Page: 2, PerPage: 20},
		Status: "active",
		Search: "john",
	}
	if userService.findAllusersQuery != expectedQuery {
		t.Fatalf("unexpected FindAll query: got %+v, want %+v", userService.findAllusersQuery, expectedQuery)
	}
}

func TestGetAllUsers_EmptyResult(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	userService := &fakeUserService{
		findAllUsersPagination: &pagination.Meta{
			Page:       1,
			PerPage:    10,
			Total:      0,
			TotalPages: 0,
		},
	}
	handler := NewUserAdminHandler(
		&fakeUserAdminService{},
		userService,
		logger,
	)
	router := gin.New()
	router.GET("/admin/users", Handle(logger, handler.GetAllUsers))

	request := httptest.NewRequest(http.MethodGet, "/admin/users", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, response.Code)
	}

	var body ResponseWithMeta[[]ListUsersResponse, pagination.Meta]
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Data == nil || len(body.Data) != 0 {
		t.Fatalf("expected an empty non-nil data array, got %#v", body.Data)
	}
	if body.Meta != *userService.findAllUsersPagination {
		t.Fatalf("unexpected pagination metadata: %+v", body.Meta)
	}

	expectedQuery := service.FindAllUsersQuery{
		Param: pagination.Params{Page: 1, PerPage: 10},
	}
	if userService.findAllusersQuery != expectedQuery {
		t.Fatalf("unexpected default query: %+v", userService.findAllusersQuery)
	}
}
