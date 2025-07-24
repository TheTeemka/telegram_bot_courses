package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
)

func NewSQLiteDB(dbPath string) *sql.DB {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		panic(fmt.Errorf("creating directory for database: %w", err))
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		panic(fmt.Errorf("opening database: %w", err))
	}

	if err := db.Ping(); err != nil {
		panic(fmt.Errorf("pinging database: %w", err))
	}
	return db
}
