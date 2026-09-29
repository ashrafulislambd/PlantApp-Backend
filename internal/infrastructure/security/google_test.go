package security

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testAud = "web-client.apps.googleusercontent.com"

var testNow = time.Date(2026, 9, 29, 12, 0, 0, 0, time.UTC)

func newKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	k, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	return k
}

// certsServer serves a JWKS containing the given keys (kid -> key).
func certsServer(t *testing.T, keys map[string]*rsa.PrivateKey, hits *int) *httptest.Server {
	t.Helper()
	type jwk struct {
		Kid string `json:"kid"`
		Kty string `json:"kty"`
		N   string `json:"n"`
		E   string `json:"e"`
	}
	var out struct {
		Keys []jwk `json:"keys"`
	}
	for kid, k := range keys {
		out.Keys = append(out.Keys, jwk{
			Kid: kid, Kty: "RSA",
			N: base64.RawURLEncoding.EncodeToString(k.N.Bytes()),
			E: base64.RawURLEncoding.EncodeToString(big.NewInt(int64(k.E)).Bytes()),
		})
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hits != nil {
			*hits++
		}
		w.Header().Set("Cache-Control", "public, max-age=3600")
		_ = json.NewEncoder(w).Encode(out)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func newVerifier(srv *httptest.Server) *GoogleVerifier {
	v := NewGoogleVerifier([]string{testAud})
	v.certsURL = srv.URL
	v.now = func() time.Time { return testNow }
	return v
}

func signToken(t *testing.T, key *rsa.PrivateKey, kid string, override jwt.MapClaims) string {
	t.Helper()
	claims := jwt.MapClaims{
		"iss":            "https://accounts.google.com",
		"aud":            testAud,
		"sub":            "1234567890",
		"email":          "Rahim@Example.com",
		"email_verified": true,
		"name":           "Rahim Uddin",
		"iat":            testNow.Add(-time.Minute).Unix(),
		"exp":            testNow.Add(time.Hour).Unix(),
	}
	for k, v := range override {
		if v == nil {
			delete(claims, k)
		} else {
			claims[k] = v
		}
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tok.Header["kid"] = kid
	s, err := tok.SignedString(key)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestGoogleVerify_Valid(t *testing.T) {
	key := newKey(t)
	v := newVerifier(certsServer(t, map[string]*rsa.PrivateKey{"k1": key}, nil))

	id, err := v.Verify(context.Background(), signToken(t, key, "k1", nil))
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if id.Subject != "1234567890" || id.Email != "Rahim@Example.com" || id.Name != "Rahim Uddin" {
		t.Errorf("unexpected identity: %+v", id)
	}
}

func TestGoogleVerify_AcceptsBareIssuer(t *testing.T) {
	key := newKey(t)
	v := newVerifier(certsServer(t, map[string]*rsa.PrivateKey{"k1": key}, nil))
	if _, err := v.Verify(context.Background(), signToken(t, key, "k1", jwt.MapClaims{"iss": "accounts.google.com"})); err != nil {
		t.Errorf("Verify() error = %v", err)
	}
}

func TestGoogleVerify_Rejects(t *testing.T) {
	key := newKey(t)
	other := newKey(t)
	v := newVerifier(certsServer(t, map[string]*rsa.PrivateKey{"k1": key}, nil))

	cases := map[string]string{
		"wrong audience":   signToken(t, key, "k1", jwt.MapClaims{"aud": "someone-else"}),
		"wrong issuer":     signToken(t, key, "k1", jwt.MapClaims{"iss": "https://evil.example.com"}),
		"expired":          signToken(t, key, "k1", jwt.MapClaims{"exp": testNow.Add(-time.Hour).Unix()}),
		"no expiry":        signToken(t, key, "k1", jwt.MapClaims{"exp": nil}),
		"unverified email": signToken(t, key, "k1", jwt.MapClaims{"email_verified": false}),
		"missing email":    signToken(t, key, "k1", jwt.MapClaims{"email": nil}),
		"missing subject":  signToken(t, key, "k1", jwt.MapClaims{"sub": nil}),
		"unknown kid":      signToken(t, key, "nope", nil),
		"bad signature":    signToken(t, other, "k1", nil),
		"garbage":          "not-a-jwt",
	}
	for name, tok := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := v.Verify(context.Background(), tok)
			if !errors.Is(err, ErrInvalidGoogleToken) {
				t.Errorf("Verify() error = %v, want ErrInvalidGoogleToken", err)
			}
		})
	}
}

func TestGoogleVerify_RejectsHMACToken(t *testing.T) {
	key := newKey(t)
	v := newVerifier(certsServer(t, map[string]*rsa.PrivateKey{"k1": key}, nil))
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss": "https://accounts.google.com", "aud": testAud, "sub": "1", "email": "a@b.c",
		"email_verified": true, "exp": testNow.Add(time.Hour).Unix(),
	})
	tok.Header["kid"] = "k1"
	s, _ := tok.SignedString([]byte("secret"))
	if _, err := v.Verify(context.Background(), s); !errors.Is(err, ErrInvalidGoogleToken) {
		t.Errorf("Verify() error = %v, want ErrInvalidGoogleToken", err)
	}
}

func TestGoogleVerify_CachesKeys(t *testing.T) {
	key := newKey(t)
	hits := 0
	v := newVerifier(certsServer(t, map[string]*rsa.PrivateKey{"k1": key}, &hits))
	for i := 0; i < 3; i++ {
		if _, err := v.Verify(context.Background(), signToken(t, key, "k1", nil)); err != nil {
			t.Fatal(err)
		}
	}
	if hits != 1 {
		t.Errorf("certs fetched %d times, want 1", hits)
	}
}

func TestGoogleVerify_KeyEndpointDownIsNotInvalidToken(t *testing.T) {
	key := newKey(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()
	v := newVerifier(srv)

	_, err := v.Verify(context.Background(), signToken(t, key, "k1", nil))
	if err == nil || errors.Is(err, ErrInvalidGoogleToken) {
		t.Errorf("Verify() error = %v, want a non-ErrInvalidGoogleToken error", err)
	}
}
