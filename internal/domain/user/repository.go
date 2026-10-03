package user

import "context"

// Repository persists Users. The MongoDB implementation lives in
// internal/infrastructure/repository/mongo.
type Repository interface {
	Create(ctx context.Context, u *User) error
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id string) (*User, error)
	GetByGoogleID(ctx context.Context, googleID string) (*User, error)
	// LinkGoogle attaches a Google account to an existing user.
	LinkGoogle(ctx context.Context, id, googleID string) error
	// UpdatePasswordHash sets (or, for a Google-only account, adds) a
	// user's password. Used by Register and by the forgot/reset-password
	// flow.
	UpdatePasswordHash(ctx context.Context, id, passwordHash string) error
}
