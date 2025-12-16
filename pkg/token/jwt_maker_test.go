package token

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTMaker(t *testing.T) {
	// Common Setup
	maker, err := NewJWTMaker("12345678901234567890123456789012") // 32 chars
	require.NoError(t, err)

	username := "test_user"
	userID := uint64(101)
	duration := time.Minute

	issuedAt := time.Now()
	expiredAt := issuedAt.Add(duration)

	tests := []struct {
		name          string
		setupToken    func(t *testing.T) string
		checkResponse func(t *testing.T, payload *Payload, err error)
	}{
		{
			name: "Success",
			setupToken: func(t *testing.T) string {
				token, _, err := maker.CreateToken(userID, username, duration)
				require.NoError(t, err)
				return token
			},
			checkResponse: func(t *testing.T, payload *Payload, err error) {
				require.NoError(t, err)
				require.NotEmpty(t, payload)
				
				assert.Equal(t, userID, payload.UserID)
				assert.Equal(t, username, payload.Username)
				assert.WithinDuration(t, issuedAt, payload.IssuedAt, time.Second)
				assert.WithinDuration(t, expiredAt, payload.ExpiredAt, time.Second)
				assert.NotZero(t, payload.ID)
			},
		},
		{
			name: "ExpiredToken",
			setupToken: func(t *testing.T) string {
				token, _, err := maker.CreateToken(userID, username, -time.Minute)
				require.NoError(t, err)
				return token
			},
			checkResponse: func(t *testing.T, payload *Payload, err error) {
				require.Error(t, err)
				assert.EqualError(t, err, ErrExpiredToken.Error())
				require.Nil(t, payload)
			},
		},
		{
			name: "InvalidTokenAlg",
			setupToken: func(t *testing.T) string {
				payload, err := NewPayload(userID, username, duration)
				require.NoError(t, err)

				jwtToken := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{
					"id":         payload.ID,
					"user_id":    payload.UserID,
					"username":   payload.Username,
					"issued_at":  payload.IssuedAt,
					"expired_at": payload.ExpiredAt,
				})
				token, err := jwtToken.SignedString(jwt.UnsafeAllowNoneSignatureType)
				require.NoError(t, err)
				return token
			},
			checkResponse: func(t *testing.T, payload *Payload, err error) {
				require.Error(t, err)
				assert.EqualError(t, err, ErrInvalidToken.Error())
				require.Nil(t, payload)
			},
		},
		{
			name: "TamperedToken",
			setupToken: func(t *testing.T) string {
				token, _, err := maker.CreateToken(userID, username, duration)
				require.NoError(t, err)
				// Tamper with the token by modifying the last character
				return token[0:len(token)-1] + "x"
			},
			checkResponse: func(t *testing.T, payload *Payload, err error) {
				require.Error(t, err)
				assert.EqualError(t, err, ErrInvalidToken.Error())
				require.Nil(t, payload)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := tt.setupToken(t)
			payload, err := maker.VerifyToken(token)
			tt.checkResponse(t, payload, err)
		})
	}
}

func TestNewJWTMaker(t *testing.T) {
	tests := []struct {
		name      string
		secretKey string
		wantErr   bool
	}{
		{
			name:      "ValidKeySize",
			secretKey: "12345678901234567890123456789012",
			wantErr:   false,
		},
		{
			name:      "InvalidKeySizeTooShort",
			secretKey: "short_key",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			maker, err := NewJWTMaker(tt.secretKey)
			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, maker)
			} else {
				require.NoError(t, err)
				require.NotNil(t, maker)
			}
		})
	}
}