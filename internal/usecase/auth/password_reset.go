package auth

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
	"time"

	"plantpal-backend/internal/domain/apperr"
	"plantpal-backend/internal/infrastructure/security"
)

// resetCode is a short-lived, single-use code emailed to the user for the
// forgot-password flow. Kept in memory: fine for the single backend
// instance this deployment runs, same tradeoff as the Google OAuth session
// store (see google_oauth_handler.go).
type resetCode struct {
	code      string
	expiresAt time.Time
}

const resetCodeTTL = 15 * time.Minute

// ForgotPassword issues a reset code for email, if that account can
// actually use one. To avoid leaking which emails are registered, an
// unknown email is treated the same as a successful send (no error). A
// Google-only account (no password set) gets a distinct, explicit error
// instead - that detail is worth surfacing to the user.
func (s *Service) ForgotPassword(ctx context.Context, email string) error {
	email = normalizeEmail(email)
	u, err := s.users.GetByEmail(ctx, email)
	if errors.Is(err, apperr.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if u.PasswordHash == "" && u.GoogleID != "" {
		return fmt.Errorf("%w: this account uses Google sign-in - there's no password to reset", apperr.ErrInvalidInput)
	}

	code, err := generateResetCode()
	if err != nil {
		return err
	}

	s.resetMu.Lock()
	s.resetCodes[email] = resetCode{code: code, expiresAt: time.Now().Add(resetCodeTTL)}
	s.resetMu.Unlock()

	// TODO: send via a real email provider. Logged for now so the flow is
	// testable end-to-end before that infrastructure exists.
	log.Printf("password reset code for %s: %s (expires in %s)", email, code, resetCodeTTL)
	return nil
}

// ResetPassword consumes a code issued by ForgotPassword and sets a new
// password. The code is removed on the first attempt regardless of
// outcome - a wrong code doesn't get a second guess.
func (s *Service) ResetPassword(ctx context.Context, email, code, newPassword string) error {
	email = normalizeEmail(email)
	if len(newPassword) < 8 {
		return fmt.Errorf("%w: password must be at least 8 characters", apperr.ErrInvalidInput)
	}

	s.resetMu.Lock()
	stored, ok := s.resetCodes[email]
	delete(s.resetCodes, email)
	s.resetMu.Unlock()

	if !ok || time.Now().After(stored.expiresAt) || stored.code != code {
		return fmt.Errorf("%w: invalid or expired reset code", apperr.ErrUnauthorized)
	}

	u, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return err
	}
	hash, err := security.HashPassword(newPassword)
	if err != nil {
		return err
	}
	return s.users.UpdatePasswordHash(ctx, u.ID, hash)
}

func generateResetCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}
