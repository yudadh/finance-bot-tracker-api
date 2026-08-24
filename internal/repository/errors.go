package repository

import (
	"errors"

	"github.com/yudadh/finance-bot-tracker-api/internal/domain"
	"gorm.io/gorm"
)

func translateError(err error) error {
	switch {
	case err == nil:
		return nil

	case errors.Is(err, gorm.ErrDuplicatedKey):
		return domain.ErrAlreadyExists

	case errors.Is(err, gorm.ErrRecordNotFound):
		return domain.ErrNotFound

	case errors.Is(err, gorm.ErrForeignKeyViolated):
		return domain.ErrConflict

	default:
		return err
	}
}