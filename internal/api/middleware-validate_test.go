package api

import (
	"broccoli/internal/auth"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMiddlewareValidateUser(t *testing.T) {
	t.Parallel()

	secret := "super-secret-middleware-key-123!"
	cfg := &Router{JwtSecret: secret}
	expectedUserID := int64(999)

	validToken := auth.MakeJWT(expectedUserID, "", 15*time.Minute, secret)
	invalidSignatureToken := auth.MakeJWT(expectedUserID, "", 15*time.Minute, "wrong-hacker-secret")

	var nextHandlerWasCalled bool
	var extractedUserID int64

	mockNextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextHandlerWasCalled = true

		val := r.Context().Value(userIdKey)
		if id, ok := val.(int64); ok {
			extractedUserID = id
		}

		w.WriteHeader(http.StatusOK)
	})

	protectedHandler := cfg.MiddlewareValidateUser(mockNextHandler)

	tests := []struct {
		name                 string
		authHeader           string
		expectHandlerReached bool
		expectedStatusCode   int
	}{
		{
			name:                 "Happy Path - Valid Token",
			authHeader:           "Bearer " + validToken,
			expectHandlerReached: true,
			expectedStatusCode:   http.StatusOK,
		},
		{
			name:                 "Failure - Missing Header",
			authHeader:           "",
			expectHandlerReached: false,
			expectedStatusCode:   http.StatusUnauthorized,
		},
		{
			name:                 "Failure - Bad Prefix (Basic Auth)",
			authHeader:           "Basic " + validToken,
			expectHandlerReached: false,
			expectedStatusCode:   http.StatusUnauthorized,
		},
		{
			name:                 "Failure - Invalid JWT Signature",
			authHeader:           "Bearer " + invalidSignatureToken,
			expectHandlerReached: false,
			expectedStatusCode:   http.StatusUnauthorized,
		},
		{
			name:                 "Failure - Garbage Token",
			authHeader:           "Bearer not.a.real.jwt",
			expectHandlerReached: false,
			expectedStatusCode:   http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			nextHandlerWasCalled = false
			extractedUserID = 0

			req := httptest.NewRequest(http.MethodGet, "/protected-route", nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}
			rr := httptest.NewRecorder()

			protectedHandler.ServeHTTP(rr, req)

			if rr.Code != tc.expectedStatusCode {
				t.Errorf("Expected status code %d, got %d", tc.expectedStatusCode, rr.Code)
			}

			if nextHandlerWasCalled != tc.expectHandlerReached {
				t.Errorf("Expected next handler reached: %v, but got: %v", tc.expectHandlerReached, nextHandlerWasCalled)
			}

			if tc.expectHandlerReached && extractedUserID != expectedUserID {
				t.Errorf("Middleware failed to inject context correctly. Expected UserID %d, got %d", expectedUserID, extractedUserID)
			}
		})
	}
}
