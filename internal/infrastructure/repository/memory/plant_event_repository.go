package memory

import (
	"context"
	"sort"
	"sync"

	"plantpal-backend/internal/domain/plant"
)

type PlantEventRepository struct {
	mu    sync.RWMutex
	items map[string]*plant.PlantEvent
}

func NewPlantEventRepository() *PlantEventRepository {
	return &PlantEventRepository{items: make(map[string]*plant.PlantEvent)}
}

func (r *PlantEventRepository) Create(_ context.Context, event *plant.PlantEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[event.ID] = event
	return nil
}

func (r *PlantEventRepository) ListByPlant(_ context.Context, plantID, userID string) ([]*plant.PlantEvent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*plant.PlantEvent, 0)
	for _, event := range r.items {
		if event.PlantID == plantID && event.UserID == userID {
			out = append(out, event)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

func (r *PlantEventRepository) ListByUser(_ context.Context, userID string) ([]*plant.PlantEvent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*plant.PlantEvent, 0)
	for _, event := range r.items {
		if event.UserID == userID {
			out = append(out, event)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}
