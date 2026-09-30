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