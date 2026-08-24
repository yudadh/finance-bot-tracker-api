package service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/yudadh/finance-bot-tracker-api/internal/domain"
)

type userRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByTelegramID(ctx context.Context, telegramID int64) (*domain.User, error)
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
		TelegramID: userInput.TelegramID,
		TelegramUsername: userInput.TelegramUsername,
		FirstName: userInput.FirstName,
		LastName: userInput.LastName,
		LanguageCode: userInput.LanguageCode,
		Timezone: "Asia/Makassar",
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
