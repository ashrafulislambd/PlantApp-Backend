package fertilizer

import "testing"

func TestLocalized(t *testing.T) {
	f := Fertilizer{
		Name:           "Nitrogen",
		Category:       "Nutrient",
		Instructions:   "English instructions",
		NameBn:         "নাইট্রোজেন",
		CategoryBn:     "পুষ্টি উপাদান",
		InstructionsBn: "বাংলা নির্দেশনা",
	}

	if got := f.Localized("bn"); got.Name != f.NameBn || got.Category != f.CategoryBn || got.Instructions != f.InstructionsBn {
		t.Fatalf("Localized(\"bn\") = %+v, want Bengali fields", got)
	}
	if got := f.Localized("en"); got.Name != f.Name || got.Category != f.Category || got.Instructions != f.Instructions {
		t.Fatalf("Localized(\"en\") = %+v, want original fields", got)
	}

	withoutTranslation := Fertilizer{Name: "Custom fertilizer"}
	if got := withoutTranslation.Localized("bn"); got.Name != withoutTranslation.Name {
		t.Fatalf("Localized(\"bn\") without translation = %+v, want original fields", got)
	}
}
