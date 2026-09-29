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
	ID        string    `json:"id" bson:"_id"`
	UserID    string    `json:"userId" bson:"userId"`
	SessionID string    `json:"sessionId" bson:"sessionId"`
	Role      Role      `json:"role" bson:"role"`
	Content   string    `json:"content" bson:"content"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
	Provider  string    `json:"provider,omitempty" bson:"provider,omitempty"`
}
