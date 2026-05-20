package db

import (
	"database/sql"
	"log"
)

func initialize(db *sql.DB) {
	queries := []string{
		`DROP TABLE IF EXISTS ledger`,
		`DROP TABLE IF EXISTS transactions`,
		`DROP TABLE IF EXISTS accounts`,
		`DROP TABLE IF EXISTS users`,
		`CREATE TABLE users (
					id INTEGER PRIMARY KEY AUTO_INCREMENT,
					created_at TIMESTAMP(6) NOT NULL,
					updated_at TIMESTAMP(6) NOT NULL,
					username VARCHAR(20) NOT NULL,
					hashed_password TEXT NOT NULL,
			
					UNIQUE (username(20)) )`,
		`CREATE TABLE accounts (
			id INTEGER PRIMARY KEY AUTO_INCREMENT,
			created_at TIMESTAMP(6) NOT NULL,
			updated_at TIMESTAMP(6) NOT NULL,
			alias VARCHAR(20) NOT NULL,
			user_id INTEGER NOT NULL,
			balance INTEGER NOT NULL,

			UNIQUE (alias),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE)`,
		`CREATE TABLE transactions (
			id INTEGER PRIMARY KEY AUTO_INCREMENT,
			created_at TIMESTAMP(6) NOT NULL,
			amount INTEGER NOT NULL,
			account_origin INTEGER NOT NULL,
			account_destination INTEGER NOT NULL,
			description TEXT NOT NULL,
			
			FOREIGN KEY (account_origin) REFERENCES accounts(id),
			FOREIGN KEY (account_destination) REFERENCES accounts(id))`,
		`CREATE TABLE ledger (
			id INTEGER PRIMARY KEY AUTO_INCREMENT,
			created_at TIMESTAMP(6) NOT NULL,
			amount INTEGER NOT NULL,
			transaction_type ENUM('deposit', 'withdraw', 'transfer_in', 'transfer_out') NOT NULL,
			account INTEGER NOT NULL,
			transaction_id INTEGER NOT NULL,
			
			FOREIGN KEY (account) REFERENCES accounts(id),
			FOREIGN KEY (transaction_id) REFERENCES transactions(id))`,
	}

	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			log.Println("[Error DB] Initializing: ", query)
			log.Fatal(err)
		}
	}

	sysUser, err := Users.Insert("broccoli", "admin")
	if err != nil {
		log.Fatal("Error creating system user: ", err)
	}

	Accounts.Insert("broccoli", sysUser.Id)

}
