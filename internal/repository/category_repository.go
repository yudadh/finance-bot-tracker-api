package repository

import (
	"context"

	"github.com/yudadh/finance-bot-tracker-api/internal/domain"
	"gorm.io/gorm"
)

type CategoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) Create(ctx context.Context, category *domain.Category) error {
	return r.db.WithContext(ctx).Create(category).Error
}

func (r *CategoryRepository) FindByID(ctx context.Context, id uint64) (*domain.Category, error) {
	var category domain.Category

	err := r.db.
		WithContext(ctx).
		Where("id = ?", id).
		First(&category).
		Error

	if err != nil {
		return nil, err
	}

	return &category, nil
}

func (r *CategoryRepository) FindAll(ctx context.Context) ([]domain.Category, error) {
	var categories []domain.Category

	err := r.db.
		WithContext(ctx).
		Find(&categories).
		Error

	if err != nil {
		return nil, err
	}

	return categories, nil
}

func (r *CategoryRepository) Delete(ctx context.Context, category *domain.Category) error {
	return r.db.WithContext(ctx).Delete(category).Error
}
