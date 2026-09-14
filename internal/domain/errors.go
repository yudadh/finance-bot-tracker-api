package domain

import "errors"

var (
	ErrNotFound = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrInvalidInput = errors.New("invalid input")
	ErrConflict = errors.New("conflict")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidDateRange = errors.New("invalid date range")
	ErrInvalidID = errors.New("invalid id")
)

type ResourceType string

const (
	ResourceUser ResourceType = "user"
	ResourceTransaction ResourceType = "transaction"
	ResourceBudget ResourceType = "budget"
	ResourceCategory ResourceType = "category"
)

type NotFoundError struct {
	Resource ResourceType
}

func (e *NotFoundError) Error() string {
	return string(e.Resource) + " not found"
}

func (e *NotFoundError) Unwrap() error {
	return ErrNotFound
}

