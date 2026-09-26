package memory_test

import (
	"errors"
	"testing"

	"myplantpal-backend/internal/domain/apperr"
	"myplantpal-backend/internal/domain/product"
	"myplantpal-backend/internal/infrastructure/repository/memory"
	productseed "myplantpal-backend/internal/infrastructure/repository/memory/seed/products"
)

func newFixtureRepo() *memory.ProductRepository {
	cats := []*product.Category{
		{ID: "a", Name: "A"},
		{ID: "b", Name: "B"},
		{ID: "empty", Name: "Empty"},
	}
	prods := []*product.Product{
		{ID: "a1", CategoryID: "a"},
		{ID: "a2", CategoryID: "a"},
		{ID: "b1", CategoryID: "b"},
	}
	return memory.NewProductRepository(prods, cats)
}

func TestProductRepositoryListFiltersByCategory(t *testing.T) {
	repo := newFixtureRepo()

	tests := []struct {
		category string
		want     int
	}{
		{"", 3},
		{"a", 2},
		{"b", 1},
		{"empty", 0},
		{"unknown", 0},
	}
	for _, tc := range tests {
		got, err := repo.List(tc.category)
		if err != nil {
			t.Fatalf("List(%q): %v", tc.category, err)
		}
		if got == nil {
			t.Errorf("List(%q) returned nil; want a non-nil slice so JSON is []", tc.category)
		}
		if len(got) != tc.want {
			t.Errorf("List(%q) = %d products, want %d", tc.category, len(got), tc.want)
		}
		for _, p := range got {
			if tc.category != "" && p.CategoryID != tc.category {
				t.Errorf("List(%q) returned product %s from category %s", tc.category, p.ID, p.CategoryID)
			}
		}
	}
}

func TestProductRepositoryListReturnsCopy(t *testing.T) {
	repo := newFixtureRepo()
	first, _ := repo.List("")
	first[0] = nil

	second, _ := repo.List("")
	if second[0] == nil {
		t.Fatal("mutating the returned slice changed the repository contents")
	}
}

func TestProductRepositoryGetByID(t *testing.T) {
	repo := newFixtureRepo()

	p, err := repo.GetByID("b1")
	if err != nil || p.ID != "b1" {
		t.Fatalf("GetByID(b1) = %v, %v", p, err)
	}
	if _, err := repo.GetByID("missing"); !errors.Is(err, apperr.ErrNotFound) {
		t.Fatalf("GetByID(missing) error = %v, want ErrNotFound", err)
	}
}

// With the real seed, filtering by every category must partition the full
// catalog: each product appears in exactly one category list.
func TestSeedCategoriesPartitionCatalog(t *testing.T) {
	repo := memory.NewProductRepository(productseed.Products(), productseed.Categories())

	all, err := repo.List("")
	if err != nil {
		t.Fatal(err)
	}
	cats, err := repo.ListCategories()
	if err != nil {
		t.Fatal(err)
	}
	total := 0
	for _, c := range cats {
		items, err := repo.List(c.ID)
		if err != nil {
			t.Fatal(err)
		}
		total += len(items)
	}
	if total != len(all) {
		t.Fatalf("category lists sum to %d products, but the full list has %d", total, len(all))
	}
}
