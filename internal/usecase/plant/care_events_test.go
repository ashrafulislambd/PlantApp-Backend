package plant

import (
	"context"
	"testing"
	"time"

	"plantpal-backend/internal/domain/plant"
	"plantpal-backend/internal/infrastructure/repository/memory"
)

func TestMarkWateredAwardsPointsOnlyWhenDueAndRecordsEvents(t *testing.T) {
	svc := newTestService()
	events := memory.NewPlantEventRepository()
	svc.SetEventRepository(events)
	ctx := context.Background()

	p, err := svc.Create(ctx, CreateInput{UserID: "user_1", Name: "Fern", WateringFrequencyDays: 4})
	if err != nil {
		t.Fatal(err)
	}

	first, err := svc.MarkWateredWithPoints(ctx, p.ID, "user_1")
	if err != nil {
		t.Fatal(err)
	}
	if first.PointsEarned != 10 {
		t.Fatalf("first watering points = %d, want 10", first.PointsEarned)
	}

	second, err := svc.MarkWateredWithPoints(ctx, p.ID, "user_1")
	if err != nil {
		t.Fatal(err)
	}
	if second.PointsEarned != 0 {
		t.Fatalf("early repeat watering points = %d, want 0", second.PointsEarned)
	}

	logged, err := events.ListByPlant(ctx, p.ID, "user_1")
	if err != nil {
		t.Fatal(err)
	}
	if len(logged) != 2 {
		t.Fatalf("watering event count = %d, want 2", len(logged))
	}
	points := logged[0].Points + logged[1].Points
	if logged[0].EventType != plant.EventWater || logged[1].EventType != plant.EventWater || points != 10 {
		t.Fatalf("watering events = %#v, want two water records totaling 10 points", logged)
	}
}

func TestMarkFertilizedAwardsPointsOnlyWhenDue(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()
	p, err := svc.Create(ctx, CreateInput{UserID: "user_1", Name: "Rose", AgeStage: "mature"})
	if err != nil {
		t.Fatal(err)
	}
	due := time.Now().UTC().Add(-time.Hour)
	p.NextFertilizingAt = &due
	if err := svc.repo.Update(ctx, p); err != nil {
		t.Fatal(err)
	}

	result, err := svc.MarkFertilizedWithPoints(ctx, p.ID, "user_1")
	if err != nil {
		t.Fatal(err)
	}
	if result.PointsEarned != 20 {
		t.Fatalf("fertilizing points = %d, want 20", result.PointsEarned)
	}
}

func TestSkipReschedulesAndRecordsReasonWithoutChangingLastWatered(t *testing.T) {
	svc := newTestService()
	events := memory.NewPlantEventRepository()
	svc.SetEventRepository(events)
	ctx := context.Background()
	p, err := svc.Create(ctx, CreateInput{UserID: "user_1", Name: "Fern", WateringFrequencyDays: 5})
	if err != nil {
		t.Fatal(err)
	}
	lastWatered := time.Now().UTC().Add(-2 * 24 * time.Hour)
	p.LastWateredAt = &lastWatered
	if err := svc.repo.Update(ctx, p); err != nil {
		t.Fatal(err)
	}

	updated, err := svc.Skip(ctx, p.ID, "user_1", "rained", 3)
	if err != nil {
		t.Fatal(err)
	}
	if updated.LastWateredAt == nil || !updated.LastWateredAt.Equal(lastWatered) {
		t.Fatalf("last watered changed to %v, want %v", updated.LastWateredAt, lastWatered)
	}
	if updated.NextWateringAt.Before(time.Now().AddDate(0, 0, 2)) {
		t.Fatalf("next watering = %v, want at least two days from now", updated.NextWateringAt)
	}
	logged, err := events.ListByPlant(ctx, p.ID, "user_1")
	if err != nil {
		t.Fatal(err)
	}
	if len(logged) != 1 || logged[0].EventType != plant.EventSkip || logged[0].Reason != "rained" {
		t.Fatalf("skip event = %#v, want reason rained", logged)
	}

	if _, err := svc.Skip(ctx, p.ID, "user_1", "not-a-reason", 1); err == nil {
		t.Fatal("Skip() accepted an unsupported reason")
	}
}
