// Package memory provides in-memory implementations of the domain
// repository interfaces, guarded by mutexes. These are a stand-in for the
// eventual MongoDB-backed implementations described in CLAUDE.md — swap
// them out by constructing a different type that satisfies the same
// domain.Repository interface; no usecase or handler code needs to change.
package memory

import (
	"context"
	"sort"
	"sync"

	"myplantpal-backend/internal/domain/apperr"
	"myplantpal-backend/internal/domain/plant"
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

func (r *PlantRepository) GetByID(_ context.Context, id, userID string) (*plant.Plant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.items[id]
	if !ok || p.UserID != userID {
		return nil, apperr.ErrNotFound
	}
	return p, nil
}

func (r *PlantRepository) List(_ context.Context, userID string) ([]*plant.Plant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*plant.Plant, 0, len(r.items))
	for _, p := range r.items {
		if p.UserID != userID {
			continue
		}
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out, nil
}

func (r *PlantRepository) Delete(_ context.Context, id, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.items[id]
	if !ok || p.UserID != userID {
		return apperr.ErrNotFound
	}
	delete(r.items, id)
	return nil
}
