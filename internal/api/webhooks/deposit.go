package webhooks

import (
	"broccoli/internal/auth"
	"broccoli/internal/db"
	"encoding/json"
	"net/http"
	"strconv"
)

type DepositRequest struct {
	Event string `json:"event"`
	Data  struct {
		UserId      string `json:"user_id"`
		Amount      int    `json:"amount"`
		Description string `json:"description"`
	} `json:"data"`
}

func (ctx *Webhook) Deposit(w http.ResponseWriter, r *http.Request) {
	token, ok := auth.GetBearerToken(r.Header)
	if !ok || token != ctx.DepositToken {
		http.Error(w, "Token doesn't match", http.StatusUnauthorized)
		return
	}

	var req DepositRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if req.Event != "user.deposit" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var userId int64
	userId, err = strconv.ParseInt(req.Data.UserId, 10, 64)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	_, err = db.Users.FindByID(userId)
	if err != nil {
		http.Error(w, "Not Founded", http.StatusNotFound)
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

	amount := int64(req.Data.Amount)
	account.Balance += amount

	err = db.Accounts.UpdateBalance(tx, &account)
	if err != nil {
		http.Error(w, "DB not responding", http.StatusInternalServerError)
		return
	}

	// Note: Don't update sysaccount.balance
	// 	that will force a lock in the db for a shared account
	//	sysaccount.balance is instead calculated dynamically by using the ledger.

	trx, err := db.Transactions.Insert(tx, amount, sysAccount.Id, account.Id, req.Data.Description)
	if err != nil {
		http.Error(w, "DB not responding", http.StatusInternalServerError)
		return
	}

	_, err = db.Ledger.Insert(tx, sysAccount.Id, -amount, db.Transaction_Deposit, trx.Id)
	if err != nil {
		http.Error(w, "DB not responding", http.StatusInternalServerError)
		return
	}

	_, err = db.Ledger.Insert(tx, account.Id, amount, db.Transaction_Deposit, trx.Id)
	if err != nil {
		http.Error(w, "DB not responding", http.StatusInternalServerError)
		return
	}

	err = tx.Commit()
	if err != nil {
		http.Error(w, "DB not responding", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
