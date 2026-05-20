package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"broccoli/internal/db" // Adjust to your actual module path
)

func TestHandleWithdrawal_EdgeCases(t *testing.T) {
	t.Parallel()
	cfg := &Router{}

	// ---------------------------------------------------------
	// Setup: Prerequisite Data
	// ---------------------------------------------------------
	
	// 1. Ensure the system account exists for the handler to find
	sysUser, _ := db.Users.Insert("WithdrawalSysUser", "hash")
	_, err := db.Accounts.Insert("broccoli", sysUser.Id)
	if err != nil && err.Error() != "UNIQUE constraint failed" { 
		// Ignore if it already exists from another test
	}

	// 2. Create the target user and account for our edge cases
	user, err := db.Users.Insert("WithdrawEdgeUser", "hash")
	if err != nil {
		t.Fatalf("Failed to create prerequisite user: %v", err)
	}
	_, err = db.Accounts.Insert("WithdrawEdgeAlias", user.Id)
	if err != nil {
		t.Fatalf("Failed to create prerequisite account: %v", err)
	}

	// ---------------------------------------------------------
	// Table-Driven Edge Cases
	// ---------------------------------------------------------
	tests := []struct {
		name           string
		injectUserID   int64 // Simulated ID extracted from context
		payload        interface{}
		expectedStatus int
	}{
		{
			name:         "Failure - Malformed JSON",
			injectUserID: user.Id,
			payload:      "not a valid json string",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:         "Failure - User Account Not Found",
			injectUserID: -1, // Ghost user ID
			payload: withdrawalRequest{
				Amount:      50,
				Destination: "External Bank",
				Description: "Withdrawal",
			},
			// Based on current handler code, if FindForUpdate fails, it returns 500
			expectedStatus: http.StatusInternalServerError, 
		},
		{
			name:         "Failure - Zero Amount",
			injectUserID: user.Id,
			payload: withdrawalRequest{
				Amount:      0,
				Destination: "External Bank",
				Description: "Withdrawal",
			},
			expectedStatus: http.StatusExpectationFailed,
		},
		{
			name:         "Failure - Negative Amount",
			injectUserID: user.Id,
			payload: withdrawalRequest{
				Amount:      -500,
				Destination: "External Bank",
				Description: "Withdrawal",
			},
			expectedStatus: http.StatusExpectationFailed,
		},
		{
			name:         "Failure - Insufficient Funds",
			injectUserID: user.Id,
			payload: withdrawalRequest{
				Amount:      50, // Account was just created, balance is 0
				Destination: "External Bank",
				Description: "Withdrawal",
			},
			expectedStatus: http.StatusExpectationFailed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Convert payload to bytes
			var bodyBytes []byte
			if str, ok := tc.payload.(string); ok {
				bodyBytes = []byte(str)
			} else {
				bodyBytes, _ = json.Marshal(tc.payload)
			}

			// Construct Request
			req := httptest.NewRequest(http.MethodPost, "/api/withdrawal", bytes.NewBuffer(bodyBytes))

			// Mock the middleware context
			ctx := context.WithValue(req.Context(), userIdKey, tc.injectUserID)
			req = req.WithContext(ctx)

			// Record Response
			rr := httptest.NewRecorder()
			cfg.HandleWithdrawal(rr, req)

			// Assert Status
			if rr.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tc.expectedStatus, rr.Code, rr.Body.String())
			}
		})
	}
}