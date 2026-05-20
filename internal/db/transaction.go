package db

import (
	"database/sql"
	"time"
)

type transactions struct {
}

var Transactions transactions

type Transaction struct {
	Id                 int64
	CreatedAt          time.Time
	Amount             int64
	AccountOrigin      int64
	AccountDestination int64
	Description        string
}

var transactionQuery = "id, created_at, amount, account_origin, account_destination, description"

func populateTransactionWithRow(row RowScanner) (Transaction, error) {
	var transaction Transaction
	err := row.Scan(&transaction.Id, &transaction.CreatedAt, &transaction.Amount, &transaction.AccountOrigin, &transaction.AccountDestination, &transaction.Description)
	return transaction, err
}
func (_ transactions) Insert(tx *sql.Tx, amount, accountOrigin, accountDestination int64, description string) (Transaction, error) {
	transaction := Transaction{
		Amount:             amount,
		AccountOrigin:      accountOrigin,
		AccountDestination: accountDestination,
		Description:        description,
	}

	var result sql.Result
	err := handleRetries(func() error {
		transaction.CreatedAt = time.Now().UTC().Truncate(time.Microsecond)
		var err error
		result, err = tx.Exec("INSERT INTO transactions (created_at, amount, account_origin, account_destination, description) VALUES (?, ?, ?, ?, ?)",
			transaction.CreatedAt, transaction.Amount, transaction.AccountOrigin, transaction.AccountDestination, transaction.Description)
		return err
	})

	if err != nil {
		return transaction, err
	}

	transaction.Id, _ = result.LastInsertId()

	return transaction, err
}

func (_ transactions) GetAllByAccountOrigin(accountOrigin int64) ([]Transaction, error) {
	var transactions []Transaction
	var rows *sql.Rows
	err := handleRetries(func() error {
		var err error
		rows, err = db.Query("SELECT "+transactionQuery+" FROM transactions WHERE account_origin = ?", accountOrigin)
		return err
	})
	if err != nil {
		return transactions, err
	}
	defer rows.Close()

	for rows.Next() {
		u, err := populateTransactionWithRow(rows)
		if err != nil {
			return transactions, err
		}
		transactions = append(transactions, u)
	}

	return transactions, err
}
func (_ transactions) GetAllByAccountDestination(accountDestination int64) ([]Transaction, error) {
	var transactions []Transaction
	var rows *sql.Rows
	err := handleRetries(func() error {
		var err error
		rows, err = db.Query("SELECT "+transactionQuery+" FROM transactions WHERE account_destination = ?", accountDestination)
		return err
	})
	if err != nil {
		return transactions, err
	}
	defer rows.Close()

	for rows.Next() {
		u, err := populateTransactionWithRow(rows)
		if err != nil {
			return transactions, err
		}
		transactions = append(transactions, u)
	}

	return transactions, err
}
