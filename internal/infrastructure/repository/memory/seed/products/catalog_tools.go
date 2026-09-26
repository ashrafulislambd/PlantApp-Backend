package seed

import "myplantpal-backend/internal/domain/product"

// toolProducts returns the "gardening-tools" category. Prices are Daraz
// Bangladesh selling prices in BDT checked on 2026-09-24; USD is derived.
func toolProducts() []*product.Product {
	return []*product.Product{
		{
			// Sale price; regular price ৳250.
			ID: "tool-mini-kodal-daraz", Name: "Garden Hoe Spade / Kodal with Wooden Handle",
			CategoryID:   "gardening-tools",
			Description:  "Compact garden hoe/spade with a wooden handle for loosening soil, weeding, and planting in small beds and pots. Lightweight and easy to carry.",
			ImageURL:     "https://img.drz.lazcdn.com/static/bd/p/3aa1a8d5330701baa96f2a99365f6d62.jpg_720x720q80.jpg_.webp",
			PriceBDT:     165, PriceUSD: usdFromBDT(165),
			Vendor: "Daraz Bangladesh", BuyURL: "https://www.daraz.com.bd/products/garden-hoe-spade-kodal-kodal-garden-hoe-spade-with-wooden-handle-agricultural-tool-mini-kodal-1-pieces-gardening-tools-i325992607.html",
			LastVerified: verified("2026-09-24"),
		},
		{
			// Variant listing with 8 brand/size options from ৳490 to ৳730;
			// recorded the lowest-priced default option (TOTAL brand, 8-inch,
			// 15mm cutting capacity).
			ID: "tool-pruning-shear-ongkoor", Name: "Pruning Shears",
			CategoryID:   "gardening-tools",
			Description:  "A stainless-steel hand pruning shear for trimming dead branches, spent flowers, and stems up to roughly 1 inch thick, keeping plants tidy and encouraging new growth.",
			ImageURL:     "https://ongkoor.com/wp-content/uploads/2026/01/Pruning-Shears.webp",
			PriceBDT:     490, PriceUSD: usdFromBDT(490),
			Vendor: "Ongkoor.com", BuyURL: "https://ongkoor.com/product/pruning-shears-garden-tools-cutting-scissors/",
			LastVerified: verified("2026-09-24"),
		},
		{
			ID: "tool-hand-trowel-ongkoor", Name: "Hand Trowel",
			CategoryID:   "gardening-tools",
			Description:  "A small hand-held digging tool for transplanting seedlings, loosening potting soil, and mixing in fertilizer in pots or raised beds.",
			ImageURL:     "https://ongkoor.com/wp-content/uploads/2026/01/Trowe.webp",
			PriceBDT:     250, PriceUSD: usdFromBDT(250),
			Vendor: "Ongkoor.com", BuyURL: "https://ongkoor.com/product/hand-trowel-%e0%a6%9f%e0%a7%8d%e0%a6%b0%e0%a6%be%e0%a6%93%e0%a6%af%e0%a6%bc%e0%a7%87%e0%a6%b2-%e0%a6%9b%e0%a7%8b%e0%a6%9f-%e0%a6%95%e0%a7%8b%e0%a6%a6%e0%a6%be%e0%a6%b2/",
			LastVerified: verified("2026-09-24"),
		},
		{
			ID: "tool-garden-gloves-ongkoor", Name: "Garden Gloves",
			CategoryID:   "gardening-tools",
			Description:  "Protective fabric gardening gloves that shield hands from soil, thorns, and fertilizer while weeding, pruning, or repotting.",
			ImageURL:     "https://ongkoor.com/wp-content/uploads/2026/01/Garden-Gloves-2.webp",
			PriceBDT:     690, PriceUSD: usdFromBDT(690),
			Vendor: "Ongkoor.com", BuyURL: "https://ongkoor.com/product/garden-gloves-%e0%a6%97%e0%a6%be%e0%a6%b0%e0%a7%8d%e0%a6%a1%e0%a7%87%e0%a6%a8-%e0%a6%97%e0%a7%8d%e0%a6%b2%e0%a6%be%e0%a6%ad%e0%a6%b8/",
			LastVerified: verified("2026-09-24"),
		},
	}
}
