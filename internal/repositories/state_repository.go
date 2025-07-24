package repositories

import (
	"database/sql"
	"fmt"
	"time"
)

type StateRepository interface {
	Upsert(telegram_id int64, state string) error
	GetState(telegram_id int64) (string, error)
}

type stateRepository struct {
	db *sql.DB
}

func NewStateRepository(db *sql.DB) StateRepository {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS chat_states (
			telegram_id INTEGER NOT NULL, 
			state TEXT , 
			created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ,
			updated_at DATETIME ,
			PRIMARY KEY (telegram_id)
		)
	`)
	if err != nil {
		panic(err)
	}
	return &stateRepository{db: db}
}

func (r *stateRepository) Upsert(telegram_id int64, state string) error {
	query := `
		INSERT OR REPLACE INTO chat_states (telegram_id, state, updated_at)
        VALUES (?, ?, ?)`

	_, err := r.db.Exec(query, telegram_id, state, time.Now())
	if err != nil {
		return fmt.Errorf("upserting chat state: %w", err)
	}
	return nil
}

func (r *stateRepository) GetState(telegram_id int64) (string, error) {
	query := `
		SELECT state FROM chat_states 
		WHERE telegram_id = ?`

	var state string
	err := r.db.QueryRow(query, telegram_id).Scan(&state)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("upserting chat state: %w", err)
	}
	return state, nil
}
