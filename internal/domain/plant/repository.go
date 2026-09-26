package plant

import "context"

// Repository persists Plants. The MongoDB implementation lives in
// internal/infrastructure/repository/mongo; an in-memory implementation
// also exists in internal/infrastructure/repository/memory for local dev
// without a database.
//
// Every method besides Create is scoped to a single owning userID: a plant
// that exists but belongs to a different user is treated the same as one
// that doesn't exist (apperr.ErrNotFound), so callers can't probe for other
// users' data.
type Repository interface {
	Create(ctx context.Context, p *Plant) error
	GetByID(ctx context.Context, id, userID string) (*Plant, error)
	List(ctx context.Context, userID string) ([]*Plant, error)
	Delete(ctx context.Context, id, userID string) error
}
