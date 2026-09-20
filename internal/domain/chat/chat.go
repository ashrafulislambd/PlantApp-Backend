// Package chat holds the domain for the "AI Chat Box" component ("Chat
// with expert" / "What would you like to know?"): a simple per-session
// message history with an AI-generated reply.
package chat

import "time"

type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

// Message is a single turn in a chat session.
type Message struct {
	ID        string    `json:"id"`
	SessionID string    `json:"sessionId"`
	Role      Role      `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}
