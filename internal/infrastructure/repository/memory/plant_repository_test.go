package memory

import (
	"context"
	"errors"
	"testing"
	"time"

	"myplantpal-backend/internal/domain/apperr"
	"myplantpal-backend/internal/domain/plant"
)

func TestPlantRepository_CreateAndGet(t *testing.T) {
	repo := NewPlantRepository()
	ctx := context.Background()
	p := &plant.Plant{ID: "pl_1", Name: "Rose"}

	if err := repo.Create(ctx, p); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := repo.GetByID(ctx, "pl_1")
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.Name != "Rose" {
		t.Errorf("GetByID().Name = %q, want %q", got.Name, "Rose")
	}
}

func TestPlantRepository_GetByID_NotFound(t *testing.T) {
	repo := NewPlantRepository()
	_, err := repo.GetByID(context.Background(), "missing")
	if !errors.Is(err, apperr.ErrNotFound) {
		t.Errorf("GetByID() error = %v, want %v", err, apperr.ErrNotFound)
	}
}

func TestPlantRepository_List_SortedByCreatedAt(t *testing.T) {
	repo := NewPlantRepository()
	ctx := context.Background()
	now := time.Now().UTC()

	newer := &plant.Plant{ID: "pl_newer", Name: "Newer", CreatedAt: now}
	older := &plant.Plant{ID: "pl_older", Name: "Older", CreatedAt: now.Add(-time.Hour)}

	if err := repo.Create(ctx, newer); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if err := repo.Create(ctx, older); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	list, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("List() returned %d items, want 2", len(list))
	}
	if list[0].ID != "pl_older" || list[1].ID != "pl_newer" {
		t.Errorf("List() not sorted oldest-first: got [%s, %s]", list[0].ID, list[1].ID)
	}
}

func TestPlantRepository_Delete(t *testing.T) {
	repo := NewPlantRepository()
	ctx := context.Background()
	p := &plant.Plant{ID: "pl_1", Name: "Rose"}
	_ = repo.Create(ctx, p)

	if err := repo.Delete(ctx, "pl_1"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := repo.GetByID(ctx, "pl_1"); !errors.Is(err, apperr.ErrNotFound) {
		t.Errorf("expected deleted plant to be not found, got err = %v", err)
	}
}

func TestPlantRepository_Delete_NotFound(t *testing.T) {
	repo := NewPlantRepository()
	err := repo.Delete(context.Background(), "missing")
	if !errors.Is(err, apperr.ErrNotFound) {
		t.Errorf("Delete() error = %v, want %v", err, apperr.ErrNotFound)
	}
}
