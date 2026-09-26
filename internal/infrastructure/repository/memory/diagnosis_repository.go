package memory

import (
	"context"
	"sort"
	"sync"

	"myplantpal-backend/internal/domain/apperr"
	"myplantpal-backend/internal/domain/diagnosis"
)

type DiagnosisRepository struct {
	mu    sync.RWMutex
	items map[string]*diagnosis.Diagnosis
}

func NewDiagnosisRepository() *DiagnosisRepository {
	return &DiagnosisRepository{items: make(map[string]*diagnosis.Diagnosis)}
}

func (r *DiagnosisRepository) Create(_ context.Context, d *diagnosis.Diagnosis) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[d.ID] = d
	return nil
}

func (r *DiagnosisRepository) GetByID(_ context.Context, id, userID string) (*diagnosis.Diagnosis, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	d, ok := r.items[id]
	if !ok || d.UserID != userID {
		return nil, apperr.ErrNotFound
	}
	return d, nil
}

func (r *DiagnosisRepository) List(_ context.Context, userID string, plantID *string) ([]*diagnosis.Diagnosis, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*diagnosis.Diagnosis, 0, len(r.items))
	for _, d := range r.items {
		if d.UserID != userID {
			continue
		}
		if plantID != nil {
			if d.PlantID == nil || *d.PlantID != *plantID {
				continue
			}
		}
		out = append(out, d)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out, nil
}
