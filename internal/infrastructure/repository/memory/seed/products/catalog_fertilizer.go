package seed

import "myplantpal-backend/internal/domain/product"

// fertilizerProducts returns the BDT-priced part of the "fertilizer" category.
// Prices are selling prices checked on 2026-09-24; USD is derived from BDT.
func fertilizerProducts() []*product.Product {
	return []*product.Product{
		{
			// Sale price; regular price ৳124.
			ID: "fert-npk-american-daraz", Name: "NPK Fertilizer (American NPKS 8:20:14:5, 100 g)",
			CategoryID:   "fertilizer",
			Description:  "A water-soluble NPK fertilizer supplied as a 100 g pack. Dissolves quickly; the higher phosphorus share suits flowering and fruiting plants.",
			ImageURL:     "https://img.drz.lazcdn.com/static/bd/p/c73b928da091a83fe99d51d1265b1e80.jpg_720x720q80.jpg",
			PriceBDT:     107, PriceUSD: usdFromBDT(107),
			Vendor: "Daraz Bangladesh", BuyURL: "https://www.daraz.com.bd/products/npk-npks-820145-50-50-100-i292940214.html",
			LastVerified: verified("2026-09-24"),
		},
		{
			ID: "fert-epsom-salt-ongkoor", Name: "Magnesium Sulfate (Epsom Salt)",
			CategoryID:   "fertilizer",
			Description:  "Agriculture-grade Epsom salt used as a diluted foliar spray or soil drench to correct magnesium deficiency, greening up yellowing leaves on roses, chilies, and potted plants.",
			ImageURL:     "https://ongkoor.com/wp-content/uploads/2026/01/Magnesium-Sulfate.webp",
			PriceBDT:     100, PriceUSD: usdFromBDT(100),
			Vendor: "Ongkoor.com", BuyURL: "https://ongkoor.com/product/magnesium-sulfate-%e0%a6%87%e0%a6%aa%e0%a6%b8%e0%a6%ae-%e0%a6%b8%e0%a6%b2%e0%a7%8d%e0%a6%9f/",
			LastVerified: verified("2026-09-24"),
		},
		{
			ID: "fert-bone-meal-ongkoor", Name: "Bone Meal",
			CategoryID:   "fertilizer",
			Description:  "A slow-release organic fertilizer made from processed animal bone, supplying phosphorus and calcium to strengthen roots and boost flowering and fruiting.",
			ImageURL:     "https://ongkoor.com/wp-content/uploads/2026/01/Bone-Meal.webp",
			PriceBDT:     120, PriceUSD: usdFromBDT(120),
			Vendor: "Ongkoor.com", BuyURL: "https://ongkoor.com/product/bone-meal-%e0%a6%b9%e0%a6%be%e0%a6%a1%e0%a6%bc%e0%a7%87%e0%a6%b0-%e0%a6%97%e0%a7%81%e0%a6%a1%e0%a6%bc%e0%a6%be/",
			LastVerified: verified("2026-09-24"),
		},
		{
			ID: "fert-mustard-cake-ongkoor", Name: "Mustard Oil Cake",
			CategoryID:   "fertilizer",
			Description:  "A traditional organic fertilizer left over from pressing mustard seeds for oil, mixed into soil or soaked in water to make a nitrogen-rich liquid feed for leafy growth.",
			ImageURL:     "https://ongkoor.com/wp-content/uploads/2026/01/Mustard-Oil-Cake.webp",
			PriceBDT:     75, PriceUSD: usdFromBDT(75),
			Vendor: "Ongkoor.com", BuyURL: "https://ongkoor.com/product/mustard-oil-cake-%e0%a6%b8%e0%a6%b0%e0%a6%bf%e0%a6%b7%e0%a6%be%e0%a6%b0-%e0%a6%96%e0%a7%88%e0%a6%b2/",
			LastVerified: verified("2026-09-24"),
		},
	}
}
