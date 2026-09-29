// Package user holds the User entity and its repository contract.
package user

import "time"

// User is an app account, authenticated by email + password and/or Google.
// A Google-only account has an empty PasswordHash (password login can never
// succeed for it); GoogleID is Google's stable "sub" for the linked account.
type User struct {
	ID           string    `json:"id" bson:"_id"`
	Email        string    `json:"email" bson:"email"`
	Name         string    `json:"name,omitempty" bson:"name,omitempty"`
	PasswordHash string    `json:"-" bson:"passwordHash"`
	GoogleID     string    `json:"-" bson:"googleId,omitempty"`
	CreatedAt    time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt" bson:"updatedAt"`
}
