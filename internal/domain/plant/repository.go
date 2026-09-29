package plant

import (
	"context"
	"time"
)

type Repository interface {
	Create(ctx context.Context, p *Plant) error
	GetByID(ctx context.Context, id string) (*Plant, error)
	List(ctx context.Context) ([]*Plant, error)
	Update(ctx context.Context, p *Plant) error
	Delete(ctx context.Context, id string) error
	// ListDue returns plants whose next watering/fertilizing is due at
	// or before `before`. Used by reminders and the notifications module.
	ListDue(ctx context.Context, before time.Time) ([]*Plant, error)
}