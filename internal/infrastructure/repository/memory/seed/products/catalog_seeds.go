package seed

import "myplantpal-backend/internal/domain/product"

// seedProducts returns the "seeds" category. Prices are Daraz Bangladesh
// selling prices in BDT checked on 2026-09-24; USD is derived from BDT.
func seedProducts() []*product.Product {
	return []*product.Product{
		{
			// Sale price; regular price ৳95.
			ID: "seed-hibiscus-daraz", Name: "Hibiscus / Joba Flower Seeds – 15+ pcs",
			ScientificName: "Hibiscus rosa-sinensis", CategoryID: "seeds",
			Description:  "Mixed-color hibiscus seeds that produce red, pink, orange, white, and yellow blooms. Suitable for balconies and rooftops; reported germination rate 80–95%.",
			ImageURL:     "https://img.drz.lazcdn.com/static/bd/p/a82e02fff644732fe69686791637bfcb.jpg_720x720q80.jpg_.webp",
			PriceBDT:     35, PriceUSD: usdFromBDT(35),
			Vendor: "Daraz Bangladesh", BuyURL: "https://www.daraz.com.bd/products/hibiscus-flower-1530-pcs-seed-mixed-color-hibiscus-moscheutos-i177626484.html",
			LastVerified: verified("2026-09-24"),
		},
		{
			// Sale price; regular price ৳99.
			ID: "seed-chili-ornamental-daraz", Name: "All Season Ornamental Chili Seeds – 20 pcs",
			ScientificName: "Capsicum annuum", CategoryID: "seeds",
			Description:  "Pack of 20 ornamental chili seeds suitable for indoor or outdoor growing. Germination rate reported at 80–95%; plants produce small, colorful chilies.",
			ImageURL:     "https://img.drz.lazcdn.com/g/kf/S0a353a4099d043b9a115ecee48b3398eA.jpg_720x720q80.jpg",
			PriceBDT:     46, PriceUSD: usdFromBDT(46),
			Vendor: "Daraz Bangladesh", BuyURL: "https://www.daraz.com.bd/products/all-season-ornamental-chili-seeds-20-pcs-seeds-i203760286.html",
			LastVerified: verified("2026-09-24"),
		},
		{
			// Sale price; regular price ৳113.
			ID: "seed-china-chili-daraz", Name: "China Hybrid Green Long Chili Seeds – 30 pcs + Gift",
			ScientificName: "Capsicum annuum", CategoryID: "seeds",
			Description:  "Pack of 30 long green chili seeds from a Chinese hybrid variety, suitable for indoor or outdoor cultivation. Reported germination rate 80–95%; includes a random free seed gift.",
			ImageURL:     "https://img.drz.lazcdn.com/static/bd/p/eaf43ebad3b42dabf70fb06d2b8650d7.jpg_720x720q80.jpg_.webp",
			PriceBDT:     51, PriceUSD: usdFromBDT(51),
			Vendor: "Daraz Bangladesh", BuyURL: "https://www.daraz.com.bd/products/china-hybrid-confirm-green-long-chili-seed-30-pcs-seedsgift-i325311798.html",
			LastVerified: verified("2026-09-24"),
		},
		{
			// Sale price; regular price ৳125.
			ID: "seed-cherry-tomato-yellow-daraz", Name: "Cherry Tomato Yellow Seeds F1 Hybrid – 20 pcs",
			ScientificName: "Solanum lycopersicum", CategoryID: "seeds",
			Description:  "Pack of 20 F1 hybrid yellow cherry tomato seeds suitable for rooftop or balcony gardens. Reported germination rate 80–95%; plants begin fruiting in about 60–70 days.",
			ImageURL:     "https://img.drz.lazcdn.com/static/bd/p/bd63cb0b1efcca4b976c84015d5b9cab.jpg_720x720q80.jpg",
			PriceBDT:     39, PriceUSD: usdFromBDT(39),
			Vendor: "Daraz Bangladesh", BuyURL: "https://www.daraz.com.bd/products/20-cherry-tomato-yellow-i124402940.html",
			LastVerified: verified("2026-09-24"),
		},
	}
}
