package repository

import (
	"context"

	"github.com/yudadh/finance-bot-tracker-api/internal/domain"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	err := r.db.WithContext(ctx).Create(user).Error
	return translateError(err)
}

func (r *UserRepository) FindByID(ctx context.Context, id uint64) (*domain.User, error) {
	var user domain.User

	err := r.db.
		WithContext(ctx).
		Where("id = ?", id).
		First(&user).
		Error

	if err != nil {
		return nil, translateError(err)
	}

	return &user, err
}

func (r *UserRepository) FindByTelegramID(ctx context.Context, telegramId int64) (*domain.User, error) {
	var user domain.User

	err := r.db.
		WithContext(ctx).
		Where("telegram_id = ?", telegramId).
		First(&user).
		Error

	if err != nil {
		return nil, translateError(err)
	}

	return &user, err
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	err := r.db.WithContext(ctx).Save(user).Error
	return translateError(err)
}
