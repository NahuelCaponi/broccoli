package auth

import (
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err.Error())
	}
	return string(bytes)
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil && err != bcrypt.ErrMismatchedHashAndPassword {
		panic(err.Error())
	}
	return err != bcrypt.ErrMismatchedHashAndPassword
}
