package api

import (
	"broccoli/internal/auth"
	"broccoli/internal/db"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type userRequest struct {
	UserName string `json:"username"`
	Password string `json:"password"`
}
type userResponse struct {
	UpdatedAt string `json:"updated_at"`
	UserName  string `json:"username"`
}

func (cfg *Router) HandleUserCreate(w http.ResponseWriter, r *http.Request) {
	var req userRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Bad request: Need email and password fields", http.StatusBadRequest)
		return
	}

	req.UserName = strings.ToLower(req.UserName)

	if len([]rune(req.UserName)) > 20 {
		http.Error(w, "username should be less than 20 character", http.StatusBadRequest)
		return
	}
	if len(req.Password) > 72 {
		http.Error(w, "Password should be less than 72 bytes", http.StatusBadRequest)
		return
	}

	prevUser, err := db.Users.FindByUserName(req.UserName)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "DB not responding", http.StatusInternalServerError)
		return
	}
	if prevUser.Id > 0 {
		http.Error(w, "Login to access your account", http.StatusConflict)
		return
	}

	user, err := db.Users.Insert(req.UserName, auth.HashPassword(req.Password))
	if err != nil {
		http.Error(w, "DB not responding", http.StatusInternalServerError)
		return
	}

	writeResponse(w, http.StatusCreated, userResponse{
		UserName:  user.UserName,
		UpdatedAt: user.CreatedAt.Format(time.RFC3339),
	})
}

func (cfg *Router) HandleUserUpdate(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value(userIdKey).(int64)
	if !ok {
		panic("ValidateUser changed UserId but didn't update user")
	}
	var req userRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Bad request: Need username and password fields", http.StatusBadRequest)
		return
	}
	if len([]rune(req.UserName)) > 20 {
		http.Error(w, "username should be less than 20 character", http.StatusBadRequest)
		return
	}
	if len(req.Password) > 72 {
		http.Error(w, "Password should be less than 72 bytes", http.StatusBadRequest)
		return
	}

	user, err := db.Users.FindByID(userId)
	if err != nil {
		http.Error(w, "User not founded", http.StatusConflict)
		return
	}

	user.UserName = strings.ToLower(req.UserName)
	user.HashedPassword = auth.HashPassword(req.Password)

	err = db.Users.Update(&user)
	if err != nil {
		http.Error(w, "DB not responding", http.StatusInternalServerError)
		return
	}

	writeResponse(w, http.StatusOK, userResponse{
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	})
}
