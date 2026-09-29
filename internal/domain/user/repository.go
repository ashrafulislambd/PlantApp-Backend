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
}
