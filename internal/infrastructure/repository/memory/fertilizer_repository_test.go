package memory

import (
	"context"
	"errors"
	"testing"

	"myplantpal-backend/internal/domain/apperr"
	"myplantpal-backend/internal/domain/fertilizer"
)

func seedFertilizers() []*fertilizer.Fertilizer {
	return []*fertilizer.Fertilizer{
		{ID: "fert_n", Name: "Nitrogen (leaf growth)", Category: "Nitrogen", Instructions: "..."},
		{ID: "fert_p", Name: "Phosphorus (root growth)", Category: "Phosphorus", Instructions: "..."},
	}
}

func TestFertilizerRepository_SeedAndGet(t *testing.T) {
	repo := NewFertilizerRepository(seedFertilizers()...)
	got, err := repo.GetByID(context.Background(), "fert_n")
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.Name != "Nitrogen (leaf growth)" {
		t.Errorf("GetByID().Name = %q, want seeded name", got.Name)
	}
}

func TestFertilizerRepository_GetByID_NotFound(t *testing.T) {
	repo := NewFertilizerRepository()
	_, err := repo.GetByID(context.Background(), "missing")
	if !errors.Is(err, apperr.ErrNotFound) {
		t.Errorf("GetByID() error = %v, want %v", err, apperr.ErrNotFound)
	}
}

func TestFertilizerRepository_List_EmptyQueryReturnsAll(t *testing.T) {
	repo := NewFertilizerRepository(seedFertilizers()...)
	list, err := repo.List(context.Background(), "")
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("List(\"\") returned %d items, want 2", len(list))
	}
}

func TestFertilizerRepository_List_FiltersCaseInsensitiveByNameOrCategory(t *testing.T) {
	repo := NewFertilizerRepository(seedFertilizers()...)

	byName, err := repo.List(context.Background(), "NITROGEN")
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(byName) != 1 || byName[0].ID != "fert_n" {
		t.Errorf("List(\"NITROGEN\") = %+v, want just fert_n", byName)
	}

	byCategory, err := repo.List(context.Background(), "phosphorus")
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(byCategory) != 1 || byCategory[0].ID != "fert_p" {
		t.Errorf("List(\"phosphorus\") = %+v, want just fert_p", byCategory)
	}

	none, err := repo.List(context.Background(), "potassium")
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(none) != 0 {
		t.Errorf("List(\"potassium\") = %+v, want no matches", none)
	}
}

func TestFertilizerRepository_List_SortedByName(t *testing.T) {
	repo := NewFertilizerRepository(seedFertilizers()...)
	list, err := repo.List(context.Background(), "")
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if list[0].Name != "Nitrogen (leaf growth)" || list[1].Name != "Phosphorus (root growth)" {
		t.Errorf("List() not sorted by name: got [%s, %s]", list[0].Name, list[1].Name)
	}
}

func TestFertilizerRepository_Create(t *testing.T) {
	repo := NewFertilizerRepository()
	f := &fertilizer.Fertilizer{ID: "fert_new", Name: "Custom", Instructions: "..."}
	if err := repo.Create(context.Background(), f); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	got, err := repo.GetByID(context.Background(), "fert_new")
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.Name != "Custom" {
		t.Errorf("GetByID().Name = %q, want %q", got.Name, "Custom")
	}
}
