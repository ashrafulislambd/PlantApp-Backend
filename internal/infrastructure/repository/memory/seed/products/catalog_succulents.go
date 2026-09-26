package seed

import "myplantpal-backend/internal/domain/product"

// succulentProducts returns the "succulents-cacti" category. Prices are BDT
// selling prices checked on 2026-09-24; USD is derived from BDT.
func succulentProducts() []*product.Product {
	return []*product.Product{
		{
			ID: "succ-ball-cactus-sobujghor", Name: "Ball Cactus (Parodia magnifica)",
			ScientificName: "Parodia magnifica", CategoryID: "succulents-cacti",
			Description:  "A small, spherical cactus with pale yellow-white spines, about 10–15 cm. Thrives in bright indirect light and needs very little water.",
			ImageURL:     "https://sobujghor.com/wp-content/uploads/2025/08/ball-cactus.png",
			PriceBDT:     300, PriceUSD: usdFromBDT(300),
			Vendor: "Sobujghor.com", BuyURL: "https://sobujghor.com/product/%E0%A6%AC%E0%A6%B2-%E0%A6%95%E0%A7%8D%E0%A6%AF%E0%A6%BE%E0%A6%95%E0%A6%9F%E0%A6%BE%E0%A6%B8-parodia-magnifica/",
			LastVerified: verified("2026-09-24"),
		},
	}
}
