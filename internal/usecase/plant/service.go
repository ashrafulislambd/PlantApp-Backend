// Package plant implements the application logic for registering plants,
// generating their care roadmap, and tracking watering/fertilizing.
package plant

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"plantpal-backend/internal/domain/apperr"
	"plantpal-backend/internal/domain/plant"
	"plantpal-backend/internal/idgen"
)

type Service struct {
	repo       plant.Repository
	events     plant.EventRepository
	ids        idgen.Generator
	images     plant.ImageStore
	identifier plant.Identifier
	scans      plant.ScanSource
}

func NewService(repo plant.Repository, ids idgen.Generator) *Service {
	return &Service{repo: repo, ids: ids}
}

func (s *Service) SetEventRepository(repo plant.EventRepository) { s.events = repo }

func (s *Service) SetImageStore(store plant.ImageStore) { s.images = store }
func (s *Service) SetIdentifier(id plant.Identifier)    { s.identifier = id }

// SetScanSource lets health take the latest AI Doctor scan of each plant into
// account. Optional: without it, health ignores scans.
func (s *Service) SetScanSource(src plant.ScanSource) { s.scans = src }

// Upper bounds for the optional AI-identify values accepted on create.
const (
	maxWaterAmountMl = 5000
	maxCareTipsRunes = 1000
)

type CreateInput struct {
	UserID                string
	Name                  string
	Type                  string
	AgeStage              string
	Location              string
	Sunlight              string
	WateringFrequencyDays int
	// Outdoor marks a plant that lives outside (garden, balcony).
	Outdoor bool
	Image   string
	// LastWateredAt is when the user last watered the plant. Omitted (nil)
	// means "not watered / unknown": the plant is due right now.
	LastWateredAt *time.Time
	// LastWatered is an alias accepted for Flutter client compatibility.
	LastWatered *time.Time
	// WaterAmountMl and CareTips are optional values kept from the AI
	// identify step. When set they replace the rule-based roadmap values, so
	// the saved plan matches the preview the user saw.
	WaterAmountMl int
	CareTips      string
	// Lang is the requester's UI language ("en" or "bn"), used to pick
	// which language the generated tips/fertilizer text comes back in.
	Lang string
}

func (s *Service) Create(ctx context.Context, in CreateInput) (*plant.Plant, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, fmt.Errorf("%w: name is required", apperr.ErrInvalidInput)
	}
	if strings.TrimSpace(in.UserID) == "" {
		return nil, fmt.Errorf("%w: userID is required", apperr.ErrInvalidInput)
	}

	days := in.WateringFrequencyDays
	if days < 0 {
		days = 0
	}

	lastWatered := in.LastWateredAt
	if lastWatered == nil {
		lastWatered = in.LastWatered
	}

	now := time.Now().UTC()
	roadmap := buildRoadmap(in.Type, in.AgeStage, in.Lang)
	if in.WaterAmountMl > 0 && in.WaterAmountMl <= maxWaterAmountMl {
		roadmap.WaterAmountMl = in.WaterAmountMl
	}
	if tips := strings.TrimSpace(in.CareTips); tips != "" {
		if r := []rune(tips); len(r) > maxCareTipsRunes {
			tips = string(r[:maxCareTipsRunes])
		}
		roadmap.Tips = tips
	}
	p := &plant.Plant{
		ID:                    s.ids.New("pl"),
		UserID:                in.UserID,
		Name:                  name,
		Type:                  strings.TrimSpace(in.Type),
		AgeStage:              strings.TrimSpace(in.AgeStage),
		Location:              strings.TrimSpace(in.Location),
		Sunlight:              strings.TrimSpace(in.Sunlight),
		Outdoor:               in.Outdoor,
		ImageURL:              in.Image,
		WateringFrequencyDays: days,
		CareRoadmap:           roadmap,
		FertilizerNote:        roadmap.FertilizerRecommendation,
		NextWateringAt:        now, // not watered / unknown: due now
		CreatedAt:             now,
		UpdatedAt:             now,
	}
	if lastWatered != nil {
		last := lastWatered.UTC()
		if last.After(now) {
			last = now // a watering cannot be in the future
		}
		p.LastWateredAt = &last
		p.NextWateringAt = scheduleNextWatering(ctx, p, last)
	}
	p.WaterLevel = waterLevelFor(p.NextWateringAt, now)
	setNextFertilizing(p, now)
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	return s.withHealth(ctx, p), nil
}

func (s *Service) Get(ctx context.Context, id, userID string) (*plant.Plant, error) {
	p, err := s.repo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	return s.withHealth(ctx, p), nil
}

func (s *Service) List(ctx context.Context, userID string) ([]*plant.Plant, error) {
	items, err := s.repo.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.withHealthAll(ctx, userID, items), nil
}

func (s *Service) Delete(ctx context.Context, id, userID string) error {
	p, err := s.repo.GetByID(ctx, id, userID)
	if err != nil {
		return err
	}
	if s.images != nil && p.ImageKey != "" {
		_ = s.images.Delete(ctx, p.ImageKey)
	}
	return s.repo.Delete(ctx, id, userID)
}

var imageExt = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

func (s *Service) UploadImage(ctx context.Context, id, userID string, data []byte, contentType string) (*plant.Plant, error) {
	p, err := s.repo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if s.images == nil {
		return nil, fmt.Errorf("%w: image storage is not enabled", apperr.ErrInvalidInput)
	}
	ext, ok := imageExt[contentType]
	if !ok {
		ext = ".jpg"
	}
	key := id + ext
	if err := s.images.Save(ctx, key, data); err != nil {
		return nil, fmt.Errorf("save plant image: %w", err)
	}
	p.ImageKey = key
	p.ImageURL = "/api/v1/plants/" + id + "/image"
	p.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}
	return s.withHealth(ctx, p), nil
}

func (s *Service) Image(ctx context.Context, id, userID string) (io.ReadCloser, string, error) {
	p, err := s.repo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, "", err
	}
	if s.images == nil || p.ImageKey == "" {
		return nil, "", apperr.ErrNotFound
	}
	rc, err := s.images.Open(ctx, p.ImageKey)
	if err != nil {
		return nil, "", err
	}
	ct := "image/jpeg"
	if strings.HasSuffix(p.ImageKey, ".png") {
		ct = "image/png"
	} else if strings.HasSuffix(p.ImageKey, ".webp") {
		ct = "image/webp"
	}
	return rc, ct, nil
}

func (s *Service) Identify(ctx context.Context, imageData []byte) (plant.IdentificationResult, error) {
	if s.identifier == nil {
		return plant.IdentificationResult{}, fmt.Errorf("%w: plant identifier is not configured", apperr.ErrInvalidInput)
	}
	return s.identifier.Identify(ctx, imageData)
}
