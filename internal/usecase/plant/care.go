package plant

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"plantpal-backend/internal/domain/apperr"
	"plantpal-backend/internal/domain/plant"
)

type CareResult struct {
	Plant        *plant.Plant
	PointsEarned int
}

// MarkWatered keeps the original use-case API for existing callers.
func (s *Service) MarkWatered(ctx context.Context, id, userID string) (*plant.Plant, error) {
	result, err := s.MarkWateredWithPoints(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	return result.Plant, nil
}

// MarkWateredWithPoints logs a watering and awards points only when it was due.
func (s *Service) MarkWateredWithPoints(ctx context.Context, id, userID string) (CareResult, error) {
	p, err := s.repo.GetByID(ctx, id, userID)
	if err != nil {
		return CareResult{}, err
	}
	now := time.Now().UTC()
	points := 0
	if !p.NextWateringAt.After(now) {
		points = 10
	}
	p.LastWateredAt = &now
	p.NextWateringAt = scheduleNextWatering(ctx, p, now)
	p.UpdatedAt = now
	if err := s.repo.Update(ctx, p); err != nil {
		return CareResult{}, err
	}
	if err := s.recordEvent(ctx, &plant.PlantEvent{PlantID: p.ID, UserID: userID, EventType: plant.EventWater, Points: points, CreatedAt: now}); err != nil {
		return CareResult{}, err
	}
	return CareResult{Plant: s.withHealth(ctx, p), PointsEarned: points}, nil
}

// MarkFertilized logs fertilizing now and rolls NextFertilizingAt forward.
// Errors if the roadmap doesn't recommend fertilizing yet (e.g. seedlings).
func (s *Service) MarkFertilized(ctx context.Context, id, userID string) (*plant.Plant, error) {
	result, err := s.MarkFertilizedWithPoints(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	return result.Plant, nil
}

func (s *Service) MarkFertilizedWithPoints(ctx context.Context, id, userID string) (CareResult, error) {
	p, err := s.repo.GetByID(ctx, id, userID)
	if err != nil {
		return CareResult{}, err
	}
	if p.CareRoadmap.FertilizingIntervalDays <= 0 {
		return CareResult{}, fmt.Errorf("%w: fertilizing not recommended yet", apperr.ErrInvalidInput)
	}
	now := time.Now().UTC()
	points := 0
	if p.NextFertilizingAt != nil && !p.NextFertilizingAt.After(now) {
		points = 20
	}
	p.LastFertilizedAt = &now
	next := now.AddDate(0, 0, p.CareRoadmap.FertilizingIntervalDays)
	p.NextFertilizingAt, p.UpdatedAt = &next, now
	if err := s.repo.Update(ctx, p); err != nil {
		return CareResult{}, err
	}
	if err := s.recordEvent(ctx, &plant.PlantEvent{PlantID: p.ID, UserID: userID, EventType: plant.EventFertilize, Points: points, CreatedAt: now}); err != nil {
		return CareResult{}, err
	}
	return CareResult{Plant: s.withHealth(ctx, p), PointsEarned: points}, nil
}

func (s *Service) Skip(ctx context.Context, id, userID, reason string, days int) (*plant.Plant, error) {
	switch reason {
	case "soil_still_wet", "rained", "other":
	default:
		return nil, fmt.Errorf("%w: unsupported skip reason", apperr.ErrInvalidInput)
	}
	if days == 0 {
		days = 1
	}
	if days < 1 || days > 30 {
		return nil, fmt.Errorf("%w: snooze days must be from 1 to 30", apperr.ErrInvalidInput)
	}
	p, err := s.repo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	p.NextWateringAt = now.AddDate(0, 0, days)
	p.UpdatedAt = now
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}
	if err := s.recordEvent(ctx, &plant.PlantEvent{
		PlantID: p.ID, UserID: userID, EventType: plant.EventSkip,
		Reason: reason, Metadata: map[string]any{"snoozeDays": days}, CreatedAt: now,
	}); err != nil {
		return nil, err
	}
	return s.withHealth(ctx, p), nil
}

func (s *Service) AddNote(ctx context.Context, id, userID, note string) error {
	note = strings.TrimSpace(note)
	if note == "" {
		return fmt.Errorf("%w: note is required", apperr.ErrInvalidInput)
	}
	if len([]rune(note)) > 1000 {
		note = string([]rune(note)[:1000])
	}
	p, err := s.repo.GetByID(ctx, id, userID)
	if err != nil {
		return err
	}
	return s.recordEvent(ctx, &plant.PlantEvent{
		PlantID: p.ID, UserID: userID, EventType: plant.EventNote, Note: note, CreatedAt: time.Now().UTC(),
	})
}

func (s *Service) PlantEvents(ctx context.Context, id, userID string) ([]*plant.PlantEvent, error) {
	if _, err := s.repo.GetByID(ctx, id, userID); err != nil {
		return nil, err
	}
	if s.events == nil {
		return []*plant.PlantEvent{}, nil
	}
	return s.events.ListByPlant(ctx, id, userID)
}

func (s *Service) GardenEvents(ctx context.Context, userID string) ([]*plant.PlantEvent, error) {
	if s.events == nil {
		return []*plant.PlantEvent{}, nil
	}
	return s.events.ListByUser(ctx, userID)
}

// RecordScanEvent is called by the diagnosis use case after saving a linked scan.
func (s *Service) RecordScanEvent(ctx context.Context, userID, plantID, imageURL, note string, metadata map[string]any) error {
	if _, err := s.repo.GetByID(ctx, plantID, userID); err != nil {
		return err
	}
	return s.recordEvent(ctx, &plant.PlantEvent{
		PlantID: plantID, UserID: userID, EventType: plant.EventScan,
		ImageURL: imageURL, Note: note, Metadata: metadata, CreatedAt: time.Now().UTC(),
	})
}

func (s *Service) recordEvent(ctx context.Context, event *plant.PlantEvent) error {
	if s.events == nil {
		return nil
	}
	event.ID = s.ids.New("evt")
	return s.events.Create(ctx, event)
}

// Due returns the user's plants with a watering/fertilizing reminder due at
// or before `before`, soonest first. Polled by the notifications module.
func (s *Service) Due(ctx context.Context, userID string, before time.Time) ([]*plant.Plant, error) {
	items, err := s.repo.ListDue(ctx, userID, before)
	if err != nil {
		return nil, err
	}
	sort.Slice(items, func(i, j int) bool { return items[i].NextWateringAt.Before(items[j].NextWateringAt) })
	return s.withHealthAll(ctx, userID, items), nil
}

// scheduleNextWatering returns when the plant is next due after a watering at
// `from`: from + wateringFrequencyDays. Only when the interval is 0 does it
// fall back to the roadmap's clock times, read in the user's time zone
// (plant.LocationFrom(ctx), UTC when the client did not send one).
func scheduleNextWatering(ctx context.Context, p *plant.Plant, from time.Time) time.Time {
	if p.WateringFrequencyDays > 0 {
		return from.AddDate(0, 0, p.WateringFrequencyDays)
	}
	local := from.In(plant.LocationFrom(ctx))
	return nextWateringTime(p.CareRoadmap.WateringTimes, local).UTC()
}

// setNextFertilizing (re)computes NextFertilizingAt. It anchors on
// LastFertilizedAt when known, otherwise on `now`.
func setNextFertilizing(p *plant.Plant, now time.Time) {
	p.NextFertilizingAt = nil
	if p.CareRoadmap.FertilizingIntervalDays > 0 {
		anchor := now
		if p.LastFertilizedAt != nil {
			anchor = *p.LastFertilizedAt
		}
		next := anchor.AddDate(0, 0, p.CareRoadmap.FertilizingIntervalDays)
		p.NextFertilizingAt = &next
	}
}
