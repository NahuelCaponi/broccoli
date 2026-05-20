package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"broccoli/internal/auth"
	"broccoli/internal/db"
)

func TestHandleLogin(t *testing.T) {
	t.Parallel()

	secret := "test-login-jwt-secret-123!"
	cfg := &Router{JwtSecret: secret}

	rawPassword := "MySecurePass123!"
	hashedPassword := auth.HashPassword(rawPassword)
	testUsername := "LoginAPIUser"

	user, err := db.Users.Insert(testUsername, hashedPassword)
	if err != nil {
		t.Fatalf("Failed prerequisite: could not create test user: %v", err)
	}

	tests := []struct {
		name           string
		payload        interface{}
		expectedStatus int
	}{
		{
			name: "Happy Path - Valid Credentials",
			payload: loginRequest{
				Username: testUsername,
				Password: rawPassword,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Failure - Wrong Password",
			payload: loginRequest{
				Username: testUsername,
				Password: "WrongPassword999!",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Failure - Unknown User",
			payload: loginRequest{
				Username: "GhostUserThatDoesNotExist",
				Password: "AnyPassword123!",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Failure - Bad Request (Malformed JSON)",
			payload:        "this { is ] not valid json",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var bodyBytes []byte
			if str, ok := tc.payload.(string); ok {
				bodyBytes = []byte(str)
			} else {
				bodyBytes, _ = json.Marshal(tc.payload)
			}

			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(bodyBytes))
			rr := httptest.NewRecorder()

			cfg.HandleLogin(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tc.expectedStatus, rr.Code, rr.Body.String())
			}

			if rr.Code == http.StatusOK {
				var resp loginResponse
				err := json.NewDecoder(rr.Body).Decode(&resp)
				if err != nil {
					t.Fatalf("Failed to decode successful JSON response: %v", err)
				}

				if resp.Token == "" {
					t.Error("Expected a JWT token in the response, got an empty string")
				} else {
					extractedID, _, err := auth.ValidateJWT(resp.Token, secret)
					if err != nil {
						t.Errorf("Handler generated an invalid token: %v", err)
					}
					if extractedID != user.Id {
						t.Errorf("Token contains wrong user ID. Expected %d, got %d", user.Id, extractedID)
					}
				}

				if resp.UpdatedAt == "" {
					t.Error("Expected updated_at field, got an empty string")
				}
			}
		})
	}
}
