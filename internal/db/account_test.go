package db

import (
	"testing"
)

func TestAccountLifecycle(t *testing.T) {
	t.Parallel()

	type data struct {
		username string
		alias    string
	}
	dataCases := []data{
		{username: "AccountDBLifecycle1", alias: "AliasLifecycle1"},
		{username: "AccountDBLifecycle2", alias: "AliasLifecycle2"},
		{username: "AccountDBLifecycle3", alias: "AliasLifecycle3"},
	}

	accounts := make([]Account, 0)

	for _, dataCase := range dataCases {
		user, err := Users.Insert(dataCase.username, "password")
		if err != nil {
			t.Fatalf("Failed prerequisite: could not create user %s: %v", dataCase.username, err)
		}

		account, err := Accounts.Insert(dataCase.alias, user.Id)
		if err != nil {
			t.Fatalf("Error when creating account %s: %v", dataCase.alias, err)
		}
		
		if account.Balance != 0 {
			t.Fatalf("New account %s was created with non-zero balance: %d", dataCase.alias, account.Balance)
		}

		accounts = append(accounts, account)
	}

	accounts[1].Alias = "AliasLifecycle2B"
	
	err := Accounts.UpdateAlias(&accounts[1])
	if err != nil {
		t.Fatalf("Error updating account: %v", err)
	}

	for _, acc := range accounts {
		foundAccount, err := Accounts.FindByAlias(acc.Alias)
		if err != nil {
			t.Fatalf("Error finding account by alias %v: %v", acc.Alias, err)
		}

		if foundAccount.Id != acc.Id {
			t.Errorf("Account %v changed id. Expected %d | got %d", acc.Alias, acc.Id, foundAccount.Id)
		}
		if !foundAccount.CreatedAt.Equal(acc.CreatedAt) {
			t.Errorf("Account %v changed creation date. Expected %v | got %v", acc.Alias, acc.CreatedAt, foundAccount.CreatedAt)
		}
		if !foundAccount.UpdatedAt.Equal(acc.UpdatedAt) {
			t.Errorf("Account %v changed updated date. Expected %v | got %v", acc.Alias, acc.UpdatedAt, foundAccount.UpdatedAt)
		}
		if foundAccount.Alias != acc.Alias {
			t.Errorf("Account alias mismatch. Expected %s | got %s", acc.Alias, foundAccount.Alias)
		}
		if foundAccount.UserID != acc.UserID {
			t.Errorf("Account %v changed user_id. Expected %d | got %d", acc.Alias, acc.UserID, foundAccount.UserID)
		}
		if foundAccount.Balance != acc.Balance {
			t.Errorf("Account %v balance changed. Expected %d | got %d", acc.Alias, acc.Balance, foundAccount.Balance)
		}

		foundByID, err := Accounts.FindByID(int(acc.Id))
		if err != nil {
			t.Fatalf("Error finding account by ID %d: %v", acc.Id, err)
		}
		if foundByID.Id != acc.Id {
			t.Errorf("FindByID returned wrong account. Expected ID %d | got %d", acc.Id, foundByID.Id)
		}
	}
}