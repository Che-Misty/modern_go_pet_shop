package storage

import "errors"

var (
	ErrNotFound          = errors.New("not found")
	ErrInvalidInput      = errors.New("invalid input")
	ErrEmailExists       = errors.New("email already exists")
	ErrInsufficientStock = errors.New("insufficient stock")
)
