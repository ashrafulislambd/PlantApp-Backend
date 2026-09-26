package seed

import "myplantpal-backend/internal/domain/product"

// lightProducts returns the "grow-lights" category. Prices are Daraz
// Bangladesh selling prices in BDT checked on 2026-09-24; USD is derived.
func lightProducts() []*product.Product {
	return []*product.Product{
		{
			// Sale price; regular price ৳2,251.
			ID: "light-led-grow-50w-daraz", Name: "Full Spectrum LED Grow Light 50W with EU Plug",
			CategoryID:   "grow-lights",
			Description:  "50W full-spectrum LED panel with 660 nm red light for indoor plant growth. Includes a bendable gooseneck and EU plug; suitable for succulents, herbs, and flowering houseplants.",
			ImageURL:     "https://img.drz.lazcdn.com/static/bd/p/dd2afb16e09ff0aa0d27b199e56ac06c.jpg_720x720q80.jpg_.webp",
			PriceBDT:     2071, PriceUSD: usdFromBDT(2071),
			Vendor: "Daraz Bangladesh", BuyURL: "https://www.daraz.com.bd/products/2025-new-led-220v-50w-eu-i571210348.html",
			LastVerified: verified("2026-09-24"),
		},
	}
}
