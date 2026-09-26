package seed

import "myplantpal-backend/internal/domain/product"

// herbProducts returns the "herbs-vegetables" category. Prices are BDT
// selling prices from Bangladeshi nurseries checked on 2026-09-24; USD is
// derived from BDT.
func herbProducts() []*product.Product {
	return []*product.Product{
		{
			// Priced higher than the common seasonal herbs from the same
			// vendor: rosemary is a slower-growing specialty herb here, not a
			// listing error.
			ID: "herb-rosemary-sobujghor", Name: "Rosemary Plant",
			ScientificName: "Salvia rosmarinus", CategoryID: "herbs-vegetables",
			Description:  "A fragrant, needle-leaved evergreen herb used fresh or dried in cooking, grown well in pots on sunny balconies. Tolerates dry spells once established.",
			ImageURL:     "https://sobujghor.com/wp-content/uploads/2025/08/%e0%a6%b0%e0%a7%8b%e0%a6%9c-%e0%a6%ae%e0%a7%87%e0%a6%b0%e0%a6%bf-.jpg",
			PriceBDT:     700, PriceUSD: usdFromBDT(700),
			Vendor: "Sobujghor.com", BuyURL: "https://sobujghor.com/product/%e0%a6%b0%e0%a7%8b%e0%a6%9c%e0%a6%ae%e0%a7%87%e0%a6%b0%e0%a6%bf-%e0%a6%97%e0%a6%be%e0%a6%9b-rosemary-plant/",
			LastVerified: verified("2026-09-24"),
		},
		{
			ID: "herb-mint-ongkoor", Name: "Mint (Pudina)",
			ScientificName: "Mentha spicata", CategoryID: "herbs-vegetables",
			Description:  "A fast-spreading culinary herb with cooling, aromatic leaves used in teas, chutneys, and drinks. Thrives in small pots on a sunny balcony with regular watering.",
			ImageURL:     "https://ongkoor.com/wp-content/uploads/2026/01/Mint.webp",
			PriceBDT:     100, PriceUSD: usdFromBDT(100),
			Vendor: "Ongkoor.com", BuyURL: "https://ongkoor.com/product/mint-%e0%a6%aa%e0%a7%81%e0%a6%a6%e0%a6%bf%e0%a6%a8%e0%a6%be-%e0%a6%97%e0%a6%be%e0%a6%9b/",
			LastVerified: verified("2026-09-24"),
		},
		{
			ID: "herb-bayleaf-ongkoor", Name: "Bay Leaf (Tejpata)",
			ScientificName: "Cinnamomum tamala", CategoryID: "herbs-vegetables",
			Description:  "A slow-growing aromatic tree grown for its dried leaves, a staple seasoning in Bengali biryani, curries, and stews. Prefers full sun and well-drained loamy soil.",
			ImageURL:     "https://ongkoor.com/wp-content/uploads/2026/01/Cinnamomum-Tamala.webp",
			PriceBDT:     300, PriceUSD: usdFromBDT(300),
			Vendor: "Ongkoor.com", BuyURL: "https://ongkoor.com/product/bay-leaf-%e0%a6%a4%e0%a7%87%e0%a6%9c%e0%a6%aa%e0%a6%be%e0%a6%a4%e0%a6%be-cinnamomum-tamala/",
			LastVerified: verified("2026-09-24"),
		},
	}
}
