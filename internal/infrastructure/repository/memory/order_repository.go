package memory

import (
	"context"
	"sort"
	"sync"

	"plantpal-backend/internal/domain/apperr"
	"plantpal-backend/internal/domain/order"
)

// OrderRepository is an in-memory implementation of order.Repository, used
// by usecase tests so they don't need a live MongoDB.
type OrderRepository struct {
	mu    sync.RWMutex
	items map[string]*order.Order
}

func NewOrderRepository() *OrderRepository {
	return &OrderRepository{items: make(map[string]*order.Order)}
}

func (r *OrderRepository) Create(_ context.Context, o *order.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[o.ID] = o
	return nil
}

func (r *OrderRepository) GetByID(_ context.Context, id, userID string) (*order.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	o, ok := r.items[id]
	if !ok || o.UserID != userID {
		return nil, apperr.ErrNotFound
	}
	return o, nil
}

// List returns userID's orders, newest first — matching the Mongo
// repository's sort order for order history.
func (r *OrderRepository) List(_ context.Context, userID string) ([]*order.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*order.Order, 0, len(r.items))
	for _, o := range r.items {
		if o.UserID == userID {
			out = append(out, o)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}
