package seed

import "myplantpal-backend/internal/domain/product"

// wateringProducts returns the "watering-equipment" category. Prices are BDT
// selling prices checked on 2026-09-24; USD is derived from BDT.
func wateringProducts() []*product.Product {
	return []*product.Product{
		{
			ID: "water-can-ongkoor", Name: "Watering Can",
			CategoryID:   "watering-equipment",
			Description:  "A lightweight plastic watering can with a detachable rose-style spout that delivers a gentle, even shower, suited to rooftop gardens, indoor pots, and nursery seedlings.",
			ImageURL:     "https://ongkoor.com/wp-content/uploads/2026/01/Watering-Can.webp",
			PriceBDT:     550, PriceUSD: usdFromBDT(550),
			Vendor: "Ongkoor.com", BuyURL: "https://ongkoor.com/product/watering-can-%e0%a6%aa%e0%a6%be%e0%a6%a8%e0%a6%bf-%e0%a6%a6%e0%a7%87%e0%a6%93%e0%a6%af%e0%a6%bc%e0%a6%be%e0%a6%b0-%e0%a6%95%e0%a7%8d%e0%a6%af%e0%a6%be%e0%a6%a8/",
			LastVerified: verified("2026-09-24"),
		},
		{
			// Sale price; regular price ৳350. Marketplace listing; seller shop
			// "Ali Arshed Store".
			ID: "water-rfl-can-6l-othoba", Name: "RFL Plastic Watering Can, 6 Liters",
			CategoryID:   "watering-equipment",
			Description:  "A 6-liter teal-blue plastic watering can with a detachable sprinkler rose and dual-handle design, sized to water multiple potted plants without frequent refilling.",
			ImageURL:     "https://images.othoba.com/images/thumbs/2252392_rfl-plastic-gardening-watering-can-jar-6-liters.webp",
			PriceBDT:     308, PriceUSD: usdFromBDT(308),
			Vendor: "Othoba.com", BuyURL: "https://othoba.com/rfl-plastic-gardening-watering-can-jar-6-liters-ali-arshed-store-911764",
			LastVerified: verified("2026-09-24"),
		},
		{
			// Sale price; regular price ৳612 (44% off — a steep discount, but
			// it is clearly shown on the vendor's own page). Marketplace
			// listing; seller shop "Creative Power Engineering".
			ID: "water-spray-bottle-othoba", Name: "900ml Garden Sprayer Bottle",
			CategoryID:   "watering-equipment",
			Description:  "A 900ml hand-pump spray bottle with an adjustable mist-to-stream nozzle, used for misting succulents and herbs or applying diluted fertilizer and pesticide.",
			ImageURL:     "https://images.othoba.com/images/thumbs/2124819_900ml-garden-sprayer-bottle-perfect-for-salon-home-use.jpeg",
			PriceBDT:     340, PriceUSD: usdFromBDT(340),
			Vendor: "Othoba.com", BuyURL: "https://othoba.com/900ml-garden-sprayer-bottle-perfect-for-salon-home-use-creative-power-engineering-870375",
			LastVerified: verified("2026-09-24"),
		},
	}
}
