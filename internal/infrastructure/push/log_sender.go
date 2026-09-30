// Package push holds Sender implementations that don't talk to a real push
// service.
package push

import (
"context"
"log"

"plantpal-backend/internal/domain/notification"
)

// LogSender is the dev fallback used when FCM credentials aren't configured:
// it logs the push instead of sending it, so the whole flow stays runnable.
type LogSender struct{}

func (LogSender) Send(_ context.Context, token string, msg notification.Message) error {
short := token
if len(short) > 8 {
short = short[:8]
}
log.Printf("push (log-only) -> %s...: %s - %s", short, msg.Title, msg.Body)
return nil
}