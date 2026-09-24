// Package plant implements the application logic for registering plants
// and generating their care roadmap.
package plant

import (
	"context"
	"fmt"
	"strings"
	"time"

	"myplantpal-backend/internal/domain/apperr"
	"myplantpal-backend/internal/domain/plant"
	"myplantpal-backend/internal/idgen"
)

type Service struct {
	repo plant.Repository
	ids  idgen.Generator
}

func NewService(repo plant.Repository, ids idgen.Generator) *Service {
	return &Service{repo: repo, ids: ids}
}

type CreateInput struct {
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

	now := time.Now().UTC()
	p := &plant.Plant{
		ID:          s.ids.New("pl"),
		Name:        name,
		Type:        strings.TrimSpace(in.Type),
		AgeStage:    strings.TrimSpace(in.AgeStage),
		CareRoadmap: buildRoadmap(in.Type, in.AgeStage, in.Lang),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) Get(ctx context.Context, id string) (*plant.Plant, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]*plant.Plant, error) {
	return s.repo.List(ctx)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// buildRoadmap derives a watering schedule, tips, and a fertilizer
// suggestion from the plant's type and age stage. This is a simple
// rule-based placeholder for the "Create My Roadmap" action — swap it for
// an AI-driven recommendation later behind the same signature.
//
// The generated tips/fertilizer text is picked in the requester's
// language at creation time; it isn't re-translated later if the plant is
// fetched again under a different language.
func buildRoadmap(plantType, ageStage, lang string) plant.CareRoadmap {
	bn := lang == "bn"

	waterAmountMl := 200
	wateringTimes := []string{"08:00", "18:00"}
	tips := "Don't expose to excess sun; find a cool, dry place with plenty of indirect sunlight."
	fertilizer := "Use a balanced NPK fertilizer every 2 weeks."
	if bn {
		tips = "অতিরিক্ত রোদ এড়িয়ে চলুন; পর্যাপ্ত পরোক্ষ আলোযুক্ত একটি ঠান্ডা, শুকনো জায়গা বেছে নিন।"
		fertilizer = "প্রতি ২ সপ্তাহে একটি সুষম NPK সার ব্যবহার করুন।"
	}

	switch strings.ToLower(strings.TrimSpace(ageStage)) {
	case "seed", "seedling":
		waterAmountMl = 100
		wateringTimes = []string{"08:00"}
		fertilizer = "Avoid fertilizer until the first true leaves appear."
		if bn {
			fertilizer = "প্রথম প্রকৃত পাতা না গজানো পর্যন্ত সার ব্যবহার এড়িয়ে চলুন।"
		}
	case "mature", "adult":
		waterAmountMl = 300
		wateringTimes = []string{"08:00", "12:00", "18:00"}
		fertilizer = "Use Potassium (K) based fertilizer to support blooming."
		if bn {
			fertilizer = "ফুল ফোটাতে সাহায্য করতে পটাশিয়াম (K) ভিত্তিক সার ব্যবহার করুন।"
		}
	}

	if strings.Contains(strings.ToLower(plantType), "water") {
		waterAmountMl += 100
	}

	return plant.CareRoadmap{
		WateringTimes:            wateringTimes,
		WaterAmountMl:            waterAmountMl,
		Tips:                     tips,
		FertilizerRecommendation: fertilizer,
	}
}
