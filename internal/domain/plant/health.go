package plant

import (
	"context"
	"strings"
	"time"
)

// Health states, from best to worst.
const (
	HealthThriving  = "thriving"   // score >= 85
	HealthOkay      = "okay"       // 65-84
	HealthNeedsCare = "needs_care" // 40-64
	HealthCritical  = "critical"   // < 40
)

// Health reason codes. The server sends codes only; the client localises.
const (
	ReasonWaterLate      = "water_late"      // Value = whole days late
	ReasonFertilizerLate = "fertilizer_late" // Value = whole days late
	ReasonScanIssue      = "scan_issue"      // Detail = issue, Severity = mild|moderate|severe
)

// Penalty tuning (see PLAN.md section 3.2).
const (
	minHealth = 10
	maxHealth = 100

	waterGraceDays     = 1
	waterPenaltyPerDay = 12
	waterPenaltyCap    = 70

	fertilizerGraceDays      = 3
	fertilizerPenaltyPerWeek = 5
	fertilizerPenaltyCap     = 15

	scanFullDays = 7  // full penalty up to and including day 7
	scanHalfDays = 14 // half penalty for days 8-14, none afterwards
)

// LastScan is the most recent AI Doctor scan attached to a plant.
type LastScan struct {
	ID       string    `json:"id,omitempty"`
	Issue    string    `json:"issue"`
	Severity string    `json:"severity"`
	At       time.Time `json:"at"`
	// Treated is set once the user marked the scan as treated. A treated scan
	// no longer lowers health, but stays visible as the plant's last scan.
	Treated bool `json:"treated,omitempty"`
	// IssueBn is the Bengali issue text; only used by Localized, never sent.
	IssueBn string `json:"-"`
}

// ScanSource gives the plant module read access to scans without importing
// the diagnosis module. The returned map holds, per plant id, that plant's
// most recent scan; plants without a scan are simply absent.
type ScanSource interface {
	LatestScans(ctx context.Context, userID string, plantIDs []string) (map[string]LastScan, error)
}

// HealthReason explains one deduction from the health score.
type HealthReason struct {
	Code     string `json:"code"`
	Value    int    `json:"value,omitempty"`
	Detail   string `json:"detail,omitempty"`
	Severity string `json:"severity,omitempty"`
	Impact   int    `json:"impact"` // negative number of points
}

// Health is the computed health of a plant at a moment in time.
type Health struct {
	Score   int
	State   string
	Reasons []HealthReason
}

// StateForScore maps a score to its state.
func StateForScore(score int) string {
	switch {
	case score >= 85:
		return HealthThriving
	case score >= 65:
		return HealthOkay
	case score >= 40:
		return HealthNeedsCare
	default:
		return HealthCritical
	}
}

// wholeDaysLate returns how many full 24h periods `now` is past `due`
// (0 when not yet due or due is unset).
func wholeDaysLate(due, now time.Time) int {
	if due.IsZero() || !now.After(due) {
		return 0
	}
	return int(now.Sub(due) / (24 * time.Hour))
}

func scanBasePenalty(severity string) int {
	switch strings.ToLower(strings.TrimSpace(severity)) {
	case "mild":
		return 10
	case "moderate":
		return 25
	case "severe":
		return 45
	}
	return 0 // "none", "", or anything unknown
}

// ComputeHealth is a pure function: the same plant, scan and clock always
// give the same result. Health is never stored; it is recomputed on read.
func ComputeHealth(p *Plant, scan *LastScan, now time.Time) Health {
	penalty := 0
	reasons := []HealthReason{}

	if late := wholeDaysLate(p.NextWateringAt, now); late > waterGraceDays {
		pen := (late - waterGraceDays) * waterPenaltyPerDay
		if pen > waterPenaltyCap {
			pen = waterPenaltyCap
		}
		penalty += pen
		reasons = append(reasons, HealthReason{Code: ReasonWaterLate, Value: late, Impact: -pen})
	}

	if p.NextFertilizingAt != nil {
		if late := wholeDaysLate(*p.NextFertilizingAt, now); late > fertilizerGraceDays {
			extra := late - fertilizerGraceDays
			weeks := (extra + 6) / 7 // started weeks
			pen := weeks * fertilizerPenaltyPerWeek
			if pen > fertilizerPenaltyCap {
				pen = fertilizerPenaltyCap
			}
			penalty += pen
			reasons = append(reasons, HealthReason{Code: ReasonFertilizerLate, Value: late, Impact: -pen})
		}
	}

	if scan != nil && !scan.Treated {
		pen := scanBasePenalty(scan.Severity)
		ageDays := 0
		if now.After(scan.At) {
			ageDays = int(now.Sub(scan.At) / (24 * time.Hour))
		}
		switch {
		case ageDays <= scanFullDays:
		case ageDays <= scanHalfDays:
			pen /= 2
		default:
			pen = 0
		}
		if pen > 0 {
			penalty += pen
			reasons = append(reasons, HealthReason{
				Code:     ReasonScanIssue,
				Detail:   scan.Issue,
				Severity: strings.ToLower(strings.TrimSpace(scan.Severity)),
				Impact:   -pen,
			})
		}
	}

	score := maxHealth - penalty
	if score < minHealth {
		score = minHealth
	}
	if score > maxHealth {
		score = maxHealth
	}
	return Health{Score: score, State: StateForScore(score), Reasons: reasons}
}

// Localized returns a copy with the scan issue text in the requested language
// ("bn" when a Bengali text exists), both on LastScan and in the scan_issue
// health reason. The input is not modified.
func (p *Plant) Localized(lang string) *Plant {
	if lang != "bn" || p.LastScan == nil || p.LastScan.IssueBn == "" {
		return p
	}
	cp := *p
	scan := *p.LastScan
	scan.Issue = scan.IssueBn
	cp.LastScan = &scan
	cp.HealthReasons = make([]HealthReason, len(p.HealthReasons))
	copy(cp.HealthReasons, p.HealthReasons)
	for i := range cp.HealthReasons {
		if cp.HealthReasons[i].Code == ReasonScanIssue {
			cp.HealthReasons[i].Detail = scan.IssueBn
		}
	}
	return &cp
}

// LocalizedAll applies Localized to every plant.
func LocalizedAll(items []*Plant, lang string) []*Plant {
	out := make([]*Plant, len(items))
	for i, p := range items {
		out[i] = p.Localized(lang)
	}
	return out
}

// WithHealth returns a copy of p with the computed health fields filled in.
// The input (which may be the repository's own instance) is not modified.
func WithHealth(p *Plant, scan *LastScan, now time.Time) *Plant {
	h := ComputeHealth(p, scan, now)
	cp := *p
	cp.Health = h.Score
	cp.HealthState = h.State
	cp.HealthReasons = h.Reasons
	if scan != nil {
		s := *scan
		cp.LastScan = &s
	} else {
		cp.LastScan = nil
	}
	return &cp
}
