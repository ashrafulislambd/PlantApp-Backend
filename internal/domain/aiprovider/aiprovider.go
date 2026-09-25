// Package aiprovider names the AI backends that can answer a chat or
// diagnosis request, so callers can report which one actually responded.
package aiprovider

// Name identifies which AI backend produced a result.
type Name string

const (
	Gemini Name = "gemini"
	Groq   Name = "groq"
	Mock   Name = "mock"
)
