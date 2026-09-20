package httpapi

import (
	"net/http"

	"myplantpal-backend/internal/interface/http/respond"
	v1 "myplantpal-backend/internal/interface/http/v1"
)

// NewRouter builds the full HTTP handler: unversioned root/health routes
// plus every registered API version, wrapped in shared middleware.
func NewRouter(v1Deps v1.Dependencies) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", healthCheck)
	mux.HandleFunc("GET /", apiInfo)
	mux.HandleFunc("GET /docs", docsHandler)
	mux.HandleFunc("GET /openapi.yaml", openAPISpecHandler)

	v1.RegisterRoutes(mux, v1Deps)

	return withMiddleware(mux)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	respond.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func apiInfo(w http.ResponseWriter, r *http.Request) {
	respond.JSON(w, http.StatusOK, map[string]any{
		"name":     "MyPlantPal API",
		"versions": []string{"v1"},
		"docs":     "/docs",
	})
}
