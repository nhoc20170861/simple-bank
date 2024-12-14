package token

import "time"

// Maker is an interface for managing token
type Maker interface {
	// This method creates a new token for a specific username and duration
	CreateToken(username string, duration time.Duration) (string, error)
	// The method checks if the token is valid or not
	VerifyToken(token string) (*Payload, error)
}