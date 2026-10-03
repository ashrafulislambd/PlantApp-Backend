package plant

import (
	"context"
	"errors"
	"strings"
	"testing"

	"plantpal-backend/internal/domain/apperr"
	"plantpal-backend/internal/idgen"
	"plantpal-backend/internal/infrastructure/repository/memory"
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
	if p.CareRoadmap.FertilizingIntervalDays != 0 {
		t.Errorf("FertilizingIntervalDays = %d, want 0 for seed stage (no fertilizing yet)", p.CareRoadmap.FertilizingIntervalDays)
	}
}

func TestCreate_RoadmapByMatureAgeStage(t *testing.T) {
	svc := newTestService()
	p, err := svc.Create(context.Background(), CreateInput{UserID: "user_1", Name: "Rose", Type: "Water based", AgeStage: "mature"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if p.CareRoadmap.FertilizingIntervalDays != 7 {
		t.Errorf("FertilizingIntervalDays = %d, want 7 for mature stage", p.CareRoadmap.FertilizingIntervalDays)
	}
	if !strings.Contains(p.FertilizerNote, "Potassium") {
		t.Errorf("FertilizerNote = %q, want it to mention Potassium for mature stage", p.FertilizerNote)
	}
}

func TestCreate_RoadmapDefaultAgeStage(t *testing.T) {
	svc := newTestService()
	p, err := svc.Create(context.Background(), CreateInput{UserID: "user_1", Name: "Fern"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if p.CareRoadmap.FertilizingIntervalDays != 14 {
		t.Errorf("FertilizingIntervalDays = %d, want 14 default", p.CareRoadmap.FertilizingIntervalDays)
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
	if strings.Contains(p.FertilizerNote, "Potassium") {
		t.Errorf("FertilizerNote = %q, expected Bengali text, got English", p.FertilizerNote)
	}
}

func TestCreate_DefaultsWateringFrequencyAndComputesNextWatering(t *testing.T) {
	svc := newTestService()
	p, err := svc.Create(context.Background(), CreateInput{UserID: "user_1", Name: "Fern"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if p.WateringFrequencyDays != 7 {
		t.Errorf("WateringFrequencyDays = %d, want 7 default", p.WateringFrequencyDays)
	}
	// Never watered -> base is "now", so next watering is 7 days out.
	if p.WaterLevel != "In 7 days" {
		t.Errorf("WaterLevel = %q, want %q for a freshly-added, never-watered plant", p.WaterLevel, "In 7 days")
	}
}

func TestMarkWatered_RollsNextWateringForward(t *testing.T) {
	svc := newTestService()
	p, err := svc.Create(context.Background(), CreateInput{UserID: "user_1", Name: "Fern", WateringFrequencyDays: 3})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	updated, err := svc.MarkWatered(context.Background(), p.ID, "user_1")
	if err != nil {
		t.Fatalf("MarkWatered() error = %v", err)
	}
	if updated.LastWateredAt == nil {
		t.Fatal("LastWateredAt was not set")
	}
	if updated.WaterLevel != "In 3 days" {
		t.Errorf("WaterLevel = %q, want %q", updated.WaterLevel, "In 3 days")
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
