package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/yudadh/finance-bot-tracker-api/internal/config"
	"github.com/yudadh/finance-bot-tracker-api/internal/domain"
	"github.com/yudadh/finance-bot-tracker-api/internal/security"
)

type AdminUserRepository interface {
	Create(ctx context.Context, adminUser *domain.AdminUser) error
	FindByID(ctx context.Context, id uint64) (*domain.AdminUser, error)
	FindByEmail(ctx context.Context, email string) (*domain.AdminUser, error)
	Update(ctx context.Context, adminUser *domain.AdminUser) error
}

type AdminUserService struct {
	adminUserRepo AdminUserRepository
	jwtConfig     config.JWTConfig
}

type RegisterAdminUserInput struct {
	Email    string
	Password string
	Name     string
}

type RegisterAdminUserResult struct {
	ID     uint64
	Email  string
	Name   string
	Status domain.AdminUserStatus
}

type FindByIDResult struct {
	ID     uint64
	Email  string
	Name   string
	Status domain.AdminUserStatus
}

func NewAdminUserService(
	adminUserRepo AdminUserRepository,
	jwtConfig config.JWTConfig,
) *AdminUserService {
	return &AdminUserService{
		adminUserRepo: adminUserRepo,
		jwtConfig:     jwtConfig,
	}
}

func (s *AdminUserService) Register(ctx context.Context, adminUserInput RegisterAdminUserInput) (*RegisterAdminUserResult, error) {
	passwordHash, err := security.HashPassword(adminUserInput.Password)
	if err != nil {
		return nil, fmt.Errorf(
			"hash admin user password: %w",
			err,
		)
	}

	adminUser := &domain.AdminUser{
		Email:        adminUserInput.Email,
		PasswordHash: passwordHash,
		Name:         adminUserInput.Name,
		Status:       domain.AdminUserStatusActive,
	}

	err = s.adminUserRepo.Create(ctx, adminUser)
	if err != nil {
		return nil, fmt.Errorf(
			"created registered admin user: %w",
			err,
		)
	}

	return &RegisterAdminUserResult{
		ID: adminUser.ID,
		Email:  adminUser.Email,
		Name:   adminUser.Name,
		Status: adminUser.Status,
	}, nil
}

func (s *AdminUserService) Login(ctx context.Context, email string, password string) (string, error) {
	adminUser, err := s.adminUserRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return "", domain.ErrInvalidCredentials
		}
		return "", fmt.Errorf(
			"load admin user for login: %w",
			err,
		)
	}

	if !security.ValidatePasswordHash(adminUser.PasswordHash, password) {
		return "", domain.ErrInvalidCredentials
	}

	token, err := security.GenerateAdminToken(s.jwtConfig, adminUser.ID, adminUser.Email)
	if err != nil {
		return "", fmt.Errorf(
			"generate admin user jwt token: %w",
			err,
		)
	}

	return token, nil
}

func (s *AdminUserService) FindByID(ctx context.Context, id uint64) (*FindByIDResult, error) {
	adminUser, err := s.adminUserRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf(
			"load admin user by id: %w",
			err,
		)
	}

	return &FindByIDResult{
		ID:     adminUser.ID,
		Email:  adminUser.Email,
		Name:   adminUser.Name,
		Status: adminUser.Status,
	}, nil
}
