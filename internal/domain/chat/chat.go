// Package chat holds the domain for the "AI Chat Box" component ("Chat
// with expert" / "What would you like to know?"): a simple per-session
// message history with an AI-generated reply.
package chat

import (
	"encoding/json"
	"strings"
	"time"
	"unicode/utf8"
)

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

	// DiagnosisID links a turn to a plant-photo scan (a saved Diagnosis and
	// its photo). It is set on the user turn that carried the photo and on
	// the assistant turn that summarised the result, and on a user turn that
	// was sent with a scan as context.
	DiagnosisID string `json:"diagnosisId,omitempty" bson:"diagnosisId,omitempty"`
}

// DiagnosisImagePath is where the photo of a diagnosis is served (Bearer
// auth required).
func DiagnosisImagePath(diagnosisID string) string {
	return "/api/v1/diagnoses/" + diagnosisID + "/image"
}

// MarshalJSON adds imageUrl, derived from DiagnosisID, so history can show
// the photo without the client building the URL itself.
func (m Message) MarshalJSON() ([]byte, error) {
	type plain Message
	out := struct {
		plain
		ImageURL string `json:"imageUrl,omitempty"`
	}{plain: plain(m)}
	if m.DiagnosisID != "" {
		out.ImageURL = DiagnosisImagePath(m.DiagnosisID)
	}
	return json.Marshal(out)
}

// Session is a summary of one conversation, as shown in the chat history
// menu. It is derived from the stored messages; there is no separate
// session record.
type Session struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	LastMessageAt time.Time `json:"lastMessageAt"`
	MessageCount  int       `json:"messageCount"`
}

// MaxTitleRunes is the longest session title, in characters.
const MaxTitleRunes = 40

// TitleFromContent turns the first message of a conversation into a short
// single-line title: whitespace collapsed, cut at MaxTitleRunes with an
// ellipsis. Counting runes (not bytes) keeps Bengali text intact.
func TitleFromContent(content string) string {
	t := strings.Join(strings.Fields(content), " ")
	if utf8.RuneCountInString(t) <= MaxTitleRunes {
		return t
	}
	r := []rune(t)
	return strings.TrimSpace(string(r[:MaxTitleRunes])) + "\u2026"
}

// PhotoTitle is the session title when a conversation opens with a photo and
// no caption.
const PhotoTitle = "\U0001F4F7"

// SessionTitle is the title for a session whose first message has this
// content and diagnosisID.
func SessionTitle(content, diagnosisID string) string {
	if t := TitleFromContent(content); t != "" {
		return t
	}
	if diagnosisID != "" {
		return PhotoTitle
	}
	return ""
}
