package token

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const minSecretKeySize = 32

// JWTMaker is a JSON Web Token maker
type JWTMaker struct {
	secretKey string
}

// NewJWTMaker creates a new JWTMaker
func NewJWTMaker(secretKey string) (Maker, error) {
	if len(secretKey) < minSecretKeySize {
		return nil, fmt.Errorf("invalid key size: must be at least %d characters", minSecretKeySize)
	}
	return &JWTMaker{secretKey: secretKey}, nil
}

// CreateToken creates a new token for a specific username and duration
func (maker *JWTMaker) CreateToken(userID uint64, username string, duration time.Duration) (string, *Payload, error) {
	payload, err := NewPayload(userID, username, duration)
	if err != nil {
		return "", payload, err
	}

	// Use JSON Marshal/Unmarshal to convert Payload struct to jwt.MapClaims
	// This ensures consistency and flexibility with struct fields
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", payload, fmt.Errorf("failed to marshal payload: %w", err)
	}

	var claims jwt.MapClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return "", payload, fmt.Errorf("failed to unmarshal payload into claims: %w", err)
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := jwtToken.SignedString([]byte(maker.secretKey))
	return token, payload, err
}

// VerifyToken checks if the token is valid or not
func (maker *JWTMaker) VerifyToken(token string) (*Payload, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, ErrInvalidToken
		}
		return []byte(maker.secretKey), nil
	}

	jwtToken, err := jwt.Parse(token, keyFunc)
	if err != nil {
		// jwt.Parse can return an error if the token is expired if standard "exp" claim is used.
		// Since we use custom fields, standard validation might be skipped or fail differently.
		// However, we should wrap the error cleanly.
		return nil, ErrInvalidToken
	}

	claims, ok := jwtToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	// Data Mapping: Convert jwt.MapClaims back to Payload struct using JSON
	// Note: jwt-go parses numbers as float64 by default.
	// JSON round-trip handles the type conversion cleanly.
	jsonBody, err := json.Marshal(claims)
	if err != nil {
		return nil, ErrInvalidToken
	}

	payload := &Payload{}
	if err := json.Unmarshal(jsonBody, payload); err != nil {
		return nil, ErrInvalidToken
	}

	// Check if the token is expired using our custom logic
	if err := payload.Valid(); err != nil {
		return nil, err
	}

	return payload, nil
}
