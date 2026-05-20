package auth

import (
	"strings"
	"testing"
)

func TestPasswordLifecycle(t *testing.T) {
	t.Parallel()

	validPassword := "superSecretPortfolioPass123!"
	wrongPassword := "imposterPassword456!"

	hash := HashPassword(validPassword)
	if hash == "" {
		t.Fatal("HashPassword returned an empty string")
	}

	isValid := CheckPasswordHash(validPassword, hash)
	if !isValid {
		t.Error("CheckPasswordHash rejected a valid password match")
	}

	isInvalid := CheckPasswordHash(wrongPassword, hash)
	if isInvalid {
		t.Error("CRITICAL: CheckPasswordHash accepted an incorrect password")
	}
}

func TestPasswordLengthPanic(t *testing.T) {
	t.Parallel()

	// Bcrypt has a strict 72-byte limit.
	// We create a 73-byte string to deliberately trigger the failure.
	maliciousPassword := strings.Repeat("a", 73)

	// We use defer and recover to catch the panic before it crashes the test suite.
	defer func() {
		r := recover()
		if r == nil {
			t.Error("Expected HashPassword to panic with a >72 byte password, but it survived.")
		}
	}()

	// This function call will crash. The defer block above catches it.
	HashPassword(maliciousPassword)
}
