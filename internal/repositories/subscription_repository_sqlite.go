package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/TheTeemka/telegram_bot_cources/internal/models"
	_ "github.com/mattn/go-sqlite3"
)

type sqliteSubscriptionRepo struct {
	db *sql.DB
}

func NewSQLiteSubscriptionRepo(dbPath string) CourseSubscriptionRepository {
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

	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS subscriptions (
            telegram_id INTEGER NOT NULL,
            course TEXT NOT NULL,
			section TEXT NOT NULL, 
            added_at DATETIME NOT NULL,
            PRIMARY KEY (telegram_id, course, section)
        )
    `)

	if err != nil {
		panic(fmt.Errorf("creating subscriptions table: %w", err))
	}

	return &sqliteSubscriptionRepo{db: db}
}

func (r *sqliteSubscriptionRepo) Subscribe(userID int64, course string, sections []string) error {
	query := `
		INSERT OR REPLACE INTO subscriptions (telegram_id, course, section,  added_at)
        VALUES (?, ?, ?, ?)
    `
	tx, err := r.db.BeginTx(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}

	for _, sect := range sections {
		_, err = tx.Exec(query, userID, course, sect, time.Now())
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("inserting subscription: %w", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}
	return nil
}

func (r *sqliteSubscriptionRepo) UnSubscribe(userID int64, course string) error {
	query := `
		DELETE FROM subscriptions 
		WHERE telegram_id = ? AND course = ?
    `

	_, err := r.db.Exec(query, userID, course, time.Now())
	if err != nil {
		return fmt.Errorf("unsubscring subscription from all sections: %w", err)
	}

	return nil
}

func (r *sqliteSubscriptionRepo) ClearSubscriptions(userID int64) error {
	query := `
		DELETE FROM subscriptions 
		WHERE telegram_id = ? 
    `

	_, err := r.db.Exec(query, userID, time.Now())
	if err != nil {
		return fmt.Errorf("clearing subscription from all cources: %w", err)
	}

	return nil
}

func (r *sqliteSubscriptionRepo) GetSubscriptions(userID int64) ([]models.CourseSubscription, error) {
	rows, err := r.db.Query(`
        SELECT telegram_id, course, section, added_at
        FROM subscriptions
        WHERE telegram_id = ?
        ORDER BY added_at DESC
    `, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []models.CourseSubscription
	for rows.Next() {
		var sub models.CourseSubscription
		err := rows.Scan(&sub.UserID, &sub.Course, &sub.Section, &sub.AddedAt)
		if err != nil {
			return nil, err
		}
		subs = append(subs, sub)
	}

	return subs, nil
}

func (r *sqliteSubscriptionRepo) Close() error {
	return r.db.Close()
}
