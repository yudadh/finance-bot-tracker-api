package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/yudadh/finance-bot-tracker-api/internal/domain"
	"github.com/yudadh/finance-bot-tracker-api/internal/pagination"
	"github.com/yudadh/finance-bot-tracker-api/internal/repository"
)

type userRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByTelegramID(ctx context.Context, telegramID int64) (*domain.User, error)
	FindAll(ctx context.Context, query repository.FindAllUsersQuery) ([]domain.User, error)
	CountAll(ctx context.Context, query repository.FindAllUsersQuery) (*int64, error)
}

type UserService struct {
	userRepo userRepository
	logger   *slog.Logger
}

type FindOrCreateUserInput struct {
	TelegramID       int64
	TelegramUsername string
	FirstName        string
	LastName         string
	LanguageCode     string
}

type FindAllUserResult struct {
	ID               uint64
	TelegramID       int64
	TelegramUsername string
	FirstName        string
	LastName         string
	LanguageCode     string
	Timezone         string
	Status           domain.UserStatus
}

type FindAllUsersQuery struct {
	Param pagination.Params
	Status string
	Search string
}

func NewUserService(userRepo userRepository, logger *slog.Logger) *UserService {
	return &UserService{
		userRepo: userRepo,
		logger:   logger,
	}
}

func (s *UserService) FindOrCreate(
	ctx context.Context,
	userInput *FindOrCreateUserInput,
) (*domain.User, error) {
	user, err := s.userRepo.FindByTelegramID(ctx, userInput.TelegramID)
	if err == nil {
		return user, nil
	}

	if !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}

	user = &domain.User{
		TelegramID:       userInput.TelegramID,
		TelegramUsername: userInput.TelegramUsername,
		FirstName:        userInput.FirstName,
		LastName:         userInput.LastName,
		LanguageCode:     userInput.LanguageCode,
		Timezone:         "Asia/Makassar",
		Status:           domain.UserStatusActive,
	}

	err = s.userRepo.Create(ctx, user)

	if err != nil {
		if errors.Is(err, domain.ErrAlreadyExists) {
			return nil, domain.ErrAlreadyExists
		}
		return nil, err
	}

	return user, nil
}

func (s *UserService) FindAll(ctx context.Context, query FindAllUsersQuery) ([]FindAllUserResult, *pagination.Meta, error) {
	queryValue := repository.FindAllUsersQuery{
		Limit: query.Param.Limit(),
		Offset: query.Param.Offset(),
		Status: domain.UserStatus(query.Status),
		Search: query.Search,
	}
	users, err := s.userRepo.FindAll(ctx, queryValue)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"load users %v",
			err,
		)
	}

	total, err := s.userRepo.CountAll(ctx, queryValue)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"count users %v",
			err,
		)
	}

	result := []FindAllUserResult{}
	for _, user := range users {
		result = append(result, FindAllUserResult{
			ID: user.ID,
			TelegramID: user.TelegramID,
			TelegramUsername: user.TelegramUsername,
			FirstName: user.FirstName,
			LastName: user.LastName,
			Timezone: user.Timezone,
			Status: user.Status,
		})
	}

	meta := pagination.NewMeta(query.Param, *total)

	return result, &meta, nil
}
