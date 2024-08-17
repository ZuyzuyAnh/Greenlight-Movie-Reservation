package repository

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
)

var db *sql.DB

func TestMain(m *testing.M) {
	code, err := run(m)
	if err != nil {
		fmt.Println(err)
	}
	os.Exit(code)
}

func run(m *testing.M) (code int, err error) {
	db, err = sql.Open("postgres", os.Getenv("GREENLIGHT_DB_DSN"))
	if err != nil {
		return -1, fmt.Errorf("could not connect to database: %w", err)
	}

	defer db.Close()

	return m.Run(), nil
}
