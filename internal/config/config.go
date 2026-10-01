package config

import (
"os"
"strconv"
"strings"
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

// GoogleClientIDs are the OAuth client IDs whose Google ID tokens the
// backend accepts (the *Web* client ID the Flutter app uses as its
// serverClientId). Empty disables POST /api/v1/auth/google.
GoogleClientIDs []string

// GoogleClientSecret and GoogleRedirectURI are only needed for the
// browser-based authorization-code flow (GET /auth/google/callback,
// matching lib/features/auth/data/datasources/auth_remote_data_source.dart's
// googleLogin()) - the server-side half of exchanging Google's `code` for
// an ID token. GoogleClientSecret belongs to the SAME Web OAuth client as
// GoogleClientIDs[0] (Google Cloud Console -> that client -> Client
// secret). GoogleRedirectURI must exactly match AppConfig.googleRedirectUri
// on the Flutter side AND the redirect URI registered on that OAuth
// client in the Console - Google rejects the exchange otherwise. Either
// empty disables these two routes (the ID-token-only path above still
// works independently).
GoogleClientSecret string
GoogleRedirectURI  string

// Diagnosis photo uploads.
UploadDir      string
MaxUploadBytes int64

// Push notifications. Provide the Firebase service-account key either as
// a file path or as raw JSON; with neither, pushes are only logged.
FCMCredentialsFile string
FCMCredentialsJSON string
// ReminderInterval is how often the scheduler checks for due reminders.
// 0 disables the scheduler.
ReminderInterval time.Duration
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

GoogleClientIDs:    splitCSV(os.Getenv("GOOGLE_CLIENT_ID")),
GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
GoogleRedirectURI:  os.Getenv("GOOGLE_REDIRECT_URI"),

UploadDir:      getenv("UPLOAD_DIR", "uploads"),
MaxUploadBytes: getInt64("MAX_UPLOAD_BYTES", 8<<20),

FCMCredentialsFile: os.Getenv("FCM_CREDENTIALS_FILE"),
FCMCredentialsJSON: os.Getenv("FCM_CREDENTIALS_JSON"),
ReminderInterval:   getDuration("REMINDER_INTERVAL", time.Minute),
}
}

func splitCSV(s string) []string {
var out []string
for _, p := range strings.Split(s, ",") {
if p = strings.TrimSpace(p); p != "" {
out = append(out, p)
}
}
return out
}

func getDuration(key string, fallback time.Duration) time.Duration {
if v := os.Getenv(key); v != "" {
if d, err := time.ParseDuration(v); err == nil {
return d
}
}
return fallback
}

func getInt64(key string, fallback int64) int64 {
if v := os.Getenv(key); v != "" {
if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
return n
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