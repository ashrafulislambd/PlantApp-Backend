package memory

import (
	"context"
	"sort"
	"strings"
	"sync"

	"myplantpal-backend/internal/domain/apperr"
	"myplantpal-backend/internal/domain/fertilizer"
)

type FertilizerRepository struct {
	mu    sync.RWMutex
	items map[string]*fertilizer.Fertilizer
}

// NewFertilizerRepository optionally seeds initial data (e.g. the recipes
// already present in the Figma design).
func NewFertilizerRepository(seed ...*fertilizer.Fertilizer) *FertilizerRepository {
	items := make(map[string]*fertilizer.Fertilizer, len(seed))
	for _, f := range seed {
		items[f.ID] = f
	}
	return &FertilizerRepository{items: items}
}

func (r *FertilizerRepository) Create(_ context.Context, f *fertilizer.Fertilizer) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[f.ID] = f
	return nil
}

func (r *FertilizerRepository) GetByID(_ context.Context, id string) (*fertilizer.Fertilizer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	f, ok := r.items[id]
	if !ok {
		return nil, apperr.ErrNotFound
	}
	return f, nil
}

func (r *FertilizerRepository) List(_ context.Context, query string) ([]*fertilizer.Fertilizer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	q := strings.ToLower(strings.TrimSpace(query))
	out := make([]*fertilizer.Fertilizer, 0, len(r.items))
	for _, f := range r.items {
		if q == "" || strings.Contains(strings.ToLower(f.Name), q) || strings.Contains(strings.ToLower(f.Category), q) {
			out = append(out, f)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}
