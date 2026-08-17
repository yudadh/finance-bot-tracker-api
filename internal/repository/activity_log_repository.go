package repository

import (
	"context"

	"github.com/yudadh/finance-bot-tracker-api/internal/domain"
	"gorm.io/gorm"
)

type ActivityLogRepository struct {
	db *gorm.DB
}

func NewActivityLogRepository(db *gorm.DB) *ActivityLogRepository {
	return &ActivityLogRepository{db: db}
}

func (r *ActivityLogRepository) Create(ctx context.Context, activityLog *domain.ActivityLog) error {
	return r.db.WithContext(ctx).Create(activityLog).Error
}

func (r *ActivityLogRepository) FindByID(ctx context.Context, id uint64) (*domain.ActivityLog, error) {
	var activityLog domain.ActivityLog

	err := r.db.
		WithContext(ctx).
		Where("id = ?", id).
		First(&activityLog).
		Error

	if err != nil {
		return nil, err
	}

	return &activityLog, nil
}

func (r *ActivityLogRepository) FindAll(ctx context.Context, limit int, offset int) ([]domain.ActivityLog, error) {
	var activityLogs []domain.ActivityLog

	err := r.db.
		WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Find(&activityLogs).
		Error

	if err != nil {
		return nil, err
	}

	return activityLogs, nil
}