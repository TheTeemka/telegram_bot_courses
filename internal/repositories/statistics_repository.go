package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"
)

type StatisticsRepository struct {
	db    *sql.DB
	Stats map[string]int64
}

func NewStatisticsRepository(db *sql.DB) *StatisticsRepository {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS statistics (
			action TEXT , 
			count int64,
			PRIMARY KEY (action)
		)
	`)
	if err != nil {
		panic(err)
	}
	return &StatisticsRepository{
		db:    db,
		Stats: map[string]int64{},
	}
}

func (r *StatisticsRepository) AddOne(action string) {
	r.Stats[action]++
}

func (r *StatisticsRepository) Run(ctx context.Context) {
	ticker := time.NewTicker(6 * time.Hour)

	for {
		select {
		case <-ctx.Done():
			err := r.Upsert()
			if err != nil {
				slog.Error("Failed to upsert statistics", "error", err)
			}
			return
		case <-ticker.C:
			err := r.Upsert()
			if err != nil {
				slog.Error("Failed to upsert statistics", "error", err)
			}
		}
	}
}

func (r *StatisticsRepository) Upsert() error {
	slog.Info("Upserting statistics", "len", len(r.Stats))
	query := `
        INSERT INTO statistics (action, count)
        VALUES ($1, $2)
		ON CONFLICT(action) DO UPDATE SET count = count + $2;
`

	tx, err := r.db.BeginTx(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	for action, count := range r.Stats {
		_, err := tx.Exec(query, action, count)
		if err != nil {
			if err := tx.Rollback(); err != nil {
				return fmt.Errorf("rolling back transaction: %w", err)
			}
			return fmt.Errorf("upserting chat state: %w", err)
		}
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}
	return nil
}

// func (r *StatisticsRepository) GetAll() error {
// 	query := `
// 		SELECT action, count FROM statistics`
// 	rows, err := r.db.Query(query)
// 	if err != nil {
// 		return fmt.Errorf("querying statistics: %w", err)
// 	}
// 	defer rows.Close()

// 	stats := make(map[string]int64)
// 	var action string
// 	var count int64
// 	for rows.Next() {
// 		if err := rows.Scan(&action, &count); err != nil {
// 			return fmt.Errorf("scanning row: %w", err)
// 		}
// 		stats[action] = count
// 	}

// 	if err := rows.Err(); err != nil {
// 		return fmt.Errorf("error in rows: %w", err)
// 	}
// 	r.Stats = stats
// 	return nil
// }
