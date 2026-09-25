// Package config loads runtime configuration from the environment.
package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port string

	GeminiAPIKey string
	GeminiModel  string

	// GroqAPIKey enables real AI Doctor / AI Chat replies via Groq
	// (https://console.groq.com), used as a fallback when Gemini is unset
	// or fails. When both are empty, main.go falls back to the mock
	// providers so the API stays runnable without any key.
	GroqAPIKey      string
	GroqChatModel   string
	GroqVisionModel string

	MongoURI    string
	MongoDBName string

	JWTSecret     string
	JWTAccessTTL  time.Duration
	JWTRefreshTTL time.Duration
}

func Load() Config {
	// Load local development settings without overriding explicitly exported variables.
	_ = godotenv.Load()

	return Config{
		Port: getenv("PORT", "8081"),

		GeminiAPIKey: os.Getenv("GEMINI_API_KEY"),
		GeminiModel:  getenv("GEMINI_MODEL", "gemini-3.6-flash"),

		GroqAPIKey: os.Getenv("GROQ_API_KEY"),
		// Model defaults as of Sept 2026 (see console.groq.com/docs/deprecations —
		// llama-3.3-70b-versatile and llama-4-scout were both decommissioned).
		GroqChatModel:   getenv("GROQ_CHAT_MODEL", "openai/gpt-oss-120b"),
		GroqVisionModel: getenv("GROQ_VISION_MODEL", "qwen/qwen3.8-27b"),

		MongoURI:    getenv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDBName: getenv("MONGO_DB_NAME", "myplantpal"),

		JWTSecret:     os.Getenv("JWT_SECRET"),
		JWTAccessTTL:  getDuration("JWT_ACCESS_TTL", 15*time.Minute),
		JWTRefreshTTL: getDuration("JWT_REFRESH_TTL", 30*24*time.Hour),
	}
}

func getDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}