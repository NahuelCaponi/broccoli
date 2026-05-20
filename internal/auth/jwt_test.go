package auth

import (
	"testing"
	"time"
)

func TestJWTLifecycle(t *testing.T) {
	t.Parallel()

	secret := "super-secret-portfolio-key-123!"
	expectedUserID := int64(8675309)

	token := MakeJWT(expectedUserID, "sub", 15*time.Minute, secret)
	if token == "" {
		t.Fatal("MakeJWT returned an empty string instead of a valid token")
	}

	extractedID, subject, err := ValidateJWT(token, secret)
	if err != nil {
		t.Fatalf("ValidateJWT rejected a freshly minted, valid token: %v", err)
	}

	if extractedID != expectedUserID {
		t.Errorf("ValidateJWT extracted the wrong User ID. Expected %d, got %d", expectedUserID, extractedID)
	}
	if subject!= "sub" {
		t.Errorf("ValidateJWT extracted the wrong subject. Expected %s, got %s", "sub", subject)
	}
}

func TestJWTSecurityAndFailures(t *testing.T) {
	t.Parallel()

	validSecret := "correct-server-secret"
	hackerSecret := "guessed-server-secret"
	userID := int64(42)

	expiredToken := MakeJWT(userID,"", -1*time.Hour, validSecret)
	_,_, err := ValidateJWT(expiredToken, validSecret)
	if err == nil {
		t.Error("CRITICAL: ValidateJWT accepted a token that is already expired")
	}

	validToken := MakeJWT(userID,"", 1*time.Hour, validSecret)
	_, _, err = ValidateJWT(validToken, hackerSecret)
	if err == nil {
		t.Error("CRITICAL: ValidateJWT accepted a token signed with the wrong secret key")
	}

	malformedToken := "this.is.not.a.valid.jwt.string"
	
	_, _, err = ValidateJWT(malformedToken, validSecret)
	if err == nil {
		t.Error("ValidateJWT accepted a completely malformed string as a valid token")
	}
}