// Package product holds the Product entity and its repository contract.
package product

import "time"

// Product is a single item in the PlantPal shop catalog.
type Product struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	ScientificName  string     `json:"scientificName,omitempty"`
	CategoryID      string     `json:"categoryId"`
	Description     string     `json:"description"`
	ImageURL        string     `json:"imageUrl"`
	PriceBDT        float64    `json:"priceBdt"`
	PriceUSD        float64    `json:"priceUsd"`
	Vendor          string     `json:"vendor"`
	BuyURL          string     `json:"buyUrl"`
	LastVerified    time.Time  `json:"lastVerified"`
	RefreshedAt     *time.Time `json:"refreshedAt,omitempty"`
	RefreshedSource string     `json:"refreshedSource,omitempty"`
}

// IsStale reports whether the price is older than 30 days.
func (p *Product) IsStale() bool {
	return time.Since(p.LastVerified) > 30*24*time.Hour
}

// Category groups products in the catalog.
type Category struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// RefreshResult is the AI-sourced price check returned to the caller.
type RefreshResult struct {
	ProductID string    `json:"productId"`
	PriceBDT  float64   `json:"priceBdt"`
	PriceUSD  float64   `json:"priceUsd"`
	InStock   bool      `json:"inStock"`
	AsOf      time.Time `json:"asOf"`
	Source    string    `json:"source"`
	Note      string    `json:"note,omitempty"`
}

// Repository is the port the usecase calls.
type Repository interface {
	List(categoryID string) ([]*Product, error)
	GetByID(id string) (*Product, error)
	ListCategories() ([]*Category, error)
}

// PriceRefresher is the port for AI-driven live price checks.
type PriceRefresher interface {
	RefreshPrice(p *Product) (*RefreshResult, error)
}
