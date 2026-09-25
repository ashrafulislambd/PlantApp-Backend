// Package refreshtoken holds the RefreshToken entity and its repository
// contract. A refresh token is deliberately NOT a JWT — it's an opaque random
// string, looked up by its hash, which is what makes it revocable.
package refreshtoken

import "time"

type RefreshToken struct {
	ID        string     `json:"id" bson:"_id"`
	UserID    string     `json:"userId" bson:"userId"`
	TokenHash string     `json:"-" bson:"tokenHash"`
	ExpiresAt time.Time  `json:"expiresAt" bson:"expiresAt"`
	CreatedAt time.Time  `json:"createdAt" bson:"createdAt"`
	RevokedAt *time.Time `json:"revokedAt,omitempty" bson:"revokedAt,omitempty"`
}
