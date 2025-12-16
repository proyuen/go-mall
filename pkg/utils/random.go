package utils

import (
	"math/rand/v2"
	"strings"
)

const alphabet = "abcdefghijklmnopqrstuvwxyz"

// RandomInt generates a random integer between min and max (inclusive).
func RandomInt(min, max int) int {
	if min >= max {
		return min
	}
	return rand.IntN(max-min+1) + int(min)
}

// RandomString generates a random string of length n.
func RandomString(n int) string {
	var sb strings.Builder
	sb.Grow(n)
	k := len(alphabet)

	for i := 0; i < n; i++ {
		c := alphabet[rand.IntN(k)]
		sb.WriteByte(c)
	}

	return sb.String()
}

// RandomOwner generates a random owner name.
func RandomOwner() string {
	length := RandomInt(3, 10)
	return RandomString(length)
}

// RandomEmail generates a random email address.
func RandomEmail(domain string) string {
	if domain == "" {
		domain = "test.com"
	}
	length := RandomInt(3, 10)
	return RandomString(length) + "@" + domain
}
