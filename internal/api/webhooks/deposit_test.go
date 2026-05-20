package webhooks

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"broccoli/internal/db"
)

func TestMain(m *testing.M) {
	db.Connect("broccoli_test")
	defer db.Close()

	os.Exit(m.Run())
}

func TestDepositWebhook(t *testing.T) {
	t.Parallel()

	targetUser, err := db.Users.Insert("WebhookTargetUser", "pass")
	if err != nil {
		t.Fatalf("Failed to create target user: %v", err)
	}
	targetAccount, err := db.Accounts.Insert("WebhookTargetAlias", targetUser.Id)
	if err != nil {
		t.Fatalf("Failed to create target account: %v", err)
	}

	secretToken := "test-secret-token-123"
	wh := &Webhook{DepositToken: secretToken}

	tests := []struct {
		name           string
		authHeader     string
		payload        interface{}
		expectedStatus int
	}{
		{
			name:       "Happy Path - Valid Deposit",
			authHeader: "Bearer " + secretToken,
			payload: DepositRequest{
				Event: "user.deposit",
				Data: struct {
					UserId      string `json:"user_id"`
					Amount      int    `json:"amount"`
					Description string `json:"description"`
				}{
					UserId:      strconv.FormatInt(targetUser.Id, 10),
					Amount:      5000, // $50.00
					Description: "Stripe Deposit",
				},
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:       "Failure - Unauthorized (Wrong Token)",
			authHeader: "Bearer wrong-token-456",
			payload: DepositRequest{
				Event: "user.deposit",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Failure - Unauthorized (Missing Header)",
			authHeader:     "",
			payload:        DepositRequest{},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:       "Ignored - Wrong Event Type",
			authHeader: "Bearer " + secretToken,
			payload: DepositRequest{
				Event: "user.withdrawal", // Should return 204 but do nothing
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:       "Failure - User Not Found",
			authHeader: "Bearer " + secretToken,
			payload: DepositRequest{
				Event: "user.deposit",
				Data: struct {
					UserId      string `json:"user_id"`
					Amount      int    `json:"amount"`
					Description string `json:"description"`
				}{
					UserId:      "-1",
					Amount:      5000,
					Description: "Ghost Deposit",
				},
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Failure - Bad Payload (Malformed JSON)",
			authHeader:     "Bearer " + secretToken,
			payload:        "this is not json",
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

			req := httptest.NewRequest(http.MethodPost, "/webhooks/deposit", bytes.NewBuffer(bodyBytes))
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}
			rr := httptest.NewRecorder()

			wh.Deposit(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tc.expectedStatus, rr.Code, rr.Body.String())
			}
		})
	}


	updatedAccount, err := db.Accounts.FindByID(int(targetAccount.Id))
	if err != nil {
		t.Fatalf("Failed to fetch updated account: %v", err)
	}

	expectedBalance := int64(5000)
	if updatedAccount.Balance != expectedBalance {
		t.Errorf("Database state mismatch! Expected account balance to be %d, got %d", expectedBalance, updatedAccount.Balance)
	}

	ledgerEntries, err := db.Ledger.GetAllByAccount(updatedAccount.Id)
	if err != nil {
		t.Fatalf("Failed to fetch ledger entries: %v", err)
	}

	if len(ledgerEntries) > 1 {
		t.Error("Database state mismatch! Expected 1 ledger entry to be created, but found none.")
	}
}
