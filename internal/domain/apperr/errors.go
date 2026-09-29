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
	// ErrRateLimited means an upstream service (e.g. an AI provider) rejected
	// the call because its quota was exhausted; the client should retry later.
	ErrRateLimited = errors.New("rate limit reached, please try again shortly")
	// ErrUnavailable means a feature is switched off by server configuration
	// (e.g. Google sign-in without GOOGLE_CLIENT_ID).
	ErrUnavailable = errors.New("feature unavailable")
)
