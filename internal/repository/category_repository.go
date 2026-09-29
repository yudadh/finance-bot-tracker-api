package repository

import (
	"context"

	"github.com/yudadh/finance-bot-tracker-api/internal/domain"
	"gorm.io/gorm"
)

type CategoryRepository struct {
	db *gorm.DB
}

type FindAllCategoriesQuery struct {
	Limit int
	Offset int
	TransactionType domain.TransactionType
	Name string
}

func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) Create(ctx context.Context, category *domain.Category) error {
	err := r.db.WithContext(ctx).Create(category).Error
	return translateError(err)
}

func (r *CategoryRepository) FindByID(ctx context.Context, id uint64) (*domain.Category, error) {
	var category domain.Category

	err := r.db.
		WithContext(ctx).
		Where("id = ?", id).
		First(&category).
		Error

	if err != nil {
		return nil, translateError(err)
	}

	return &category, nil
}

func (r *CategoryRepository) FindAllWithoutPagination(ctx context.Context) ([]domain.Category, error) {
	var categories []domain.Category

	err := r.db.
		WithContext(ctx).
		Find(&categories).
		Error

	if err != nil {
		return nil, translateError(err)
	}

	return categories, nil
}

func (r *CategoryRepository) FindAll(ctx context.Context, query FindAllCategoriesQuery) ([]domain.Category, error) {
	var categories []domain.Category

	err := r.db.
		WithContext(ctx).
		Where("type = ?", query.TransactionType).
		Where("name LIKE ?", "%"+query.Name+"%").
		Limit(query.Limit).
		Offset(query.Offset).
		Find(&categories).
		Error

	if err != nil {
		return nil, translateError(err)
	}

	return categories, nil
}

func (r *CategoryRepository) CountAll(ctx context.Context, query FindAllCategoriesQuery) (*int64, error) {
	var total int64

	err := r.db.
		WithContext(ctx).
		Where("type = ?", query.TransactionType).
		Where("name LIKE ?", "%"+query.Name+"%").
		Count(&total).
		Error

	if err != nil {
		return nil, translateError(err)
	}

	return &total, nil
}

func (r *CategoryRepository) RestoreByNameAndType(
	ctx context.Context,
	name string,
	transactionType domain.TransactionType, 
) (*domain.Category, error) {
	var category domain.Category

	err := r.db.
		WithContext(ctx).
		Unscoped().
		Where("name = ?", name).
		Where("type = ?", transactionType).
		First(&category).
		Error
	
	if err != nil {
		return nil, translateError(err)
	}

	if err = r.db.
		WithContext(ctx).
		Unscoped().
		Model(&category).
		Update("deleted_at", nil).Error; err != nil {
		return nil, translateError(err)
	}

	return &category, nil
}

func (r *CategoryRepository) Update(ctx context.Context, category *domain.Category) error {
	err := r.db.
		WithContext(ctx).
		Save(category).
		Error

	return translateError(err)
}

func (r *CategoryRepository) Delete(ctx context.Context, category *domain.Category) error {
	err := r.db.WithContext(ctx).Delete(category).Error
	return translateError(err)
}

func (r *CategoryRepository) HardDelete(ctx context.Context, category *domain.Category) error {
	err := r.db.WithContext(ctx).Unscoped().Delete(category).Error
	return translateError(err)
}
