package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port string

	// GeminiAPIKey is the primary AI Doctor / AI Chat provider.
	GeminiAPIKey string
	GeminiModel  string

	// GroqAPIKey enables AI Doctor / AI Chat replies via Groq
	// (https://console.groq.com), used as a fallback when Gemini is unset
	// or fails, and powers the shop price-refresh feature. When both keys
	// are empty, main.go falls back to the mock providers so the API stays
	// runnable without any key.
	GroqAPIKey      string
	GroqChatModel   string
	GroqVisionModel string

	// MongoDB
	MongoURI    string
	MongoDBName string

	// Auth (JWT)
	JWTSecret     string
	JWTAccessTTL  time.Duration
	JWTRefreshTTL time.Duration
}

func Load() Config {
	// Load local development settings without overriding explicitly exported variables.
	_ = godotenv.Load()

	chatModel := os.Getenv("GROQ_CHAT_MODEL")
	if chatModel == "" {
		chatModel = getenv("GROQ_MODEL", "openai/gpt-oss-120b")
	}

	return Config{
		Port: getenv("PORT", "8080"),

		GeminiAPIKey: os.Getenv("GEMINI_API_KEY"),
		GeminiModel:  getenv("GEMINI_MODEL", "gemini-3.6-flash"),

		GroqAPIKey: os.Getenv("GROQ_API_KEY"),
		// Model defaults as of Sept 2026 (see console.groq.com/docs/deprecations -
		// llama-3.3-70b-versatile and llama-4-scout were both decommissioned).
		// chatModel also honours the legacy GROQ_MODEL variable.
		GroqChatModel:   chatModel,
		GroqVisionModel: getenv("GROQ_VISION_MODEL", "qwen/qwen3.8-27b"),

		MongoURI:    getenv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDBName: getenv("MONGO_DB_NAME", "plantpal"),

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
