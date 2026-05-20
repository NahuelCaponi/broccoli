package auth

import (
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func MakeJWT(userID int64, subject string, expiresIn time.Duration, tokenSecret string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "broccoli-access",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		ID:        strconv.FormatInt(userID, 10),
		Subject:   subject,
	})
	jwt, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		panic("Error while signing jwt" + err.Error())
	}

	return jwt
}

func ValidateJWT(tokenString, tokenSecret string) (int64, string, error) {
	claims := jwt.RegisteredClaims{}
	_, err := jwt.ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (any, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return 0, "", err
	}
	id, err := strconv.ParseInt(claims.ID, 10, 64)
	return id, claims.Subject, err
}
