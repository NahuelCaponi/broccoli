package db

import (
	"testing"
)

func TestLedgerTransactions(t *testing.T) {
	t.Parallel()

	// ==== Pre-requisits ====
	user, err := Users.Insert("LedgerTestUser", "pass")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	accountOrigin, err := Accounts.Insert("LedgerOriginAlias", user.Id)
	if err != nil {
		t.Fatalf("Failed to create origin account: %v", err)
	}
	accountDest, err := Accounts.Insert("LedgerDestAlias", user.Id)
	if err != nil {
		t.Fatalf("Failed to create dest account: %v", err)
	}

	
	// ==== Happy path ====
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin db transaction: %v", err)
	}

	amount := int64(10000)
	trx, err := Transactions.Insert(tx, amount, accountOrigin.Id, accountDest.Id, "Ledger Test Transfer")
	if err != nil {
		t.Fatalf("Failed to create transaction header: %v", err)
	}

	debitEntry, err := Ledger.Insert(tx, accountOrigin.Id, -amount, Transaction_TransferOut, trx.Id)
	if err != nil {
		tx.Rollback() // Safety net
		t.Fatalf("Failed to insert debit: %v", err)
	}
	creditEntry, err := Ledger.Insert(tx, accountDest.Id, amount, Transaction_TransferIn, trx.Id)
	if err != nil {
		tx.Rollback()
		t.Fatalf("Failed to insert credit: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	originLedger, _ := Ledger.GetAllByAccount(accountOrigin.Id)
	if len(originLedger) != 1 || originLedger[0].Id != debitEntry.Id {
		t.Errorf("Expected 1 committed debit entry for origin account. Got %d", len(originLedger))
	}

	destLedger, _ := Ledger.GetAllByAccount(accountDest.Id)
	if len(destLedger) != 1 || destLedger[0].Id != creditEntry.Id {
		t.Errorf("Expected 1 committed credit entry for dest account. Got %d", len(destLedger))
	}

	// ==== Error Path ====
	txError, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin disaster transaction: %v", err)
	}

	_, err = Ledger.Insert(txError, accountOrigin.Id, -5000, Transaction_TransferOut, trx.Id)
	if err != nil {
		txError.Rollback()
		t.Fatalf("Disaster debit failed unexpectedly: %v", err)
	}

	// Crash
	err = txError.Rollback()
	if err != nil {
		t.Fatalf("Failed to rollback transaction: %v", err)
	}

	originLedgerAfterCrash, _ := Ledger.GetAllByAccount(accountOrigin.Id)
	if len(originLedgerAfterCrash) != 1 {
		t.Errorf("ROLLBACK FAILED! Expected origin account to still have 1 entry, but found %d.", len(originLedgerAfterCrash))
	}
}
