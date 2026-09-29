package v1

import (
	"net/http"

	"plantpal-backend/internal/interface/http/respond"
	authuc "plantpal-backend/internal/usecase/auth"
	chatuc "plantpal-backend/internal/usecase/chat"
	diagnosisuc "plantpal-backend/internal/usecase/diagnosis"
	fertilizeruc "plantpal-backend/internal/usecase/fertilizer"
	plantuc "plantpal-backend/internal/usecase/plant"
	productuc "plantpal-backend/internal/usecase/product"
)

// Dependencies are the usecases this API version's handlers call into.
type Dependencies struct {
	PlantService      *plantuc.Service
	FertilizerService *fertilizeruc.Service
	DiagnosisService  *diagnosisuc.Service
	ChatService       *chatuc.Service
	ProductService    *productuc.Service
	AuthService       *authuc.Service
	RequireAuth       func(http.HandlerFunc) http.HandlerFunc
}

const basePath = "/api/v1"

// RegisterRoutes mounts every v1 endpoint on mux.
func RegisterRoutes(mux *http.ServeMux, deps Dependencies) {
	mux.HandleFunc("GET "+basePath+"/health", HealthCheck)

	plantH := NewPlantHandler(deps.PlantService)
	mux.HandleFunc("POST "+basePath+"/plants", deps.RequireAuth(plantH.Create))
	mux.HandleFunc("GET "+basePath+"/plants", deps.RequireAuth(plantH.List))
	mux.HandleFunc("GET "+basePath+"/plants/due", deps.RequireAuth(plantH.Due))
	mux.HandleFunc("GET "+basePath+"/plants/{id}", deps.RequireAuth(plantH.Get))
	mux.HandleFunc("PATCH "+basePath+"/plants/{id}", deps.RequireAuth(plantH.Update))
	mux.HandleFunc("DELETE "+basePath+"/plants/{id}", deps.RequireAuth(plantH.Delete))
	mux.HandleFunc("POST "+basePath+"/plants/{id}/water", deps.RequireAuth(plantH.MarkWatered))
	mux.HandleFunc("POST "+basePath+"/plants/{id}/fertilize", deps.RequireAuth(plantH.MarkFertilized))

	authH := NewAuthHandler(deps.AuthService)
	mux.HandleFunc("POST "+basePath+"/auth/register", authH.Register)
	mux.HandleFunc("POST "+basePath+"/auth/login", authH.Login)
	mux.HandleFunc("POST "+basePath+"/auth/refresh", authH.Refresh)
	mux.HandleFunc("POST "+basePath+"/auth/logout", authH.Logout)
	mux.HandleFunc("GET "+basePath+"/auth/me", deps.RequireAuth(authH.Me))

	fertH := NewFertilizerHandler(deps.FertilizerService)
	mux.HandleFunc("POST "+basePath+"/fertilizers", deps.RequireAuth(fertH.Create))
	mux.HandleFunc("GET "+basePath+"/fertilizers", deps.RequireAuth(fertH.List))
	mux.HandleFunc("GET "+basePath+"/fertilizers/{id}", deps.RequireAuth(fertH.Get))

	diagH := NewDiagnosisHandler(deps.DiagnosisService)
	mux.HandleFunc("POST "+basePath+"/diagnoses", deps.RequireAuth(diagH.Create))
	mux.HandleFunc("GET "+basePath+"/diagnoses", deps.RequireAuth(diagH.List))
	mux.HandleFunc("GET "+basePath+"/diagnoses/{id}", deps.RequireAuth(diagH.Get))

	chatH := NewChatHandler(deps.ChatService)
	mux.HandleFunc("POST "+basePath+"/chat/messages", deps.RequireAuth(chatH.Send))
	mux.HandleFunc("GET "+basePath+"/chat/messages", deps.RequireAuth(chatH.List))

	prodH := NewProductHandler(deps.ProductService)
	mux.HandleFunc("GET "+basePath+"/products", deps.RequireAuth(prodH.List))
	mux.HandleFunc("GET "+basePath+"/products/{id}", deps.RequireAuth(prodH.Get))
	mux.HandleFunc("POST "+basePath+"/products/{id}/refresh", deps.RequireAuth(prodH.Refresh))
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	respond.JSON(w, http.StatusOK, map[string]string{"status": "ok", "version": "v1"})
}
