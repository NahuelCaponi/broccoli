package db

import "testing"

func TestUserLifecycle(t *testing.T) {
	t.Parallel()

	type data struct {
		username string
	}
	dataCases := []data{
		{username: "UserDBLifecycle1"},
		{username: "UserDBLifecycle2"},
		{username: "UserDBLifecycle3"},
	}

	users := make([]User, 0)
	for _, dataCases := range dataCases {
		user, err := Users.Insert(dataCases.username, "password")
		if err != nil {
			t.Fatal("Error when creating first user")
		}
		users = append(users, user)
	}
	users[1].UserName = "UserLifecycle2B"
	users[1].HashedPassword = "passwordB"
	err := Users.Update(&users[1])
	if err != nil {
		t.Error("Error updating user: ", err)
	}

	for _, user := range users {
		foundUser, err := Users.FindByUserName(user.UserName)
		if err != nil {
			t.Fatalf("Error finding user %v", user.UserName)
		}
		if foundUser.Id != user.Id {
			t.Errorf("User %v changed id", user.UserName)
		}
		if !foundUser.CreatedAt.Equal(user.CreatedAt) {
			t.Errorf("User %v changed creation date. Expected %v | got %v", user.UserName, foundUser.CreatedAt, user.CreatedAt)
		}
		if !foundUser.UpdatedAt.Equal(user.UpdatedAt) {
			t.Errorf("User %v changed updated date. Expected %v | got %v", user.UserName, foundUser.UpdatedAt, user.UpdatedAt)
		}
		if foundUser.UserName != user.UserName {
			t.Errorf("User %v changed username", user.UserName)
		}
		if foundUser.HashedPassword != user.HashedPassword {
			t.Errorf("User %v changed password", user.UserName)
		}

	}

}
