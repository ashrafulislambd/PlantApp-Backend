package memory

import (
	"context"
	"sort"
	"sync"
	"time"

	"plantpal-backend/internal/domain/apperr"
	"plantpal-backend/internal/domain/plant"
)

type PlantRepository struct {
	mu    sync.RWMutex
	items map[string]*plant.Plant
}

func NewPlantRepository() *PlantRepository {
	return &PlantRepository{items: make(map[string]*plant.Plant)}
}

func (r *PlantRepository) Create(_ context.Context, p *plant.Plant) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[p.ID] = p
	return nil
}

func (r *PlantRepository) GetByID(_ context.Context, id string) (*plant.Plant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.items[id]
	if !ok {
		return nil, apperr.ErrNotFound
	}
	return p, nil
}

func (r *PlantRepository) List(_ context.Context) ([]*plant.Plant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*plant.Plant, 0, len(r.items))
	for _, p := range r.items {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out, nil
}

func (r *PlantRepository) Update(_ context.Context, p *plant.Plant) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[p.ID]; !ok {
		return apperr.ErrNotFound
	}
	r.items[p.ID] = p
	return nil
}

func (r *PlantRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[id]; !ok {
		return apperr.ErrNotFound
	}
	delete(r.items, id)
	return nil
}

func (r *PlantRepository) ListDue(_ context.Context, before time.Time) ([]*plant.Plant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*plant.Plant, 0)
	for _, p := range r.items {
		watering := !p.NextWateringAt.IsZero() && !p.NextWateringAt.After(before)
		fertilizing := p.NextFertilizingAt != nil && !p.NextFertilizingAt.After(before)
		if watering || fertilizing {
			out = append(out, p)
		}
	}
	return out, nil
} 
