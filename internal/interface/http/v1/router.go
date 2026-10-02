// Package v1 is the first version of the HTTP API surface, mounted under
// /api/v1. A future breaking change gets its own v2 package mounted
// alongside it, so both can be served at once during a migration.
package v1

import (
"net/http"

"plantpal-backend/internal/interface/http/respond"
authuc "plantpal-backend/internal/usecase/auth"
chatuc "plantpal-backend/internal/usecase/chat"
diagnosisuc "plantpal-backend/internal/usecase/diagnosis"
fertilizeruc "plantpal-backend/internal/usecase/fertilizer"
notificationuc "plantpal-backend/internal/usecase/notification"
orderuc "plantpal-backend/internal/usecase/order"
plantuc "plantpal-backend/internal/usecase/plant"
productuc "plantpal-backend/internal/usecase/product"
)

// Dependencies are the usecases this API version's handlers call into.
type Dependencies struct {
PlantService        *plantuc.Service
FertilizerService   *fertilizeruc.Service
DiagnosisService    *diagnosisuc.Service
ChatService         *chatuc.Service
ProductService      *productuc.Service
OrderService        *orderuc.Service
AuthService         *authuc.Service
NotificationService *notificationuc.Service
// MaxImageBytes caps diagnosis photo uploads; 0 uses the 8 MB default.
MaxImageBytes int64
RequireAuth   func(http.HandlerFunc) http.HandlerFunc
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
mux.HandleFunc("POST "+basePath+"/auth/google", authH.Google)
mux.HandleFunc("POST "+basePath+"/auth/refresh", authH.Refresh)
mux.HandleFunc("POST "+basePath+"/auth/logout", authH.Logout)
mux.HandleFunc("GET "+basePath+"/auth/me", deps.RequireAuth(authH.Me))

fertH := NewFertilizerHandler(deps.FertilizerService)
mux.HandleFunc("POST "+basePath+"/fertilizers", deps.RequireAuth(fertH.Create))
mux.HandleFunc("GET "+basePath+"/fertilizers", deps.RequireAuth(fertH.List))
mux.HandleFunc("GET "+basePath+"/fertilizers/{id}", deps.RequireAuth(fertH.Get))

diagH := NewDiagnosisHandler(deps.DiagnosisService, deps.MaxImageBytes)
mux.HandleFunc("POST "+basePath+"/diagnoses", deps.RequireAuth(diagH.Create))
mux.HandleFunc("GET "+basePath+"/diagnoses", deps.RequireAuth(diagH.List))
mux.HandleFunc("GET "+basePath+"/diagnoses/{id}", deps.RequireAuth(diagH.Get))
mux.HandleFunc("GET "+basePath+"/diagnoses/{id}/image", deps.RequireAuth(diagH.Image))

deviceH := NewDeviceHandler(deps.NotificationService)
mux.HandleFunc("POST "+basePath+"/devices", deps.RequireAuth(deviceH.Register))
mux.HandleFunc("DELETE "+basePath+"/devices", deps.RequireAuth(deviceH.Unregister))

chatH := NewChatHandler(deps.ChatService)
mux.HandleFunc("POST "+basePath+"/chat/messages", deps.RequireAuth(chatH.Send))
mux.HandleFunc("GET "+basePath+"/chat/messages", deps.RequireAuth(chatH.List))
mux.HandleFunc("POST "+basePath+"/chat/messages/stream", deps.RequireAuth(chatH.Stream))
mux.HandleFunc("GET "+basePath+"/chat/sessions", deps.RequireAuth(chatH.Sessions))
mux.HandleFunc("DELETE "+basePath+"/chat/sessions/{id}", deps.RequireAuth(chatH.DeleteSession))

prodH := NewProductHandler(deps.ProductService)
mux.HandleFunc("GET "+basePath+"/products", deps.RequireAuth(prodH.List))
mux.HandleFunc("GET "+basePath+"/products/{id}", deps.RequireAuth(prodH.Get))
mux.HandleFunc("POST "+basePath+"/products/{id}/refresh", deps.RequireAuth(prodH.Refresh))

orderH := NewOrderHandler(deps.OrderService)
mux.HandleFunc("POST "+basePath+"/orders", deps.RequireAuth(orderH.Create))
mux.HandleFunc("GET "+basePath+"/orders", deps.RequireAuth(orderH.List))
mux.HandleFunc("GET "+basePath+"/orders/{id}", deps.RequireAuth(orderH.Get))
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
respond.JSON(w, http.StatusOK, map[string]string{"status": "ok", "version": "v1"})
}