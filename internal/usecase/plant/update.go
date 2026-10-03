package plant

import (
	"context"
	"fmt"
	"strings"
	"time"

	"plantpal-backend/internal/domain/apperr"
	"plantpal-backend/internal/domain/plant"
)

type UpdateInput struct {
	Name, Type, AgeStage  *string
	Location, Sunlight    *string
	Outdoor               *bool
	WateringFrequencyDays *int
	Lang                  string
}

// Update edits name/type/age-stage/location/sunlight/watering-frequency.
//
// The care roadmap (and the fertilizing schedule) is regenerated only when
// Type or AgeStage really changed. Changing the watering interval re-anchors
// the next watering on LastWateredAt, never on "now"; a plant that was never
// watered keeps its due date.
func (s *Service) Update(ctx context.Context, id, userID string, in UpdateInput) (*plant.Plant, error) {
	p, err := s.repo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	regenerate := false
	if in.Name != nil {
		name := strings.TrimSpace(*in.Name)
		if name == "" {
			return nil, fmt.Errorf("%w: name is required", apperr.ErrInvalidInput)
		}
		p.Name = name
	}
	if in.Type != nil {
		t := strings.TrimSpace(*in.Type)
		if !strings.EqualFold(t, p.Type) {
			regenerate = true
		}
		p.Type = t
	}
	if in.AgeStage != nil {
		a := strings.TrimSpace(*in.AgeStage)
		if !strings.EqualFold(a, p.AgeStage) {
			regenerate = true
		}
		p.AgeStage = a
	}
	if in.Location != nil {
		p.Location = strings.TrimSpace(*in.Location)
	}
	if in.Sunlight != nil {
		p.Sunlight = strings.TrimSpace(*in.Sunlight)
	}
	if in.Outdoor != nil {
		p.Outdoor = *in.Outdoor
	}
	now := time.Now().UTC()
	if in.WateringFrequencyDays != nil && *in.WateringFrequencyDays > 0 && *in.WateringFrequencyDays != p.WateringFrequencyDays {
		p.WateringFrequencyDays = *in.WateringFrequencyDays
		if p.LastWateredAt != nil {
			p.NextWateringAt = scheduleNextWatering(ctx, p, *p.LastWateredAt)
		}
		// never watered: keep the existing due date
	}
	if regenerate {
		p.CareRoadmap = buildRoadmap(p.Type, p.AgeStage, in.Lang)
		if p.WateringFrequencyDays <= 0 && p.LastWateredAt != nil {
			p.NextWateringAt = scheduleNextWatering(ctx, p, *p.LastWateredAt)
		}
		setNextFertilizing(p, now)
	}
	p.UpdatedAt = now
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}
	return s.withHealth(ctx, p), nil
}
