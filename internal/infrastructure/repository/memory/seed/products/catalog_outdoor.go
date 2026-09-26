package seed

import "myplantpal-backend/internal/domain/product"

// outdoorProducts returns the "outdoor-plants" category. Prices are BDT
// selling prices from Bangladeshi nurseries checked on 2026-09-24; USD is
// derived from BDT.
func outdoorProducts() []*product.Product {
	return []*product.Product{
		{
			// Sale price; regular price ৳500. Listing also offers a without-pot
			// variant (৳350) and a 10-inch-pot variant (৳550); ৳400 is the
			// page's default current price.
			ID: "outdoor-bougainvillea-pink-sobujghor", Name: "Bougainvillea Pink",
			ScientificName: "Bougainvillea glabra", CategoryID: "outdoor-plants",
			Description:  "A fast-growing, thorny climbing shrub covered in long-lasting pink papery bracts. Needs full sun and can be trained along a fence, gate, or trellis.",
			ImageURL:     "https://sobujghor.com/wp-content/uploads/2025/01/bgh-1.jpg",
			PriceBDT:     400, PriceUSD: usdFromBDT(400),
			Vendor: "Sobujghor.com", BuyURL: "https://sobujghor.com/product/bougainvillea-pink-baganbilas-get-ful-kagoj-ful-%e0%a6%ac%e0%a6%97%e0%a6%be%e0%a6%a8%e0%a6%ac%e0%a6%bf%e0%a6%b2%e0%a6%be%e0%a6%b8-%e0%a6%97%e0%a7%8b%e0%a6%b2%e0%a6%be%e0%a6%aa%e0%a6%bf/",
			LastVerified: verified("2026-09-24"),
		},
		{
			ID: "outdoor-areca-palm-sobujghor", Name: "Areca Palm",
			ScientificName: "Dypsis lutescens", CategoryID: "outdoor-plants",
			Description:  "A feathery, clustering palm often used as a privacy screen on rooftops, balconies, or entryways. Grows well in bright indirect light with weekly watering.",
			ImageURL:     "https://sobujghor.com/wp-content/uploads/2025/08/arika-plum.webp",
			PriceBDT:     600, PriceUSD: usdFromBDT(600),
			Vendor: "Sobujghor.com", BuyURL: "https://sobujghor.com/product/%e0%a6%8f%e0%a6%b0%e0%a6%bf%e0%a6%95%e0%a6%be-%e0%a6%aa%e0%a6%be%e0%a6%ae-areca-palm/",
			LastVerified: verified("2026-09-24"),
		},
		{
			// Sale price; regular price ৳500.
			ID: "outdoor-cassia-fistula-sobujghor", Name: "Cassia Fistula (Golden Shower Tree / Sonalu)",
			ScientificName: "Cassia fistula", CategoryID: "outdoor-plants",
			Description:  "A hardy flowering shade tree that bursts into cascades of golden-yellow blooms in summer, commonly planted along roadsides and gardens across Bangladesh. Seed pods are not meant to be eaten.",
			ImageURL:     "https://sobujghor.com/wp-content/uploads/2026/06/sonalu-flower.jpeg",
			PriceBDT:     400, PriceUSD: usdFromBDT(400),
			Vendor: "Sobujghor.com", BuyURL: "https://sobujghor.com/product/cassia-fistula-%e0%a6%b8%e0%a7%8b%e0%a6%a8%e0%a6%be%e0%a6%b2%e0%a7%81-%e0%a6%ab%e0%a7%81%e0%a6%b2%e0%a5%a4/",
			LastVerified: verified("2026-09-24"),
		},
		{
			// Sale price; regular price ৳400.
			ID: "outdoor-hazari-rose-sobujghor", Name: "Hazari Rose (Thousand-Petal Rose)",
			ScientificName: "Rosa hybrid (multi-petaled garden rose cultivar)", CategoryID: "outdoor-plants",
			Description:  "A many-petaled shrub rose available in red, pink, white, or yellow, valued in Bangladesh for its long repeat-blooming season. Needs full sun and well-drained soil.",
			ImageURL:     "https://sobujghor.com/wp-content/uploads/2025/01/Hazari-Rose-tree.jpg",
			PriceBDT:     350, PriceUSD: usdFromBDT(350),
			Vendor: "Sobujghor.com", BuyURL: "https://sobujghor.com/product/hazari-rose-%e0%a6%b9%e0%a6%be%e0%a6%9c%e0%a6%be%e0%a6%b0%e0%a6%bf-%e0%a6%97%e0%a7%8b%e0%a6%b2%e0%a6%be%e0%a6%aa%e0%a5%a4/",
			LastVerified: verified("2026-09-24"),
		},
	}
}
