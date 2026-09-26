package diagnosis

import "context"

// Repository persists Diagnoses (the "Add to Log" action on the Diseases
// Detection screen). The MongoDB implementation lives in
// internal/infrastructure/repository/mongo; an in-memory implementation
// also exists in internal/infrastructure/repository/memory for local dev
// without a database.
//
// GetByID and List are scoped to a single owning userID: a diagnosis
// belonging to another user is treated as not found.
type Repository interface {
	Create(ctx context.Context, d *Diagnosis) error
	GetByID(ctx context.Context, id, userID string) (*Diagnosis, error)
	// List returns userID's diagnoses, optionally filtered by plantID.
	List(ctx context.Context, userID string, plantID *string) ([]*Diagnosis, error)
}
