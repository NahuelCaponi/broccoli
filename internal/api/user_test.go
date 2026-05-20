package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"broccoli/internal/db" // Adjust to your actual module path
)

func TestHandleUserCreate_EdgeCases(t *testing.T) {
	t.Parallel()
	cfg := &Router{}

	// Setup: We need one existing user in the database to test the Conflict edge case
	existingUsername := "existingedgeuser"
	_, err := db.Users.Insert(existingUsername, "dummyhash")
	if err != nil {
		t.Fatalf("Failed to create prerequisite user for edge tests: %v", err)
	}

	tests := []struct {
		name           string
		payload        interface{}
		expectedStatus int
	}{
		{
			name: "Failure - User Already Exists",
			payload: userRequest{
				UserName: existingUsername,
				Password: "ValidPassword123!",
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name: "Failure - Password Too Long (> 72 bytes)",
			payload: userRequest{
				UserName: "LongPassUser",
				Password: strings.Repeat("a", 73), // Triggers the bcrypt protection limit
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Failure - Malformed JSON",
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

			req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBuffer(bodyBytes))
			rr := httptest.NewRecorder()

			cfg.HandleUserCreate(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tc.expectedStatus, rr.Code, rr.Body.String())
			}
		})
	}
}

func TestHandleUserUpdate_EdgeCases(t *testing.T) {
	t.Parallel()
	cfg := &Router{}

	tests := []struct {
		name           string
		injectUserID   int64 // Simulated ID extracted from context
		payload        interface{}
		expectedStatus int
	}{
		{
			name:         "Failure - User Not Found (Deleted or invalid JWT payload)",
			injectUserID: 99999999, // ID that does not exist in the test DB
			payload: userRequest{
				UserName: "GhostUser",
				Password: "password123!",
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name:         "Failure - Password Too Long (> 72 bytes)",
			injectUserID: 1, // Doesn't matter if it exists, it should fail before DB check
			payload: userRequest{
				UserName: "SomeUser",
				Password: strings.Repeat("x", 75),
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Failure - Malformed JSON",
			injectUserID:   1,
			payload:        "invalid payload data",
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

			req := httptest.NewRequest(http.MethodPut, "/api/users", bytes.NewBuffer(bodyBytes))
			
			// Mock the middleware by directly injecting our unexported context key
			ctx := context.WithValue(req.Context(), userIdKey, tc.injectUserID)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()

			cfg.HandleUserUpdate(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tc.expectedStatus, rr.Code, rr.Body.String())
			}
		})
	}
}