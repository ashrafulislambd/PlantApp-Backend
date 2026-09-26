// Package seed provides the initial product catalog data.
//
// The catalog is split by category across catalog_*.go files in this package;
// each file exposes one function returning that category's products, and
// Products() concatenates them. To add a batch of products, extend (or add) the
// matching catalog_<category>.go file and register its function in Products().
package seed

import (
	"math"
	"time"

	"myplantpal-backend/internal/domain/product"
)

// bdtPerUSD is the rate used to derive USD prices for products that are listed
// in BDT by Bangladeshi vendors (BDT is the source of truth for those). Rate
// checked 2026-09-24.
const bdtPerUSD = 123.0

// usdFromBDT converts a BDT selling price to USD, rounded to cents.
func usdFromBDT(taka float64) float64 { return math.Round(taka/bdtPerUSD*100) / 100 }

func verified(date string) time.Time {
	t, _ := time.Parse("2006-01-02", date)
	return t
}

// Products returns the seed product catalog.
func Products() []*product.Product {
	var all []*product.Product
	all = append(all, plantProducts()...)
	all = append(all, outdoorProducts()...)
	all = append(all, floweringProducts()...)
	all = append(all, succulentProducts()...)
	all = append(all, herbProducts()...)
	all = append(all, seedProducts()...)
	all = append(all, fertilizerProducts()...)
	all = append(all, potProducts()...)
	all = append(all, toolProducts()...)
	all = append(all, wateringProducts()...)
	all = append(all, lightProducts()...)
	return all
}

// Categories returns the catalog categories. IDs are stable, lower-case
// kebab-case and are the values used by ?categoryId= and Product.CategoryID.
func Categories() []*product.Category {
	return []*product.Category{
		{ID: "plants", Name: "Houseplants"},
		{ID: "outdoor-plants", Name: "Outdoor Plants"},
		{ID: "flowering-plants", Name: "Flowering Plants"},
		{ID: "succulents-cacti", Name: "Succulents and Cacti"},
		{ID: "herbs-vegetables", Name: "Herbs and Vegetables"},
		{ID: "seeds", Name: "Seeds"},
		{ID: "pots-planters", Name: "Pots and Planters"},
		{ID: "soil-potting-mix", Name: "Soil and Potting Mix"},
		{ID: "fertilizer", Name: "Fertilizers"},
		{ID: "care", Name: "Plant Care and Pest Control"},
		{ID: "gardening-tools", Name: "Gardening Tools"},
		{ID: "watering-equipment", Name: "Watering Equipment"},
		{ID: "grow-lights", Name: "Grow Lights and Accessories"},
	}
}
