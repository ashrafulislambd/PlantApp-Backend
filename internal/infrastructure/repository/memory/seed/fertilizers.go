// Package seed provides initial in-memory data so the API is useful
// out of the box, without a real database behind it yet.
package seed

import (
	"time"

	"myplantpal-backend/internal/domain/fertilizer"
)

// Fertilizers returns the starter recipes shown in the Figma design plus a
// couple of siblings covering the other two primary nutrients. Each is
// seeded with both English and Bengali text.
func Fertilizers() []*fertilizer.Fertilizer {
	now := time.Now().UTC()
	return []*fertilizer.Fertilizer{
		{
			ID:       "fert_seed_nitrogen",
			Name:     "Nitrogen (leaf growth)",
			Category: "Nitrogen",
			Instructions: "Put 1 banana peel + used tea leaves + 1 liter water in a bottle/jar. " +
				"Keep it in a shady, cool place (not under direct sun).\n" +
				"Soak for 2 days.\n" +
				"Then pour a little around the soil near the roots, not directly on the trunk.\n" +
				"Use it once every 10-15 days.",
			NameBn:     "নাইট্রোজেন (পাতার বৃদ্ধি)",
			CategoryBn: "নাইট্রোজেন",
			InstructionsBn: "একটি বোতল/জারে অথবা কলসিয়ায় 1টি কলার খোসা + ব্যবহৃত চাপাতা + 1 লিটার পানি রাখুন। " +
				"একটি ছায়াযুক্ত, ঠান্ডা জায়গায় রাখুন (সরাসরি রোদের নীচে নয়)।\n" +
				"2 দিন ভিজিয়ে রাখুন।\n" +
				"তারপর মূলের কাছাকাছি মাটিতে অল্প করে ঢালুন, সরাসরি কাণ্ডে নয়।\n" +
				"প্রতি 10-15 দিন অন্তর একবার ব্যবহার করুন।",
			CreatedAt: now,
		},
		{
			ID:       "fert_seed_phosphorus",
			Name:     "Phosphorus (root & flower growth)",
			Category: "Phosphorus",
			Instructions: "Crush dried eggshells into a fine powder and mix into the top layer of soil. " +
				"Alternatively, soak rice water for 24 hours and use it to water the plant once a week.",
			NameBn:     "ফসফরাস (মূল ও ফুল বৃদ্ধি)",
			CategoryBn: "ফসফরাস",
			InstructionsBn: "শুকনো ডিমের খোসা গুড়ো করে মাটির উপরের স্তরে মিশিয়ে দিন। " +
				"বিকল্পভাবে, 24 ঘণ্টা চাল ভিজিয়ে রাখা পানি সপ্তাহে একবার গাছে দিন।",
			CreatedAt: now,
		},
		{
			ID:       "fert_seed_potassium",
			Name:     "Potassium (blooming & disease resistance)",
			Category: "Potassium",
			Instructions: "Save banana peels and dry them in the sun for a few days, then bury the dried " +
				"peels a few centimeters deep near the plant's roots. Water lightly afterward.",
			NameBn:     "পটাশিয়াম (ফুল ও রোগ প্রতিরোধ)",
			CategoryBn: "পটাশিয়াম",
			InstructionsBn: "কলার খোসা সংরক্ষণ করে কয়েক দিন রোদে শুকান, তারপর গাছের মূলের কাছে " +
				"কয়েক সেন্টিমিটার গভীরে পুঁতে দিন। তারপর হালকা পানি দিন।",
			CreatedAt: now,
		},
	}
}
