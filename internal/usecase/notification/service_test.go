package notification

import (
"context"
"errors"
"sync"
"testing"
"time"

"plantpal-backend/internal/domain/apperr"
"plantpal-backend/internal/domain/notification"
"plantpal-backend/internal/domain/plant"
"plantpal-backend/internal/infrastructure/repository/memory"
)

type fakeSender struct {
mu   sync.Mutex
sent []string
err  error
}

func (f *fakeSender) Send(_ context.Context, token string, _ notification.Message) error {
f.mu.Lock()
defer f.mu.Unlock()
if f.err != nil {
return f.err
}
f.sent = append(f.sent, token)
return nil
}

func setup(t *testing.T, sender notification.Sender, nextWatering time.Time) (*Service, *memory.NotificationRepository) {
t.Helper()
ctx := context.Background()
now := time.Date(2026, 9, 30, 9, 0, 0, 0, time.UTC)

plants := memory.NewPlantRepository()
if err := plants.Create(ctx, &plant.Plant{ID: "pl_1", UserID: "u1", Name: "Rose", NextWateringAt: nextWatering}); err != nil {
t.Fatalf("create plant: %v", err)
}
repo := memory.NewNotificationRepository()
svc := NewService(repo, sender, plants)
svc.now = func() time.Time { return now }
if err := svc.RegisterDevice(ctx, "u1", "tok_1", "android"); err != nil {
t.Fatalf("RegisterDevice: %v", err)
}
return svc, repo
}

var testNow = time.Date(2026, 9, 30, 9, 0, 0, 0, time.UTC)

func TestSendDueReminders_SendsOncePerDueTime(t *testing.T) {
sender := &fakeSender{}
svc, _ := setup(t, sender, testNow.Add(-time.Hour))

for i := 0; i < 2; i++ {
if _, err := svc.SendDueReminders(context.Background()); err != nil {
t.Fatalf("SendDueReminders: %v", err)
}
}
if len(sender.sent) != 1 {
t.Errorf("sent %d pushes, want exactly 1", len(sender.sent))
}
}

func TestSendDueReminders_SkipsStaleReminders(t *testing.T) {
sender := &fakeSender{}
svc, _ := setup(t, sender, testNow.Add(-48*time.Hour))

if _, err := svc.SendDueReminders(context.Background()); err != nil {
t.Fatalf("SendDueReminders: %v", err)
}
if len(sender.sent) != 0 {
t.Errorf("sent %d pushes for a stale reminder, want 0", len(sender.sent))
}
}

func TestSendDueReminders_DropsInvalidTokens(t *testing.T) {
sender := &fakeSender{err: notification.ErrInvalidToken}
svc, repo := setup(t, sender, testNow.Add(-time.Hour))

if _, err := svc.SendDueReminders(context.Background()); err != nil {
t.Fatalf("SendDueReminders: %v", err)
}
devices, _ := repo.ListByUser(context.Background(), "u1")
if len(devices) != 0 {
t.Errorf("dead token was kept: %+v", devices)
}
}

func TestSendDueReminders_RetriesAfterTransientFailure(t *testing.T) {
sender := &fakeSender{err: errors.New("fcm down")}
svc, _ := setup(t, sender, testNow.Add(-time.Hour))

if _, err := svc.SendDueReminders(context.Background()); err != nil {
t.Fatalf("SendDueReminders: %v", err)
}
sender.err = nil
if _, err := svc.SendDueReminders(context.Background()); err != nil {
t.Fatalf("SendDueReminders: %v", err)
}
if len(sender.sent) != 1 {
t.Errorf("sent %d pushes after recovery, want 1", len(sender.sent))
}
}

func TestRegisterDevice_RejectsBadInput(t *testing.T) {
svc := NewService(memory.NewNotificationRepository(), &fakeSender{}, memory.NewPlantRepository())
ctx := context.Background()
if err := svc.RegisterDevice(ctx, "u1", "tok", "symbian"); !errors.Is(err, apperr.ErrInvalidInput) {
t.Errorf("bad platform: err = %v, want ErrInvalidInput", err)
}
if err := svc.RegisterDevice(ctx, "u1", "  ", "android"); !errors.Is(err, apperr.ErrInvalidInput) {
t.Errorf("empty token: err = %v, want ErrInvalidInput", err)
}
}

func intp(n int) *int { return &n }

func TestSendDueReminders_HeldDuringDevicesLocalNight(t *testing.T) {
sender := &fakeSender{}
svc, repo := setup(t, sender, testNow.Add(-time.Hour))
ctx := context.Background()
// Replace the default device by one at UTC-10: 09:00 UTC is 23:00 there.
if err := repo.DeleteToken(ctx, "u1", "tok_1"); err != nil {
t.Fatalf("DeleteToken: %v", err)
}
if err := svc.RegisterDeviceInZone(ctx, "u1", "tok_night", "android", "", intp(-600)); err != nil {
t.Fatalf("RegisterDeviceInZone: %v", err)
}

n, err := svc.SendDueReminders(ctx)
if err != nil || n != 0 || len(sender.sent) != 0 {
t.Fatalf("at local 23:00 delivered=%d sent=%v err=%v, want nothing sent", n, sender.sent, err)
}

// Eight hours later it is 07:00 local and the held reminder goes out once.
svc.now = func() time.Time { return testNow.Add(8 * time.Hour) }
for i := 0; i < 2; i++ {
if _, err := svc.SendDueReminders(ctx); err != nil {
t.Fatalf("SendDueReminders: %v", err)
}
}
if len(sender.sent) != 1 || sender.sent[0] != "tok_night" {
t.Errorf("sent = %v, want exactly one push to tok_night after quiet hours", sender.sent)
}
}

func TestSendDueReminders_OnlyAwakeDevicesGetThePush(t *testing.T) {
sender := &fakeSender{}
svc, _ := setup(t, sender, testNow.Add(-time.Hour))
ctx := context.Background()
// tok_1 has no zone (UTC, 09:00 = awake); add one in its local night.
if err := svc.RegisterDeviceInZone(ctx, "u1", "tok_asleep", "ios", "Pacific/Honolulu", nil); err != nil {
t.Fatalf("RegisterDeviceInZone: %v", err)
}
if _, err := svc.SendDueReminders(ctx); err != nil {
t.Fatalf("SendDueReminders: %v", err)
}
if len(sender.sent) != 1 || sender.sent[0] != "tok_1" {
t.Errorf("sent = %v, want only tok_1", sender.sent)
}
}

func TestRegisterDeviceInZone_IgnoresInvalidZone(t *testing.T) {
svc, repo := setup(t, &fakeSender{}, testNow.Add(-time.Hour))
ctx := context.Background()
if err := svc.RegisterDeviceInZone(ctx, "u1", "tok_bad", "android", "Mars/Olympus", intp(9999)); err != nil {
t.Fatalf("RegisterDeviceInZone: %v", err)
}
devices, _ := repo.ListByUser(ctx, "u1")
for _, d := range devices {
if d.Token == "tok_bad" && (d.Timezone != "" || d.UTCOffsetMinutes != nil) {
t.Errorf("device = %+v, want invalid zone and offset dropped", d)
}
}
}
