package utils

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandomInt(t *testing.T) {
	min := int64(10)
	max := int64(20)

	// Run multiple times to increase confidence
	for i := 0; i < 100; i++ {
		val := RandomInt(min, max)
		assert.GreaterOrEqual(t, val, min, "Value should be greater or equal to min")
		assert.LessOrEqual(t, val, max, "Value should be less or equal to max")
	}
}

func TestRandomString(t *testing.T) {
	n := 10
	regex := regexp.MustCompile(`^[a-z]+$`)

	for i := 0; i < 100; i++ {
		str := RandomString(n)
		assert.Len(t, str, n, "String length should match n")
		assert.True(t, regex.MatchString(str), "String should only contain lowercase alphabet characters")
	}
}

func TestRandomOwner(t *testing.T) {
	for i := 0; i < 100; i++ {
		owner := RandomOwner()
		assert.NotEmpty(t, owner)
		assert.Len(t, owner, 6, "RandomOwner should return string of length 6")
	}
}

func TestRandomEmail(t *testing.T) {
	regex := regexp.MustCompile(`^[a-z]{6}@email\.com$`)

	for i := 0; i < 100; i++ {
		email := RandomEmail()
		assert.NotEmpty(t, email)
		assert.True(t, regex.MatchString(email), "RandomEmail should match expected format")
	}
}
