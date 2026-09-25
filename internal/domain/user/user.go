// Package user holds the User entity and its repository contract.
package user

import "time"

// User is an app account, authenticated by email + password.
type User struct {
	ID           string    `json:"id" bson:"_id"`
	Email        string    `json:"email" bson:"email"`
	Name         string    `json:"name,omitempty" bson:"name,omitempty"`
	PasswordHash string    `json:"-" bson:"passwordHash"`
	CreatedAt    time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt" bson:"updatedAt"`
}
