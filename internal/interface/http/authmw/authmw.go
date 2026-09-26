// Package authmw provides the Bearer-JWT auth middleware shared by every
// API version.
package authmw

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"myplantpal-backend/internal/domain/apperr"
	"myplantpal-backend/internal/interface/http/respond"
)

type ctxKey int

const userIDKey ctxKey = 0

type TokenParser interface {
	ParseAccessToken(token string) (string, error)
}

func RequireAuth(parser TokenParser) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			token, ok := strings.CutPrefix(header, "Bearer ")
			if !ok || token == "" {
				respond.Error(w, fmt.Errorf("%w: missing bearer token", apperr.ErrUnauthorized))
				return
			}
			userID, err := parser.ParseAccessToken(token)
			if err != nil {
				respond.Error(w, fmt.Errorf("%w: invalid or expired token", apperr.ErrUnauthorized))
				return
			}
			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next(w, r.WithContext(ctx))
		}
	}
}

func UserID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIDKey).(string)
	return id, ok
}

// WithUserID returns a context copy with the user ID set (useful for tests and internal requests).
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}
