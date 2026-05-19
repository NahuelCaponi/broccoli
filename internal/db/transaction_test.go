package db

import (
	"testing"
)

func TestTransactionLifecycle(t *testing.T) {
	t.Parallel()

	// ==== Prerequisites ====
	originUser, err := Users.Insert("TxOriginUser", "pass")
	if err != nil {
		t.Fatalf("Prerequisite failed: could not create origin user: %v", err)
	}
	originAccount, err := Accounts.Insert("TxOriginAlias", originUser.Id)
	if err != nil {
		t.Fatalf("Prerequisite failed: could not create origin account: %v", err)
	}

	destUser, err := Users.Insert("TxDestUser", "pass")
	if err != nil {
		t.Fatalf("Prerequisite failed: could not create dest user: %v", err)
	}
	destAccount, err := Accounts.Insert("TxDestAlias", destUser.Id)
	if err != nil {
		t.Fatalf("Prerequisite failed: could not create dest account: %v", err)
	}

	amount := int64(5000) // Represeting $50.00
	desc := "Last dinner"


	// ==== Inserting ====
	session, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin db transaction: %v", err)
	}

	transaction, err := Transactions.Insert(session, amount, originAccount.Id, destAccount.Id, desc)
	if err != nil {
		t.Fatalf("Error inserting transaction: %v", err)
	}

	// Needed for ledger
	if transaction.Id <= 0 {
		t.Fatalf("Transaction was inserted but failed to capture LastInsertId. Got: %d", transaction.Id)
	}

	err = session.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}
	
	// ==== Check the get alls ====
	originTxs, err := Transactions.GetAllByAccountOrigin(originAccount.Id)
	if err != nil {
		t.Fatalf("Error fetching by origin: %v", err)
	}
	if len(originTxs) != 1 {
		t.Fatalf("Expected 1 transaction for origin, got %d", len(originTxs))
	}
	if originTxs[0].Id != transaction.Id || originTxs[0].Amount != amount {
		t.Errorf("Origin transaction data mismatch. Expected ID %d and Amount %d", transaction.Id, amount)
	}

	// Check Destination
	destTxs, err := Transactions.GetAllByAccountDestination(destAccount.Id)
	if err != nil {
		t.Fatalf("Error fetching by destination: %v", err)
	}
	if len(destTxs) != 1 {
		t.Fatalf("Expected 1 transaction for destination, got %d", len(destTxs))
	}
	if destTxs[0].Description != desc {
		t.Errorf("Destination transaction description mismatch. Expected '%s', got '%s'", desc, destTxs[0].Description)
	}
}