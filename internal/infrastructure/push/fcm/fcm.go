// Package fcm sends push notifications through the Firebase Cloud Messaging
// HTTP v1 API. It authenticates with a Google service-account key (signed
// JWT exchanged for an OAuth2 access token), so it needs no extra modules.
package fcm

import (
"bytes"
"context"
"crypto/rsa"
"encoding/json"
"errors"
"fmt"
"io"
"net/http"
"net/url"
"os"
"strings"
"sync"
"time"

"github.com/golang-jwt/jwt/v5"

"plantpal-backend/internal/domain/notification"
)

const (
scope           = "https://www.googleapis.com/auth/firebase.messaging"
defaultTokenURI = "https://oauth2.googleapis.com/token"
maxRespBytes    = 64 << 10
)

type serviceAccount struct {
ProjectID   string `json:"project_id"`
ClientEmail string `json:"client_email"`
PrivateKey  string `json:"private_key"`
TokenURI    string `json:"token_uri"`
}

type Sender struct {
email    string
tokenURI string
sendURL  string
key      *rsa.PrivateKey
http     *http.Client

mu        sync.Mutex
cached    string
expiresAt time.Time
}

func NewFromFile(path string) (*Sender, error) {
raw, err := os.ReadFile(path)
if err != nil {
return nil, fmt.Errorf("read FCM credentials: %w", err)
}
return NewFromJSON(raw)
}

func NewFromJSON(raw []byte) (*Sender, error) {
var sa serviceAccount
if err := json.Unmarshal(raw, &sa); err != nil {
return nil, fmt.Errorf("parse FCM credentials: %w", err)
}
if sa.ProjectID == "" || sa.ClientEmail == "" || sa.PrivateKey == "" {
return nil, errors.New("FCM credentials must contain project_id, client_email and private_key")
}
key, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(sa.PrivateKey))
if err != nil {
return nil, fmt.Errorf("parse FCM private key: %w", err)
}
if sa.TokenURI == "" {
sa.TokenURI = defaultTokenURI
}
return &Sender{
email:    sa.ClientEmail,
tokenURI: sa.TokenURI,
sendURL:  fmt.Sprintf("https://fcm.googleapis.com/v1/projects/%s/messages:send", sa.ProjectID),
key:      key,
http:     &http.Client{Timeout: 10 * time.Second},
}, nil
}

func (s *Sender) accessToken(ctx context.Context) (string, error) {
s.mu.Lock()
defer s.mu.Unlock()
if s.cached != "" && time.Now().Before(s.expiresAt.Add(-time.Minute)) {
return s.cached, nil
}

now := time.Now()
assertion, err := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
"iss":   s.email,
"scope": scope,
"aud":   s.tokenURI,
"iat":   now.Unix(),
"exp":   now.Add(time.Hour).Unix(),
}).SignedString(s.key)
if err != nil {
return "", fmt.Errorf("sign FCM assertion: %w", err)
}

form := url.Values{
"grant_type": {"urn:ietf:params:oauth:grant-type:jwt-bearer"},
"assertion":  {assertion},
}
req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.tokenURI, strings.NewReader(form.Encode()))
if err != nil {
return "", err
}
req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
resp, err := s.http.Do(req)
if err != nil {
return "", fmt.Errorf("fcm oauth: %w", err)
}
defer resp.Body.Close()
body, _ := io.ReadAll(io.LimitReader(resp.Body, maxRespBytes))
if resp.StatusCode != http.StatusOK {
return "", fmt.Errorf("fcm oauth: status %d: %s", resp.StatusCode, body)
}

var out struct {
AccessToken string `json:"access_token"`
ExpiresIn   int    `json:"expires_in"`
}
if err := json.Unmarshal(body, &out); err != nil || out.AccessToken == "" {
return "", errors.New("fcm oauth: unexpected token response")
}
s.cached = out.AccessToken
s.expiresAt = now.Add(time.Duration(out.ExpiresIn) * time.Second)
return s.cached, nil
}

func (s *Sender) invalidate() {
s.mu.Lock()
s.cached = ""
s.mu.Unlock()
}

// Send delivers msg to one device. A dead token yields notification.ErrInvalidToken.
func (s *Sender) Send(ctx context.Context, token string, msg notification.Message) error {
access, err := s.accessToken(ctx)
if err != nil {
return err
}

m := map[string]any{
"token":        token,
"notification": map[string]string{"title": msg.Title, "body": msg.Body},
"android":      map[string]any{"priority": "HIGH"},
}
if len(msg.Data) > 0 {
m["data"] = msg.Data
}
payload, err := json.Marshal(map[string]any{"message": m})
if err != nil {
return err
}

req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.sendURL, bytes.NewReader(payload))
if err != nil {
return err
}
req.Header.Set("Authorization", "Bearer "+access)
req.Header.Set("Content-Type", "application/json")
resp, err := s.http.Do(req)
if err != nil {
return fmt.Errorf("fcm send: %w", err)
}
defer resp.Body.Close()
body, _ := io.ReadAll(io.LimitReader(resp.Body, maxRespBytes))

if resp.StatusCode == http.StatusOK {
return nil
}
if resp.StatusCode == http.StatusNotFound ||
(resp.StatusCode == http.StatusBadRequest && bytes.Contains(body, []byte("INVALID_ARGUMENT"))) {
return fmt.Errorf("%w: %s", notification.ErrInvalidToken, bytes.TrimSpace(body))
}
if resp.StatusCode == http.StatusUnauthorized {
s.invalidate()
}
return fmt.Errorf("fcm send: status %d: %s", resp.StatusCode, bytes.TrimSpace(body))
}