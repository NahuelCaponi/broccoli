package api

import (
	"broccoli/internal/db"
	"encoding/json"
	"net/http"
	"time"
)

type withdrawalRequest struct {
	Amount      int64  `json:"amount"`
	Destination string `json:"destination"`
	Description string `json:"description"`
}
type withdrawalResponse struct {
	CreatedAt string `json:"created_at"`
}

func (cfg *Router) HandleWithdrawal(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value(userIdKey).(int64)
	if !ok {
		panic("ValidateUser changed UserId but didn't update user")
	}
	var req withdrawalRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	if req.Amount <= 0 {
		// TODO: This may be a front-end bug, but most probably a mallicious attempt
		// Log and do something to user? They do have the JWT to get here
		http.Error(w, "Not enough funds", http.StatusExpectationFailed) // Hide from response that we know it was a hack attempt
		return

	}

	sysAccount, err := db.Accounts.FindByAlias("broccoli")
	if err != nil {
		http.Error(w, "DB not responding", http.StatusInternalServerError)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, "DB not responding", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback() // Safety net

	account, err := db.Accounts.FindForUpdateByUserId(tx, userId)
	if err != nil {
		http.Error(w, "DB not responding", http.StatusInternalServerError)
		return
	}

	if account.Balance < req.Amount {
		http.Error(w, "Not enough funds", http.StatusExpectationFailed)
		return
	}
	account.Balance -= req.Amount

	err = db.Accounts.UpdateBalance(tx, &account)
	if err != nil {
		http.Error(w, "DB not responding", http.StatusInternalServerError)
		return
	}

	// Note: Don't update sysaccount.balance
	// 	that will force a lock in the db for a shared account
	//	sysaccount.balance is instead calculated dynamically by using the ledger.

	trx, err := db.Transactions.Insert(tx, req.Amount, account.Id, sysAccount.Id, req.Description)
	if err != nil {
		http.Error(w, "DB not responding", http.StatusInternalServerError)
		return
	}

	_, err = db.Ledger.Insert(tx, account.Id, -req.Amount, db.Transaction_Withdraw, trx.Id)
	if err != nil {
		http.Error(w, "DB not responding", http.StatusInternalServerError)
		return
	}

	_, err = db.Ledger.Insert(tx, sysAccount.Id, req.Amount, db.Transaction_Withdraw, trx.Id)
	if err != nil {
		http.Error(w, "DB not responding", http.StatusInternalServerError)
		return
	}

	err = tx.Commit()
	if err != nil {
		http.Error(w, "DB not responding", http.StatusInternalServerError)
		return
	}

	writeResponse(w, http.StatusOK, withdrawalResponse{
		CreatedAt: trx.CreatedAt.Format(time.RFC3339)})
}
