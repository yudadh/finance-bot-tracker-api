package middleware

import (
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
	"github.com/yudadh/finance-bot-tracker-api/internal/config"
	"github.com/yudadh/finance-bot-tracker-api/internal/domain"
	"github.com/yudadh/finance-bot-tracker-api/internal/security"
	"github.com/yudadh/finance-bot-tracker-api/internal/service"
)

type fakeAdminUserService struct {
	adminUser     *service.FindByIDResult
	findUserError error
	findUserCalls int
	requestedID   uint64
}

func TestAdminAuth_InactiveAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := testJWTConfig()
	token, err := security.GenerateAdminToken(cfg, 42, "admin@example.com")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	adminUserService := newFakeAdminUserService(nil)
	adminUserService.adminUser.Status = domain.AdminUserStatusInactive
	router := gin.New()
	handlerCalled := false
	router.Use(AdminAuth(cfg, adminUserService, logger))
	router.GET("/admin", func(c *gin.Context) {
		handlerCalled = true
		c.Status(http.StatusOK)
	})

	request := httptest.NewRequest(http.MethodGet, "/admin", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected status code %d, got %d", http.StatusForbidden, response.Code)
	}
	if handlerCalled {
		t.Fatal("expected protected route not called")
	}

	var body struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !body.Error || body.Message != "inactive account" {
		t.Fatalf("unexpected response: %+v", body)
	}
}

func TestAdminAuth_AdminLookupError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := testJWTConfig()
	token, err := security.GenerateAdminToken(cfg, 42, "admin@example.com")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	adminUserService := newFakeAdminUserService(errors.New("database unavailable"))
	router := gin.New()
	handlerCalled := false
	router.Use(AdminAuth(cfg, adminUserService, logger))
	router.GET("/admin", func(c *gin.Context) {
		handlerCalled = true
		c.Status(http.StatusOK)
	})

	request := httptest.NewRequest(http.MethodGet, "/admin", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("expected status code %d, got %d", http.StatusInternalServerError, response.Code)
	}
	if handlerCalled {
		t.Fatal("expected protected route not called")
	}

	var body struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !body.Error || body.Message != "internal server error" {
		t.Fatalf("unexpected response: %+v", body)
	}
}

func (f *fakeAdminUserService) FindByID(ctx context.Context, id uint64) (*service.FindByIDResult, error) {
	f.findUserCalls++
	f.requestedID = id
	return f.adminUser, f.findUserError
}

func testJWTConfig() config.JWTConfig {
	return config.JWTConfig{
		Secret: "test-secret",
		Issuer: "finance-bot-test",
		TTL:    time.Hour,
	}
}

func newFakeAdminUserService(err error) *fakeAdminUserService {
	return &fakeAdminUserService{
		adminUser: &service.FindByIDResult{
			ID:     42,
			Email:  "admin@example.com",
			Name:   "admin",
			Status: domain.AdminUserStatusActive,
		},
		findUserError: err,
	}
}

var logger = slog.New(slog.NewTextHandler(io.Discard, nil))

func TestAdminAuth_MissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handlerCalled := false
	cfg := testJWTConfig()
	adminUserService := newFakeAdminUserService(nil)

	router.Use(AdminAuth(cfg, adminUserService, logger))
	router.GET("/admin", func(c *gin.Context) {
		handlerCalled = true
		c.Status(http.StatusOK)
	})

	request := httptest.NewRequest(http.MethodGet, "/admin", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", response.Code)
	}

	if handlerCalled {
		t.Fatal("expected protected handler not to be called")
	}
}

func TestAdminAuth_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := testJWTConfig()
	token, err := security.GenerateAdminToken(cfg, 42, "token@example.com")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	router := gin.New()
	adminUserService := newFakeAdminUserService(nil)
	adminUserService.adminUser.Email = "db@example.com"

	router.Use(AdminAuth(cfg, adminUserService, logger))
	router.GET("/admin", func(c *gin.Context) {
		adminUserID, idExists := c.Get("admin_user_id")
		email, emailExists := c.Get("email")
		if !idExists || !emailExists || adminUserID != uint64(42) || email != "db@example.com" {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusOK)
	})

	request := httptest.NewRequest(http.MethodGet, "/admin", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	if adminUserService.findUserCalls != 1 || adminUserService.requestedID != 42 {
		t.Fatalf("unexpected FindByID call: calls=%d id=%d", adminUserService.findUserCalls, adminUserService.requestedID)
	}
}

func TestAdminAuth_MalformedHeader(t *testing.T) {
	tests := []struct {
		name   string
		header string
	}{
		{
			name:   "missing authorization header",
			header: "",
		},
		{
			name:   "missing bearer scheme",
			header: "eyJhbGciOiJIUzI1NiIs...",
		},
		{
			name:   "bearer without token",
			header: "Bearer",
		},
		{
			name:   "bearer with space but no token",
			header: "Bearer ",
		},
		{
			name:   "bearer with only whitespace token",
			header: "Bearer    ",
		},
		{
			name:   "wrong auth scheme",
			header: "Basic eyJhbGciOiJIUzI1NiIs...",
		},
		{
			name:   "missing space after bearer",
			header: "BearereyJhbGciOiJIUzI1NiIs...",
		},
		{
			name:   "extra text before bearer",
			header: "Token Bearer eyJhbGciOiJIUzI1NiIs...",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			cfg := testJWTConfig()
			router := gin.New()
			handlerCalled := false

			adminUserService := newFakeAdminUserService(nil)

			router.Use(AdminAuth(cfg, adminUserService, logger))
			router.GET("/admin", func(c *gin.Context) {
				handlerCalled = true
				c.Status(http.StatusOK)
			})

			request := httptest.NewRequest(http.MethodGet, "/admin", nil)
			request.Header.Set("Authorization", test.header)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			if response.Code != http.StatusUnauthorized {
				t.Fatalf("expected status code 401, got %d", response.Code)
			}

			if handlerCalled {
				t.Fatalf("expected protected handler not called")
			}
		})
	}
}

func TestAdminAuth_InvalidSignature(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := testJWTConfig()
	router := gin.New()
	handlerCalled := false

	token, err := security.GenerateAdminToken(cfg, 62, "admin@example.com")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	cfg.Secret = "different-secret"
	adminUserService := newFakeAdminUserService(nil)

	router.Use(AdminAuth(cfg, adminUserService, logger))
	router.GET("/admin", func(c *gin.Context) {
		handlerCalled = true
		c.Status(http.StatusOK)
	})

	request := httptest.NewRequest(http.MethodGet, "/admin", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status code 401, got %d", response.Code)
	}

	if handlerCalled {
		t.Fatalf("expected protected route not called")
	}
}

func TestAdminAuth_ExpiredToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := testJWTConfig()
	router := gin.New()
	handlerCalled := false

	cfg.TTL = -time.Minute
	token, err := security.GenerateAdminToken(cfg, 42, "admin@example.com")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	adminUserService := newFakeAdminUserService(nil)

	router.Use(AdminAuth(cfg, adminUserService, logger))
	router.GET("/admin", func(c *gin.Context) {
		handlerCalled = true
		c.Status(http.StatusOK)
	})

	request := httptest.NewRequest(http.MethodGet, "/admin", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status code 401, got %d", response.Code)
	}

	if handlerCalled {
		t.Fatalf("expected protected route not called")
	}
}

func TestAdminAuth_NotAnAuthorizedAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := testJWTConfig()
	router := gin.New()
	handlerCalled := false

	token, err := security.GenerateAdminToken(cfg, 42, "admin@example.com")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	adminUserService := newFakeAdminUserService(domain.ErrNotFound)

	router.Use(AdminAuth(cfg, adminUserService, logger))
	router.GET("/admin", func(c *gin.Context) {
		handlerCalled = true
		c.Status(http.StatusOK)
	})

	request := httptest.NewRequest(http.MethodGet, "/admin", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status code 401, got %d", response.Code)
	}

	if handlerCalled {
		t.Fatalf("expected protected route not called")
	}
}
