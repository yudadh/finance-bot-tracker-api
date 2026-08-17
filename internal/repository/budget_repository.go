package repository

import (
	"context"
	"time"

	"github.com/yudadh/finance-bot-tracker-api/internal/domain"
	"gorm.io/gorm"
)

type BudgetRepository struct {
	db *gorm.DB
}

func NewBudgetRepository(db *gorm.DB) *BudgetRepository {
	return &BudgetRepository{db: db}
}

func (r *BudgetRepository) Create(ctx context.Context, budget *domain.Budget) error {
	return r.db.WithContext(ctx).Create(budget).Error
}

func (r *BudgetRepository) FindByID(ctx context.Context, id uint64) (*domain.Budget, error) {
	var budget domain.Budget

	err := r.db.
		WithContext(ctx).
		Where("id = ?", id).
		First(&budget).Error

	if err != nil {
		return nil, err
	}

	return &budget, nil
}

func (r *BudgetRepository) FindByUserPeriod(ctx context.Context, userID uint64, period time.Time, periodType domain.BudgetPeriodType) (*domain.Budget, error) {
	var budget domain.Budget

	err := r.db.
		WithContext(ctx).
		Where("user_id = ?", userID).
		Where("period_type = ?", periodType).
		Where("period_start = ?", period).
		First(&budget).Error

	if err != nil {
		return nil, err
	}

	return &budget, nil
}

func (r *BudgetRepository) FindAll(ctx context.Context, limit int, offset int) ([]domain.Budget, error) {
	var budgets []domain.Budget

	err := r.db.
		WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Find(&budgets).Error

	if err != nil {
		return nil, err
	}

	return budgets, nil
}

func (r *BudgetRepository) Update(ctx context.Context, budget *domain.Budget) error {
	return r.db.WithContext(ctx).Save(budget).Error
}

