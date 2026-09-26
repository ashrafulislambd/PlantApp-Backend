package v1

import (
	"encoding/json"
	"fmt"
	"net/http"

	"myplantpal-backend/internal/domain/apperr"
	"myplantpal-backend/internal/domain/user"
	"myplantpal-backend/internal/interface/http/authmw"
	"myplantpal-backend/internal/interface/http/respond"
	authuc "myplantpal-backend/internal/usecase/auth"
)

type AuthHandler struct {
	svc *authuc.Service
}

func NewAuthHandler(svc *authuc.Service) *AuthHandler {
	return &AuthHandler{svc: svc}
}

type authResponse struct {
	User         *user.User `json:"user,omitempty"`
	AccessToken  string     `json:"accessToken"`
	RefreshToken string     `json:"refreshToken"`
	ExpiresIn    int64      `json:"expiresIn"`
}

func toAuthResponse(r *authuc.AuthResult) authResponse {
	return authResponse{User: r.User, AccessToken: r.AccessToken, RefreshToken: r.RefreshToken, ExpiresIn: r.ExpiresIn}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, fmt.Errorf("%w: invalid JSON body", apperr.ErrInvalidInput))
		return
	}
	res, err := h.svc.Register(r.Context(), authuc.RegisterInput{Email: req.Email, Password: req.Password, Name: req.Name})
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.JSON(w, http.StatusCreated, toAuthResponse(res))
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, fmt.Errorf("%w: invalid JSON body", apperr.ErrInvalidInput))
		return
	}
	res, err := h.svc.Login(r.Context(), authuc.LoginInput{Email: req.Email, Password: req.Password})
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, toAuthResponse(res))
}

type refreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RefreshToken == "" {
		respond.Error(w, fmt.Errorf("%w: refreshToken is required", apperr.ErrInvalidInput))
		return
	}
	res, err := h.svc.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, toAuthResponse(res))
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RefreshToken == "" {
		respond.Error(w, fmt.Errorf("%w: refreshToken is required", apperr.ErrInvalidInput))
		return
	}
	if err := h.svc.Logout(r.Context(), req.RefreshToken); err != nil {
		respond.Error(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, _ := authmw.UserID(r.Context())
	u, err := h.svc.Me(r.Context(), userID)
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, u)
}
