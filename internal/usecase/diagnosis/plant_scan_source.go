package diagnosis

import (
	"context"

	"plantpal-backend/internal/domain/diagnosis"
	"plantpal-backend/internal/domain/plant"
)

// PlantScanSource adapts the diagnosis repository to plant.ScanSource, so the
// plant module can read each plant's latest scan for its health score.
type PlantScanSource struct {
	repo diagnosis.Repository
}

func NewPlantScanSource(repo diagnosis.Repository) *PlantScanSource {
	return &PlantScanSource{repo: repo}
}

// LatestScans returns, per requested plant id, the most recent scan that was
// saved against that plant. One repository query serves all plants.
func (a *PlantScanSource) LatestScans(ctx context.Context, userID string, plantIDs []string) (map[string]plant.LastScan, error) {
	want := make(map[string]struct{}, len(plantIDs))
	for _, id := range plantIDs {
		want[id] = struct{}{}
	}
	all, err := a.repo.List(ctx, userID, nil)
	if err != nil {
		return nil, err
	}
	out := make(map[string]plant.LastScan)
	for _, d := range all {
		if d.PlantID == nil {
			continue
		}
		if _, ok := want[*d.PlantID]; !ok {
			continue
		}
		if cur, seen := out[*d.PlantID]; seen && !d.CreatedAt.After(cur.At) {
			continue
		}
		// The newest scan wins even when treated: a treated latest scan must
		// not let an older, untreated one count again. Health skips treated
		// scans itself (plant.ComputeHealth).
		out[*d.PlantID] = plant.LastScan{
			ID:       d.ID,
			Issue:    d.Issue,
			IssueBn:  d.IssueBn,
			Severity: d.Severity,
			At:       d.CreatedAt,
			Treated:  d.Treated,
		}
	}
	return out, nil
}
