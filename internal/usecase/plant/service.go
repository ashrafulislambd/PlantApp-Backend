// Package plant implements the application logic for registering plants,
// generating their care roadmap, and tracking watering/fertilizing.
package plant

import (
	"context"
	"fmt"
	"strings"
	"time"

	"plantpal-backend/internal/domain/apperr"
	"plantpal-backend/internal/domain/plant"
	"plantpal-backend/internal/idgen"
)

type Service struct {
	repo plant.Repository
	ids  idgen.Generator
}

func NewService(repo plant.Repository, ids idgen.Generator) *Service {
	return &Service{repo: repo, ids: ids}
}

type CreateInput struct {
	UserID   string
	Name     string
	Type     string
	AgeStage string
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

	now := time.Now().UTC()
	roadmap := buildRoadmap(in.Type, in.AgeStage, in.Lang)
	p := &plant.Plant{
		ID:             s.ids.New("pl"),
		UserID:         in.UserID,
		Name:           name,
		Type:           strings.TrimSpace(in.Type),
		AgeStage:       strings.TrimSpace(in.AgeStage),
		CareRoadmap:    roadmap,
		NextWateringAt: nextWateringTime(roadmap.WateringTimes, now),
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	setNextFertilizing(p, now)
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) Get(ctx context.Context, id, userID string) (*plant.Plant, error) {
	return s.repo.GetByID(ctx, id, userID)
}

func (s *Service) List(ctx context.Context, userID string) ([]*plant.Plant, error) {
	return s.repo.List(ctx, userID)
}

func (s *Service) Delete(ctx context.Context, id, userID string) error {
	return s.repo.Delete(ctx, id, userID)
}
