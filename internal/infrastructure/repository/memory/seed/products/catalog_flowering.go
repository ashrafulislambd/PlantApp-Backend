package seed

import "myplantpal-backend/internal/domain/product"

// floweringProducts returns the "flowering-plants" category. Prices are BDT
// selling prices from Bangladeshi nurseries checked on 2026-09-24; USD is
// derived from BDT.
func floweringProducts() []*product.Product {
	return []*product.Product{
		{
			ID: "flower-kamini-ongkoor", Name: "Kamini",
			ScientificName: "Murraya paniculata", CategoryID: "flowering-plants",
			Description:  "A compact shrub with glossy green leaves and clusters of small, fragrant white flowers, grown indoors or outdoors in Bangladesh. Tolerates low care once established.",
			ImageURL:     "https://ongkoor.com/wp-content/uploads/2026/01/Kamini.webp",
			PriceBDT:     300, PriceUSD: usdFromBDT(300),
			Vendor: "Ongkoor.com", BuyURL: "https://ongkoor.com/product/kamini-fragrant-premium-floral-plant/",
			LastVerified: verified("2026-09-24"),
		},
		{
			ID: "flower-petunia-ongkoor", Name: "Petunia",
			ScientificName: "Petunia × hybrida", CategoryID: "flowering-plants",
			Description:  "A bushy, brightly colored winter-flowering annual popular in Bangladesh for hanging baskets, window boxes, and rooftop gardens. Needs 5–6 hours of direct sun to bloom well.",
			ImageURL:     "https://ongkoor.com/wp-content/uploads/2025/12/Petunia-9.webp",
			PriceBDT:     120, PriceUSD: usdFromBDT(120),
			Vendor: "Ongkoor.com", BuyURL: "https://ongkoor.com/product/petunia-%e0%a6%aa%e0%a6%bf%e0%a6%9f%e0%a7%81%e0%a6%a8%e0%a6%bf%e0%a6%af%e0%a6%bc%e0%a6%be-%e0%a6%ab%e0%a7%81%e0%a6%b2-%e0%a6%97%e0%a6%be%e0%a6%9b/",
			LastVerified: verified("2026-09-24"),
		},
		{
			ID: "flower-crown-of-thorns-ongkoor", Name: "Crown of Thorns",
			ScientificName: "Euphorbia milii", CategoryID: "flowering-plants",
			Description:  "A thorny, drought-tolerant succulent shrub that flowers nearly year-round in small red, pink, or yellow blooms. Its milky sap can irritate skin, so handle with care.",
			ImageURL:     "https://ongkoor.com/wp-content/uploads/2026/01/Kata-Mukut.webp",
			PriceBDT:     390, PriceUSD: usdFromBDT(390),
			Vendor: "Ongkoor.com", BuyURL: "https://ongkoor.com/product/crown-of-thorns-%e0%a6%95%e0%a6%be%e0%a6%81%e0%a6%9f%e0%a6%be-%e0%a6%ae%e0%a7%81%e0%a6%95%e0%a7%81%e0%a6%9f/",
			LastVerified: verified("2026-09-24"),
		},
		{
			ID: "flower-chandra-mallika-ongkoor", Name: "Chandra Mallika (Chrysanthemum)",
			ScientificName: "Chrysanthemum × morifolium", CategoryID: "flowering-plants",
			Description:  "A winter-blooming chrysanthemum grown for its dense, colorful flower heads in pots and rooftop gardens. Pinching young shoots encourages bushier growth and more blooms.",
			ImageURL:     "https://ongkoor.com/wp-content/uploads/2026/01/Chandra-Mallika.webp",
			PriceBDT:     200, PriceUSD: usdFromBDT(200),
			Vendor: "Ongkoor.com", BuyURL: "https://ongkoor.com/product/chandra-mallika-%e0%a6%9a%e0%a6%a8%e0%a7%8d%e0%a6%a6%e0%a7%8d%e0%a6%b0%e0%a6%ae%e0%a6%b2%e0%a7%8d%e0%a6%b2%e0%a6%bf%e0%a6%95%e0%a6%be-%e0%a6%ab%e0%a7%81%e0%a6%b2-%e0%a6%97%e0%a6%be%e0%a6%9b/",
			LastVerified: verified("2026-09-24"),
		},
	}
}
