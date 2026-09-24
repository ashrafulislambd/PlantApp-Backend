// Package config loads runtime configuration from the environment.
package config

import "os"

type Config struct {
	Port string

	// GroqAPIKey enables real AI Doctor / AI Chat replies via Groq
	// (https://console.groq.com). When empty, main.go falls back to the
	// mock providers so the API stays runnable without a key.
	GroqAPIKey      string
	GroqChatModel   string
	GroqVisionModel string
}

func Load() Config {
	loadDotenv(".env")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Model defaults as of Sept 2026 (see console.groq.com/docs/deprecations —
	// llama-3.3-70b-versatile and llama-4-scout were both decommissioned).
	chatModel := os.Getenv("GROQ_CHAT_MODEL")
	if chatModel == "" {
		chatModel = "openai/gpt-oss-120b"
	}

	visionModel := os.Getenv("GROQ_VISION_MODEL")
	if visionModel == "" {
		visionModel = "qwen/qwen3.8-27b"
	}

	return Config{
		Port:            port,
		GroqAPIKey:      os.Getenv("GROQ_API_KEY"),
		GroqChatModel:   chatModel,
		GroqVisionModel: visionModel,
	}
}
