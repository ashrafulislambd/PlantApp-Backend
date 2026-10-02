package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"plantpal-backend/internal/domain/apperr"
	"plantpal-backend/internal/domain/user"
	"plantpal-backend/internal/idgen"
	"plantpal-backend/internal/infrastructure/security"
)

func newPasswordService(users *fakeUsers) *Service {
	return NewService(users, &fakeTokens{}, idgen.New(),
		security.NewJWTIssuer("test-secret-test-secret-test-secret", "test", 15*time.Minute), time.Hour)
}

func TestForgotPassword_UnknownEmailDoesNotError(t *testing.T) {
	svc := newPasswordService(newFakeUsers())
	if err := svc.ForgotPassword(context.Background(), "nobody@example.com"); err != nil {
		t.Errorf("error = %v, want nil (don't leak whether the email is registered)", err)
	}
}

func TestForgotPassword_GoogleOnlyAccountIsRejected(t *testing.T) {
	existing := &user.User{ID: "usr_1", Email: "rahim@example.com", GoogleID: "g-123"}
	svc := newPasswordService(newFakeUsers(existing))
	err := svc.ForgotPassword(context.Background(), "rahim@example.com")
	if !errors.Is(err, apperr.ErrInvalidInput) {
		t.Errorf("error = %v, want ErrInvalidInput", err)
	}
}

func TestForgotPasswordThenResetPassword(t *testing.T) {
	hash, _ := security.HashPassword("old-password")
	existing := &user.User{ID: "usr_1", Email: "rahim@example.com", PasswordHash: hash}
	users := newFakeUsers(existing)
	svc := newPasswordService(users)

	if err := svc.ForgotPassword(context.Background(), "rahim@example.com"); err != nil {
		t.Fatalf("ForgotPassword() error = %v", err)
	}
	svc.resetMu.Lock()
	code := svc.resetCodes["rahim@example.com"].code
	svc.resetMu.Unlock()
	if code == "" {
		t.Fatal("no reset code was issued")
	}

	if err := svc.ResetPassword(context.Background(), "rahim@example.com", code, "new-password"); err != nil {
		t.Fatalf("ResetPassword() error = %v", err)
	}
	if err := security.ComparePassword(users.byID["usr_1"].PasswordHash, "new-password"); err != nil {
		t.Error("password was not updated to the new value")
	}

	// The code is single-use.
	err := svc.ResetPassword(context.Background(), "rahim@example.com", code, "another-password")
	if !errors.Is(err, apperr.ErrUnauthorized) {
		t.Errorf("second use: error = %v, want ErrUnauthorized", err)
	}
}

func TestResetPassword_WrongCodeFails(t *testing.T) {
	existing := &user.User{ID: "usr_1", Email: "rahim@example.com", PasswordHash: "hash"}
	users := newFakeUsers(existing)
	svc := newPasswordService(users)

	if err := svc.ForgotPassword(context.Background(), "rahim@example.com"); err != nil {
		t.Fatalf("ForgotPassword() error = %v", err)
	}
	err := svc.ResetPassword(context.Background(), "rahim@example.com", "000000", "new-password-123")
	if !errors.Is(err, apperr.ErrUnauthorized) {
		t.Errorf("error = %v, want ErrUnauthorized", err)
	}
}

func TestResetPassword_ExpiredCodeFails(t *testing.T) {
	existing := &user.User{ID: "usr_1", Email: "rahim@example.com", PasswordHash: "hash"}
	users := newFakeUsers(existing)
	svc := newPasswordService(users)

	svc.resetCodes["rahim@example.com"] = resetCode{code: "123456", expiresAt: time.Now().Add(-time.Minute)}
	err := svc.ResetPassword(context.Background(), "rahim@example.com", "123456", "new-password-123")
	if !errors.Is(err, apperr.ErrUnauthorized) {
		t.Errorf("error = %v, want ErrUnauthorized", err)
	}
}

func TestResetPassword_TooShortPasswordRejected(t *testing.T) {
	svc := newPasswordService(newFakeUsers())
	err := svc.ResetPassword(context.Background(), "rahim@example.com", "123456", "short")
	if !errors.Is(err, apperr.ErrInvalidInput) {
		t.Errorf("error = %v, want ErrInvalidInput", err)
	}
}
