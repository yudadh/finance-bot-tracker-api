package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/yudadh/finance-bot-tracker-api/internal/domain"
	"github.com/yudadh/finance-bot-tracker-api/internal/pagination"
	"github.com/yudadh/finance-bot-tracker-api/internal/service"
)

type fakeCategoryService struct {
	createCategoryCalls int
	createCategoryInput service.CategoryInput
	createCategoryResult *service.CategoryResult
	createCategoryErr error

	findByIDCalls int
	categoryID uint64
	findByIDResult *service.CategoryResult
	findByIDErr error

	findAllCategoryCalls int
	findAllCategoryQuery service.CategoryQuery
	findAllCategoryResult []service.CategoryResult
	findAllCategoryResultTotal int64
	findAllCategoryErr error

	updateCategoryCalls int
	updateCategoryInput service.UpdateCategoryInput
	updateCategoryResult *service.CategoryResult
	updateCategoryErr error

	hardDeleteCategoryCalls int
	hardDeleteCategoryID uint64
	hardDeleteCategoryErr error
}

func (f *fakeCategoryService) Create(
	_ context.Context, 
	categoryInput service.CategoryInput,
) (*service.CategoryResult, error) {
	f.createCategoryCalls++
	f.createCategoryInput = categoryInput
	
	return f.createCategoryResult, f.createCategoryErr
}

func (f *fakeCategoryService) FindByID(
	_ context.Context, 
	id uint64,
) (*service.CategoryResult, error) {
	f.findByIDCalls++
	f.categoryID = id

	return f.findByIDResult, f.findByIDErr
}

func (f *fakeCategoryService) FindAll(
	_ context.Context, 
	query service.CategoryQuery,
) ([]service.CategoryResult, int64, error) {
	f.findAllCategoryCalls++
	f.findAllCategoryQuery = query

	return f.findAllCategoryResult, f.findAllCategoryResultTotal, f.findAllCategoryErr
}

func (f *fakeCategoryService) Update(
	_ context.Context, 
	data service.UpdateCategoryInput,
) (*service.CategoryResult, error) {
	f.updateCategoryCalls++

	return f.updateCategoryResult, f.updateCategoryErr
}

func (f *fakeCategoryService) HardDelete(
	_ context.Context,
	id uint64,
) error {
	f.hardDeleteCategoryCalls++
	f.hardDeleteCategoryID = id

	return f.hardDeleteCategoryErr
}

func TestCreateCategory_InvalidRequestBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	categoryService := &fakeCategoryService{
		createCategoryResult: &service.CategoryResult{
			ID: 1,
			Name: "food",
			Type: domain.TransactionTypeExpense,
			Keywords: []string{"makan", "ayam"},
			IsDefault: true,
		},
	}
	handler := NewCategoryHandler(categoryService)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	router := gin.New()
	router.POST("/api/category", Handle(logger, handler.Create))

	request := httptest.NewRequest(
		http.MethodPost, 
		"/api/category",
		 strings.NewReader(`{
		 	"name": "toy",
			"type": "expense",
			"keywords": [1, 2],
			"is_default": true
		 }`),
	)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected response code %d, got %d", http.StatusUnprocessableEntity, response.Code)
	}

}

func TestCreateCategory_ResponseShape(t *testing.T) {
	gin.SetMode(gin.TestMode)
	categoryService := &fakeCategoryService{
		createCategoryResult: &service.CategoryResult{
			ID: 1, 
			Name: "food",
			Type: domain.TransactionTypeExpense,
			Keywords: []string{"makan", "ayam"},
			IsDefault: true,
		},
	}
	handler := NewCategoryHandler(categoryService)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	router := gin.New()
	router.POST("/api/category", Handle(logger, handler.Create))

	request := httptest.NewRequest(
		http.MethodPost, 
		"/api/category",
		 strings.NewReader(`{
		 	"name": "food",
			"type": "expense",
			"keywords": ["makan", "ayam"],
			"is_default": true
		 }`),
	)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	var bodyField map[string]any

	if err := json.NewDecoder(response.Body).Decode(&bodyField); err != nil {
		t.Fatalf("decode body field error: %v", err)
	}

	if _, ok := bodyField["error"]; !ok {
		t.Fatal(`response body must contain "error" field`)
	}

	if _, ok := bodyField["message"]; !ok {
		t.Fatal(`response body must contain "message" field`)
	}

	if _, ok := bodyField["data"]; !ok {
		t.Fatal(`response body must contain "data" field`)
	}

	data := bodyField["data"].(map[string]any)

	if _, ok := data["id"]; !ok {
		t.Fatal(`response body data must contain "id" field`)
	}

	if _, ok := data["name"]; !ok {
		t.Fatal(`response body data must contain "name" field`)
	}

	if _, ok := data["type"]; !ok {
		t.Fatal(`response body data must contain "type" field`)
	}

	if _, ok := data["keywords"]; !ok {
		t.Fatal(`response body data must contain "keywords" field`)
	}

	if _, ok := data["is_default"]; !ok {
		t.Fatal(`response body data must contain "is_default" field`)
	}

}

func TestCreateCategory_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	categoryService := &fakeCategoryService{
		createCategoryResult: &service.CategoryResult{
			ID: 1,
			Name: "food",
			Type: domain.TransactionTypeExpense,
			Keywords: []string{"makan", "ayam"},
			IsDefault: true,
		},
	}
	handler := NewCategoryHandler(categoryService)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	router := gin.New()
	router.POST("/api/category", Handle(logger, handler.Create))

	request := httptest.NewRequest(
		http.MethodPost, 
		"/api/category",
		 strings.NewReader(`{
		 	"name": "food",
			"type": "expense",
			"keywords": ["makan", "ayam"],
			"is_default": true
		 }`),
	)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request) 

	var body Response[CategoryResponse]

	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode body error: %v", err)
	}

	if response.Code != http.StatusCreated {
		t.Fatalf("expected response code %d, got %d", http.StatusCreated, response.Code)
	}

	if body.Error || body.Message != "category successfully created" {
		t.Fatalf("unexpected response body %+v", body)
	}

	data := body.Data
	if data.ID != categoryService.createCategoryResult.ID {
		t.Fatalf("expected id %d, got %d", categoryService.createCategoryResult.ID, data.ID)
	}

	if data.Name != categoryService.createCategoryResult.Name {
		t.Fatalf("expected name %s, got %s", categoryService.createCategoryResult.Name, data.Name)
	}

	if data.Type != string(categoryService.createCategoryResult.Type) {
		t.Fatalf("expected type %s, got %s", categoryService.createCategoryResult.Type, data.Type)
	}

	if !slices.Equal(data.Keywords, categoryService.createCategoryResult.Keywords) {
		t.Fatalf("expected keywords %v, got %v", categoryService.createCategoryResult.Keywords, data.Keywords)
	}

	if data.IsDefault != categoryService.createCategoryResult.IsDefault {
		t.Fatalf("expected is_default %v, got %v", categoryService.createCategoryResult.IsDefault, data.IsDefault)
	}

}

func TestFindByCategoryID_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	categoryService := &fakeCategoryService{
		findByIDResult: &service.CategoryResult{
			ID: 1,
			Name: "food",
			Type: domain.TransactionTypeExpense,
			Keywords: []string{"makanan"},
			IsDefault: true,
		},
	}

	handler := NewCategoryHandler(categoryService)

	router.GET("/api/category/:id", Handle(logger, handler.FindByID))
	
	request := httptest.NewRequest(http.MethodGet, "/api/category/1", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected response code %d, got %d", http.StatusOK, response.Code)
	}

	var body Response[CategoryResponse]

	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode body error: %v", err)
	}

	data := body.Data
	if data.ID != categoryService.findByIDResult.ID {
		t.Fatalf("expected id %d, got %d", categoryService.findByIDResult.ID, data.ID)
	}

	if data.Name != categoryService.findByIDResult.Name {
		t.Fatalf("expected name %s, got %s", categoryService.findByIDResult.Name, data.Name)
	}

	if data.Type != string(categoryService.findByIDResult.Type) {
		t.Fatalf("expected type %s, got %s", categoryService.findByIDResult.Type, data.Type)
	}

	if !slices.Equal(data.Keywords, categoryService.findByIDResult.Keywords) {
		t.Fatalf("expected keywords %+v, got %+v", categoryService.findByIDResult.Keywords, data.Keywords)
	}

	if data.IsDefault != categoryService.findByIDResult.IsDefault {
		t.Fatalf("expected is_default %t, got %t", categoryService.findByIDResult.IsDefault, data.IsDefault)
	}
}

func TestFindByCategoryID_InvalidRequestParam(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	categoryService := &fakeCategoryService{}

	handler := NewCategoryHandler(categoryService)

	router.GET("/api/category:id", Handle(logger, handler.FindByID))
	
	request := httptest.NewRequest(http.MethodGet, "/api/category/0", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected response code %d, got %d", http.StatusBadRequest, response.Code)
	}

	if categoryService.findByIDCalls != 0 {
		t.Fatal("expected category service not to be called")
	}

	var body map[string]any

	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode body error: %v", err)
	}

	if _, ok := body["error"]; !ok && body["error"] != true {
		t.Fatal(`response body must contain "error" field with value true`)
	}

	if _, ok := body["message"]; !ok && body["message"] != "id must be a positive integer" {
		t.Fatal(`response body must contain "message" field with value "id must be a positive integer"`)
	}
}

func TestFindAllCategory_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	expectedData := []CategoryResponse{
		{
			ID: 1,
			Name: "food",
			Type: string(domain.TransactionTypeExpense),
			Keywords: []string{"makan", "kopi"},
			IsDefault: true,
		}, 
		{
			ID: 2,
			Name: "entertaiment",
			Type: string(domain.TransactionTypeExpense),
			Keywords: []string{"netflix", "bioskop"},
			IsDefault: true,
		},
	}
	
	categoryService := &fakeCategoryService{
		findAllCategoryResult: []service.CategoryResult{
			{ID: 1, Name: "food", Type: domain.TransactionTypeExpense, Keywords: []string{"makan", "kopi"}, IsDefault: true},
			{ID: 2, Name: "entertaiment", Type: domain.TransactionTypeExpense, Keywords: []string{"netflix", "bioskop"}, IsDefault: true},
		},
		findAllCategoryResultTotal: 2,
	}

	handler := NewCategoryHandler(categoryService)

	router.GET("/api/category", Handle(logger, handler.FindAll))
	
	request := httptest.NewRequest(http.MethodGet, "/api/category", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected response code %d, got %d", http.StatusOK, response.Code)
	}

	var body struct {
		Data []CategoryResponse `json:"data"`
		Meta *pagination.Meta `json:"meta,omitempty"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode body error: %v", err)
	}

	if body.Meta == nil {
		t.Fatal(`response body must contain "meta" field`)
	}
	if body.Data == nil {
		t.Fatal(`response body must contain "data" field`)
	}

	expectedMeta := pagination.NewMeta(pagination.Params{
		Page: 1,
		PerPage: 10,
	}, categoryService.findAllCategoryResultTotal)

	if !reflect.DeepEqual(body.Meta, &expectedMeta) {
		t.Fatalf("unexpected meta: got %+v, want %+v", body.Meta, &expectedMeta)
	}

	if !reflect.DeepEqual(body.Data, expectedData) {
		t.Fatalf("unexpected data: got %+v, want %+v", body.Data, expectedData)
	}
}

func TestFindAllCategory_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	
	categoryService := &fakeCategoryService{
		findAllCategoryErr: errors.New("database connection failed"),
	}

	handler := NewCategoryHandler(categoryService)

	router.GET("/api/category", Handle(logger, handler.FindAll))
	
	request := httptest.NewRequest(http.MethodGet, "/api/category", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("expected response code %d, got %d", http.StatusInternalServerError, response.Code)
	}

	var body struct {
		Error bool `json:"error"`
		Message string `json:"message"`
		Data []CategoryResponse `json:"data"`
		Meta *pagination.Meta `json:"meta,omitempty"`
	}

	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode body error: %v", err)
	}

	if !body.Error {
		t.Fatal("expected error to be true")
	}

	if body.Message != "internal server error" {
		t.Fatal(`expected message to be "internal server error"`)
	}

	if body.Meta != nil {
		t.Fatal(`response body must not contain "meta" field`)
	}

	if body.Data != nil {
		t.Fatal(`response body must not contain "data" field`)
	}

}

func TestHardDeleteCategory_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	categoryService := &fakeCategoryService{
		
	}
	handler := NewCategoryHandler(categoryService)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	router := gin.New()
	router.DELETE("/api/category/:id", Handle(logger, handler.HardDelete))

	request := httptest.NewRequest(http.MethodDelete, "/api/category/1", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected response code %d, got %d", http.StatusOK, response.Code)
	}

	var body struct {
		Error bool `json:"error"`
		Message string `json:"message"`
	}

	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode body error: %v", err)
	}
	

	if body.Error {
		t.Fatal("expected error to be false")
	}
	
	expectedMessage := "category successfully deleted permanently"
	if body.Message != expectedMessage {
		t.Fatalf("expected message %s, got %s", expectedMessage, body.Message)
	}
}


