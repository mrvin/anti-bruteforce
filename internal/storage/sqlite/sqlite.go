package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	// Add SQLite driver without CGO for the database/sql package.
	_ "modernc.org/sqlite"
)

const (
	maxOpenConns    = 1
	maxIdleConns    = 1
	connMaxLifetime = 0
	connMaxIdleTime = 0
)

type Conf struct {
	StoragePath string
}

func New(ctx context.Context, conf *Conf) (*sql.DB, error) {
	db, err := sql.Open("sqlite", conf.StoragePath)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	if err := initSchemaUp(ctx, db); err != nil {
		return nil, fmt.Errorf("db init schema: %w", err)
	}

	// Setting db connections pool.
	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(connMaxLifetime)
	db.SetConnMaxIdleTime(connMaxIdleTime)

	return db, nil
}

func initSchemaUp(ctx context.Context, db *sql.DB) error {
	sqlCreateTableBlacklist := `
	CREATE TABLE IF NOT EXISTS blacklist (
		net TEXT NOT NULL UNIQUE PRIMARY KEY
	)`
	if _, err := db.ExecContext(ctx, sqlCreateTableBlacklist); err != nil {
		return fmt.Errorf("exec create blacklist: %w", err)
	}

	sqlCreateTableWhitelist := `
	CREATE TABLE IF NOT EXISTS whitelist (
		net TEXT NOT NULL UNIQUE PRIMARY KEY
	)`
	if _, err := db.ExecContext(ctx, sqlCreateTableWhitelist); err != nil {
		return fmt.Errorf("exec create whitelist: %w", err)
	}

	return nil
}
