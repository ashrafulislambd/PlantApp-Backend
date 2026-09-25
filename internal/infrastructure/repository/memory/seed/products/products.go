// Package seed provides the initial product catalog data.
package seed

import (
	"time"
	"myplantpal-backend/internal/domain/product"
)

// USD to BDT conversion rate (approximate, update periodically).
const usdToBDT = 110.0

func bdt(usd float64) float64 { return usd * usdToBDT }

func verified(date string) time.Time {
	t, _ := time.Parse("2006-01-02", date)
	return t
}

func unsplash(query string) string {
	return "https://source.unsplash.com/400x400/?" + query
}

// Products returns the seed product catalog.
func Products() []*product.Product {
	return []*product.Product{
		{
			ID: "plant-golden-pothos", Name: "Golden Pothos",
			ScientificName: "Epipremnum aureum", CategoryID: "plants",
			Description:  "An easy-care trailing vine that tolerates low light — a great first plant.",
			ImageURL:     unsplash("golden+pothos+plant"),
			PriceUSD:     20.99, PriceBDT: bdt(20.99),
			Vendor: "The Home Depot", BuyURL: "https://www.homedepot.com/p/202674500",
			LastVerified: verified("2026-09-22"),
		},
		{
			ID: "plant-snake-plant", Name: "Snake Plant",
			ScientificName: "Dracaena trifasciata", CategoryID: "plants",
			Description:  "Extremely low-maintenance; thrives in low light and tolerates neglect.",
			ImageURL:     unsplash("snake+plant+sansevieria"),
			PriceUSD:     21.10, PriceBDT: bdt(21.10),
			Vendor: "The Home Depot", BuyURL: "https://www.homedepot.com/p/202676724",
			LastVerified: verified("2026-09-22"),
		},
		{
			ID: "plant-peace-lily", Name: "Peace Lily",
			ScientificName: "Spathiphyllum", CategoryID: "plants",
			Description:  "Glossy leaves and white blooms; droops to tell you when it's thirsty.",
			ImageURL:     unsplash("peace+lily+plant"),
			PriceUSD:     20.17, PriceBDT: bdt(20.17),
			Vendor: "The Home Depot", BuyURL: "https://www.homedepot.com/p/202204594",
			LastVerified: verified("2026-09-22"),
		},
		{
			ID: "plant-zz-plant", Name: "ZZ Plant",
			ScientificName: "Zamioculcas zamiifolia", CategoryID: "plants",
			Description:  "Near-indestructible — its rhizomes store water, going weeks between waterings.",
			ImageURL:     unsplash("ZZ+plant+zamioculcas"),
			PriceUSD:     19.97, PriceBDT: bdt(19.97),
			Vendor: "The Home Depot", BuyURL: "https://www.homedepot.com/p/202204595",
			LastVerified: verified("2026-09-22"),
		},
		{
			ID: "fert-miracle-gro-all", Name: "Miracle-Gro All Purpose",
			CategoryID:   "fertilizer",
			Description:  "Feeds instantly; NPK 24-8-16 balanced for most houseplants.",
			ImageURL:     unsplash("miracle+gro+plant+fertilizer"),
			PriceUSD:     12.98, PriceBDT: bdt(12.98),
			Vendor: "The Home Depot", BuyURL: "https://www.homedepot.com/p/300499469",
			LastVerified: verified("2026-09-22"),
		},
		{
			ID: "fert-osmocote-plus", Name: "Osmocote Smart-Release",
			CategoryID:   "fertilizer",
			Description:  "Slow-release granules feed for up to 6 months; great for potted plants.",
			ImageURL:     unsplash("osmocote+slow+release+fertilizer"),
			PriceUSD:     16.98, PriceBDT: bdt(16.98),
			Vendor: "The Home Depot", BuyURL: "https://www.homedepot.com/p/301538765",
			LastVerified: verified("2026-09-22"),
		},
		{
			ID: "care-neem-oil", Name: "Neem Oil Spray",
			CategoryID:   "care",
			Description:  "Natural pesticide and fungicide; safe for indoor plants.",
			ImageURL:     unsplash("neem+oil+plant+spray"),
			PriceUSD:     9.97, PriceBDT: bdt(9.97),
			Vendor: "Ace Hardware", BuyURL: "https://www.acehardware.com/departments/lawn-and-garden/lawn-and-garden-chemicals/insect-killer/7464988",
			LastVerified: verified("2026-09-22"),
		},
		{
			ID: "care-moisture-meter", Name: "Soil Moisture Meter",
			CategoryID:   "care",
			Description:  "3-in-1 soil tester: moisture, light and pH. No batteries needed.",
			ImageURL:     unsplash("soil+moisture+meter+plant"),
			PriceUSD:     9.99, PriceBDT: bdt(9.99),
			Vendor: "Ace Hardware", BuyURL: "https://www.acehardware.com/departments/lawn-and-garden/gardening/planters-and-pots/7464990",
			LastVerified: verified("2026-09-22"),
		},
	}
}

// Categories returns the catalog categories.
func Categories() []*product.Category {
	return []*product.Category{
		{ID: "plants", Name: "Houseplants"},
		{ID: "fertilizer", Name: "Plant Food & Fertilizer"},
		{ID: "care", Name: "Plant Care & Pest Control"},
	}
}
