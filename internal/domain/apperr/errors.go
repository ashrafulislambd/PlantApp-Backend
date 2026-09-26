// Package apperr defines sentinel errors shared across the domain and
// usecase layers. Delivery adapters (HTTP handlers) map these to status
// codes without needing to know which repository or usecase produced them.
package apperr

import "errors"

var (
	ErrNotFound     = errors.New("resource not found")
	ErrInvalidInput = errors.New("invalid input")
	ErrConflict     = errors.New("resource already exists")
	ErrUnauthorized = errors.New("unauthorized")
)
