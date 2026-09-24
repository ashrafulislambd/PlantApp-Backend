package fertilizer

import (
	"context"
	"errors"
	"testing"

	"myplantpal-backend/internal/domain/apperr"
	"myplantpal-backend/internal/idgen"
	"myplantpal-backend/internal/infrastructure/repository/memory"
)

func newTestService() *Service {
	return NewService(memory.NewFertilizerRepository(), idgen.New())
}

func TestCreate_RequiresNameAndInstructions(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	cases := []CreateInput{
		{Name: "", Instructions: "Do a thing"},
		{Name: "Custom", Instructions: "  "},
	}
	for _, in := range cases {
		if _, err := svc.Create(ctx, in); !errors.Is(err, apperr.ErrInvalidInput) {
			t.Errorf("Create(%+v) error = %v, want %v", in, err, apperr.ErrInvalidInput)
		}
	}
}

func TestCreate_TrimsAndPersists(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	f, err := svc.Create(ctx, CreateInput{Name: " Custom Mix ", Category: " Custom ", Instructions: " Mix it. "})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if f.Name != "Custom Mix" || f.Category != "Custom" || f.Instructions != "Mix it." {
		t.Errorf("Create() = %+v, want trimmed fields", f)
	}

	got, err := svc.Get(ctx, f.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Name != f.Name {
		t.Errorf("Get() returned a different fertilizer than was created")
	}
}

func TestList_SearchQueryTrimmed(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()
	if _, err := svc.Create(ctx, CreateInput{Name: "Nitrogen Mix", Category: "Nitrogen", Instructions: "..."}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	list, err := svc.List(ctx, "  nitrogen  ")
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(list) != 1 {
		t.Errorf("List(\"  nitrogen  \") returned %d items, want 1", len(list))
	}
}

func TestGet_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.Get(context.Background(), "missing")
	if !errors.Is(err, apperr.ErrNotFound) {
		t.Errorf("Get() error = %v, want %v", err, apperr.ErrNotFound)
	}
}
