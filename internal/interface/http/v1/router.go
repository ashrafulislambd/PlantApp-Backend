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
	mux.HandleFunc("POST "+basePath+"/plants", plantH.Create)
	mux.HandleFunc("GET "+basePath+"/plants", plantH.List)
	mux.HandleFunc("GET "+basePath+"/plants/due", plantH.Due)
	mux.HandleFunc("GET "+basePath+"/plants/{id}", plantH.Get)
	mux.HandleFunc("PATCH "+basePath+"/plants/{id}", plantH.Update)
	mux.HandleFunc("DELETE "+basePath+"/plants/{id}", plantH.Delete)
	mux.HandleFunc("POST "+basePath+"/plants/{id}/water", plantH.MarkWatered)
	mux.HandleFunc("POST "+basePath+"/plants/{id}/fertilize", plantH.MarkFertilized)


	authH := NewAuthHandler(deps.AuthService)
	mux.HandleFunc("POST "+basePath+"/auth/register", authH.Register)
	mux.HandleFunc("POST "+basePath+"/auth/login", authH.Login)
	mux.HandleFunc("POST "+basePath+"/auth/refresh", authH.Refresh)
	mux.HandleFunc("POST "+basePath+"/auth/logout", authH.Logout)
	mux.HandleFunc("GET "+basePath+"/auth/me", deps.RequireAuth(authH.Me))

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

	prodH := NewProductHandler(deps.ProductService)
	mux.HandleFunc("GET "+basePath+"/products", prodH.List)
	mux.HandleFunc("GET "+basePath+"/products/{id}", prodH.Get)
	mux.HandleFunc("POST "+basePath+"/products/{id}/refresh", prodH.Refresh)
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	respond.JSON(w, http.StatusOK, map[string]string{"status": "ok", "version": "v1"})
}

