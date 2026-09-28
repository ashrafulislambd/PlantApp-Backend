package order

import "context"

// Repository persists Orders (the "Continue to Payment" -> pay action on
// the Checkout/Payment screens). The MongoDB implementation lives in
// internal/infrastructure/repository/mongo.
//
// GetByID and List are scoped to a single owning userID: an order
// belonging to another user is treated as not found.
type Repository interface {
	Create(ctx context.Context, o *Order) error
	GetByID(ctx context.Context, id, userID string) (*Order, error)
	// List returns userID's orders, newest first (order history).
	List(ctx context.Context, userID string) ([]*Order, error)
}
