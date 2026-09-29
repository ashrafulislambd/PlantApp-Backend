package security

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ErrInvalidGoogleToken means the ID token itself is bad (wrong signature,
// audience, issuer, expired, unverified email...). Any other error from
// Verify is an infrastructure problem (e.g. Google's key endpoint is down).
var ErrInvalidGoogleToken = errors.New("invalid Google ID token")

const googleCertsURL = "https://www.googleapis.com/oauth2/v3/certs"

// GoogleIdentity is the verified subset of a Google ID token we care about.
type GoogleIdentity struct {
	Subject string // Google's stable, unique account ID ("sub")
	Email   string
	Name    string
}

// GoogleVerifier verifies Google Sign-In ID tokens locally: it checks the
// RS256 signature against Google's published public keys (cached), then the
// issuer, audience, expiry and email_verified claims.
type GoogleVerifier struct {
	audiences map[string]struct{}
	certsURL  string
	client    *http.Client
	now       func() time.Time

	mu        sync.Mutex
	keys      map[string]*rsa.PublicKey
	expiresAt time.Time
	fetchedAt time.Time
}

// NewGoogleVerifier accepts tokens issued for any of clientIDs (the OAuth
// *Web* client ID that the mobile app passes as serverClientId).
func NewGoogleVerifier(clientIDs []string) *GoogleVerifier {
	aud := make(map[string]struct{}, len(clientIDs))
	for _, id := range clientIDs {
		if id = strings.TrimSpace(id); id != "" {
			aud[id] = struct{}{}
		}
	}
	return &GoogleVerifier{
		audiences: aud,
		certsURL:  googleCertsURL,
		client:    &http.Client{Timeout: 10 * time.Second},
		now:       time.Now,
		keys:      map[string]*rsa.PublicKey{},
	}
}

type googleClaims struct {
	jwt.RegisteredClaims
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
}

// Verify validates idToken and returns the identity it asserts.
func (v *GoogleVerifier) Verify(ctx context.Context, idToken string) (*GoogleIdentity, error) {
	var fetchErr error
	claims := &googleClaims{}
	_, err := jwt.ParseWithClaims(idToken, claims, func(t *jwt.Token) (any, error) {
		kid, _ := t.Header["kid"].(string)
		key, err := v.key(ctx, kid)
		if err != nil {
			var inv invalidKeyError
			if !errors.As(err, &inv) {
				fetchErr = err
			}
			return nil, err
		}
		return key, nil
	},
		jwt.WithValidMethods([]string{"RS256"}),
		jwt.WithExpirationRequired(),
		jwt.WithLeeway(time.Minute),
		jwt.WithTimeFunc(v.now),
	)
	if fetchErr != nil {
		return nil, fetchErr
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidGoogleToken, err)
	}

	if claims.Issuer != "accounts.google.com" && claims.Issuer != "https://accounts.google.com" {
		return nil, fmt.Errorf("%w: unexpected issuer", ErrInvalidGoogleToken)
	}
	audOK := false
	for _, a := range claims.Audience {
		if _, ok := v.audiences[a]; ok {
			audOK = true
			break
		}
	}
	if !audOK {
		return nil, fmt.Errorf("%w: unexpected audience", ErrInvalidGoogleToken)
	}
	if claims.Subject == "" || claims.Email == "" {
		return nil, fmt.Errorf("%w: missing subject or email", ErrInvalidGoogleToken)
	}
	if !claims.EmailVerified {
		return nil, fmt.Errorf("%w: email is not verified", ErrInvalidGoogleToken)
	}
	return &GoogleIdentity{Subject: claims.Subject, Email: claims.Email, Name: claims.Name}, nil
}

// invalidKeyError marks "this token names a key Google doesn't have" as a
// problem with the token, not with our connection to Google.
type invalidKeyError struct{ msg string }

func (e invalidKeyError) Error() string { return e.msg }

func (v *GoogleVerifier) key(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	if kid == "" {
		return nil, invalidKeyError{"token has no key id"}
	}
	v.mu.Lock()
	defer v.mu.Unlock()

	now := v.now()
	if k, ok := v.keys[kid]; ok && now.Before(v.expiresAt) {
		return k, nil
	}
	// Refetch when the cache is stale, or when the kid is unknown but we
	// haven't fetched in the last minute (Google rotates keys; the minimum
	// interval stops junk tokens from making us hammer Google).
	stale := !now.Before(v.expiresAt)
	if stale || now.Sub(v.fetchedAt) > time.Minute {
		if err := v.refresh(ctx); err != nil {
			return nil, err
		}
	}
	k, ok := v.keys[kid]
	if !ok {
		return nil, invalidKeyError{"unknown signing key"}
	}
	return k, nil
}

func (v *GoogleVerifier) refresh(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, v.certsURL, nil)
	if err != nil {
		return err
	}
	resp, err := v.client.Do(req)
	if err != nil {
		return fmt.Errorf("fetch Google signing keys: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fetch Google signing keys: status %d", resp.StatusCode)
	}
	var body struct {
		Keys []struct {
			Kid string `json:"kid"`
			Kty string `json:"kty"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return fmt.Errorf("decode Google signing keys: %w", err)
	}
	keys := make(map[string]*rsa.PublicKey, len(body.Keys))
	for _, k := range body.Keys {
		if k.Kty != "RSA" {
			continue
		}
		nb, err1 := base64.RawURLEncoding.DecodeString(k.N)
		eb, err2 := base64.RawURLEncoding.DecodeString(k.E)
		if err1 != nil || err2 != nil {
			continue
		}
		e := new(big.Int).SetBytes(eb)
		if !e.IsInt64() {
			continue
		}
		keys[k.Kid] = &rsa.PublicKey{N: new(big.Int).SetBytes(nb), E: int(e.Int64())}
	}
	if len(keys) == 0 {
		return errors.New("no usable signing keys returned by Google")
	}
	now := v.now()
	v.keys = keys
	v.fetchedAt = now
	v.expiresAt = now.Add(cacheTTL(resp.Header.Get("Cache-Control")))
	return nil
}

// cacheTTL reads max-age from a Cache-Control header, defaulting to an hour.
func cacheTTL(h string) time.Duration {
	for _, part := range strings.Split(h, ",") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "max-age=") {
			if n, err := strconv.Atoi(strings.TrimPrefix(part, "max-age=")); err == nil && n > 0 {
				return time.Duration(n) * time.Second
			}
		}
	}
	return time.Hour
}
