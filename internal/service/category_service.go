package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/yudadh/finance-bot-tracker-api/internal/domain"
	"github.com/yudadh/finance-bot-tracker-api/internal/pagination"
	"github.com/yudadh/finance-bot-tracker-api/internal/repository"
)

type CategoryRepository interface {
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
}

type CategoryService struct {
	categoryRepo CategoryRepository
}

type CategoryInput struct {
	Name      string
	Type      domain.TransactionType
	Keywords  []string
	IsDefault bool
}

type CategoryResult struct {
	ID        uint64
	Name      string
	Type      domain.TransactionType
	Keywords  []string
	IsDefault bool
}

type CategoryQuery struct {
	Param           pagination.Params
	TransactionType domain.TransactionType
	Name            string
}

type UpdateCategoryInput struct {
	ID        uint64
	Name      string
	Type      domain.TransactionType
	Keywords  []string
	IsDefault bool
}

func NewCategoryService(categoryRepo CategoryRepository) *CategoryService {
	return &CategoryService{
		categoryRepo: categoryRepo,
	}
}

func (s *CategoryService) Create(ctx context.Context, categoryInput CategoryInput) (*CategoryResult, error) {
	category, err := s.categoryRepo.RestoreByNameAndType(ctx, categoryInput.Name, categoryInput.Type)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			newCategory := &domain.Category{
				Name:      categoryInput.Name,
				Type:      categoryInput.Type,
				Keywords:  categoryInput.Keywords,
				IsDefault: categoryInput.IsDefault,
			}
			createCategoryErr := s.categoryRepo.Create(ctx, newCategory)

			if createCategoryErr != nil {
				return nil, fmt.Errorf(
					"create category: %w",
					createCategoryErr,
				)
			}

			return &CategoryResult{
				ID:        newCategory.ID,
				Name:      newCategory.Name,
				Type:      newCategory.Type,
				Keywords:  newCategory.Keywords,
				IsDefault: newCategory.IsDefault,
			}, nil
		}

		return nil, fmt.Errorf(
			"restore category: %w",
			err,
		)
	}

	return &CategoryResult{
		ID:        category.ID,
		Name:      category.Name,
		Type:      category.Type,
		Keywords:  category.Keywords,
		IsDefault: category.IsDefault,
	}, nil
}

func (s *CategoryService) FindByID(ctx context.Context, id uint64) (*CategoryResult, error) {
	category, err := s.categoryRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, &domain.NotFoundError{Resource: domain.ResourceCategory}
		}

		return nil, fmt.Errorf(
			"load category %w",
			err,
		)
	}

	return &CategoryResult{
		ID:        category.ID,
		Name:      category.Name,
		Type:      category.Type,
		Keywords:  category.Keywords,
		IsDefault: category.IsDefault,
	}, nil
}

func (s *CategoryService) FindAll(ctx context.Context, query CategoryQuery) ([]CategoryResult, int64, error) {
	repositoryQuery := repository.FindAllCategoriesQuery{
		Limit:           query.Param.Limit(),
		Offset:          query.Param.Offset(),
	}
	
	if query.TransactionType != "" {
		repositoryQuery.TransactionType = query.TransactionType
	}

	if query.Name != "" {
		repositoryQuery.Name = query.Name
	}

	categories, err := s.categoryRepo.FindAll(ctx, repositoryQuery)

	if err != nil {
		return nil, 0, fmt.Errorf(
			"load categories: %w",
			err,
		)
	}

	total, err := s.categoryRepo.CountAll(ctx, repository.FindAllCategoriesQuery{
		TransactionType: query.TransactionType,
		Name:            query.Name,
	})

	if err != nil {
		return nil, 0, fmt.Errorf(
			"count categories: %w",
			err,
		)
	}

	categoryResult := make([]CategoryResult, 0, len(categories))

	for _, category := range categories {
		categoryResult = append(categoryResult, CategoryResult{
			ID:        category.ID,
			Name:      category.Name,
			Type:      category.Type,
			Keywords:  category.Keywords,
			IsDefault: category.IsDefault,
		})
	}

	return categoryResult, *total, nil
}

func (s *CategoryService) Update(ctx context.Context, data UpdateCategoryInput) (*CategoryResult, error) {
	category, err := s.categoryRepo.FindByID(ctx, data.ID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, &domain.NotFoundError{Resource: domain.ResourceCategory}
		}
		
		return nil, fmt.Errorf(
			"load category by id: %w",
			err,
		)
	}

	category.Name = data.Name
	category.Type = data.Type
	category.Keywords = data.Keywords
	category.IsDefault = data.IsDefault

	err = s.categoryRepo.Update(ctx, category)
	if err != nil {
		return nil, fmt.Errorf(
			"update category: %w",
			err,
		)
	}

	return &CategoryResult{
		ID: category.ID,
		Name: category.Name,
		Type: category.Type,
		Keywords: category.Keywords,
		IsDefault: category.IsDefault,
	}, nil
}

func (s *CategoryService) HardDelete(ctx context.Context, id uint64) error {
	category, err := s.categoryRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return &domain.NotFoundError{Resource: domain.ResourceCategory}
		}
		
		return fmt.Errorf(
			"load category by id: %w",
			err,
		)
	}

	err = s.categoryRepo.HardDelete(ctx, category)
	if err != nil {
		return fmt.Errorf(
			"hard delete category: %w",
			err,
		)
	}

	return nil
}
