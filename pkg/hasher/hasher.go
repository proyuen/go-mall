package hasher

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

//go:generate mockgen -source=$GOFILE -destination=../../internal/mocks/hasher_mock.go -package=mocks
// PasswordHasher defines the interface for password hashing operations.
type PasswordHasher interface {
	// Hash returns the hashed version of the password.
	Hash(password string) (string, error)
	// Check compares a plaintext password with a hashed password.
	Check(password, hashedPassword string) error
}

// BcryptHasher implements PasswordHasher using the bcrypt algorithm.
type BcryptHasher struct {
	cost int
}

// NewBcryptHasher creates a new BcryptHasher with the specified cost.
// If the cost is less than bcrypt.MinCost, it defaults to bcrypt.DefaultCost.
func NewBcryptHasher(cost int) *BcryptHasher {
	if cost < bcrypt.MinCost {
		cost = bcrypt.DefaultCost
	}
	return &BcryptHasher{
		cost: cost,
	}
}

// Hash hashes a password using bcrypt.
func (h *BcryptHasher) Hash(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedPassword), nil
}

// Check compares a hashed password with a plaintext password.
func (h *BcryptHasher) Check(password, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
