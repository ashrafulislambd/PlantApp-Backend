// Package seed provides initial in-memory data so the API is useful
// out of the box, without a real database behind it yet.
package seed

import (
	"time"

	"myplantpal-backend/internal/domain/fertilizer"
)

// Fertilizers returns the starter recipes shown in the Figma design plus a
// couple of siblings covering the other two primary nutrients.
func Fertilizers() []*fertilizer.Fertilizer {
	now := time.Now().UTC()
	return []*fertilizer.Fertilizer{
		{
			ID:         "fert_seed_nitrogen",
			Name:       "Nitrogen (leaf growth)",
			Category:   "Nitrogen",
			NameBn:     "নাইট্রোজেন (পাতার বৃদ্ধি)",
			CategoryBn: "নাইট্রোজেন",
			Instructions: "Put 1 banana peel + used tea leaves + 1 liter water in a bottle/jar. " +
				"Keep it in a shady, cool place (not under direct sun).\n" +
				"Soak for 2 days.\n" +
				"Then pour a little around the soil near the roots, not directly on the trunk.\n" +
				"Use it once every 10-15 days.",
			InstructionsBn: "১টি কলার খোসা, ব্যবহৃত চায়ের পাতা এবং ১ লিটার পানি একটি বোতলে বা বয়ামে রাখুন। " +
				"ছায়াযুক্ত, ঠান্ডা জায়গায় ২ দিন ভিজিয়ে রাখুন।\n" +
				"এরপর কাণ্ডে সরাসরি না দিয়ে শিকড়ের কাছের মাটিতে অল্প পরিমাণে ঢালুন।\n" +
				"১০-১৫ দিন পরপর একবার ব্যবহার করুন।",
			CreatedAt: now,
		},
		{
			ID:         "fert_seed_phosphorus",
			Name:       "Phosphorus (root & flower growth)",
			Category:   "Phosphorus",
			NameBn:     "ফসফরাস (শিকড় ও ফুলের বৃদ্ধি)",
			CategoryBn: "ফসফরাস",
			Instructions: "Crush dried eggshells into a fine powder and mix into the top layer of soil. " +
				"Alternatively, soak rice water for 24 hours and use it to water the plant once a week.",
			InstructionsBn: "শুকনো ডিমের খোসা গুঁড়ো করে মাটির ওপরের স্তরে মিশিয়ে দিন। " +
				"অথবা, ভাত ধোয়া পানি ২৪ ঘণ্টা রেখে সপ্তাহে একবার গাছে দিন।",
			CreatedAt: now,
		},
		{
			ID:         "fert_seed_potassium",
			Name:       "Potassium (blooming & disease resistance)",
			Category:   "Potassium",
			NameBn:     "পটাশিয়াম (ফুল ফোটা ও রোগ প্রতিরোধ)",
			CategoryBn: "পটাশিয়াম",
			Instructions: "Save banana peels and dry them in the sun for a few days, then bury the dried " +
				"peels a few centimeters deep near the plant's roots. Water lightly afterward.",
			InstructionsBn: "কলার খোসা কয়েক দিন রোদে শুকিয়ে নিন, তারপর গাছের শিকড়ের কাছে কয়েক সেন্টিমিটার " +
				"গভীরে পুঁতে দিন। এরপর অল্প পানি দিন।",
			CreatedAt: now,
		},
	}
}
