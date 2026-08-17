package repository

import (
	"context"

	"github.com/yudadh/finance-bot-tracker-api/internal/domain"
	"gorm.io/gorm"
)

type ParserAttemptRepository struct {
	db *gorm.DB
}

func NewParserAttemptRepository(db *gorm.DB) *ParserAttemptRepository {
	return &ParserAttemptRepository{db: db}
}

func (r *ParserAttemptRepository) Create(ctx context.Context, parserAttempt *domain.ParserAttempt) error {
	return r.db.WithContext(ctx).Create(parserAttempt).Error
}

func (r *ParserAttemptRepository) FindAll(ctx context.Context, limit int, offset int) ([]domain.ParserAttempt, error) {
	var parserAttempts []domain.ParserAttempt

	err := r.db.
		WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Find(&parserAttempts).
		Error

	if err != nil {
		return nil, err
	}

	return parserAttempts, nil
}