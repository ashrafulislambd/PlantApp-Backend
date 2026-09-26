package memory

import (
	"fmt"
	"myplantpal-backend/internal/domain/apperr"
	"myplantpal-backend/internal/domain/product"
)

// ProductRepository is an in-memory implementation of product.Repository.
type ProductRepository struct {
	products   []*product.Product
	categories []*product.Category
	byID       map[string]*product.Product
}

// NewProductRepository builds the repository seeded with the given products and categories.
func NewProductRepository(products []*product.Product, categories []*product.Category) *ProductRepository {
	idx := make(map[string]*product.Product, len(products))
	for _, p := range products {
		idx[p.ID] = p
	}
	return &ProductRepository{products: products, categories: categories, byID: idx}
}

// List returns products in seed order. An empty categoryID returns every
// product; otherwise only products whose CategoryID matches exactly. The
// result is always a fresh, non-nil slice so it marshals as [] (never null)
// and callers cannot mutate the repository's internal ordering.
func (r *ProductRepository) List(categoryID string) ([]*product.Product, error) {
	out := make([]*product.Product, 0, len(r.products))
	for _, p := range r.products {
		if categoryID == "" || p.CategoryID == categoryID {
			out = append(out, p)
		}
	}
	return out, nil
}

func (r *ProductRepository) GetByID(id string) (*product.Product, error) {
	p, ok := r.byID[id]
	if !ok {
		return nil, fmt.Errorf("%w: product %q not found", apperr.ErrNotFound, id)
	}
	return p, nil
}

func (r *ProductRepository) ListCategories() ([]*product.Category, error) {
	return r.categories, nil
}

// UpdateRefresh stores the refreshed price data back on the product.
func (r *ProductRepository) UpdateRefresh(result *product.RefreshResult) {
	p, ok := r.byID[result.ProductID]
	if !ok {
		return
	}
	p.PriceBDT = result.PriceBDT
	p.PriceUSD = result.PriceUSD
	p.RefreshedAt = &result.AsOf
	p.RefreshedSource = result.Source
}
