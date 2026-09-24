package security

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTIssuer signs and verifies short-lived access tokens.
type JWTIssuer struct {
	Secret    []byte
	Issuer    string
	AccessTTL time.Duration
}

func NewJWTIssuer(secret, issuer string, accessTTL time.Duration) *JWTIssuer {
	return &JWTIssuer{Secret: []byte(secret), Issuer: issuer, AccessTTL: accessTTL}
}

// NewAccessToken issues a JWT whose subject is the user ID.
func (j *JWTIssuer) NewAccessToken(userID string) (token string, expiresAt time.Time, err error) {
	expiresAt = time.Now().UTC().Add(j.AccessTTL)
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		Issuer:    j.Issuer,
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err = t.SignedString(j.Secret)
	return token, expiresAt, err
}

// ParseAccessToken validates the token and returns the user ID (subject).
func (j *JWTIssuer) ParseAccessToken(token string) (string, error) {
	claims := &jwt.RegisteredClaims{}
	parsed, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return j.Secret, nil
	})
	if err != nil || !parsed.Valid || claims.Subject == "" {
		return "", errors.New("invalid or expired token")
	}
	return claims.Subject, nil
}
