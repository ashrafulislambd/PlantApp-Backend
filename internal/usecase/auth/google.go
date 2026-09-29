package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"plantpal-backend/internal/domain/apperr"
	"plantpal-backend/internal/domain/user"
	"plantpal-backend/internal/infrastructure/security"
)

// GoogleVerifier validates a Google ID token (see security.GoogleVerifier).
type GoogleVerifier interface {
	Verify(ctx context.Context, idToken string) (*security.GoogleIdentity, error)
}

// SetGoogleVerifier enables Google sign-in. Without it LoginWithGoogle
// reports the feature as unavailable.
func (s *Service) SetGoogleVerifier(v GoogleVerifier) { s.google = v }

// LoginWithGoogle exchanges a Google ID token for the app's own access +
// refresh tokens. The user is found by Google account ID, else by verified
// email (linking the Google account to an existing email/password user),
// else created.
func (s *Service) LoginWithGoogle(ctx context.Context, idToken string) (*AuthResult, error) {
	if s.google == nil {
		return nil, fmt.Errorf("%w: Google sign-in is not configured on this server", apperr.ErrUnavailable)
	}
	if strings.TrimSpace(idToken) == "" {
		return nil, fmt.Errorf("%w: idToken is required", apperr.ErrInvalidInput)
	}

	id, err := s.google.Verify(ctx, idToken)
	if err != nil {
		if errors.Is(err, security.ErrInvalidGoogleToken) {
			return nil, fmt.Errorf("%w: invalid Google token", apperr.ErrUnauthorized)
		}
		return nil, err
	}

	u, err := s.users.GetByGoogleID(ctx, id.Subject)
	if errors.Is(err, apperr.ErrNotFound) {
		u, err = s.findOrCreateByEmail(ctx, id)
	}
	if err != nil {
		return nil, err
	}
	return s.issueTokens(ctx, u)
}

func (s *Service) findOrCreateByEmail(ctx context.Context, id *security.GoogleIdentity) (*user.User, error) {
	email := normalizeEmail(id.Email)
	u, err := s.users.GetByEmail(ctx, email)
	switch {
	case err == nil:
		if u.GoogleID != "" && u.GoogleID != id.Subject {
			// The email belongs to a user already linked to a different
			// Google account; don't hand this one access to it.
			return nil, fmt.Errorf("%w: this email is linked to a different Google account", apperr.ErrConflict)
		}
		if u.GoogleID == "" {
			if err := s.users.LinkGoogle(ctx, u.ID, id.Subject); err != nil {
				return nil, err
			}
			u.GoogleID = id.Subject
		}
		return u, nil
	case errors.Is(err, apperr.ErrNotFound):
		now := time.Now().UTC()
		u = &user.User{
			ID:        s.ids.New("usr"),
			Email:     email,
			Name:      strings.TrimSpace(id.Name),
			GoogleID:  id.Subject,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := s.users.Create(ctx, u); err != nil {
			return nil, err
		}
		return u, nil
	default:
		return nil, err
	}
}
