package seed

import "myplantpal-backend/internal/domain/product"

// plantProducts returns the "plants" (Houseplants) category for BDT-priced
// Bangladeshi vendors. Prices are selling prices in BDT checked on 2026-09-24;
// USD is derived from BDT.
func plantProducts() []*product.Product {
	return []*product.Product{
		{
			ID: "plant-jade-crassula-ongkoor", Name: "Jade Plant (Crassula ovata)",
			ScientificName: "Crassula ovata", CategoryID: "plants",
			Description:  "A slow-growing succulent with thick, glossy green leaves. Prefers bright light and only occasional watering; toxic to pets if eaten.",
			ImageURL:     "https://ongkoor.com/wp-content/uploads/2025/12/Jade-Plant-1.webp",
			PriceBDT:     350, PriceUSD: usdFromBDT(350),
			Vendor: "Ongkoor.com", BuyURL: "https://ongkoor.com/product/jade-plant-crassula-ovata/",
			LastVerified: verified("2026-09-24"),
		},
		{
			ID: "plant-zz-zamioculcas-sobujghor", Name: "ZZ Plant (Zamioculcas zamiifolia)",
			ScientificName: "Zamioculcas zamiifolia", CategoryID: "plants",
			Description:  "A hardy indoor plant with glossy, dark green leaves on upright stems, about 2–3 ft tall. Tolerates low light and needs very little water; toxic to pets if eaten.",
			ImageURL:     "https://sobujghor.com/wp-content/uploads/2025/08/zz-plant.webp",
			PriceBDT:     600, PriceUSD: usdFromBDT(600),
			Vendor: "Sobujghor.com", BuyURL: "https://sobujghor.com/product/%E0%A6%9C%E0%A6%BF%E0%A6%9C%E0%A6%BF-%E0%A6%AA%E0%A7%8D%E0%A6%B2%E0%A6%BE%E0%A6%A8%E0%A7%8D%E0%A6%9F-zz-plant/",
			LastVerified: verified("2026-09-24"),
		},
		{
			ID: "plant-lucky-bamboo-950-awal", Name: "Lucky Bamboo (3-Layer Ceramic Pot)",
			ScientificName: "Dracaena sanderiana", CategoryID: "plants",
			Description:  "A 3-layer lucky bamboo arrangement in a ceramic pot with white stones. Grows in water and tolerates low light; toxic to pets if eaten.",
			ImageURL:     "https://awalexpressbd.com/wp-content/uploads/2025/05/Lucky-Bamboo-2.jpg",
			PriceBDT:     950, PriceUSD: usdFromBDT(950),
			Vendor: "Awal Express BD", BuyURL: "https://awalexpressbd.com/?product=lucky-bamboo",
			LastVerified: verified("2026-09-24"),
		},
		{
			ID: "plant-spider-plant-awal", Name: "Spider Plant (Chlorophytum comosum)",
			ScientificName: "Chlorophytum comosum", CategoryID: "plants",
			Description:  "An adaptable indoor plant with arching, variegated leaves that sends out plantlets on runners. Sold with a white ceramic pot.",
			ImageURL:     "https://awalexpressbd.com/wp-content/uploads/2025/05/WhatsApp-Image-2025-05-22-at-09.19.52_d3145284.webp",
			PriceBDT:     400, PriceUSD: usdFromBDT(400),
			Vendor: "Awal Express BD", BuyURL: "https://awalexpressbd.com/?product=spider-plant",
			LastVerified: verified("2026-09-24"),
		},
	}
}
