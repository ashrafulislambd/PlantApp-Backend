package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"plantpal-backend/internal/domain/apperr"
	"plantpal-backend/internal/domain/refreshtoken"
	"plantpal-backend/internal/domain/user"
	"plantpal-backend/internal/idgen"
	"plantpal-backend/internal/infrastructure/security"
)

type fakeUsers struct{ byID map[string]*user.User }

func newFakeUsers(us ...*user.User) *fakeUsers {
	f := &fakeUsers{byID: map[string]*user.User{}}
	for _, u := range us {
		f.byID[u.ID] = u
	}
	return f
}

func (f *fakeUsers) Create(_ context.Context, u *user.User) error {
	for _, x := range f.byID {
		if x.Email == u.Email {
			return apperr.ErrConflict
		}
	}
	f.byID[u.ID] = u
	return nil
}

func (f *fakeUsers) find(match func(*user.User) bool) (*user.User, error) {
	for _, u := range f.byID {
		if match(u) {
			c := *u
			return &c, nil
		}
	}
	return nil, apperr.ErrNotFound
}

func (f *fakeUsers) GetByEmail(_ context.Context, e string) (*user.User, error) {
	return f.find(func(u *user.User) bool { return u.Email == e })
}
func (f *fakeUsers) GetByID(_ context.Context, id string) (*user.User, error) {
	return f.find(func(u *user.User) bool { return u.ID == id })
}
func (f *fakeUsers) GetByGoogleID(_ context.Context, g string) (*user.User, error) {
	return f.find(func(u *user.User) bool { return u.GoogleID != "" && u.GoogleID == g })
}
func (f *fakeUsers) LinkGoogle(_ context.Context, id, g string) error {
	u, ok := f.byID[id]
	if !ok {
		return apperr.ErrNotFound
	}
	u.GoogleID = g
	return nil
}
func (f *fakeUsers) UpdatePasswordHash(_ context.Context, id, hash string) error {
	u, ok := f.byID[id]
	if !ok {
		return apperr.ErrNotFound
	}
	u.PasswordHash = hash
	return nil
}

type fakeTokens struct{ n int }

func (f *fakeTokens) Create(context.Context, *refreshtoken.RefreshToken) error { f.n++; return nil }
func (f *fakeTokens) GetByHash(context.Context, string) (*refreshtoken.RefreshToken, error) {
	return nil, apperr.ErrNotFound
}
func (f *fakeTokens) Revoke(context.Context, string) error { return nil }

type fakeGoogle struct {
	id  *security.GoogleIdentity
	err error
}

func (f fakeGoogle) Verify(context.Context, string) (*security.GoogleIdentity, error) {
	return f.id, f.err
}

func newGoogleService(users *fakeUsers, g GoogleVerifier) *Service {
	svc := NewService(users, &fakeTokens{}, idgen.New(),
		security.NewJWTIssuer("test-secret-test-secret-test-secret", "test", 15*time.Minute), time.Hour)
	if g != nil {
		svc.SetGoogleVerifier(g)
	}
	return svc
}

var rahim = &security.GoogleIdentity{Subject: "g-123", Email: "Rahim@Example.com", Name: "Rahim"}

func TestLoginWithGoogle_NotConfigured(t *testing.T) {
	svc := newGoogleService(newFakeUsers(), nil)
	if _, err := svc.LoginWithGoogle(context.Background(), "tok"); !errors.Is(err, apperr.ErrUnavailable) {
		t.Errorf("error = %v, want ErrUnavailable", err)
	}
}

func TestLoginWithGoogle_EmptyToken(t *testing.T) {
	svc := newGoogleService(newFakeUsers(), fakeGoogle{id: rahim})
	if _, err := svc.LoginWithGoogle(context.Background(), "  "); !errors.Is(err, apperr.ErrInvalidInput) {
		t.Errorf("error = %v, want ErrInvalidInput", err)
	}
}

func TestLoginWithGoogle_InvalidToken(t *testing.T) {
	svc := newGoogleService(newFakeUsers(), fakeGoogle{err: security.ErrInvalidGoogleToken})
	if _, err := svc.LoginWithGoogle(context.Background(), "tok"); !errors.Is(err, apperr.ErrUnauthorized) {
		t.Errorf("error = %v, want ErrUnauthorized", err)
	}
}

func TestLoginWithGoogle_InfraErrorIsNotUnauthorized(t *testing.T) {
	boom := errors.New("google down")
	svc := newGoogleService(newFakeUsers(), fakeGoogle{err: boom})
	_, err := svc.LoginWithGoogle(context.Background(), "tok")
	if !errors.Is(err, boom) || errors.Is(err, apperr.ErrUnauthorized) {
		t.Errorf("error = %v, want the underlying infra error", err)
	}
}

func TestLoginWithGoogle_CreatesNewUser(t *testing.T) {
	users := newFakeUsers()
	svc := newGoogleService(users, fakeGoogle{id: rahim})

	res, err := svc.LoginWithGoogle(context.Background(), "tok")
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if res.AccessToken == "" || res.RefreshToken == "" {
		t.Error("expected access and refresh tokens")
	}
	if res.User.Email != "rahim@example.com" || res.User.Name != "Rahim" || res.User.GoogleID != "g-123" {
		t.Errorf("unexpected user: %+v", res.User)
	}
	if res.User.PasswordHash != "" {
		t.Error("Google-only user must have no password hash")
	}
	if len(users.byID) != 1 {
		t.Errorf("users = %d, want 1", len(users.byID))
	}
}

func TestLoginWithGoogle_LinksExistingEmailUser(t *testing.T) {
	existing := &user.User{ID: "usr_1", Email: "rahim@example.com", PasswordHash: "hash"}
	users := newFakeUsers(existing)
	svc := newGoogleService(users, fakeGoogle{id: rahim})

	res, err := svc.LoginWithGoogle(context.Background(), "tok")
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if res.User.ID != "usr_1" {
		t.Errorf("user ID = %s, want usr_1", res.User.ID)
	}
	if users.byID["usr_1"].GoogleID != "g-123" {
		t.Error("Google ID was not linked")
	}
	if len(users.byID) != 1 {
		t.Errorf("users = %d, want 1 (no duplicate account)", len(users.byID))
	}
}

func TestLoginWithGoogle_ReturningUserFoundByGoogleID(t *testing.T) {
	// Email changed since the first login: the Google ID still matches.
	existing := &user.User{ID: "usr_1", Email: "old@example.com", GoogleID: "g-123"}
	users := newFakeUsers(existing)
	svc := newGoogleService(users, fakeGoogle{id: rahim})

	res, err := svc.LoginWithGoogle(context.Background(), "tok")
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if res.User.ID != "usr_1" || len(users.byID) != 1 {
		t.Errorf("expected the existing user, got %+v (users=%d)", res.User, len(users.byID))
	}
}

func TestLoginWithGoogle_EmailLinkedToDifferentGoogleAccount(t *testing.T) {
	existing := &user.User{ID: "usr_1", Email: "rahim@example.com", GoogleID: "someone-else"}
	svc := newGoogleService(newFakeUsers(existing), fakeGoogle{id: rahim})
	if _, err := svc.LoginWithGoogle(context.Background(), "tok"); !errors.Is(err, apperr.ErrConflict) {
		t.Errorf("error = %v, want ErrConflict", err)
	}
}

func TestPasswordLoginFailsForGoogleOnlyUser(t *testing.T) {
	users := newFakeUsers()
	svc := newGoogleService(users, fakeGoogle{id: rahim})
	if _, err := svc.LoginWithGoogle(context.Background(), "tok"); err != nil {
		t.Fatal(err)
	}
	_, err := svc.Login(context.Background(), LoginInput{Email: "rahim@example.com", Password: "anything-at-all"})
	if !errors.Is(err, apperr.ErrUnauthorized) {
		t.Errorf("Login() error = %v, want ErrUnauthorized", err)
	}
}
