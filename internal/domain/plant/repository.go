package plant

import "context"

// Repository persists Plants. The in-memory implementation lives in
// internal/infrastructure/repository/memory; swap it for a MongoDB-backed
// implementation later without touching the usecase or delivery layers.
type Repository interface {
	Create(ctx context.Context, p *Plant) error
	GetByID(ctx context.Context, id string) (*Plant, error)
	List(ctx context.Context) ([]*Plant, error)
	Delete(ctx context.Context, id string) error
}
