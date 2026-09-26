// Package respond provides shared JSON response helpers for HTTP handlers.
// It is a leaf package (no dependency on the router or middleware) so both
// the top-level http package and versioned handler packages (v1, ...) can
// depend on it without an import cycle.
package respond

import (
	"encoding/json"
	"errors"
	"net/http"

	"myplantpal-backend/internal/domain/apperr"
)

type envelope struct {
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(envelope{Data: data})
}

// Error maps a domain sentinel error (via errors.Is) to an HTTP status code
// and writes a JSON error envelope. Unrecognized errors become 500s.
func Error(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, apperr.ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, apperr.ErrInvalidInput):
		status = http.StatusBadRequest
	case errors.Is(err, apperr.ErrConflict):
		status = http.StatusConflict
	case errors.Is(err, apperr.ErrUnauthorized):
		status = http.StatusUnauthorized
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(envelope{Error: err.Error()})
}
