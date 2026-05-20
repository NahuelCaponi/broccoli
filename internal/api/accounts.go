package api

import (
	"broccoli/internal/db"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type accountRequest struct {
	Alias string `json:"alias"`
}
type accountResponse struct {
	UpdatedAt string `json:"updated_at"`
	Alias     string `json:"alias"`
}

func (cfg *Router) HandleAccountCreate(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value(userIdKey).(int64)
	if !ok {
		panic("ValidateUser changed UserId but didn't update accounts")
	}

	var req accountRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Bad request: Need alias", http.StatusBadRequest)
		return
	}

	_, err = db.Accounts.FindByUserId(userId)
	if err == nil {
		http.Error(w, "User already has an Account", http.StatusConflict)
		return
	}
	if err != sql.ErrNoRows {
		http.Error(w, "DB not responding", http.StatusInternalServerError)
		return
	}

	req.Alias = strings.ToLower(req.Alias)
	if len(req.Alias) > 20 {
		http.Error(w, "Alias should be less than 20 bytes", http.StatusBadRequest)
		return
	}

	account, err := db.Accounts.Insert(req.Alias, userId)
	if err != nil {
		http.Error(w, "DB not responding", http.StatusInternalServerError)
		return
	}

	writeResponse(w, http.StatusCreated, accountResponse{
		Alias:     account.Alias,
		UpdatedAt: account.CreatedAt.Format(time.RFC3339),
	})
}

func (cfg *Router) HandleAccountUpdate(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value(userIdKey).(int64)
	if !ok {
		panic("ValidateUser changed usertId but didn't update accounts")
	}

	var req accountRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Bad request: Need alias", http.StatusBadRequest)
		return
	}

	if len(req.Alias) > 20 {
		http.Error(w, "Alias should be less than 20 bytes", http.StatusBadRequest)
		return
	}

	account, err := db.Accounts.FindByUserId(userId)
	if err != nil {
		http.Error(w, "Account not founded", http.StatusConflict)
		return
	}

	account.Alias = strings.ToLower(req.Alias)

	err = db.Accounts.UpdateAlias(&account)
	if err != nil {
		http.Error(w, "DB not responding", http.StatusInternalServerError)
		return
	}

	writeResponse(w, http.StatusOK, accountResponse{
		UpdatedAt: account.UpdatedAt.Format(time.RFC3339),
		Alias:     account.Alias,
	})
}
