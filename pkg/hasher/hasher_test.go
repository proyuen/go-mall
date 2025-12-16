package hasher

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestBcryptHasher_Hash(t *testing.T) {
	tests := []struct {
		name     string
		cost     int
		password string
		wantErr  bool
	}{
		{
			name:     "default_cost",
			cost:     0, // Should default to bcrypt.DefaultCost
			password: "password123",
			wantErr:  false,
		},
		{
			name:     "valid_cost",
			cost:     bcrypt.MinCost,
			password: "securePassword",
			wantErr:  false,
		},
		{
			name:     "empty_password",
			cost:     bcrypt.DefaultCost,
			password: "",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewBcryptHasher(tt.cost)
			got, err := h.Hash(tt.password)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotEmpty(t, got)

			// Verify we can check it
			err = bcrypt.CompareHashAndPassword([]byte(got), []byte(tt.password))
			require.NoError(t, err)
		})
	}
}

func TestBcryptHasher_Check(t *testing.T) {
	h := NewBcryptHasher(bcrypt.DefaultCost)
	password := "mySecret123"
	hash, err := h.Hash(password)
	require.NoError(t, err)

	tests := []struct {
		name           string
		password       string
		hashedPassword string
		wantErr        bool
		errType        error
	}{
		{
			name:           "correct_password",
			password:       password,
			hashedPassword: hash,
			wantErr:        false,
		},
		{
			name:           "incorrect_password",
			password:       "wrong",
			hashedPassword: hash,
			wantErr:        true,
			errType:        bcrypt.ErrMismatchedHashAndPassword,
		},
		{
			name:           "invalid_hash",
			password:       password,
			hashedPassword: "not_a_hash",
			wantErr:        true,
			// bcrypt error structure for format is just an error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := h.Check(tt.password, tt.hashedPassword)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errType != nil {
					assert.Equal(t, tt.errType, err)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestNewBcryptHasher_CostFallback(t *testing.T) {
	// If we pass a cost < MinCost, it should use DefaultCost.
	// We can't easily inspect the internal field without reflection or helper,
	// but we can infer it works if Hash doesn't panic and produces a valid hash.
	// Alternatively, we could check the cost from the generated hash string.
	
	h := NewBcryptHasher(1) // Too low
	hash, err := h.Hash("test")
	require.NoError(t, err)
	
	cost, err := bcrypt.Cost([]byte(hash))
	require.NoError(t, err)
	assert.Equal(t, bcrypt.DefaultCost, cost)
}
