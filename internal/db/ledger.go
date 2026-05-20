package db

import (
	"database/sql"
	"time"
)

type ledger struct{}

var Ledger ledger

type TransactionType string

const (
	Transaction_Deposit     TransactionType = "deposit"
	Transaction_Withdraw    TransactionType = "withdraw"
	Transaction_TransferIn  TransactionType = "transfer_in"
	Transaction_TransferOut TransactionType = "transfer_out"
)

type LedgerEntry struct {
	Id              int64
	CreatedAt       time.Time
	Amount          int64
	TransactionType TransactionType
	Account         int64
	TransactionId   int64
}

var ledgerQuery = "id, created_at, amount, transaction_type, account, transaction_id"

func populateLedgerEntryWithRow(row RowScanner) (LedgerEntry, error) {
	var entry LedgerEntry
	err := row.Scan(&entry.Id, &entry.CreatedAt, &entry.Amount, &entry.TransactionType, &entry.Account, &entry.TransactionId)
	return entry, err
}

func (_ ledger) Insert(tx *sql.Tx, account, amount int64, transactionType TransactionType, transactionId int64) (LedgerEntry, error) {
	entry := LedgerEntry{
		Amount:          amount,
		TransactionType: transactionType,
		Account:         account,
		TransactionId:   transactionId,
	}

	var result sql.Result
	err := handleRetries(func() error {
		entry.CreatedAt = time.Now().UTC().Truncate(time.Microsecond)
		var err error
		result, err = tx.Exec("INSERT INTO ledger (created_at, amount, transaction_type, account, transaction_id) VALUES (?, ?, ?, ?, ?)",
			entry.CreatedAt, entry.Amount, entry.TransactionType, entry.Account, entry.TransactionId)
		return err
	})

	if err != nil {
		return entry, err
	}

	entry.Id, _ = result.LastInsertId()

	return entry, err
}

func (_ ledger) GetAllByAccount(account int64) ([]LedgerEntry, error) {
	var ledgerList []LedgerEntry
	var rows *sql.Rows
	var err = handleRetries(func() error {
		var err error
		rows, err = db.Query("SELECT "+ledgerQuery+" FROM ledger WHERE account = ?", account)
		return err
	})
	if err != nil {
		return ledgerList, err
	}
	defer rows.Close()

	for rows.Next() {
		u, err := populateLedgerEntryWithRow(rows)
		if err != nil {
			return ledgerList, err
		}
		ledgerList = append(ledgerList, u)
	}

	return ledgerList, err
}

func (_ ledger) GetAgregation() (int64, error) {
	var total int64
	var rows *sql.Rows
	var err = handleRetries(func() error {
		var err error
		rows, err = db.Query("SELECT sum(amount) FROM ledger")
		return err
	})
	if err != nil {
		return -1, err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&total)

		if err != nil {
			return -1, err
		}
	}

	return total, err
}
