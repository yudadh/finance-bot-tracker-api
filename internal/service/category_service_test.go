package service

import (
	"context"
	"errors"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/yudadh/finance-bot-tracker-api/internal/domain"
	"github.com/yudadh/finance-bot-tracker-api/internal/pagination"
	"github.com/yudadh/finance-bot-tracker-api/internal/repository"
)

/***
	Create(ctx context.Context, category *domain.Category) error
	FindByID(ctx context.Context, id uint64) (*domain.Category, error)
	RestoreByNameAndType(
		ctx context.Context,
		name string,
		transactionType domain.TransactionType,
	) (*domain.Category, error)
	FindAll(ctx context.Context, query repository.FindAllCategoriesQuery) ([]domain.Category, error)
	CountAll(ctx context.Context, query repository.FindAllCategoriesQuery) (*int64, error)
	Update(ctx context.Context, category *domain.Category) error
	HardDelete(ctx context.Context, category *domain.Category) error
**/

type fakeCategoryRepo struct {
	createCategoryInput *domain.Category
	createCategoryErr error

	id uint64
	findByIDResult *domain.Category
	findByIDErr error

	name string
	transactionType domain.TransactionType
	restoreByNameAndTypeResult *domain.Category
	restoreByNameAndTypeErr error

	findAllCategoryQuery repository.FindAllCategoriesQuery
	findAllCategoryResult []domain.Category
	findAllCategoryErr error

	countAllCategoryQuery repository.FindAllCategoriesQuery
	countAllCategoryResult *int64
	countAllCategoryErr error

	updateCategoryInput *domain.Category
	updateCategoryErr error

	hardDeleteCategoryInput *domain.Category
	hardDeleteCategoryErr error
}

func (f *fakeCategoryRepo) Create(
	_ context.Context, 
	category *domain.Category,
) error {
	f.createCategoryInput = category

	return f.createCategoryErr
}

func (f *fakeCategoryRepo) FindByID(
	_ context.Context,
	id uint64,
) (*domain.Category, error) {
	f.id = id

	return f.findByIDResult, f.findByIDErr
}

func (f *fakeCategoryRepo) RestoreByNameAndType(
	_ context.Context,
	name string,
	transctionType domain.TransactionType,
) (*domain.Category, error) {
	f.name = name
	f.transactionType = transctionType

	return f.restoreByNameAndTypeResult, f.restoreByNameAndTypeErr
}

func (f *fakeCategoryRepo) FindAll(
	_ context.Context,
	query repository.FindAllCategoriesQuery,
) ([]domain.Category, error) {
	f.findAllCategoryQuery = query

	return f.findAllCategoryResult, f.findAllCategoryErr
}

func (f *fakeCategoryRepo) CountAll(
	_ context.Context,
	query repository.FindAllCategoriesQuery,
) (*int64, error) {
	f.countAllCategoryQuery = query

	return f.countAllCategoryResult, f.countAllCategoryErr
}

func (f *fakeCategoryRepo) Update(
	_ context.Context,
	category *domain.Category,
) error {
	f.updateCategoryInput = category

	return f.updateCategoryErr
}

func (f *fakeCategoryRepo) HardDelete(
	_ context.Context,
	category *domain.Category,
) error {
	f.hardDeleteCategoryInput = category

	return f.hardDeleteCategoryErr
}

func TestCreateCategoryWithoutRestore_Success(t *testing.T) {
	ctx := context.Background()
	service := NewCategoryService(&fakeCategoryRepo{
		restoreByNameAndTypeErr: domain.ErrNotFound,
	})

	result, err := service.Create(ctx, CategoryInput{
		Name: "food",
		Type: domain.TransactionTypeExpense,
		Keywords: []string{"makan"},
		IsDefault: true,
	})

	if err != nil {
		t.Fatalf("create category error: %v", err)
	}

	if result.Name != "food" {
		t.Fatalf("expected name %s, got %s", "food", result.Name)
	}

	if result.Type != domain.TransactionTypeExpense {
		t.Fatalf("expected type %s, got %s", string(domain.TransactionTypeExpense), string(result.Type))
	}

	expectedKeywords := []string{"makan"}
	if len(result.Keywords) != 1 || !slices.Equal(result.Keywords, expectedKeywords) {
		t.Fatalf("expected keywords value %v, got %v", expectedKeywords, result.Keywords)
	}
	
	if !result.IsDefault {
		t.Fatalf("expected is_default %v, got %v", true, result.IsDefault)
	}
}
func TestCreateCategoryWithRestore_Success(t *testing.T) {
	ctx := context.Background()
	service := NewCategoryService(&fakeCategoryRepo{
		restoreByNameAndTypeResult: &domain.Category{
			ID: 1,
			Name: "food",
			Type: domain.TransactionTypeExpense,
			Keywords: []string{"makan"},
			IsDefault: true,
		},
	})

	result, err := service.Create(ctx, CategoryInput{
		Name: "food",
		Type: domain.TransactionTypeExpense,
		Keywords: []string{"makan"},
		IsDefault: true,
	})

	if err != nil {
		t.Fatalf("create category error: %v", err)
	}

	if result.ID != 1 {
		t.Fatalf("expected id %d, got %d", 1, result.ID)
	}

	if result.Name != "food" {
		t.Fatalf("expected name %s, got %s", "food", result.Name)
	}

	if result.Type != domain.TransactionTypeExpense {
		t.Fatalf("expected type %s, got %s", string(domain.TransactionTypeExpense), string(result.Type))
	}

	expectedKeywords := []string{"makan"}
	if len(result.Keywords) != 1 || !slices.Equal(result.Keywords, expectedKeywords) {
		t.Fatalf("expected keywords value %v, got %v", expectedKeywords, result.Keywords)
	}
	
	if !result.IsDefault {
		t.Fatalf("expected is_default %v, got %v", true, result.IsDefault)
	}
}

func TestCreateCategoryWithoutRestore_RepositoryError(t *testing.T) {
	ctx := context.Background()
	service := NewCategoryService(&fakeCategoryRepo{
		restoreByNameAndTypeErr: domain.ErrNotFound,
		createCategoryErr: errors.New("database connection failed"),
	})

	result, err := service.Create(ctx, CategoryInput{
		Name: "food",
		Type: domain.TransactionTypeExpense,
		Keywords: []string{"makan"},
		IsDefault: true,
	})

	if err == nil {
		t.Fatalf("expected error not to be nil")
	}

	if result != nil {
		t.Fatal("expected result to be nil")
	}
}

func TestCreateCategoryWithRestore_RepositoryError(t *testing.T) {
	ctx := context.Background()
	service := NewCategoryService(&fakeCategoryRepo{
		restoreByNameAndTypeErr: errors.New("database connection failed"),
	})

	result, err := service.Create(ctx, CategoryInput{
		Name: "food",
		Type: domain.TransactionTypeExpense,
		Keywords: []string{"makan"},
		IsDefault: true,
	})

	if err == nil {
		t.Fatalf("expected error not to be nil")
	}

	if result != nil {
		t.Fatal("expected result to be nil")
	}
}

func TestFindAllCategory_Success(t *testing.T) {
	totalCategory := int64(2)
	categoryRepo := &fakeCategoryRepo{
		findAllCategoryResult: []domain.Category{
			{
				ID: 1, 
				Name: "food", 
				Type: domain.TransactionTypeExpense, 
				Keywords: []string{"makan", "kopi"}, 
				IsDefault: true,
			},
			{
				ID: 2, 
				Name: "entertaiment", 
				Type: domain.TransactionTypeExpense, 
				Keywords: []string{"netflix", "bioskop"}, 
				IsDefault: true,
			},
		},
		countAllCategoryResult: &totalCategory,
		
	}
	service := NewCategoryService(categoryRepo)

	query := CategoryQuery{
		Param: pagination.Params{
			Page: 1,
			PerPage: 10,
		},
	}

	result, total, err := service.FindAll(context.Background(), query)
	
	if err != nil {
		t.Fatalf("find all category error: %v", err)
	}

	if total != int64(len(categoryRepo.findAllCategoryResult)) {
		t.Fatalf("expected total data %d, got %d", int64(len(categoryRepo.findAllCategoryResult)), total)
	}

	if result == nil {
		t.Fatal("expected result not to be nil")
	}

	expectedResult := []CategoryResult{
		{
			ID: 1, 
			Name: "food", 
			Type: domain.TransactionTypeExpense, 
			Keywords: []string{"makan", "kopi"}, 
			IsDefault: true,
		},
		{
			ID: 2, 
			Name: "entertaiment", 
			Type: domain.TransactionTypeExpense, 
			Keywords: []string{"netflix", "bioskop"}, 
			IsDefault: true,
		},
	}
	if !reflect.DeepEqual(result, expectedResult) {
		t.Fatalf("expected result %+v, got %+v", expectedResult, result)
	}
}

func TestFindAllCategory_RepositoryError(t *testing.T) {
	categoryRepo := &fakeCategoryRepo{
		findAllCategoryErr: errors.New("database connection failed"),
	}
	service := NewCategoryService(categoryRepo)

	query := CategoryQuery{
		Param: pagination.Params{
			Page: 1,
			PerPage: 10,
		},
	}

	result, total, err := service.FindAll(context.Background(), query)
	
	if err == nil {
		t.Fatal("expected to be error")
	}

	if result != nil {
		t.Fatal("expected result to be nil")
	}

	if total != 0 {
		t.Fatal("expected total to be 0")
	}

	if !errors.Is(err, categoryRepo.findAllCategoryErr) {
		t.Fatalf("expected err %v, got %v", categoryRepo.countAllCategoryErr, err)
	}

	if !strings.Contains(err.Error(), "load categories") {
		t.Fatalf(`expected error contains "load categories", got %s`, err.Error())
	}
	
}

func TestFindAllCategory_CountAllRepositoryError(t *testing.T) {
	categoryRepo := &fakeCategoryRepo{
		countAllCategoryErr: errors.New("database connection failed"),
	}
	service := NewCategoryService(categoryRepo)

	query := CategoryQuery{
		Param: pagination.Params{
			Page: 1,
			PerPage: 10,
		},
	}

	result, total, err := service.FindAll(context.Background(), query)
	
	if err == nil {
		t.Fatal("expected to be error")
	}

	if result != nil {
		t.Fatal("expected result to be nil")
	}

	if total != 0 {
		t.Fatal("expected total to be 0")
	}

	if !errors.Is(err, categoryRepo.countAllCategoryErr) {
		t.Fatalf("expected err %v, got %v", categoryRepo.countAllCategoryErr, err)
	}

	if !strings.Contains(err.Error(), "count categories") {
		t.Fatalf(`expected error contains "count categories", got %s`, err.Error())
	}
	
}

func TestUpdateCategory_Success(t *testing.T) {
	categoryRepo := &fakeCategoryRepo{
		findByIDResult: &domain.Category{
			ID: 1, 
			Name: "food", 
			Type: domain.TransactionTypeExpense, 
			Keywords: []string{"makan", "kopi"}, 
			IsDefault: true,
		},
	}
	service := NewCategoryService(categoryRepo)

	updatedData := UpdateCategoryInput{
		ID: 1,
		Name: "food updated",
		Type: domain.TransactionTypeExpense,
		Keywords: []string{"makan", "kopi", "beli"},
		IsDefault: true,
	}
	result, err := service.Update(context.Background(), updatedData)
	
	if err != nil {
		t.Fatalf("update category error: %v", err)
	}

	if result.ID != updatedData.ID {
		t.Errorf("expected ID %d, got %d", updatedData.ID, result.ID)
	}

	if result.Name != updatedData.Name {
		t.Errorf("expected Name %q, got %q", updatedData.Name, result.Name)
	}

	if result.Type != updatedData.Type {
		t.Errorf("expected Type %q, got %q", updatedData.Type, result.Type)
	}

	if !slices.Equal(result.Keywords, updatedData.Keywords) {
		t.Errorf("expected Keywords %v, got %v", updatedData.Keywords, result.Keywords)
	}

	if result.IsDefault != updatedData.IsDefault {
		t.Errorf("expected IsDefault %t, got %t", updatedData.IsDefault, result.IsDefault)
	}
}

func TestUpdateCategory_CategoryNotFound(t *testing.T) {
	categoryRepo := &fakeCategoryRepo{
		findByIDErr: &domain.NotFoundError{Resource: domain.ResourceCategory},
	}
	service := NewCategoryService(categoryRepo)

	updatedData := UpdateCategoryInput{
		ID: 1,
		Name: "food updated",
		Type: domain.TransactionTypeExpense,
		Keywords: []string{"makan", "kopi", "beli"},
		IsDefault: true,
	}
	result, err := service.Update(context.Background(), updatedData)
	
	if err == nil {
		t.Fatal("expected error not to be nil")
	}

	if result != nil {
		t.Fatal("expected result to be nil")
	}

	var notFound *domain.NotFoundError
	if !errors.As(err, &notFound) || notFound.Resource != domain.ResourceCategory {
		t.Fatalf("expected error %v, got %v", notFound.Error(), err)
	}
}

func TestUpdateCategory_RepositoryError(t *testing.T) {
	categoryRepo := &fakeCategoryRepo{
		findByIDResult: &domain.Category{
			ID: 1, 
			Name: "food", 
			Type: domain.TransactionTypeExpense, 
			Keywords: []string{"makan", "kopi"}, 
			IsDefault: true,
		},
		updateCategoryErr: errors.New("database connection failed"),
	}
	service := NewCategoryService(categoryRepo)

	updatedData := UpdateCategoryInput{
		ID: 1,
		Name: "food updated",
		Type: domain.TransactionTypeExpense,
		Keywords: []string{"makan", "kopi", "beli"},
		IsDefault: true,
	}
	result, err := service.Update(context.Background(), updatedData)
	
	if err == nil {
		t.Fatal("expected error not to be nil")
	}

	if result != nil {
		t.Fatal("expected result to be nil")
	}

	if !errors.Is(err, categoryRepo.updateCategoryErr) {
		t.Fatalf("expected error %v, got %v", categoryRepo.updateCategoryErr, err)
	}
}

func TestHardDeleteCategory_Success(t *testing.T) {
	categoryRepo := &fakeCategoryRepo{
		findByIDResult: &domain.Category{
			ID: 1, 
			Name: "food", 
			Type: domain.TransactionTypeExpense, 
			Keywords: []string{"makan", "kopi"}, 
			IsDefault: true,
		},
	}
	service := NewCategoryService(categoryRepo)

	err := service.HardDelete(context.Background(), 1)

	if err != nil {
		t.Fatalf("hard delete category error: %v", err)
	}
}

func TestHardDeleteCategory_CategoryNotFoundError(t *testing.T) {
	categoryRepo := &fakeCategoryRepo{
		findByIDErr: &domain.NotFoundError{Resource: domain.ResourceCategory},
	}
	service := NewCategoryService(categoryRepo)

	err := service.HardDelete(context.Background(), 1)

	if err == nil {
		t.Fatal("expected error not to be nil")
	}

	var notFound *domain.NotFoundError
	if !errors.As(err, &notFound) || notFound.Resource != domain.ResourceCategory {
		t.Fatalf(
			"expected error %v with resource %s, got error %v with resource %s", 
			notFound.Error(), 
			domain.ResourceCategory, 
			err, 
			notFound.Resource,
		)
	}
}

func TestHardDeleteCategory_RepositoryError(t *testing.T) {
	categoryRepo := &fakeCategoryRepo{
		hardDeleteCategoryErr: errors.New("database connection failed"),
	}
	service := NewCategoryService(categoryRepo)

	err := service.HardDelete(context.Background(), 1)

	if err == nil {
		t.Fatal("expected error not to be nil")
	}

	if !errors.Is(err, categoryRepo.hardDeleteCategoryErr) {
		t.Fatalf("expected error %v, got %v", categoryRepo.hardDeleteCategoryErr.Error(), err.Error())
	}
}

