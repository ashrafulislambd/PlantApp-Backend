// Package v1 is the first version of the HTTP API surface, mounted under
// /api/v1. A future breaking change gets its own v2 package mounted
// alongside it, so both can be served at once during a migration.
package v1

import (
	"net/http"

	"myplantpal-backend/internal/interface/http/respond"
	chatuc "myplantpal-backend/internal/usecase/chat"
	diagnosisuc "myplantpal-backend/internal/usecase/diagnosis"
	fertilizeruc "myplantpal-backend/internal/usecase/fertilizer"
	plantuc "myplantpal-backend/internal/usecase/plant"
)

// Dependencies are the usecases this API version's handlers call into.
type Dependencies struct {
	PlantService      *plantuc.Service
	FertilizerService *fertilizeruc.Service
	DiagnosisService  *diagnosisuc.Service
	ChatService       *chatuc.Service
}

const basePath = "/api/v1"

// RegisterRoutes mounts every v1 endpoint on mux.
func RegisterRoutes(mux *http.ServeMux, deps Dependencies) {
	mux.HandleFunc("GET "+basePath+"/health", HealthCheck)

	plantH := NewPlantHandler(deps.PlantService)
	mux.HandleFunc("POST "+basePath+"/plants", plantH.Create)
	mux.HandleFunc("GET "+basePath+"/plants", plantH.List)
	mux.HandleFunc("GET "+basePath+"/plants/{id}", plantH.Get)
	mux.HandleFunc("DELETE "+basePath+"/plants/{id}", plantH.Delete)

	fertH := NewFertilizerHandler(deps.FertilizerService)
	mux.HandleFunc("POST "+basePath+"/fertilizers", fertH.Create)
	mux.HandleFunc("GET "+basePath+"/fertilizers", fertH.List)
	mux.HandleFunc("GET "+basePath+"/fertilizers/{id}", fertH.Get)

	diagH := NewDiagnosisHandler(deps.DiagnosisService)
	mux.HandleFunc("POST "+basePath+"/diagnoses", diagH.Create)
	mux.HandleFunc("GET "+basePath+"/diagnoses", diagH.List)
	mux.HandleFunc("GET "+basePath+"/diagnoses/{id}", diagH.Get)

	chatH := NewChatHandler(deps.ChatService)
	mux.HandleFunc("POST "+basePath+"/chat/messages", chatH.Send)
	mux.HandleFunc("GET "+basePath+"/chat/messages", chatH.List)
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	respond.JSON(w, http.StatusOK, map[string]string{"status": "ok", "version": "v1"})
}
