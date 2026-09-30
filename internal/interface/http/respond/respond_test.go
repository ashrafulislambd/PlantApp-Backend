package respond

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"plantpal-backend/internal/domain/apperr"
)

func TestError_RateLimitedIs429WithRetryAfter(t *testing.T) {
	rec := httptest.NewRecorder()
	Error(rec, fmt.Errorf("gemini: status 429: %w", apperr.ErrRateLimited))

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want 429", rec.Code)
	}
	if rec.Header().Get("Retry-After") == "" {
		t.Fatal("Retry-After header missing")
	}
}
