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
	Name, Type, AgeStage, Location, Sunlight, Image, Status, Humidity *string
	Health                                                            *int
	WateringFrequencyDays                                             *int
	LastWatered                                                       *time.Time
	LastScan                                                          *time.Time
	Lang                                                              string
}

// Update edits any subset of a plant's fields. Changing Type or AgeStage
// regenerates the roadmap/fertilizer note; changing LastWatered or
// WateringFrequencyDays recomputes the next watering time, same as
// Create/MarkWatered.
func (s *Service) Update(ctx context.Context, id, userID string, in UpdateInput) (*plant.Plant, error) {
	p, err := s.repo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	regenerateRoadmap := false
	if in.Name != nil {
		name := strings.TrimSpace(*in.Name)
		if name == "" {
			return nil, fmt.Errorf("%w: name is required", apperr.ErrInvalidInput)
		}
		p.Name = name
	}
	if in.Type != nil {
		p.Type, regenerateRoadmap = strings.TrimSpace(*in.Type), true
	}
	if in.AgeStage != nil {
		p.AgeStage, regenerateRoadmap = strings.TrimSpace(*in.AgeStage), true
	}
	if in.Location != nil {
		p.Location = strings.TrimSpace(*in.Location)
	}
	if in.Sunlight != nil {
		p.Sunlight = strings.TrimSpace(*in.Sunlight)
	}
	if in.Image != nil {
		p.Image = *in.Image
	}
	if in.Status != nil {
		p.Status = strings.TrimSpace(*in.Status)
	}
	if in.Humidity != nil {
		p.Humidity = strings.TrimSpace(*in.Humidity)
	}
	if in.Health != nil {
		p.Health = in.Health
	}
	if in.LastScan != nil {
		p.LastScan = in.LastScan
	}

	recomputeWateringNeeded := false
	if in.WateringFrequencyDays != nil && *in.WateringFrequencyDays > 0 {
		p.WateringFrequencyDays = *in.WateringFrequencyDays
		recomputeWateringNeeded = true
	}
	if in.LastWatered != nil {
		p.LastWateredAt = in.LastWatered
		recomputeWateringNeeded = true
	}

	now := time.Now().UTC()
	if regenerateRoadmap {
		roadmap, fertilizerNote := buildRoadmap(p.AgeStage, in.Lang)
		p.CareRoadmap, p.FertilizerNote = roadmap, fertilizerNote
		setNextFertilizing(p, now)
	}
	if recomputeWateringNeeded {
		recomputeWatering(p, now)
	}
	p.UpdatedAt = now
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}
