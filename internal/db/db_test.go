package db

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	Connect("broccoli_test")
	defer db.Close()

	os.Exit(m.Run())
}
