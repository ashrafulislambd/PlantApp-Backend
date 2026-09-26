package seed

import "myplantpal-backend/internal/domain/product"

// potProducts returns the "pots-planters" category. Prices are BDT selling
// prices checked on 2026-09-24; USD is derived from BDT.
func potProducts() []*product.Product {
	return []*product.Product{
		{
			// Sale price; regular price ৳5,970.
			ID: "pot-geo-bag-25g-daraz", Name: "Geotextile Grow Bag (25 Gallon, 10-Pack)",
			CategoryID:   "pots-planters",
			Description:  "A 10-piece set of 25-gallon geotextile grow bags for garden planting. The breathable fabric promotes root health and drainage.",
			ImageURL:     "https://img.drz.lazcdn.com/static/bd/p/beb409c97426c99e6cefabbb1ac6b09a.jpg_720x720q80.jpg",
			PriceBDT:     2755, PriceUSD: usdFromBDT(2755),
			Vendor: "Daraz Bangladesh", BuyURL: "https://www.daraz.com.bd/products/10-25-21-16-i352344812-s1731705895.html",
			LastVerified: verified("2026-09-24"),
		},
	}
}
