package auth

import (
	"net/http"
	"strings"
)

func GetBearerToken(headers http.Header) (string, bool) {
	tokenString := headers.Get("Authorization")
	if tokenString == "" {
		return "", false
	}

	parts := strings.Split(tokenString, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", false
	}
	
	return parts[1], true
}