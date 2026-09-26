package diagnosis

import "testing"

func TestDiagnosis_Localized(t *testing.T) {
	d := Diagnosis{
		ID:           "diag_1",
		Issue:        "Phosphorus deficiency.",
		Cure:         "Use TSP.",
		Disclaimer:   Disclaimer,
		IssueBn:      "ফসফরাসের অভাব।",
		CureBn:       "TSP ব্যবহার করুন।",
		DisclaimerBn: DisclaimerBn,
	}

	t.Run("english returns original fields", func(t *testing.T) {
		got := d.Localized("en")
		if got.Issue != d.Issue || got.Cure != d.Cure || got.Disclaimer != d.Disclaimer {
			t.Errorf("Localized(en) changed fields unexpectedly: %+v", got)
		}
	})

	t.Run("bengali swaps in translated fields", func(t *testing.T) {
		got := d.Localized("bn")
		if got.Issue != d.IssueBn || got.Cure != d.CureBn || got.Disclaimer != d.DisclaimerBn {
			t.Errorf("Localized(bn) = %+v, want translated fields", got)
		}
		if got.ID != d.ID {
			t.Errorf("Localized(bn) must not change ID: got %q, want %q", got.ID, d.ID)
		}
	})

	t.Run("bengali falls back to english when no translation exists", func(t *testing.T) {
		untranslated := Diagnosis{ID: "diag_2", Issue: "Some issue", Cure: "Some cure", Disclaimer: Disclaimer}
		got := untranslated.Localized("bn")
		if got.Issue != untranslated.Issue || got.Cure != untranslated.Cure || got.Disclaimer != untranslated.Disclaimer {
			t.Errorf("Localized(bn) without translation should fall back to English, got %+v", got)
		}
	})
}
