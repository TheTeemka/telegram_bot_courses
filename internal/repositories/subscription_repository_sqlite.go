package repositories

import (
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

	// Create table if not exists
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS subscriptions (
            telegram_id INTEGER NOT NULL,
            course TEXT NOT NULL,
            added_at DATETIME NOT NULL,
            PRIMARY KEY (telegram_id, course)
        )
    `)
	if err != nil {
		panic(fmt.Errorf("creating subscriptions table: %w", err))
	}

	return &sqliteSubscriptionRepo{db: db}
}

func (r *sqliteSubscriptionRepo) Subscribe(userID int64, course string) error {
	_, err := r.db.Exec(`
        INSERT OR REPLACE INTO subscriptions (telegram_id, course, added_at)
        VALUES (?, ?, ?)
    `, userID, course, time.Now())

	if err != nil {
		return fmt.Errorf("inserting subscription: %w", err)
	}

	return nil
}

func (r *sqliteSubscriptionRepo) UnSubscribe(userID int64, course string) error {
	result, err := r.db.Exec(`
        DELETE FROM subscriptions
        WHERE telegram_id = ? AND course = ?
    `, userID, course)

	if err != nil {
		return fmt.Errorf("deleting subscription: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("subscription not found")
	}

	return nil
}

func (r *sqliteSubscriptionRepo) GetSubscription(userID int64) []models.CourseSubscription {
	rows, err := r.db.Query(`
        SELECT telegram_id, course, added_at
        FROM subscriptions
        WHERE telegram_id = ?
        ORDER BY added_at DESC
    `, userID)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var subs []models.CourseSubscription
	for rows.Next() {
		var sub models.CourseSubscription
		err := rows.Scan(&sub.UserID, &sub.Course, &sub.AddedAt)
		if err != nil {
			continue
		}
		subs = append(subs, sub)
	}

	return subs
}

func (r *sqliteSubscriptionRepo) Close() error {
	return r.db.Close()
}
