package api

import (
	"broccoli/internal/db"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

type transferRequest struct {
	Amount      int64  `json:"amount"`
	Destination string `json:"destination"`
	Description string `json:"description"`
}
type transferResponse struct {
	CreatedAt string `json:"created_at"`
}

func (cfg *Router) HandleTransfer(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value(userIdKey).(int64)
	if !ok {
		panic("ValidateUser changed UserId but didn't update user")
	}
	var req transferRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if req.Amount <= 0 {
		// TODO: This may be a front-end bug, but most probably a mallicious attempt
		// Log and do something to user?
		http.Error(w, "Not enough funds", http.StatusExpectationFailed) // Hide from response that we know it was a hack attempt
		return
	}

	_, err = db.Users.FindByID(userId)
	if err != nil {
		http.Error(w, "Not Founded", http.StatusNotFound)
		return
	}

	// Get accounts before locking
	accountPrelock, err := db.Accounts.FindByUserId(userId)
	if err != nil {
		http.Error(w, "DB not responding", http.StatusInternalServerError)
		return
	}
	if accountPrelock.Alias == req.Destination {
		// Try to transfer to self
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	destAccountPreLock, err := db.Accounts.FindByAlias(req.Destination)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Bad request", http.StatusBadRequest)
		} else {
			http.Error(w, "DB not responding", http.StatusInternalServerError)
		}
		return
	}

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, "DB not responding", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback() // Safety net

	// Check order to lock
	var originAccount db.Account
	var destAccount db.Account
	// we lock always the one with lower id to avoid deadlock in case they transfer to each other at the same time
	if accountPrelock.Id < destAccountPreLock.Id {
		originAccount, err = db.Accounts.FindForUpdateByUserId(tx, userId)
		if err != nil {
			http.Error(w, "DB not responding", http.StatusInternalServerError)
			return
		}

		destAccount, err = db.Accounts.FindForUpdateByAlias(tx, req.Destination)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Bad request", http.StatusBadRequest)
			} else {
				http.Error(w, "DB not responding", http.StatusInternalServerError)
			}
			return
		}
	} else {
		destAccount, err = db.Accounts.FindForUpdateByAlias(tx, req.Destination)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Bad request", http.StatusBadRequest)
			} else {
				http.Error(w, "DB not responding", http.StatusInternalServerError)
			}
			return
		}
		originAccount, err = db.Accounts.FindForUpdateByUserId(tx, userId)
		if err != nil {
			http.Error(w, "DB not responding", http.StatusInternalServerError)
			return
		}
	}

	if originAccount.Balance < req.Amount {
		http.Error(w, "Not enough funds", http.StatusExpectationFailed)
		return
	}
	originAccount.Balance -= req.Amount
	destAccount.Balance += req.Amount

	// Update balances
	err = db.Accounts.UpdateBalance(tx, &originAccount)
	if err != nil {
		http.Error(w, "DB not responding", http.StatusInternalServerError)
		return
	}
	err = db.Accounts.UpdateBalance(tx, &destAccount)
	if err != nil {
		http.Error(w, "DB not responding", http.StatusInternalServerError)
		return
	}

	trx, err := db.Transactions.Insert(tx, req.Amount, destAccount.Id, originAccount.Id, req.Description)
	if err != nil {
		http.Error(w, "DB not responding", http.StatusInternalServerError)
		return
	}

	// Double entry on ledger
	_, err = db.Ledger.Insert(tx, originAccount.Id, -req.Amount, db.Transaction_TransferOut, trx.Id)
	if err != nil {
		http.Error(w, "DB not responding", http.StatusInternalServerError)
		return
	}
	_, err = db.Ledger.Insert(tx, destAccount.Id, req.Amount, db.Transaction_TransferIn, trx.Id)
	if err != nil {
		http.Error(w, "DB not responding", http.StatusInternalServerError)
		return
	}

	err = tx.Commit()
	if err != nil {
		http.Error(w, "DB not responding", http.StatusInternalServerError)
		return
	}

	writeResponse(w, http.StatusOK, transferResponse{
		CreatedAt: trx.CreatedAt.Format(time.RFC3339)})
}
