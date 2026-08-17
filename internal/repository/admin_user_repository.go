package repository

import (
	"context"

	"github.com/yudadh/finance-bot-tracker-api/internal/domain"
	"gorm.io/gorm"
)

type AdminUserRepository struct {
	db *gorm.DB
}

func NewUserAdminRepository(db *gorm.DB) *AdminUserRepository {
	return &AdminUserRepository{db: db}
}

func (r *AdminUserRepository) Create(ctx context.Context, adminUser *domain.AdminUser) error {
	return r.db.WithContext(ctx).Create(adminUser).Error
}

func (r *AdminUserRepository) FindByID(ctx context.Context, id uint64) (*domain.AdminUser, error) {
	var adminUser domain.AdminUser

	err := r.db.
		WithContext(ctx).
		Where("id = ?", id).
		First(&adminUser).
		Error

	if err != nil {
		return nil, err
	}

	return &adminUser, nil
}

func (r *AdminUserRepository) FindAll(ctx context.Context, limit int, offset int) ([]domain.AdminUser, error) {
	var adminUsers []domain.AdminUser

	err := r.db.
		WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Find(&adminUsers).
		Error

	if err != nil {
		return nil, err
	}

	return adminUsers, nil
}

func (r *AdminUserRepository) Update(ctx context.Context, adminUser *domain.AdminUser) error {
	return r.db.WithContext(ctx).Save(adminUser).Error
}

func (r *AdminUserRepository) Delete(ctx context.Context, adminUser *domain.AdminUser) error {
	return r.db.WithContext(ctx).Delete(adminUser).Error
}