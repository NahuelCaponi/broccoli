package db

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func Connect(databaseName string) {
	dbConnection := os.Getenv("dbConnection")

	var err error
	db, err = sql.Open("mysql", dbConnection+"/"+databaseName+"?parseTime=true")
	if err != nil {
		log.Fatalf("Error preparing database configuration: %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}

	initialize(db)
}

func Close() {
	db.Close()
}

func Begin() (*sql.Tx, error) {
	return db.Begin()
}
