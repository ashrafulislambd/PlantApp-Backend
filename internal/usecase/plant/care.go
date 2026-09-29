package plant

import (
	"context"
	"fmt"
	"sort"
	"time"

	"plantpal-backend/internal/domain/apperr"
	"plantpal-backend/internal/domain/plant"
)

// MarkWatered logs a watering now and rolls NextWateringAt to the next
// scheduled time-of-day.
func (s *Service) MarkWatered(ctx context.Context, id string) (*plant.Plant, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	p.LastWateredAt = &now
	p.NextWateringAt = nextWateringTime(p.CareRoadmap.WateringTimes, now)
	p.UpdatedAt = now
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// MarkFertilized logs fertilizing now and rolls NextFertilizingAt forward.
// Errors if the roadmap doesn't recommend fertilizing yet (e.g. seedlings).
func (s *Service) MarkFertilized(ctx context.Context, id string) (*plant.Plant, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p.CareRoadmap.FertilizingIntervalDays <= 0 {
		return nil, fmt.Errorf("%w: fertilizing not recommended yet", apperr.ErrInvalidInput)
	}
	now := time.Now().UTC()
	p.LastFertilizedAt = &now
	next := now.AddDate(0, 0, p.CareRoadmap.FertilizingIntervalDays)
	p.NextFertilizingAt, p.UpdatedAt = &next, now
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// Due returns plants with a watering/fertilizing reminder due at or
// before `before`, soonest first. Polled by the notifications module.
func (s *Service) Due(ctx context.Context, before time.Time) ([]*plant.Plant, error) {
	items, err := s.repo.ListDue(ctx, before)
	if err != nil {
		return nil, err
	}
	sort.Slice(items, func(i, j int) bool { return items[i].NextWateringAt.Before(items[j].NextWateringAt) })
	return items, nil
}

func setNextFertilizing(p *plant.Plant, now time.Time) {
	p.NextFertilizingAt = nil
	if p.CareRoadmap.FertilizingIntervalDays > 0 {
		next := now.AddDate(0, 0, p.CareRoadmap.FertilizingIntervalDays)
		p.NextFertilizingAt = &next
	}
}
