package plant

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"plantpal-backend/internal/domain/apperr"
	"plantpal-backend/internal/domain/plant"
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

// ---- scheduling, update and health (Stage 1) ----

const testDay = 24 * time.Hour

func near(t *testing.T, got, want time.Time, label string) {
	t.Helper()
	d := got.Sub(want)
	if d < 0 {
		d = -d
	}
	if d > time.Minute {
		t.Errorf("%s = %v, want about %v", label, got, want)
	}
}

type fakeScans struct{ m map[string]plant.LastScan }

func (f fakeScans) LatestScans(context.Context, string, []string) (map[string]plant.LastScan, error) {
	return f.m, nil
}

func TestCreate_OmittedLastWateredMeansDueNow(t *testing.T) {
	svc := newTestService()
	p, err := svc.Create(context.Background(), CreateInput{UserID: "user_1", Name: "Fern", WateringFrequencyDays: 5})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	near(t, p.NextWateringAt, time.Now(), "NextWateringAt")
	if p.LastWateredAt != nil {
		t.Errorf("LastWateredAt = %v, want nil", p.LastWateredAt)
	}
}

func TestCreate_LastWateredAnchorsSchedule(t *testing.T) {
	svc := newTestService()
	last := time.Now().UTC().Add(-2 * testDay)
	p, err := svc.Create(context.Background(), CreateInput{UserID: "user_1", Name: "Fern", WateringFrequencyDays: 5, LastWateredAt: &last})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	near(t, p.NextWateringAt, last.AddDate(0, 0, 5), "NextWateringAt")
	if p.LastWateredAt == nil {
		t.Fatal("LastWateredAt not stored")
	}
}

func TestCreate_FutureLastWateredIsClamped(t *testing.T) {
	svc := newTestService()
	future := time.Now().UTC().Add(3 * testDay)
	p, err := svc.Create(context.Background(), CreateInput{UserID: "user_1", Name: "Fern", WateringFrequencyDays: 4, LastWateredAt: &future})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	near(t, *p.LastWateredAt, time.Now(), "LastWateredAt")
	near(t, p.NextWateringAt, time.Now().AddDate(0, 0, 4), "NextWateringAt")
}

func TestMarkWatered_UsesUserIntervalNotRoadmapTimes(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()
	// "mature" has 3 roadmap clock times per day; the interval must win.
	p, err := svc.Create(ctx, CreateInput{UserID: "user_1", Name: "Rose", AgeStage: "mature", WateringFrequencyDays: 5})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	got, err := svc.MarkWatered(ctx, p.ID, "user_1")
	if err != nil {
		t.Fatalf("MarkWatered() error = %v", err)
	}
	near(t, got.NextWateringAt, time.Now().AddDate(0, 0, 5), "NextWateringAt")
	near(t, *got.LastWateredAt, time.Now(), "LastWateredAt")
}

func TestMarkWatered_FallsBackToRoadmapTimesWithoutInterval(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()
	p, err := svc.Create(ctx, CreateInput{UserID: "user_1", Name: "Rose", AgeStage: "mature"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	got, err := svc.MarkWatered(ctx, p.ID, "user_1")
	if err != nil {
		t.Fatalf("MarkWatered() error = %v", err)
	}
	now := time.Now()
	if !got.NextWateringAt.After(now) || got.NextWateringAt.After(now.Add(25*time.Hour)) {
		t.Errorf("NextWateringAt = %v, want the next roadmap clock time within 24h", got.NextWateringAt)
	}
}

func TestUpdate_IntervalReanchorsOnLastWatered(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()
	last := time.Now().UTC().Add(-2 * testDay)
	p, _ := svc.Create(ctx, CreateInput{UserID: "user_1", Name: "Fern", WateringFrequencyDays: 7, LastWateredAt: &last})
	three := 3
	got, err := svc.Update(ctx, p.ID, "user_1", UpdateInput{WateringFrequencyDays: &three})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	near(t, got.NextWateringAt, last.AddDate(0, 0, 3), "NextWateringAt")
	if got.WateringFrequencyDays != 3 {
		t.Errorf("WateringFrequencyDays = %d, want 3", got.WateringFrequencyDays)
	}
}

func TestUpdate_NeverWateredKeepsDueDate(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()
	p, _ := svc.Create(ctx, CreateInput{UserID: "user_1", Name: "Fern", WateringFrequencyDays: 7})
	before := p.NextWateringAt
	three := 3
	got, err := svc.Update(ctx, p.ID, "user_1", UpdateInput{WateringFrequencyDays: &three})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if !got.NextWateringAt.Equal(before) {
		t.Errorf("NextWateringAt = %v, want unchanged %v", got.NextWateringAt, before)
	}
}

func TestUpdate_RegeneratesRoadmapOnlyWhenTypeOrStageChanged(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()
	p, _ := svc.Create(ctx, CreateInput{UserID: "user_1", Name: "Rose", Type: "Water based", AgeStage: "mature"})

	// Customise the stored plant so a regeneration would be visible.
	stored, _ := svc.repo.GetByID(ctx, p.ID, "user_1")
	stored.CareRoadmap.Tips = "my custom tip"
	fert := time.Now().UTC().Add(99 * testDay)
	stored.NextFertilizingAt = &fert

	same, other := "water based", "Rose renamed"
	got, err := svc.Update(ctx, p.ID, "user_1", UpdateInput{Name: &other, Type: &same})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if got.CareRoadmap.Tips != "my custom tip" {
		t.Errorf("Tips = %q, want custom tip kept when type is unchanged", got.CareRoadmap.Tips)
	}
	if got.NextFertilizingAt == nil || !got.NextFertilizingAt.Equal(fert) {
		t.Errorf("NextFertilizingAt = %v, want unchanged %v", got.NextFertilizingAt, fert)
	}

	changed := "Soil based"
	got, err = svc.Update(ctx, p.ID, "user_1", UpdateInput{Type: &changed})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if got.CareRoadmap.Tips == "my custom tip" {
		t.Error("roadmap must be regenerated when the type really changed")
	}
}

func TestUpdate_FertilizingAnchorsOnLastFertilized(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()
	p, _ := svc.Create(ctx, CreateInput{UserID: "user_1", Name: "Rose", AgeStage: "mature"})
	stored, _ := svc.repo.GetByID(ctx, p.ID, "user_1")
	lastFert := time.Now().UTC().Add(-3 * testDay)
	stored.LastFertilizedAt = &lastFert

	adult := "seedling"
	back := "mature"
	if _, err := svc.Update(ctx, p.ID, "user_1", UpdateInput{AgeStage: &adult}); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	got, err := svc.Update(ctx, p.ID, "user_1", UpdateInput{AgeStage: &back})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if got.NextFertilizingAt == nil {
		t.Fatal("NextFertilizingAt = nil, want a date for mature stage")
	}
	near(t, *got.NextFertilizingAt, lastFert.AddDate(0, 0, got.CareRoadmap.FertilizingIntervalDays), "NextFertilizingAt")
}

func TestHealth_FilledOnReadAndRecoversAfterWatering(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()
	last := time.Now().UTC().Add(-10 * testDay)
	p, _ := svc.Create(ctx, CreateInput{UserID: "user_1", Name: "Fern", WateringFrequencyDays: 3, LastWateredAt: &last})

	got, err := svc.Get(ctx, p.ID, "user_1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Health >= 100 || got.HealthState == "" {
		t.Errorf("Health/State = %d/%q, want a reduced score and a state", got.Health, got.HealthState)
	}
	found := false
	for _, r := range got.HealthReasons {
		if r.Code == plant.ReasonWaterLate && r.Value >= 6 {
			found = true
		}
	}
	if !found {
		t.Errorf("HealthReasons = %+v, want water_late with value >= 6", got.HealthReasons)
	}

	watered, err := svc.MarkWatered(ctx, p.ID, "user_1")
	if err != nil {
		t.Fatalf("MarkWatered() error = %v", err)
	}
	for _, r := range watered.HealthReasons {
		if r.Code == plant.ReasonWaterLate {
			t.Errorf("water_late still present after watering: %+v", r)
		}
	}
}

func TestHealth_IncludesLatestScanOnGetAndList(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()
	last := time.Now().UTC()
	p, _ := svc.Create(ctx, CreateInput{UserID: "user_1", Name: "Fern", WateringFrequencyDays: 3, LastWateredAt: &last})
	svc.SetScanSource(fakeScans{m: map[string]plant.LastScan{
		p.ID: {Issue: "Leaf spot", Severity: "Moderate", At: time.Now().UTC().Add(-time.Hour)},
	}})

	got, err := svc.Get(ctx, p.ID, "user_1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Health != 75 {
		t.Errorf("Health = %d, want 75 (moderate scan)", got.Health)
	}
	if got.LastScan == nil || got.LastScan.Issue != "Leaf spot" {
		t.Errorf("LastScan = %+v, want Leaf spot", got.LastScan)
	}

	list, _ := svc.List(ctx, "user_1")
	if len(list) != 1 || list[0].LastScan == nil || list[0].Health != 75 {
		t.Errorf("List() = %+v, want one plant with the scan applied", list)
	}
}

func TestHealth_DoesNotLeakIntoStoredPlant(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()
	p, _ := svc.Create(ctx, CreateInput{UserID: "user_1", Name: "Fern"})
	stored, _ := svc.repo.GetByID(ctx, p.ID, "user_1")
	if stored.Health != 0 || stored.HealthState != "" {
		t.Errorf("stored plant carries computed health: %d/%q", stored.Health, stored.HealthState)
	}
}

func TestCreate_KeepsAIIdentifyValues(t *testing.T) {
	svc := newTestService()
	p, err := svc.Create(context.Background(), CreateInput{
		UserID: "user_1", Name: "Fern", AgeStage: "mature",
		WaterAmountMl: 250, CareTips: "  Mist the leaves weekly.  ",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if p.CareRoadmap.WaterAmountMl != 250 {
		t.Errorf("WaterAmountMl = %d, want 250", p.CareRoadmap.WaterAmountMl)
	}
	if p.CareRoadmap.Tips != "Mist the leaves weekly." {
		t.Errorf("Tips = %q, want the trimmed AI tips", p.CareRoadmap.Tips)
	}
}

func TestCreate_IgnoresAbsurdWaterAmount(t *testing.T) {
	svc := newTestService()
	p, err := svc.Create(context.Background(), CreateInput{
		UserID: "user_1", Name: "Fern", AgeStage: "mature", WaterAmountMl: 999999,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if p.CareRoadmap.WaterAmountMl != 300 {
		t.Errorf("WaterAmountMl = %d, want the generated 300", p.CareRoadmap.WaterAmountMl)
	}
}

func TestCreate_StoresOutdoorFlagAndUpdateTogglesIt(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()
	p, err := svc.Create(ctx, CreateInput{UserID: "user_1", Name: "Tomato", Outdoor: true})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if !p.Outdoor {
		t.Fatalf("Outdoor = false after create, want true")
	}
	indoor, err := svc.Create(ctx, CreateInput{UserID: "user_1", Name: "Fern"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if indoor.Outdoor {
		t.Errorf("Outdoor = true by default, want indoor")
	}

	// Omitted keeps the value; an explicit false changes it.
	if got, _ := svc.Update(ctx, p.ID, "user_1", UpdateInput{Name: strPtr("Tomato 2")}); !got.Outdoor {
		t.Errorf("Outdoor lost on an update that did not mention it")
	}
	off := false
	got, err := svc.Update(ctx, p.ID, "user_1", UpdateInput{Outdoor: &off})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if got.Outdoor {
		t.Errorf("Outdoor = true after Update(Outdoor=false)")
	}
}

func strPtr(s string) *string { return &s }

func TestMarkWatered_RoadmapTimesUseUsersTimeZone(t *testing.T) {
	svc := newTestService()
	// A plant with no interval and a single 08:00 roadmap time (seedling).
	ctxUTC := context.Background()
	p, err := svc.Create(ctxUTC, CreateInput{UserID: "user_1", Name: "Sprout", AgeStage: "seed"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	dhaka := time.FixedZone("UTC+06:00", 6*3600)
	ctx := plant.ContextWithLocation(ctxUTC, dhaka)
	got, err := svc.MarkWatered(ctx, p.ID, "user_1")
	if err != nil {
		t.Fatalf("MarkWatered() error = %v", err)
	}
	local := got.NextWateringAt.In(dhaka)
	if local.Hour() != 8 || local.Minute() != 0 {
		t.Errorf("NextWateringAt = %v (%v local), want 08:00 in the user's zone", got.NextWateringAt, local)
	}
	if got.NextWateringAt.Location() != time.UTC {
		t.Errorf("NextWateringAt must be stored in UTC, got %v", got.NextWateringAt.Location())
	}
	if !got.NextWateringAt.After(time.Now()) || got.NextWateringAt.After(time.Now().Add(25*time.Hour)) {
		t.Errorf("NextWateringAt = %v, want within the next 24h", got.NextWateringAt)
	}
}
