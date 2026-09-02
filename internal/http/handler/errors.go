package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/yudadh/finance-bot-tracker-api/internal/domain"
)

type ValidationFieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error   bool                   `json:"error"`
	Message string                 `json:"message"`
	Errors  []ValidationFieldError `json:"errors,omitempty"`
}

func HandleError(c *gin.Context, err error) {
	var validationErrors validator.ValidationErrors

	switch {
	case errors.As(err, &validationErrors):
		errors := ParseValidationErrors(validationErrors)
		ResponseError(c, http.StatusBadRequest, "validation failed", errors)

	case errors.Is(err, domain.ErrInvalidCredentials):
		ResponseError(c, http.StatusUnauthorized, "invalid email or password", nil)

	case errors.Is(err, domain.ErrNotFound):
		ResponseError(c, http.StatusNotFound, "resource not found", nil)

	case errors.Is(err, domain.ErrAlreadyExists):
		ResponseError(c, http.StatusConflict, "resource already exists", nil)

	case errors.Is(err, domain.ErrInvalidInput):
		ResponseError(c, http.StatusBadRequest, "invalid input", nil)

	default:
		ResponseError(c, http.StatusInternalServerError, "internal server error", nil)
	}
}

func validationErrorMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email"
	case "min":
		return "must be at least " + err.Param() + " characters"
	case "max":
		return "maximum length of " + err.Param() + " characters"
	default:
		return "is invalid"
	}
}

func ParseValidationErrors(errs validator.ValidationErrors) []ValidationFieldError {
	result := make([]ValidationFieldError, 0, len(errs))

	for _, err := range errs {
		result = append(result, ValidationFieldError{
			Field: strings.ToLower(err.Field()),
			Message: validationErrorMessage(err),
		})
	}

	return result
}