package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"plantpal-backend/internal/domain/user"
	authuc "plantpal-backend/internal/usecase/auth"
)

// GoogleOAuthConfig is the server-side half of the browser-based
// authorization-code flow that
// lib/features/auth/data/datasources/auth_remote_data_source.dart's
// googleLogin() starts. Distinct from the ID-token-only POST
// /api/v1/auth/google path (v1/auth_google_handler.go), which a native
// google_sign_in integration could use instead without any of this.
type GoogleOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

func (c GoogleOAuthConfig) enabled() bool {
	return c.ClientID != "" && c.ClientSecret != "" && c.RedirectURI != ""
}

// appCallbackURI is the Flutter app's custom-scheme redirect target
// (AppConfig.appCallbackUri). Fixed, because flutter_web_auth_2's
// authenticate() only needs ANY URI matching its callbackUrlScheme to
// resolve - the actual result is fetched separately via Session, keyed by
// the `state` value the client generated and sent to Google as part of
// the authorize request.
const appCallbackURI = "plantpal://auth"

type googleOAuthHandler struct {
	auth     *authuc.Service
	cfg      GoogleOAuthConfig
	client   *http.Client
	sessions *googleSessionStore
}

func newGoogleOAuthHandler(auth *authuc.Service, cfg GoogleOAuthConfig) *googleOAuthHandler {
	return &googleOAuthHandler{
		auth:     auth,
		cfg:      cfg,
		client:   &http.Client{Timeout: 10 * time.Second},
		sessions: newGoogleSessionStore(),
	}
}

// Callback handles GET /auth/google/callback: Google lands the user's
// browser here after consent, with either `code`+`state` or `error`+
// `state`. It exchanges the code for an ID token, logs the user in exactly
// like POST /api/v1/auth/google, stashes the result under `state` for
// Session to pick up, then bounces the browser to the app's custom scheme
// so the client's authenticate() call resolves.
func (h *googleOAuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	if state == "" {
		http.Error(w, "missing state", http.StatusBadRequest)
		return
	}
	if errParam := r.URL.Query().Get("error"); errParam != "" {
		h.sessions.put(state, googleSessionResult{err: fmt.Errorf("google: %s", errParam)})
		http.Redirect(w, r, appCallbackURI, http.StatusFound)
		return
	}
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}

	idToken, err := h.exchangeCode(r.Context(), code)
	if err != nil {
		log.Printf("google oauth: token exchange failed: %v", err)
		h.sessions.put(state, googleSessionResult{err: err})
		http.Redirect(w, r, appCallbackURI, http.StatusFound)
		return
	}

	res, err := h.auth.LoginWithGoogle(r.Context(), idToken)
	if err != nil {
		log.Printf("google oauth: login failed: %v", err)
		h.sessions.put(state, googleSessionResult{err: err})
		http.Redirect(w, r, appCallbackURI, http.StatusFound)
		return
	}
	h.sessions.put(state, googleSessionResult{result: res})
	http.Redirect(w, r, appCallbackURI, http.StatusFound)
}

// sessionResponse mirrors v1's unexported authResponse (user_handler.go) -
// {accessToken, refreshToken, user, expiresIn} - so the client can parse
// this exactly like POST /api/v1/auth/login's response, rather than
// decoding the JWT for profile info (it only carries the subject/user ID,
// see security.JWTIssuer.NewAccessToken - no email or name claims).
type sessionResponse struct {
	User         *user.User `json:"user,omitempty"`
	AccessToken  string     `json:"accessToken"`
	RefreshToken string     `json:"refreshToken"`
	ExpiresIn    int64      `json:"expiresIn"`
}

// Session handles GET /auth/google/session/{redirect}: the client calls
// this immediately after authenticate() returns, passing back the same
// value it sent to Google as `state`. Single-use - the result is dropped
// after the first read.
func (h *googleOAuthHandler) Session(w http.ResponseWriter, r *http.Request) {
	result, ok := h.sessions.take(r.PathValue("redirect"))
	if !ok {
		http.Error(w, "unknown or expired session", http.StatusNotFound)
		return
	}
	if result.err != nil {
		http.Error(w, result.err.Error(), http.StatusUnauthorized)
		return
	}
	// Flat shape (no {"data": ...} envelope) to match
	// auth_remote_data_source.dart's googleLogin(), which reads the body
	// directly rather than unwrapping res.data['data'].
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(sessionResponse{
		User:         result.result.User,
		AccessToken:  result.result.AccessToken,
		RefreshToken: result.result.RefreshToken,
		ExpiresIn:    result.result.ExpiresIn,
	})
}

func (h *googleOAuthHandler) exchangeCode(ctx context.Context, code string) (string, error) {
	form := url.Values{
		"code":          {code},
		"client_id":     {h.cfg.ClientID},
		"client_secret": {h.cfg.ClientSecret},
		"redirect_uri":  {h.cfg.RedirectURI},
		"grant_type":    {"authorization_code"},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://oauth2.googleapis.com/token", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := h.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("exchange request: %w", err)
	}
	defer resp.Body.Close()
	var body struct {
		IDToken          string `json:"id_token"`
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", fmt.Errorf("decode token response: %w", err)
	}
	if resp.StatusCode != http.StatusOK || body.IDToken == "" {
		return "", fmt.Errorf("google token endpoint: %s: %s", body.Error, body.ErrorDescription)
	}
	return body.IDToken, nil
}

// googleSessionResult is what Callback hands off to Session: either a full
// login result, or the error that should surface to the client instead.
type googleSessionResult struct {
	result *authuc.AuthResult
	err    error
}

type googleSessionEntry struct {
	result    googleSessionResult
	expiresAt time.Time
}

// googleSessionStore briefly holds a Callback's result until Session picks
// it up - normally seconds, since the client fetches it right after the
// browser redirect lands. In-memory only: fine for the single backend
// instance this deployment runs; would need a shared store (e.g. Mongo or
// Redis) behind a load balancer with more than one instance.
type googleSessionStore struct {
	mu      sync.Mutex
	entries map[string]googleSessionEntry
}

func newGoogleSessionStore() *googleSessionStore {
	return &googleSessionStore{entries: map[string]googleSessionEntry{}}
}

func (s *googleSessionStore) put(key string, result googleSessionResult) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.gc()
	s.entries[key] = googleSessionEntry{result: result, expiresAt: time.Now().Add(5 * time.Minute)}
}

func (s *googleSessionStore) take(key string) (googleSessionResult, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.entries[key]
	delete(s.entries, key)
	if !ok || time.Now().After(e.expiresAt) {
		return googleSessionResult{}, false
	}
	return e.result, true
}

// gc drops stale entries (e.g. an abandoned login the client never polled
// for) so the map doesn't grow unbounded. Called while already holding mu.
func (s *googleSessionStore) gc() {
	now := time.Now()
	for k, e := range s.entries {
		if now.After(e.expiresAt) {
			delete(s.entries, k)
		}
	}
}
