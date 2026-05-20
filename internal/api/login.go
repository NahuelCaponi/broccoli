package api

import (
	"broccoli/internal/auth"
	"broccoli/internal/db"
	"encoding/json"
	"net/http"
	"time"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type loginResponse struct {
	UpdatedAt string `json:"updated_at"`
	Token     string `json:"token"`
}

var EXPIRED_PASSWORD = "expired_password"

func (cfg *Router) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Bad request: Need username and password fields", http.StatusBadRequest)
		return
	}

	user, err := db.Users.FindByUserName(req.Username)
	if err != nil || !auth.CheckPasswordHash(req.Password, user.HashedPassword) {
		http.Error(w, "Incorrect username or password", http.StatusUnauthorized)
		return
	}

	subject := ""
	if user.UpdatedAt.Add(time.Hour * 24 * 90).Before(time.Now()) {
		subject = EXPIRED_PASSWORD
	}

	token := auth.MakeJWT(user.Id, subject, time.Minute*10, cfg.JwtSecret)
	writeResponse(w, http.StatusOK, loginResponse{
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
		Token:     token,
	})
}
