package token

import (
	"time"
)

//go:generate mockgen -source=$GOFILE -destination=../../internal/mocks/token_maker_mock.go -package=mocks
// Maker is an interface for managing tokens
type Maker interface {
	// CreateToken creates a new token for a specific username and duration
	CreateToken(userID uint64, username string, duration time.Duration) (string, *Payload, error)

	// VerifyToken checks if the token is valid or not
	VerifyToken(token string) (*Payload, error)
}
