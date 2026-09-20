// Package fertilizer implements the application logic behind the
// Fertilizer screen: search recipes and add new ones.
package fertilizer

import (
	"context"
	"fmt"
	"strings"
	"time"

	"myplantpal-backend/internal/domain/apperr"
	"myplantpal-backend/internal/domain/fertilizer"
	"myplantpal-backend/internal/idgen"
)

type Service struct {
	repo fertilizer.Repository
	ids  idgen.Generator
}

func NewService(repo fertilizer.Repository, ids idgen.Generator) *Service {
	return &Service{repo: repo, ids: ids}
}

type CreateInput struct {
	Name         string
	Category     string
	Instructions string
}

func (s *Service) Create(ctx context.Context, in CreateInput) (*fertilizer.Fertilizer, error) {
	name := strings.TrimSpace(in.Name)
	instructions := strings.TrimSpace(in.Instructions)
	if name == "" || instructions == "" {
		return nil, fmt.Errorf("%w: name and instructions are required", apperr.ErrInvalidInput)
	}

	f := &fertilizer.Fertilizer{
		ID:           s.ids.New("fert"),
		Name:         name,
		Category:     strings.TrimSpace(in.Category),
		Instructions: instructions,
		CreatedAt:    time.Now().UTC(),
	}
	if err := s.repo.Create(ctx, f); err != nil {
		return nil, err
	}
	return f, nil
}

func (s *Service) Get(ctx context.Context, id string) (*fertilizer.Fertilizer, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context, query string) ([]*fertilizer.Fertilizer, error) {
	return s.repo.List(ctx, strings.TrimSpace(query))
}
