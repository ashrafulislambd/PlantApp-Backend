package diagnosis

import "context"

// Repository persists Diagnoses (the "Add to Log" action on the Diseases
// Detection screen).
type Repository interface {
	Create(ctx context.Context, d *Diagnosis) error
	GetByID(ctx context.Context, id string) (*Diagnosis, error)
	// List returns diagnoses, optionally filtered by plantID.
	List(ctx context.Context, plantID *string) ([]*Diagnosis, error)
}
