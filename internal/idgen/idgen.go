// Package idgen generates opaque, unique entity IDs.
package idgen

import (
	"crypto/rand"
	"encoding/hex"
)

type Generator struct{}

func New() Generator { return Generator{} }

// New returns a random 24-character hex ID prefixed with the given entity
// tag, e.g. New("pl") -> "pl_3f9a1c...".
func (Generator) New(prefix string) string {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		panic(err) // crypto/rand failing means the OS entropy source is broken
	}
	return prefix + "_" + hex.EncodeToString(b)
}
