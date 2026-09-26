// Package product implements the application logic for the product catalog
// and AI-powered price refresh.
package product

import (
	"fmt"
	"strings"

	"myplantpal-backend/internal/domain/apperr"
	"myplantpal-backend/internal/domain/product"
)

// PriceRefresherPort is an alias for the domain port, so main.go can
// declare a nil-able variable of this type without importing the domain.
type PriceRefresherPort = product.PriceRefresher

// Service is the product catalog usecase.
type Service struct {
	repo      product.Repository
	refresher product.PriceRefresher
}

// NewService creates a Service. refresher may be nil to disable AI refresh.
func NewService(repo product.Repository, refresher product.PriceRefresher) *Service {
	return &Service{repo: repo, refresher: refresher}
}

// List returns the catalog, optionally filtered by category ID. The ID is
// trimmed and lower-cased (category IDs are lower-case kebab-case). An empty
// ID or "all" returns every product. An unknown ID yields an empty list, not
// an error, so clients can render an empty state.
func (s *Service) List(categoryID string) ([]*product.Product, error) {
	categoryID = strings.ToLower(strings.TrimSpace(categoryID))
	if categoryID == "all" {
		categoryID = ""
	}
	return s.repo.List(categoryID)
}

func (s *Service) Get(id string) (*product.Product, error) {
	return s.repo.GetByID(id)
}

func (s *Service) ListCategories() ([]*product.Category, error) {
	return s.repo.ListCategories()
}

// RefreshPrice asks the AI refresher for a live price check.
// Results are cached 6 h server-side inside the refresher itself.
func (s *Service) RefreshPrice(id string) (*product.RefreshResult, error) {
	if s.refresher == nil {
		return nil, fmt.Errorf("%w: price refresh not configured (GROQ_API_KEY not set)", apperr.ErrInvalidInput)
	}
	p, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return s.refresher.RefreshPrice(p)
}
