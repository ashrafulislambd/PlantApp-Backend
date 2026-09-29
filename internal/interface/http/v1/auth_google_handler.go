package v1

import (
	"encoding/json"
	"fmt"
	"net/http"

	"plantpal-backend/internal/domain/apperr"
	"plantpal-backend/internal/interface/http/respond"
)

type googleLoginRequest struct {
	IDToken string `json:"idToken"`
}

// Google handles POST /auth/google: the app sends the ID token it got from
// Google Sign-In and receives the same payload as /auth/login.
func (h *AuthHandler) Google(w http.ResponseWriter, r *http.Request) {
	var req googleLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, fmt.Errorf("%w: invalid JSON body", apperr.ErrInvalidInput))
		return
	}
	res, err := h.svc.LoginWithGoogle(r.Context(), req.IDToken)
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, toAuthResponse(res))
}

