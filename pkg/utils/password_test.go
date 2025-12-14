package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPassword(t *testing.T) {
	password := "mySecretPassword123!"

	t.Run("HashPassword", func(t *testing.T) {
		hash, err := HashPassword(password)
		require.NoError(t, err, "HashPassword should not return an error")
		require.NotEmpty(t, hash, "Hash should not be empty")
		assert.NotEqual(t, password, hash, "Password hash should not be equal to the original password")
	})

	t.Run("CheckPassword_Success", func(t *testing.T) {
		hash, err := HashPassword(password)
		require.NoError(t, err, "Hashing password for success case should not fail")
		match := CheckPassword(password, hash)
		assert.True(t, match, "Password should match the hash")
	})

	t.Run("CheckPassword_Failure", func(t *testing.T) {
		hash, err := HashPassword(password)
		require.NoError(t, err, "Hashing password for failure case should not fail")
		match := CheckPassword("wrongPassword", hash)
		assert.False(t, match, "Wrong password should not match the hash")
	})

	t.Run("CheckPassword_InvalidHash", func(t *testing.T) {
		match := CheckPassword(password, "invalidHash")
		assert.False(t, match, "Should return false for an invalid hash")
	})
}