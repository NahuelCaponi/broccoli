package db

import (
	"database/sql"
	"time"
)

type accounts struct {
}

var Accounts accounts

type Account struct {
	Id        int64
	CreatedAt time.Time
	UpdatedAt time.Time
	Alias     string
	Balance   int64
	UserID    int64
}

var accountQuery = "id, created_at, updated_at, alias, user_id, balance"

func populateAccountWithRow(row RowScanner) (Account, error) {
	var account Account
	err := row.Scan(&account.Id, &account.CreatedAt, &account.UpdatedAt, &account.Alias, &account.UserID, &account.Balance)
	return account, err
}

func (_ accounts) Insert(alias string, userId int64) (Account, error) {
	account := Account{Alias: alias, UserID: userId}

	var result sql.Result
	err := handleRetries(func() error {
		account.CreatedAt = time.Now().UTC().Truncate(time.Microsecond)
		account.UpdatedAt = time.Now().UTC().Truncate(time.Microsecond)
		var err error
		result, err = db.Exec("INSERT INTO accounts (created_at, updated_at, alias, user_id, balance) VALUES (?, ?, ?, ?, 0)", account.CreatedAt, account.UpdatedAt, account.Alias, account.UserID)

		return err
	})
	if err != nil {
		return account, err
	}

	err = handleRetries(func() error {
		var err error
		account.Id, err = result.LastInsertId()
		return err
	})

	return account, err
}

func (_ accounts) UpdateAlias(acc *Account) error {
	err := handleRetries(func() error {
		acc.UpdatedAt = time.Now().UTC().Truncate(time.Microsecond)
		_, err := db.Exec("Update accounts SET updated_at = ?, alias = ? WHERE id = ?", acc.UpdatedAt, acc.Alias, acc.Id)

		return err
	})

	return err
}
func (_ accounts) FindByAlias(alias string) (Account, error) {
	var account Account
	err := handleRetries(func() error {
		row := db.QueryRow("SELECT "+accountQuery+" FROM accounts WHERE alias = ?", alias)
		var err error
		account, err = populateAccountWithRow(row)
		return err
	})
	return account, err
}
func (_ accounts) FindByUserId(userId int64) (Account, error) {
	var account Account
	err := handleRetries(func() error {
		row := db.QueryRow("SELECT "+accountQuery+" FROM accounts WHERE user_id = ?", userId)
		var err error
		account, err = populateAccountWithRow(row)
		return err
	})
	return account, err
}

// ==== For transactions ====
func (_ accounts) FindForUpdateByUserId(tx *sql.Tx, userId int64) (Account, error) {
	var account Account
	err := handleRetries(func() error {
		row := tx.QueryRow("SELECT "+accountQuery+" FROM accounts WHERE user_id = ? FOR UPDATE", userId)
		var err error
		account, err = populateAccountWithRow(row)
		return err
	})
	return account, err
}
func (_ accounts) FindForUpdateByAlias(tx *sql.Tx, alias string) (Account, error) {
	var account Account
	err := handleRetries(func() error {
		row := tx.QueryRow("SELECT "+accountQuery+" FROM accounts WHERE alias = ? FOR UPDATE", alias)
		var err error
		account, err = populateAccountWithRow(row)
		return err
	})
	return account, err
}
func (_ accounts) UpdateBalance(tx *sql.Tx, acc *Account) error {
	err := handleRetries(func() error {
		acc.UpdatedAt = time.Now().UTC().Truncate(time.Microsecond)
		_, err := tx.Exec("Update accounts SET updated_at = ?, balance = ? WHERE id = ?", acc.UpdatedAt, acc.Balance, acc.Id)

		return err
	})

	return err
}

func (_ accounts) FindByID(id int) (Account, error) {
	var account Account
	err := handleRetries(func() error {
		row := db.QueryRow("SELECT "+accountQuery+" FROM accounts WHERE id = ?", id)
		var err error
		account, err = populateAccountWithRow(row)
		return err
	})
	return account, err
}
func (_ accounts) GetAll() ([]Account, error) {
	var accounts []Account
	var rows *sql.Rows
	var err = handleRetries(func() error {
		var err error
		rows, err = db.Query("SELECT " + accountQuery + " FROM accounts")
		return err
	})
	if err != nil {
		return accounts, err
	}
	defer rows.Close()

	for rows.Next() {
		u, err := populateAccountWithRow(rows)
		if err != nil {
			return accounts, err
		}
		accounts = append(accounts, u)
	}

	return accounts, err
}
