package memory

import (
"context"
"time"

"plantpal-backend/internal/domain/plant"
)

// ListDueBetween returns plants of every user whose next watering or
// fertilizing falls in (after, before]. Used by the reminder scheduler.
func (r *PlantRepository) ListDueBetween(_ context.Context, after, before time.Time) ([]*plant.Plant, error) {
r.mu.RLock()
defer r.mu.RUnlock()
in := func(t time.Time) bool { return t.After(after) && !t.After(before) }
out := make([]*plant.Plant, 0)
for _, p := range r.items {
if in(p.NextWateringAt) || (p.NextFertilizingAt != nil && in(*p.NextFertilizingAt)) {
out = append(out, p)
}
}
return out, nil
}