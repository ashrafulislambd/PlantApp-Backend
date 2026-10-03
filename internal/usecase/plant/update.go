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
	WateringFrequencyDays *int
	Lang                  string
}

// Update edits name/type/age-stage/location/sunlight/watering-frequency.
// Changing Type or AgeStage regenerates the roadmap and recomputes the next-due times, same as Create.
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
		p.Type, regenerate = strings.TrimSpace(*in.Type), true
	}
	if in.AgeStage != nil {
		p.AgeStage, regenerate = strings.TrimSpace(*in.AgeStage), true
	}
	if in.Location != nil {
		p.Location = strings.TrimSpace(*in.Location)
	}
	if in.Sunlight != nil {
		p.Sunlight = strings.TrimSpace(*in.Sunlight)
	}
	now := time.Now().UTC()
	if in.WateringFrequencyDays != nil && *in.WateringFrequencyDays > 0 {
		p.WateringFrequencyDays = *in.WateringFrequencyDays
		p.NextWateringAt = now.AddDate(0, 0, *in.WateringFrequencyDays)
	}
	if regenerate {
		p.CareRoadmap = buildRoadmap(p.Type, p.AgeStage, in.Lang)
		if p.WateringFrequencyDays <= 0 {
			p.NextWateringAt = nextWateringTime(p.CareRoadmap.WateringTimes, now)
		}
		setNextFertilizing(p, now)
	}
	p.UpdatedAt = now
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}
