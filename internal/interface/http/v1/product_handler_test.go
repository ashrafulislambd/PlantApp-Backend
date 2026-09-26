package v1

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"myplantpal-backend/internal/domain/product"
	"myplantpal-backend/internal/infrastructure/repository/memory"
	productuc "myplantpal-backend/internal/usecase/product"
)

type listEnvelope struct {
	Data struct {
		Categories []product.Category `json:"categories"`
		Products   []product.Product  `json:"products"`
	} `json:"data"`
}

func newProductMux() *http.ServeMux {
	repo := memory.NewProductRepository(
		[]*product.Product{
			{ID: "p1", Name: "P1", CategoryID: "seeds"},
			{ID: "p2", Name: "P2", CategoryID: "seeds"},
			{ID: "p3", Name: "P3", CategoryID: "tools"},
		},
		[]*product.Category{
			{ID: "seeds", Name: "Seeds"},
			{ID: "tools", Name: "Tools"},
			{ID: "empty", Name: "Empty"},
		},
	)
	h := NewProductHandler(productuc.NewService(repo, nil))
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/products", h.List)
	mux.HandleFunc("GET /api/v1/products/{id}", h.Get)
	return mux
}

func doGet(mux *http.ServeMux, target string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, nil))
	return rec
}

func TestProductListFiltering(t *testing.T) {
	mux := newProductMux()

	tests := []struct {
		name string
		url  string
		want int
	}{
		{"all products", "/api/v1/products", 3},
		{"categoryId filter", "/api/v1/products?categoryId=seeds", 2},
		{"legacy category param", "/api/v1/products?category=seeds", 2},
		{"categoryId wins over category", "/api/v1/products?categoryId=tools&category=seeds", 1},
		{"case-insensitive", "/api/v1/products?categoryId=SEEDS", 2},
		{"all sentinel", "/api/v1/products?categoryId=all", 3},
		{"category with no products", "/api/v1/products?categoryId=empty", 0},
		{"unknown category", "/api/v1/products?categoryId=nope", 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := doGet(mux, tc.url)
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
			}
			var env listEnvelope
			if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if len(env.Data.Products) != tc.want {
				t.Errorf("got %d products, want %d", len(env.Data.Products), tc.want)
			}
			if len(env.Data.Categories) != 3 {
				t.Errorf("got %d categories, want all 3 on every response", len(env.Data.Categories))
			}
			if tc.want == 0 && !strings.Contains(rec.Body.String(), `"products":[]`) {
				t.Errorf("empty result must serialise as [], body=%s", rec.Body.String())
			}
		})
	}
}

func TestProductGet(t *testing.T) {
	mux := newProductMux()

	rec := doGet(mux, "/api/v1/products/p1")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var env struct {
		Data product.Product `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if env.Data.ID != "p1" {
		t.Errorf("got id %q, want p1", env.Data.ID)
	}

	if rec := doGet(mux, "/api/v1/products/missing"); rec.Code != http.StatusNotFound {
		t.Errorf("missing product status = %d, want 404", rec.Code)
	}
}

func TestProductListBengaliCategories(t *testing.T) {
	mux := newProductMux()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/products", nil)
	req.Header.Set("Accept-Language", "bn")
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var env listEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("decode: %v", err)
	}
	found := false
	for _, cat := range env.Data.Categories {
		if cat.ID == "seeds" && cat.Name == "বীজ" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected seeds category to be localized to 'বীজ', got categories: %+v", env.Data.Categories)
	}
}
