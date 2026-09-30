package v1

import (
"encoding/json"
"fmt"
"net/http"

"plantpal-backend/internal/domain/apperr"
"plantpal-backend/internal/interface/http/authmw"
"plantpal-backend/internal/interface/http/respond"
notificationuc "plantpal-backend/internal/usecase/notification"
)

type DeviceHandler struct {
svc *notificationuc.Service
}

func NewDeviceHandler(svc *notificationuc.Service) *DeviceHandler {
return &DeviceHandler{svc: svc}
}

type deviceRequest struct {
Token    string `json:"token"`
Platform string `json:"platform"`
}

func decodeDeviceRequest(w http.ResponseWriter, r *http.Request) (deviceRequest, error) {
var req deviceRequest
r.Body = http.MaxBytesReader(w, r.Body, 8<<10)
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
return req, fmt.Errorf("%w: invalid JSON body", apperr.ErrInvalidInput)
}
return req, nil
}

// Register stores the caller's FCM token (POST /devices).
func (h *DeviceHandler) Register(w http.ResponseWriter, r *http.Request) {
req, err := decodeDeviceRequest(w, r)
if err != nil {
respond.Error(w, err)
return
}
userID, _ := authmw.UserID(r.Context())
if err := h.svc.RegisterDevice(r.Context(), userID, req.Token, req.Platform); err != nil {
respond.Error(w, err)
return
}
respond.JSON(w, http.StatusCreated, map[string]bool{"registered": true})
}

// Unregister forgets the caller's FCM token, e.g. on logout (DELETE /devices).
func (h *DeviceHandler) Unregister(w http.ResponseWriter, r *http.Request) {
req, err := decodeDeviceRequest(w, r)
if err != nil {
respond.Error(w, err)
return
}
userID, _ := authmw.UserID(r.Context())
if err := h.svc.UnregisterDevice(r.Context(), userID, req.Token); err != nil {
respond.Error(w, err)
return
}
w.WriteHeader(http.StatusNoContent)
}