package httpapi

import (
	"log"
	"net/http"

	"plantpal-backend/internal/interface/http/respond"
	v1 "plantpal-backend/internal/interface/http/v1"
)

// NewRouter builds the full HTTP handler: unversioned root/health routes
// plus every registered API version, wrapped in shared middleware.
// googleOAuth configures the browser-redirect Google sign-in path (see
// google_oauth_handler.go); leave it zero-valued to disable those two
// routes without affecting the ID-token-only /api/v1/auth/google path.
func NewRouter(v1Deps v1.Dependencies, googleOAuth GoogleOAuthConfig) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", healthCheck)
	mux.HandleFunc("GET /", apiInfo)
	mux.HandleFunc("GET /docs", docsHandler)
	mux.HandleFunc("GET /openapi.yaml", openAPISpecHandler)

	aiH := &aiHandler{chatSvc: v1Deps.ChatService, diagnosisSvc: v1Deps.DiagnosisService}
	mux.HandleFunc("POST /ai/chat", v1Deps.RequireAuth(aiH.Chat))
	mux.HandleFunc("POST /ai/diagnose", v1Deps.RequireAuth(aiH.Diagnose))

	if googleOAuth.enabled() {
		googleH := newGoogleOAuthHandler(v1Deps.AuthService, googleOAuth)
		mux.HandleFunc("GET /auth/google/callback", googleH.Callback)
		mux.HandleFunc("GET /auth/google/session/{redirect}", googleH.Session)
		log.Println("Google sign-in (browser redirect): enabled")
	} else {
		log.Println("Google sign-in (browser redirect): disabled (set GOOGLE_CLIENT_SECRET and GOOGLE_REDIRECT_URI to enable)")
	}

	v1.RegisterRoutes(mux, v1Deps)

	return withMiddleware(mux)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	respond.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func apiInfo(w http.ResponseWriter, r *http.Request) {
	respond.JSON(w, http.StatusOK, map[string]any{
		"name":     "PlantPal API",
		"versions": []string{"v1"},
		"docs":     "/docs",
	})
}
