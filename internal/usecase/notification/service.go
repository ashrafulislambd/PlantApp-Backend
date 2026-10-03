package notification

import (
"context"
"errors"
"fmt"
"log"
"strings"
"time"

"plantpal-backend/internal/domain/apperr"
"plantpal-backend/internal/domain/notification"
"plantpal-backend/internal/domain/plant"
)

const (
maxTokenLen = 4096
// Reminders that became due longer ago than this are skipped, so a long
// outage (or the first deploy) doesn't blast stale pushes.
lookback = 24 * time.Hour

// Quiet hours in the device owner's local time: no reminder pushes from
// 22:00 until 07:00. A reminder held back this way is sent on the first tick
// after quiet hours end, as long as it is still inside the lookback window.
quietFromHour = 22
quietToHour   = 7
)

var platforms = map[string]bool{"android": true, "ios": true, "web": true}

// DueSource lists plants of all users with a reminder due in (after, before].
type DueSource interface {
ListDueBetween(ctx context.Context, after, before time.Time) ([]*plant.Plant, error)
}

type Service struct {
repo   notification.Repository
sender notification.Sender
plants DueSource
now    func() time.Time
}

func NewService(repo notification.Repository, sender notification.Sender, plants DueSource) *Service {
return &Service{repo: repo, sender: sender, plants: plants, now: time.Now}
}

func (s *Service) RegisterDevice(ctx context.Context, userID, token, platform string) error {
return s.RegisterDeviceInZone(ctx, userID, token, platform, "", nil)
}

// RegisterDeviceInZone is RegisterDevice plus where the device is: an IANA
// time zone name and/or a UTC offset in minutes (both optional, invalid
// values are ignored).
func (s *Service) RegisterDeviceInZone(ctx context.Context, userID, token, platform, zone string, offsetMinutes *int) error {
token = strings.TrimSpace(token)
platform = strings.ToLower(strings.TrimSpace(platform))
if userID == "" {
return fmt.Errorf("%w: userID is required", apperr.ErrInvalidInput)
}
if token == "" || len(token) > maxTokenLen {
return fmt.Errorf("%w: token is required", apperr.ErrInvalidInput)
}
if !platforms[platform] {
return fmt.Errorf("%w: platform must be android, ios or web", apperr.ErrInvalidInput)
}
now := s.now().UTC()
zone = strings.TrimSpace(zone)
if _, err := time.LoadLocation(zone); err != nil || len(zone) > 64 || zone == "Local" {
zone = ""
}
if offsetMinutes != nil && (*offsetMinutes < -14*60 || *offsetMinutes > 14*60) {
offsetMinutes = nil
}
return s.repo.Upsert(ctx, &notification.Device{
Token: token, UserID: userID, Platform: platform, Timezone: zone, UTCOffsetMinutes: offsetMinutes,
CreatedAt: now, UpdatedAt: now,
})
}

func (s *Service) UnregisterDevice(ctx context.Context, userID, token string) error {
token = strings.TrimSpace(token)
if token == "" {
return fmt.Errorf("%w: token is required", apperr.ErrInvalidInput)
}
return s.repo.DeleteToken(ctx, userID, token)
}

// SendDueReminders pushes every watering/fertilizing reminder that just
// became due, once each, and returns how many device deliveries succeeded.
func (s *Service) SendDueReminders(ctx context.Context) (int, error) {
now := s.now().UTC()
plants, err := s.plants.ListDueBetween(ctx, now.Add(-lookback), now)
if err != nil {
return 0, err
}
delivered := 0
for _, p := range plants {
if inWindow(p.NextWateringAt, now) {
delivered += s.notify(ctx, p, "watering", p.NextWateringAt, "Time to water "+p.Name+".")
}
if p.NextFertilizingAt != nil && inWindow(*p.NextFertilizingAt, now) {
delivered += s.notify(ctx, p, "fertilizing", *p.NextFertilizingAt, "Time to fertilize "+p.Name+".")
}
}
return delivered, nil
}

// inQuietHours reports whether the local clock reading of t is within quiet hours.
func inQuietHours(t time.Time) bool {
h := t.Hour()
return h >= quietFromHour || h < quietToHour
}

func awakeDevices(devices []*notification.Device, now time.Time) []*notification.Device {
out := make([]*notification.Device, 0, len(devices))
for _, d := range devices {
if !inQuietHours(now.In(d.Location())) {
out = append(out, d)
}
}
return out
}

func inWindow(t, now time.Time) bool {
return !t.IsZero() && t.After(now.Add(-lookback)) && !t.After(now)
}

func (s *Service) notify(ctx context.Context, p *plant.Plant, kind string, due time.Time, body string) int {
devices, err := s.repo.ListByUser(ctx, p.UserID)
if err != nil {
log.Printf("reminders: list devices for %s: %v", p.UserID, err)
return 0
}
if len(devices) == 0 {
return 0 // nothing to push to; leave unmarked so a later registration can still get it
}
// Only wake devices whose local time is outside quiet hours. When every
// device is in its night the reminder stays unmarked and goes out later.
devices = awakeDevices(devices, s.now())
if len(devices) == 0 {
return 0
}

key := fmt.Sprintf("%s:%s:%d", kind, p.ID, due.Unix())
fresh, err := s.repo.MarkSent(ctx, key)
if err != nil {
log.Printf("reminders: mark sent %s: %v", key, err)
return 0
}
if !fresh {
return 0
}

msg := notification.Message{
Title: "PlantPal",
Body:  body,
Data:  map[string]string{"type": kind, "plantId": p.ID},
}
delivered, transient := 0, false
for _, d := range devices {
err := s.sender.Send(ctx, d.Token, msg)
switch {
case err == nil:
delivered++
case errors.Is(err, notification.ErrInvalidToken):
if delErr := s.repo.DeleteToken(ctx, d.UserID, d.Token); delErr != nil {
log.Printf("reminders: drop dead token: %v", delErr)
}
default:
transient = true
log.Printf("reminders: send failed: %v", err)
}
}
if delivered == 0 && transient {
if err := s.repo.ReleaseSent(ctx, key); err != nil {
log.Printf("reminders: release %s: %v", key, err)
}
}
return delivered
}

// Run checks for due reminders now and then every interval until ctx ends.
func (s *Service) Run(ctx context.Context, interval time.Duration) {
tick := func() {
n, err := s.SendDueReminders(ctx)
if err != nil {
log.Printf("reminders: %v", err)
return
}
if n > 0 {
log.Printf("reminders: delivered %d notification(s)", n)
}
}
tick()
t := time.NewTicker(interval)
defer t.Stop()
for {
select {
case <-ctx.Done():
return
case <-t.C:
tick()
}
}
}