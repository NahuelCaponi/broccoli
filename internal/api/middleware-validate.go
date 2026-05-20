package api

import (
	"broccoli/internal/auth"
	"context"
	"net/http"
	"strings"
)

type contextKey string

var userIdKey contextKey = "userId"

func (cfg *Router) MiddlewareValidateUser(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := auth.GetBearerToken(r.Header)
		if !ok {
			http.Error(w, "Unathorized", http.StatusUnauthorized)
			return
		}

		userId, subject, err := auth.ValidateJWT(token, cfg.JwtSecret)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if subject == EXPIRED_PASSWORD && http.MethodPut == r.Method &&
			strings.Compare(r.RequestURI, "/api/users") != 0 {
			http.Error(w, "Expired Password", http.StatusUpgradeRequired)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, userIdKey, userId)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
