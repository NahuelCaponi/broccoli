package main

import (
	"broccoli/internal/api"
	"broccoli/internal/api/webhooks"
	"broccoli/internal/db"
	"log"
	"net/http"
	"os"
)

func main() {
	db.Connect("broccoli")
	defer db.Close()

	router := api.Router{JwtSecret: os.Getenv("JWT-Secret")}
	webhook := webhooks.Webhook{DepositToken: os.Getenv("Deposit-Secret")}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", api.HandleHealth)

	mux.HandleFunc("POST /api/login", router.HandleLogin)
	
	mux.HandleFunc("POST /api/users", router.HandleUserCreate)
	mux.HandleFunc("PUT /api/users", router.MiddlewareValidateUser(router.HandleUserUpdate))

	mux.HandleFunc("POST /api/accounts", router.MiddlewareValidateUser(router.HandleAccountCreate))
	mux.HandleFunc("PUT /api/accounts", router.MiddlewareValidateUser(router.HandleAccountUpdate))

	mux.HandleFunc("POST /api/deposit", webhook.Deposit)
	mux.HandleFunc("POST /api/transfer", router.MiddlewareValidateUser(router.HandleTransfer))
	mux.HandleFunc("POST /api/withdrawal", router.MiddlewareValidateUser(router.HandleWithdrawal))

	serv := http.Server{Handler: mux, Addr: ":8080"}

	log.Fatal(serv.ListenAndServe())
}
