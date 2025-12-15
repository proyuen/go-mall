package utils

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWT(t *testing.T) {
	secret := "test_secret_key_12345"
	wrongSecret := "wrong_secret_key_67890"
	userID := uint(101)
	username := "testuser"

	t.Run("GenerateAndParse_Success", func(t *testing.T) {
		token, err := GenerateToken(userID, username, secret)
		require.NoError(t, err)
		require.NotEmpty(t, token)

		claims, err := ParseToken(token, secret)
		require.NoError(t, err)
		require.NotNil(t, claims)

		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, username, claims.Username)
		assert.Equal(t, "go-mall", claims.Issuer)
	})

	t.Run("Parse_WithWrongSecret_Failure", func(t *testing.T) {
		token, err := GenerateToken(userID, username, secret)
		require.NoError(t, err)

		claims, err := ParseToken(token, wrongSecret)
		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.Contains(t, err.Error(), "signature is invalid")
	})

	t.Run("Parse_ExpiredToken_Failure", func(t *testing.T) {
		// Manually generate an expired token
		expiredClaims := UserClaims{
			UserID:   userID,
			Username: username,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // Expired 1 hour ago
				Issuer:    "go-mall",
			},
		}
		tokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
		expiredToken, err := tokenObj.SignedString([]byte(secret))
		require.NoError(t, err)

		claims, err := ParseToken(expiredToken, secret)
		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.Contains(t, err.Error(), "token has invalid claims: token is expired")
	})

	t.Run("Parse_InvalidFormat_Failure", func(t *testing.T) {
		claims, err := ParseToken("invalid.token.string", secret)
		assert.Error(t, err)
		assert.Nil(t, claims)
	})
}
