package fertilizer_test

import (
	"testing"

	"myplantpal-backend/internal/domain/fertilizer"
)

func TestFertilizer_Localized(t *testing.T) {
	f := fertilizer.Fertilizer{
		ID:             "fert_1",
		Name:           "Nitrogen",
		Category:       "Nitrogen",
		Instructions:   "Do the English thing.",
		NameBn:         "নাইট্রোজেন",
		CategoryBn:     "নাইট্রোজেন",
		InstructionsBn: "বাংলায় করুন।",
	}

	t.Run("english returns original fields", func(t *testing.T) {
		got := f.Localized("en")
		if got.Name != f.Name || got.Category != f.Category || got.Instructions != f.Instructions {
			t.Errorf("Localized(en) changed fields unexpectedly: %+v", got)
		}
	})

	t.Run("bengali swaps in translated fields", func(t *testing.T) {
		got := f.Localized("bn")
		if got.Name != f.NameBn || got.Category != f.CategoryBn || got.Instructions != f.InstructionsBn {
			t.Errorf("Localized(bn) = %+v, want translated fields", got)
		}
		if got.ID != f.ID {
			t.Errorf("Localized(bn) must not change ID: got %q, want %q", got.ID, f.ID)
		}
	})

	t.Run("bengali falls back to english when no translation exists", func(t *testing.T) {
		untranslated := fertilizer.Fertilizer{
			ID:           "fert_2",
			Name:         "User Added Mix",
			Category:     "Custom",
			Instructions: "Whatever the user typed.",
		}
		got := untranslated.Localized("bn")
		if got.Name != untranslated.Name || got.Category != untranslated.Category || got.Instructions != untranslated.Instructions {
			t.Errorf("Localized(bn) without translation should fall back to English, got %+v", got)
		}
	})
}
