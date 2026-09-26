// Package diagnosis implements the application logic behind the AI Doctor
// / Diseases Detection screen: submit a photo, get an issue and cure back.
package diagnosis

import (
	"context"
	"fmt"
	"time"

	"myplantpal-backend/internal/domain/apperr"
	"myplantpal-backend/internal/domain/diagnosis"
	"myplantpal-backend/internal/idgen"
)

type Service struct {
	repo     diagnosis.Repository
	provider diagnosis.Provider
	ids      idgen.Generator
}

func NewService(repo diagnosis.Repository, provider diagnosis.Provider, ids idgen.Generator) *Service {
	return &Service{repo: repo, provider: provider, ids: ids}
}

type AnalyzeInput struct {
	UserID    string
	PlantID   *string
	ImageData []byte
}

func (s *Service) Analyze(ctx context.Context, in AnalyzeInput) (*diagnosis.Diagnosis, error) {
	if len(in.ImageData) == 0 {
		return nil, fmt.Errorf("%w: imageBase64 is required", apperr.ErrInvalidInput)
	}
	if in.UserID == "" {
		return nil, fmt.Errorf("%w: userID is required", apperr.ErrInvalidInput)
	}

	result, err := s.provider.Analyze(ctx, in.ImageData)
	if err != nil {
		return nil, err
	}

	d := &diagnosis.Diagnosis{
		ID:           s.ids.New("diag"),
		UserID:       in.UserID,
		PlantID:      in.PlantID,
		Issue:        result.Issue,
		Cure:         result.Cure,
		Disclaimer:   diagnosis.Disclaimer,
		IssueBn:      result.IssueBn,
		CureBn:       result.CureBn,
		DisclaimerBn: diagnosis.DisclaimerBn,
		CreatedAt:    time.Now().UTC(),
		Provider:     string(result.Provider),
	}
	if err := s.repo.Create(ctx, d); err != nil {
		return nil, err
	}
	return d, nil
}

func (s *Service) Get(ctx context.Context, id, userID string) (*diagnosis.Diagnosis, error) {
	return s.repo.GetByID(ctx, id, userID)
}

func (s *Service) List(ctx context.Context, userID string, plantID *string) ([]*diagnosis.Diagnosis, error) {
	return s.repo.List(ctx, userID, plantID)
}
