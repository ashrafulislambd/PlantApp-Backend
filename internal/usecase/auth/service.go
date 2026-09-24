// Package auth implements registration, login, token refresh, and logout.
package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"myplantpal-backend/internal/domain/apperr"
	"myplantpal-backend/internal/domain/refreshtoken"
	"myplantpal-backend/internal/domain/user"
	"myplantpal-backend/internal/idgen"
	"myplantpal-backend/internal/infrastructure/security"
)

type Service struct {
	users      user.Repository
	tokens     refreshtoken.Repository
	ids        idgen.Generator
	jwt        *security.JWTIssuer
	refreshTTL time.Duration
}

func NewService(users user.Repository, tokens refreshtoken.Repository, ids idgen.Generator, jwtIssuer *security.JWTIssuer, refreshTTL time.Duration) *Service {
	return &Service{users: users, tokens: tokens, ids: ids, jwt: jwtIssuer, refreshTTL: refreshTTL}
}

type AuthResult struct {
	User         *user.User
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

type RegisterInput struct {
	Email    string
	Password string
	Name     string
}

func (s *Service) Register(ctx context.Context, in RegisterInput) (*AuthResult, error) {
	email := normalizeEmail(in.Email)
	if email == "" || !strings.Contains(email, "@") {
		return nil, fmt.Errorf("%w: a valid email is required", apperr.ErrInvalidInput)
	}
	if len(in.Password) < 8 {
		return nil, fmt.Errorf("%w: password must be at least 8 characters", apperr.ErrInvalidInput)
	}

	hash, err := security.HashPassword(in.Password)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	u := &user.User{
		ID:           s.ids.New("usr"),
		Email:        email,
		Name:         strings.TrimSpace(in.Name),
		PasswordHash: hash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.users.Create(ctx, u); err != nil {
		return nil, err
	}
	return s.issueTokens(ctx, u)
}

type LoginInput struct {
	Email    string
	Password string
}

func (s *Service) Login(ctx context.Context, in LoginInput) (*AuthResult, error) {
	email := normalizeEmail(in.Email)
	u, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			return nil, fmt.Errorf("%w: invalid email or password", apperr.ErrUnauthorized)
		}
		return nil, err
	}
	if err := security.ComparePassword(u.PasswordHash, in.Password); err != nil {
		return nil, fmt.Errorf("%w: invalid email or password", apperr.ErrUnauthorized)
	}
	return s.issueTokens(ctx, u)
}

func (s *Service) Refresh(ctx context.Context, rawToken string) (*AuthResult, error) {
	hash := hashToken(rawToken)
	rt, err := s.tokens.GetByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			return nil, fmt.Errorf("%w: invalid refresh token", apperr.ErrUnauthorized)
		}
		return nil, err
	}
	if rt.RevokedAt != nil || time.Now().UTC().After(rt.ExpiresAt) {
		return nil, fmt.Errorf("%w: refresh token expired", apperr.ErrUnauthorized)
	}
	u, err := s.users.GetByID(ctx, rt.UserID)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid refresh token", apperr.ErrUnauthorized)
	}
	if err := s.tokens.Revoke(ctx, rt.ID); err != nil {
		return nil, err
	}
	return s.issueTokens(ctx, u)
}

func (s *Service) Logout(ctx context.Context, rawToken string) error {
	hash := hashToken(rawToken)
	rt, err := s.tokens.GetByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			return nil
		}
		return err
	}
	return s.tokens.Revoke(ctx, rt.ID)
}

func (s *Service) Me(ctx context.Context, userID string) (*user.User, error) {
	return s.users.GetByID(ctx, userID)
}

func (s *Service) issueTokens(ctx context.Context, u *user.User) (*AuthResult, error) {
	access, expiresAt, err := s.jwt.NewAccessToken(u.ID)
	if err != nil {
		return nil, err
	}
	rawRefresh, err := security.GenerateOpaqueToken()
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	rt := &refreshtoken.RefreshToken{
		ID:        s.ids.New("rt"),
		UserID:    u.ID,
		TokenHash: hashToken(rawRefresh),
		ExpiresAt: now.Add(s.refreshTTL),
		CreatedAt: now,
	}
	if err := s.tokens.Create(ctx, rt); err != nil {
		return nil, err
	}
	return &AuthResult{
		User:         u,
		AccessToken:  access,
		RefreshToken: rawRefresh,
		ExpiresIn:    int64(time.Until(expiresAt).Seconds()),
	}, nil
}

func normalizeEmail(e string) string { return strings.ToLower(strings.TrimSpace(e)) }

func hashToken(t string) string {
	sum := sha256.Sum256([]byte(t))
	return hex.EncodeToString(sum[:])
}
