package security

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateOpaqueToken returns a high-entropy random token used for refresh
// tokens.
func GenerateOpaqueToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
