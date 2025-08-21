package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func Connect(connStr string) error {
	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	if err := DB.Ping(); err != nil {
		return err
	}
	fmt.Println("PostgreSQL connected!")
	return nil
}
