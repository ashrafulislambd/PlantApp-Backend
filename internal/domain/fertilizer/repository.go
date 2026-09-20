package fertilizer

import "context"

// Repository persists Fertilizers and supports the search bar on the
// Fertilizer screen ("Find your homemade fertilizer").
type Repository interface {
	Create(ctx context.Context, f *Fertilizer) error
	GetByID(ctx context.Context, id string) (*Fertilizer, error)
	// List returns fertilizers whose name or category contains query
	// (case-insensitive). An empty query returns all fertilizers.
	List(ctx context.Context, query string) ([]*Fertilizer, error)
}
