package plant

import (
	"context"
	"errors"
	"strings"
	"testing"

	"myplantpal-backend/internal/domain/apperr"
	"myplantpal-backend/internal/idgen"
	"myplantpal-backend/internal/infrastructure/repository/memory"
)

func newTestService() *Service {
	return NewService(memory.NewPlantRepository(), idgen.New())
}

func TestCreate_RequiresName(t *testing.T) {
	svc := newTestService()
	_, err := svc.Create(context.Background(), CreateInput{UserID: "user_1", Name: "  "})
	if !errors.Is(err, apperr.ErrInvalidInput) {
		t.Errorf("Create() error = %v, want %v", err, apperr.ErrInvalidInput)
	}
}

func TestCreate_TrimsNameAndPersists(t *testing.T) {
	svc := newTestService()
	p, err := svc.Create(context.Background(), CreateInput{UserID: "user_1", Name: "  Rose  ", Type: "Water based", AgeStage: "mature"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if p.Name != "Rose" {
		t.Errorf("Name = %q, want %q", p.Name, "Rose")
	}

	fetched, err := svc.Get(context.Background(), p.ID, "user_1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if fetched.ID != p.ID {
		t.Errorf("Get() returned a different plant than was created")
	}
}

func TestCreate_RoadmapBySeedAgeStage(t *testing.T) {
	svc := newTestService()
	p, err := svc.Create(context.Background(), CreateInput{UserID: "user_1", Name: "Sprout", AgeStage: "seed"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if p.CareRoadmap.WaterAmountMl != 100 {
		t.Errorf("WaterAmountMl = %d, want 100 for seed stage", p.CareRoadmap.WaterAmountMl)
	}
	if len(p.CareRoadmap.WateringTimes) != 1 {
		t.Errorf("WateringTimes = %v, want exactly 1 entry for seed stage", p.CareRoadmap.WateringTimes)
	}
}

func TestCreate_RoadmapByMatureAgeStageAndWaterType(t *testing.T) {
	svc := newTestService()
	p, err := svc.Create(context.Background(), CreateInput{UserID: "user_1", Name: "Rose", Type: "Water based", AgeStage: "mature"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if p.CareRoadmap.WaterAmountMl != 400 {
		t.Errorf("WaterAmountMl = %d, want 400 for mature + water-based", p.CareRoadmap.WaterAmountMl)
	}
	if len(p.CareRoadmap.WateringTimes) != 3 {
		t.Errorf("WateringTimes = %v, want 3 entries for mature stage", p.CareRoadmap.WateringTimes)
	}
	if !strings.Contains(p.CareRoadmap.FertilizerRecommendation, "Potassium") {
		t.Errorf("FertilizerRecommendation = %q, want it to mention Potassium for mature stage", p.CareRoadmap.FertilizerRecommendation)
	}
}

func TestCreate_RoadmapDefaultAgeStage(t *testing.T) {
	svc := newTestService()
	p, err := svc.Create(context.Background(), CreateInput{UserID: "user_1", Name: "Fern"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if p.CareRoadmap.WaterAmountMl != 200 {
		t.Errorf("WaterAmountMl = %d, want 200 default", p.CareRoadmap.WaterAmountMl)
	}
}

func TestCreate_RoadmapLocalizedToBengali(t *testing.T) {
	svc := newTestService()
	p, err := svc.Create(context.Background(), CreateInput{UserID: "user_1", Name: "Rose", AgeStage: "mature", Lang: "bn"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if strings.Contains(p.CareRoadmap.Tips, "sun") {
		t.Errorf("Tips = %q, expected Bengali text, got English", p.CareRoadmap.Tips)
	}
	if strings.Contains(p.CareRoadmap.FertilizerRecommendation, "Potassium") {
		t.Errorf("FertilizerRecommendation = %q, expected Bengali text, got English", p.CareRoadmap.FertilizerRecommendation)
	}
}

func TestDelete_NotFound(t *testing.T) {
	svc := newTestService()
	err := svc.Delete(context.Background(), "missing", "user_1")
	if !errors.Is(err, apperr.ErrNotFound) {
		t.Errorf("Delete() error = %v, want %v", err, apperr.ErrNotFound)
	}
}

func TestList_ReturnsCreatedPlants(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()
	userID := "user_1"
	if _, err := svc.Create(ctx, CreateInput{UserID: userID, Name: "Rose"}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if _, err := svc.Create(ctx, CreateInput{UserID: userID, Name: "Fern"}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	list, err := svc.List(ctx, userID)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(list) != 2 {
		t.Errorf("List() returned %d plants, want 2", len(list))
	}
}
