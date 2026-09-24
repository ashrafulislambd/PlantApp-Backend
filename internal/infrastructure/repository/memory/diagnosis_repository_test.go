package memory

import (
	"context"
	"testing"

	"myplantpal-backend/internal/domain/diagnosis"
)

func strPtr(s string) *string { return &s }

func TestDiagnosisRepository_List_FiltersByPlantID(t *testing.T) {
	repo := NewDiagnosisRepository()
	ctx := context.Background()

	forPlantA := &diagnosis.Diagnosis{ID: "diag_a", PlantID: strPtr("pl_a"), Issue: "A"}
	forPlantB := &diagnosis.Diagnosis{ID: "diag_b", PlantID: strPtr("pl_b"), Issue: "B"}
	noPlant := &diagnosis.Diagnosis{ID: "diag_c", Issue: "C"}

	for _, d := range []*diagnosis.Diagnosis{forPlantA, forPlantB, noPlant} {
		if err := repo.Create(ctx, d); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	all, err := repo.List(ctx, nil)
	if err != nil {
		t.Fatalf("List(nil) error = %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("List(nil) returned %d items, want 3", len(all))
	}

	filtered, err := repo.List(ctx, strPtr("pl_a"))
	if err != nil {
		t.Fatalf("List(pl_a) error = %v", err)
	}
	if len(filtered) != 1 || filtered[0].ID != "diag_a" {
		t.Errorf("List(pl_a) = %+v, want just diag_a", filtered)
	}

	none, err := repo.List(ctx, strPtr("pl_unknown"))
	if err != nil {
		t.Fatalf("List(pl_unknown) error = %v", err)
	}
	if len(none) != 0 {
		t.Errorf("List(pl_unknown) = %+v, want no matches", none)
	}
}

func TestDiagnosisRepository_GetByID_NotFound(t *testing.T) {
	repo := NewDiagnosisRepository()
	if _, err := repo.GetByID(context.Background(), "missing"); err == nil {
		t.Error("GetByID() error = nil, want not-found error")
	}
}
