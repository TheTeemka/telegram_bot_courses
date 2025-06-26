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
            added_at  DATETIME NOT NULL,
			updated_at DATETIME ,
			is_full BOOLEAN DEFAULT FALSE,
            PRIMARY KEY (telegram_id, course, section)
        )
    `)

	if err != nil {
		panic(fmt.Errorf("creating subscriptions table: %w", err))
	}

	return &sqliteSubscriptionRepo{db: db}
}

func (r *sqliteSubscriptionRepo) Subscribe(telegramID int64, course string, sections []string) error {
	query := `
		INSERT OR REPLACE INTO subscriptions (telegram_id, course, section,  added_at)
        VALUES (?, ?, ?, ?)
    `
	tx, err := r.db.BeginTx(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}

	for _, sect := range sections {
		_, err = tx.Exec(query, telegramID, course, sect, time.Now())
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

func (r *sqliteSubscriptionRepo) GetSubscriptions(userID int64) ([]*models.CourseSubscription, error) {
	rows, err := r.db.Query(`
        SELECT telegram_id, course, section, is_full
        FROM subscriptions
        WHERE telegram_id = ?
        ORDER BY added_at DESC
    `, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []*models.CourseSubscription
	for rows.Next() {
		var sub models.CourseSubscription
		err := rows.Scan(&sub.TelegramID, &sub.Course, &sub.Section, &sub.IsFull)
		if err != nil {
			return nil, err
		}
		subs = append(subs, &sub)
	}

	return subs, nil
}

func (r *sqliteSubscriptionRepo) GetAll() ([]*models.CourseSubscription, error) {
	rows, err := r.db.Query(`
        SELECT telegram_id, course, section, is_full
        FROM subscriptions
    `)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []*models.CourseSubscription
	for rows.Next() {
		var sub models.CourseSubscription
		err := rows.Scan(&sub.TelegramID, &sub.Course, &sub.Section, &sub.IsFull)
		if err != nil {
			return nil, err
		}
		subs = append(subs, &sub)
	}

	return subs, nil
}

func (r *sqliteSubscriptionRepo) Update(sub *models.CourseSubscription) error {
	query := `
        UPDATE subscriptions
        SET updated_at = ?, is_full = ?
        WHERE telegram_id = ? AND course = ? AND section = ?
    `

	_, err := r.db.Exec(query,
		time.Now(),
		sub.IsFull,
		sub.TelegramID,
		sub.Course,
		sub.Section,
	)
	return err
}

func (r *sqliteSubscriptionRepo) Close() error {
	return r.db.Close()
}
