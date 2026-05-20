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

func TestHandleTransfer_EdgeCases(t *testing.T) {
	t.Parallel()
	cfg := &Router{}

	// ---------------------------------------------------------
	// Setup: We need two real accounts in the DB to test transfers
	// ---------------------------------------------------------
	senderUser, err := db.Users.Insert("TransferEdgeSender", "hash")
	if err != nil {
		t.Fatalf("Failed to create sender user: %v", err)
	}
	senderAccount, err := db.Accounts.Insert("sender_alias", senderUser.Id)
	if err != nil {
		t.Fatalf("Failed to create sender account: %v", err)
	}

	receiverUser, err := db.Users.Insert("TransferEdgeReceiver", "hash")
	if err != nil {
		t.Fatalf("Failed to create receiver user: %v", err)
	}
	// Note: We don't technically need the receiver's ID in the handler, 
	// just the alias to exist in the database.
	_, err = db.Accounts.Insert("receiver_alias", receiverUser.Id)
	if err != nil {
		t.Fatalf("Failed to create receiver account: %v", err)
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
			injectUserID: senderUser.Id,
			payload:      "not valid json format",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:         "Failure - User Not Found (Bad Context ID)",
			injectUserID: 99999999, // Ghost user ID
			payload: transferRequest{
				Amount:      50,
				Destination: "receiver_alias",
			},
			expectedStatus: http.StatusNotFound, // Based on your handler's specific check
		},
		{
			name:         "Failure - Transfer to Self",
			injectUserID: senderUser.Id,
			payload: transferRequest{
				Amount:      50,
				Destination: senderAccount.Alias, // Sending to own alias
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:         "Failure - Destination Account Does Not Exist",
			injectUserID: senderUser.Id,
			payload: transferRequest{
				Amount:      50,
				Destination: "ghost_alias_nowhere",
			},
			expectedStatus: http.StatusBadRequest, // Based on your sql.ErrNoRows check
		},
		{
			name:         "Failure - Zero Amount",
			injectUserID: senderUser.Id,
			payload: transferRequest{
				Amount:      0,
				Destination: "receiver_alias",
			},
			expectedStatus: http.StatusExpectationFailed,
		},
		{
			name:         "Failure - Negative Amount",
			injectUserID: senderUser.Id,
			payload: transferRequest{
				Amount:      -100,
				Destination: "receiver_alias",
			},
			expectedStatus: http.StatusExpectationFailed,
		},
		{
			name:         "Failure - Insufficient Funds",
			injectUserID: senderUser.Id,
			payload: transferRequest{
				Amount:      50, // Sender's default balance is 0, so this must fail
				Destination: "receiver_alias",
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
			req := httptest.NewRequest(http.MethodPost, "/api/transfer", bytes.NewBuffer(bodyBytes))

			// Mock the middleware context
			ctx := context.WithValue(req.Context(), userIdKey, tc.injectUserID)
			req = req.WithContext(ctx)

			// Record Response
			rr := httptest.NewRecorder()
			cfg.HandleTransfer(rr, req)

			// Assert Status
			if rr.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tc.expectedStatus, rr.Code, rr.Body.String())
			}
		})
	}
}