package plant

import "context"

// EventRepository stores immutable plant-care events, scoped to their owner.
type EventRepository interface {
	Create(ctx context.Context, event *PlantEvent) error
	ListByPlant(ctx context.Context, plantID, userID string) ([]*PlantEvent, error)
	ListByUser(ctx context.Context, userID string) ([]*PlantEvent, error)
}
