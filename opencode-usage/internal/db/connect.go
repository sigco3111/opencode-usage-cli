package db

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

func Connect(dbPath string) (*sql.DB, error) {
	dsn := fmt.Sprintf("file:%s?mode=ro&_busy_timeout=5000", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func Close(db *sql.DB) {
	if db != nil {
		db.Close()
	}
}
