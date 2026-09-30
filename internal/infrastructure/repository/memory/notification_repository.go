package memory

import (
"context"
"sort"
"sync"
"time"

"plantpal-backend/internal/domain/notification"
)

type NotificationRepository struct {
mu      sync.Mutex
devices map[string]*notification.Device
sent    map[string]struct{}
}

func NewNotificationRepository() *NotificationRepository {
return &NotificationRepository{
devices: make(map[string]*notification.Device),
sent:    make(map[string]struct{}),
}
}

func (r *NotificationRepository) Upsert(_ context.Context, d *notification.Device) error {
r.mu.Lock()
defer r.mu.Unlock()
cp := *d
if existing, ok := r.devices[d.Token]; ok {
cp.CreatedAt = existing.CreatedAt
}
cp.UpdatedAt = time.Now().UTC()
r.devices[cp.Token] = &cp
return nil
}

func (r *NotificationRepository) ListByUser(_ context.Context, userID string) ([]*notification.Device, error) {
r.mu.Lock()
defer r.mu.Unlock()
out := make([]*notification.Device, 0)
for _, d := range r.devices {
if d.UserID == userID {
cp := *d
out = append(out, &cp)
}
}
sort.Slice(out, func(i, j int) bool { return out[i].Token < out[j].Token })
return out, nil
}

func (r *NotificationRepository) DeleteToken(_ context.Context, userID, token string) error {
r.mu.Lock()
defer r.mu.Unlock()
if d, ok := r.devices[token]; ok && d.UserID == userID {
delete(r.devices, token)
}
return nil
}

func (r *NotificationRepository) MarkSent(_ context.Context, key string) (bool, error) {
r.mu.Lock()
defer r.mu.Unlock()
if _, ok := r.sent[key]; ok {
return false, nil
}
r.sent[key] = struct{}{}
return true, nil
}

func (r *NotificationRepository) ReleaseSent(_ context.Context, key string) error {
r.mu.Lock()
defer r.mu.Unlock()
delete(r.sent, key)
return nil
}