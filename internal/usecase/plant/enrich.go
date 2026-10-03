package plant

import (
	"context"
	"time"

	"plantpal-backend/internal/domain/plant"
)

// withHealth returns a copy of p with health, healthState, healthReasons and
// lastScan filled in. A failing scan lookup never fails the request; health is
// then computed without the scan.
func (s *Service) withHealth(ctx context.Context, p *plant.Plant) *plant.Plant {
	return s.withHealthAll(ctx, p.UserID, []*plant.Plant{p})[0]
}

// withHealthAll does the same for many plants with a single scan lookup.
func (s *Service) withHealthAll(ctx context.Context, userID string, items []*plant.Plant) []*plant.Plant {
	now := time.Now().UTC()
	var scans map[string]plant.LastScan
	if s.scans != nil && len(items) > 0 {
		ids := make([]string, 0, len(items))
		for _, p := range items {
			ids = append(ids, p.ID)
		}
		if m, err := s.scans.LatestScans(ctx, userID, ids); err == nil {
			scans = m
		}
	}
	out := make([]*plant.Plant, 0, len(items))
	for _, p := range items {
		var scan *plant.LastScan
		if sc, ok := scans[p.ID]; ok {
			sc := sc
			scan = &sc
		}
		out = append(out, plant.WithHealth(p, scan, now))
	}
	return out
}
