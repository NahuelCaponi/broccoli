package api

import (
	"broccoli/internal/api/webhooks"
	"broccoli/internal/db"
	"bytes"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
)

func TestMain(m *testing.M) {
	db.Connect("broccoli_test")
	defer db.Close()

	os.Exit(m.Run())
}

func TestEndToEnd(t *testing.T) {
	router := Router{JwtSecret: "secret jwt"}
	webhook := webhooks.Webhook{DepositToken: "tokenized"}

	tokens := make([]string, 0)

	// Create 3 users: E2E-N
	for i := range 3 {
		username := "E2E-" + strconv.Itoa(i)
		req := createHttpRequest("", userRequest{UserName: username, Password: "123"})
		rr := httptest.NewRecorder()
		router.HandleUserCreate(rr, req)
		if rr.Code != http.StatusCreated {
			t.Fatalf("Expected 201, got %d", rr.Code)
		}

		req = createHttpRequest("", loginRequest{Username: username, Password: "123"})
		rr = httptest.NewRecorder()
		router.HandleLogin(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("Expected 200, got %d", rr.Code)
		}

		var loginData loginResponse
		json.NewDecoder(rr.Body).Decode(&loginData)
		tokens = append(tokens, loginData.Token)

		rr = httptest.NewRecorder()
		req = createHttpRequest(loginData.Token, accountRequest{Alias: "AccE2E-" + strconv.Itoa(i)})
		router.MiddlewareValidateUser(router.HandleAccountCreate)(rr, req)
		if rr.Code != http.StatusCreated {
			t.Fatalf("Expected 201, got %d", rr.Code)
		}

		user, _ := db.Users.FindByUserName(username)
		rr = httptest.NewRecorder()
		req = createHttpRequest("tokenized", webhooks.DepositRequest{Event: "user.deposit",
			Data: struct {
				UserId      string `json:"user_id"`
				Amount      int    `json:"amount"`
				Description string `json:"description"`
			}{
				UserId:      strconv.FormatInt(user.Id, 10),
				Description: "Deposit",
				Amount:      int(math.Pow(10, float64(i+1)))}})

		webhook.Deposit(rr, req)
		if rr.Code != http.StatusNoContent {
			t.Fatalf("Expected 204, got %d", rr.Code)
		}
	}

	rr := httptest.NewRecorder()
	req := createHttpRequest(tokens[0], transferRequest{Amount: 2, Destination: "AccE2E-1", Description: ""})
	router.MiddlewareValidateUser(router.HandleTransfer)(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", rr.Code)
	}
	rr = httptest.NewRecorder()
	req = createHttpRequest(tokens[0], transferRequest{Amount: 3, Destination: "AccE2E-2", Description: ""})
	router.MiddlewareValidateUser(router.HandleTransfer)(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", rr.Code)
	}

	rr = httptest.NewRecorder()
	req = createHttpRequest(tokens[1], transferRequest{Amount: 10, Destination: "AccE2E-0", Description: ""})
	router.MiddlewareValidateUser(router.HandleTransfer)(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", rr.Code)
	}
	rr = httptest.NewRecorder()
	req = createHttpRequest(tokens[1], transferRequest{Amount: 30, Destination: "AccE2E-2", Description: ""})
	router.MiddlewareValidateUser(router.HandleTransfer)(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", rr.Code)
	}

	rr = httptest.NewRecorder()
	req = createHttpRequest(tokens[2], transferRequest{Amount: 100, Destination: "AccE2E-0", Description: ""})
	router.MiddlewareValidateUser(router.HandleTransfer)(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", rr.Code)
	}
	rr = httptest.NewRecorder()
	req = createHttpRequest(tokens[2], transferRequest{Amount: 200, Destination: "AccE2E-1", Description: ""})
	router.MiddlewareValidateUser(router.HandleTransfer)(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", rr.Code)
	}

	for i := range 3 {
		rr = httptest.NewRecorder()
		req = createHttpRequest(tokens[i], withdrawalRequest{Amount: int64(i + 1), Destination: "Out there", Description: "bye"})
		router.MiddlewareValidateUser(router.HandleWithdrawal)(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("Expected 200, got %d", rr.Code)
		}
	}

	// ==== Check integrity ====

	// Check ledger adds to 0
	aggregation, _ := db.Ledger.GetAgregation()
	if aggregation != 0 {
		t.Errorf("The ledger doesn't add up to zero. Got: %v", aggregation)
	}

	// Check ledger aggregation matches balance
	for i := range 3 {
		account, _ := db.Accounts.FindByAlias("AccE2E-" + strconv.Itoa(i))

		entries, _ := db.Ledger.GetAllByAccount(account.Id)
		var total int64 = 0
		for _, entry := range entries {
			total += entry.Amount
		}
		if total != account.Balance {
			t.Errorf("Ledger and balance doesn't match for account %v. Expected %v | got %v", i, account.Balance, total)
		}
	}

	// Check balance is what was expected:
	// #1 10 deposit - transfer out 2 & 3 + transfer in 10 and 100  - widthdrawl 1 = 114
	account1, _ := db.Accounts.FindByAlias("AccE2E-0")
	if account1.Balance != 114 {
		t.Errorf("Ledger and balance doesn't match for account 1. Expected %v | got %v", 114, account1.Balance)

	}

	// #2 100 deposit - transfer out 10 & 30 + transfer in 2 and 200 - widthdrawl 2  = 260
	account2, _ := db.Accounts.FindByAlias("AccE2E-1")
	if account2.Balance != 260 {
		t.Errorf("Ledger and balance doesn't match for account 2. Expected %v | got %v", 260, account2.Balance)

	}
	// #3 1000 deposit - transfer out 100 & 200 + transfer in 3 and 30 - widthdrawl 3 = 730
	account3, _ := db.Accounts.FindByAlias("AccE2E-2")
	if account3.Balance != 730 {
		t.Errorf("Ledger and balance doesn't match for account 1. Expected %v | got %v", 730, account3.Balance)

	}
}

func createHttpRequest(token string, body any) *http.Request {
	bodyJson, _ := json.Marshal(body)
	request := httptest.NewRequest(http.MethodPost, "/api/users/", bytes.NewBuffer(bodyJson))
	request.Header.Add("Authorization", "Bearer "+token)
	return request
}
