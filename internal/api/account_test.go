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

func TestHandleAccountCreate_EdgeCases(t *testing.T) {
	t.Parallel()
	cfg := &Router{}

	// Setup: We need a real user and account in the database to test the 
	// "User already has an account" conflict scenario.
	user, err := db.Users.Insert("AccountEdgeUser", "hash")
	if err != nil {
		t.Fatalf("Failed to create prerequisite user: %v", err)
	}
	_, err = db.Accounts.Insert("EdgeAlias", user.Id)
	if err != nil {
		t.Fatalf("Failed to create prerequisite account: %v", err)
	}

	tests := []struct {
		name           string
		injectUserID   int64
		payload        interface{}
		expectedStatus int
	}{
		{
			name:         "Failure - User Already Has Account",
			injectUserID: user.Id, // Inject the ID of the user we just created
			payload: accountRequest{
				Alias: "NewAlias",
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name:         "Failure - Alias Too Long (> 20 bytes)",
			// We must use a ghost ID here so it passes the DB check and hits the length check
			injectUserID: 9999999, 
			payload: accountRequest{
				Alias: strings.Repeat("a", 21), 
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Failure - Malformed JSON",
			injectUserID:   9999999,
			payload:        "invalid payload { ]",
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

			req := httptest.NewRequest(http.MethodPost, "/api/accounts", bytes.NewBuffer(bodyBytes))
			
			// Mock the middleware context
			ctx := context.WithValue(req.Context(), userIdKey, tc.injectUserID)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			cfg.HandleAccountCreate(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tc.expectedStatus, rr.Code, rr.Body.String())
			}
		})
	}
}

func TestHandleAccountUpdate_EdgeCases(t *testing.T) {
	t.Parallel()
	cfg := &Router{}

	tests := []struct {
		name           string
		injectUserID   int64
		payload        interface{}
		expectedStatus int
	}{
		{
			name:         "Failure - Account Not Found",
			injectUserID: 9999999, // Ghost user ID means no account exists
			payload: accountRequest{
				Alias: "ValidAlias",
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name:         "Failure - Alias Too Long (> 20 bytes)",
			injectUserID: 1, // Doesn't matter because the length check happens before the DB query
			payload: accountRequest{
				Alias: strings.Repeat("x", 25),
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Failure - Malformed JSON",
			injectUserID:   1,
			payload:        "{ bad json format",
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

			req := httptest.NewRequest(http.MethodPut, "/api/accounts", bytes.NewBuffer(bodyBytes))
			
			// Mock the middleware context
			ctx := context.WithValue(req.Context(), userIdKey, tc.injectUserID)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			cfg.HandleAccountUpdate(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tc.expectedStatus, rr.Code, rr.Body.String())
			}
		})
	}
}