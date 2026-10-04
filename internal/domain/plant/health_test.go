package plant

import (
	"testing"
	"time"
)

var healthNow = time.Date(2026, 10, 3, 12, 0, 0, 0, time.UTC)

func days(n int) time.Duration { return time.Duration(n) * 24 * time.Hour }

// plantDue builds a plant whose watering was due `lateDays` days before healthNow.
func plantDue(lateDays int) *Plant {
	return &Plant{NextWateringAt: healthNow.Add(-days(lateDays))}
}

func TestComputeHealth_Watering(t *testing.T) {
	tests := []struct {
		name      string
		lateDays  int
		wantScore int
		wantCode  bool
		wantValue int
	}{
		{"not yet due", -2, 100, false, 0},
		{"due today", 0, 100, false, 0},
		{"within grace (1 day)", 1, 100, false, 0},
		{"2 days late", 2, 88, true, 2},
		{"3 days late", 3, 76, true, 3},
		{"5 days late", 5, 52, true, 5},
		{"cap reached at 7 days late", 7, 30, true, 7},
		{"cap holds at 30 days late", 30, 30, true, 30},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := plantDue(tt.lateDays)
			if tt.lateDays < 0 {
				p.NextWateringAt = healthNow.Add(days(-tt.lateDays))
			}
			h := ComputeHealth(p, nil, healthNow)
			if h.Score != tt.wantScore {
				t.Errorf("Score = %d, want %d", h.Score, tt.wantScore)
			}
			if got := len(h.Reasons) == 1 && h.Reasons[0].Code == ReasonWaterLate; got != tt.wantCode {
				t.Errorf("water_late reason present = %v, want %v (reasons %+v)", got, tt.wantCode, h.Reasons)
			}
			if tt.wantCode && h.Reasons[0].Value != tt.wantValue {
				t.Errorf("reason value = %d, want %d", h.Reasons[0].Value, tt.wantValue)
			}
		})
	}
}

func TestComputeHealth_PartialDayIsNotALateDay(t *testing.T) {
	p := &Plant{NextWateringAt: healthNow.Add(-(days(2) - time.Hour))} // 1 day 23h late
	if h := ComputeHealth(p, nil, healthNow); h.Score != 100 {
		t.Errorf("Score = %d, want 100 (1 whole day late is still inside grace)", h.Score)
	}
}

func TestComputeHealth_Fertilizing(t *testing.T) {
	tests := []struct {
		name      string
		lateDays  int
		wantScore int
	}{
		{"within 3 day grace", 3, 100},
		{"4 days late = 1 started week", 4, 95},
		{"10 days late = 1 started week", 10, 95},
		{"11 days late = 2 started weeks", 11, 90},
		{"cap at 15", 60, 85},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			next := healthNow.Add(-days(tt.lateDays))
			p := &Plant{NextWateringAt: healthNow.Add(days(3)), NextFertilizingAt: &next}
			h := ComputeHealth(p, nil, healthNow)
			if h.Score != tt.wantScore {
				t.Errorf("Score = %d, want %d", h.Score, tt.wantScore)
			}
		})
	}
}

func TestComputeHealth_ScanPenaltyAndFade(t *testing.T) {
	base := &Plant{NextWateringAt: healthNow.Add(days(3))}
	tests := []struct {
		name      string
		severity  string
		ageDays   int
		wantScore int
	}{
		{"mild fresh", "Mild", 0, 90},
		{"moderate fresh", "moderate", 3, 75},
		{"severe fresh", "SEVERE", 7, 55},
		{"severe halves from day 8", "severe", 8, 78}, // 45/2 = 22
		{"moderate half day 14", "moderate", 14, 88},  // 25/2 = 12
		{"gone after day 14", "severe", 15, 100},
		{"none severity", "none", 0, 100},
		{"unknown severity", "weird", 0, 100},
		{"empty severity", "", 0, 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scan := &LastScan{Issue: "Leaf spot", Severity: tt.severity, At: healthNow.Add(-days(tt.ageDays))}
			h := ComputeHealth(base, scan, healthNow)
			if h.Score != tt.wantScore {
				t.Errorf("Score = %d, want %d", h.Score, tt.wantScore)
			}
		})
	}
}

func TestComputeHealth_ScanReason(t *testing.T) {
	scan := &LastScan{Issue: "Leaf spot", Severity: "Moderate", At: healthNow.Add(-days(1))}
	h := ComputeHealth(&Plant{NextWateringAt: healthNow.Add(days(1))}, scan, healthNow)
	if len(h.Reasons) != 1 {
		t.Fatalf("Reasons = %+v, want exactly one", h.Reasons)
	}
	r := h.Reasons[0]
	if r.Code != ReasonScanIssue || r.Detail != "Leaf spot" || r.Severity != "moderate" || r.Impact != -25 {
		t.Errorf("reason = %+v, want scan_issue / Leaf spot / moderate / -25", r)
	}
}

func TestComputeHealth_FloorAndCombination(t *testing.T) {
	next := healthNow.Add(-days(60))
	p := &Plant{NextWateringAt: healthNow.Add(-days(30)), NextFertilizingAt: &next}
	scan := &LastScan{Issue: "Rot", Severity: "severe", At: healthNow}
	h := ComputeHealth(p, scan, healthNow)
	if h.Score != 10 {
		t.Errorf("Score = %d, want floor of 10", h.Score)
	}
	if h.State != HealthCritical {
		t.Errorf("State = %q, want critical", h.State)
	}
	if len(h.Reasons) != 3 {
		t.Errorf("len(Reasons) = %d, want 3", len(h.Reasons))
	}
}

func TestStateForScore(t *testing.T) {
	cases := map[int]string{
		100: HealthThriving, 85: HealthThriving,
		84: HealthOkay, 65: HealthOkay,
		64: HealthNeedsCare, 40: HealthNeedsCare,
		39: HealthCritical, 10: HealthCritical,
	}
	for score, want := range cases {
		if got := StateForScore(score); got != want {
			t.Errorf("StateForScore(%d) = %q, want %q", score, got, want)
		}
	}
}

func TestComputeHealth_ReasonsNeverNil(t *testing.T) {
	h := ComputeHealth(&Plant{NextWateringAt: healthNow.Add(days(2))}, nil, healthNow)
	if h.Reasons == nil {
		t.Error("Reasons must be an empty slice (so JSON is [] not null)")
	}
}

func TestWithHealth_CopiesAndDoesNotMutateInput(t *testing.T) {
	orig := plantDue(5)
	scan := &LastScan{Issue: "Leaf spot", Severity: "mild", At: healthNow}
	got := WithHealth(orig, scan, healthNow)
	if got == orig {
		t.Fatal("WithHealth must return a copy")
	}
	if orig.Health != 0 || orig.HealthState != "" || orig.LastScan != nil || orig.HealthReasons != nil {
		t.Errorf("input was mutated: %+v", orig)
	}
	if got.Health != 42 || got.HealthState != HealthNeedsCare {
		t.Errorf("Health/State = %d/%q, want 42/needs_care", got.Health, got.HealthState)
	}
	if got.LastScan == nil || got.LastScan.Issue != "Leaf spot" {
		t.Errorf("LastScan = %+v, want the scan", got.LastScan)
	}
	if WithHealth(orig, nil, healthNow).LastScan != nil {
		t.Error("LastScan must be nil when there is no scan")
	}
}

func TestComputeHealth_TreatedScanHasNoPenalty(t *testing.T) {
	scan := &LastScan{Issue: "Leaf spot", Severity: "Severe", At: healthNow.Add(-days(1)), Treated: true}
	h := ComputeHealth(plantDue(-1), scan, healthNow)
	if h.Score != 100 || len(h.Reasons) != 0 {
		t.Errorf("treated scan: score %d reasons %v, want 100 and none", h.Score, h.Reasons)
	}
	// Still reported as the plant's last scan.
	if got := WithHealth(plantDue(-1), scan, healthNow); got.LastScan == nil || !got.LastScan.Treated {
		t.Errorf("LastScan = %+v, want the treated scan", got.LastScan)
	}
}

func TestLocalized_SwapsScanIssueForBengali(t *testing.T) {
	scan := &LastScan{Issue: "Leaf spot", IssueBn: "পাতায় দাগ", Severity: "Mild", At: healthNow}
	p := WithHealth(plantDue(-1), scan, healthNow)

	bn := p.Localized("bn")
	if bn.LastScan.Issue != "পাতায় দাগ" {
		t.Errorf("LastScan.Issue = %q, want Bengali", bn.LastScan.Issue)
	}
	found := false
	for _, r := range bn.HealthReasons {
		if r.Code == ReasonScanIssue {
			found = true
			if r.Detail != "পাতায় দাগ" {
				t.Errorf("reason detail = %q, want Bengali", r.Detail)
			}
		}
	}
	if !found {
		t.Fatal("no scan_issue reason")
	}
	if p.LastScan.Issue != "Leaf spot" || p.HealthReasons[0].Detail != "Leaf spot" {
		t.Error("Localized must not modify the original")
	}
	if p.Localized("en") != p {
		t.Error("English should return the plant unchanged")
	}
}
