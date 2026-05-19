package db

import (
	"database/sql"
	"time"
)

type users struct {
}

var Users users

type User struct {
	Id             int64
	CreatedAt      time.Time
	UpdatedAt      time.Time
	UserName       string
	HashedPassword string
}

var userQuery = "id, created_at, updated_at, username, hashed_password"

func populateUserWithRow(row RowScanner) (User, error) {
	var user User
	err := row.Scan(&user.Id, &user.CreatedAt, &user.UpdatedAt, &user.UserName, &user.HashedPassword)
	return user, err
}

func (_ users) Insert(username, hashedPassword string) (User, error) {
	u := User{UserName: username, HashedPassword: hashedPassword}

	var result sql.Result
	err := handleRetries(func() error {
		u.CreatedAt = time.Now().UTC().Truncate(time.Microsecond)
		u.UpdatedAt = time.Now().UTC().Truncate(time.Microsecond)
		var err error
		result, err = db.Exec("INSERT INTO users (created_at, updated_at, username, hashed_password) VALUES (?, ?, ?, ?)", u.CreatedAt, u.UpdatedAt, u.UserName, u.HashedPassword)

		return err
	})
	if err != nil {
		return u, err
	}

	err = handleRetries(func() error {
		var err error
		u.Id, err = result.LastInsertId()
		return err
	})

	return u, err
}

func (_ users) Update(u *User) error {
	err := handleRetries(func() error {
		u.UpdatedAt = time.Now().UTC().Truncate(time.Microsecond)
		_, err := db.Exec("Update users SET updated_at = ?, username = ?, hashed_password = ? WHERE id = ?", u.UpdatedAt, u.UserName, u.HashedPassword, u.Id)

		return err
	})

	return err
}
func (_ users) FindByUserName(username string) (User, error) {
	var user User
	err := handleRetries(func() error {
		row := db.QueryRow("SELECT "+userQuery+" FROM users WHERE username = ?", username)
		var err error
		user, err = populateUserWithRow(row)
		return err
	})
	return user, err
}

func (_ users) FindByID(id int) (User, error) {
	var user User
	err := handleRetries(func() error {
		row := db.QueryRow("SELECT "+userQuery+" FROM users WHERE id = ?", id)
		var err error
		user, err = populateUserWithRow(row)
		return err
	})
	return user, err
}
func (_ users) GetAll() ([]User, error) {
	var users []User
	var rows *sql.Rows
	var err = handleRetries(func() error {
		var err error
		rows, err = db.Query("SELECT " + userQuery + " FROM users")
		return err
	})
	if err != nil {
		return users, err
	}
	defer rows.Close()

	for rows.Next() {
		u, err := populateUserWithRow(rows)
		if err != nil {
			return users, err
		}
		users = append(users, u)
	}

	return users, err
}
