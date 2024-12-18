package db

import (
    "database/sql"
    "fmt"
    "time"

    _ "github.com/mattn/go-sqlite3"
)

const (
    maxOpenConns    = 25
    maxIdleConns    = 25
    connMaxLifetime = 5 * time.Minute
)

func InitDB(dbPath string) (*sql.DB, error) {
    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        return nil, fmt.Errorf("failed to open database: %w", err)
    }

    db.SetMaxOpenConns(maxOpenConns)
    db.SetMaxIdleConns(maxIdleConns)
    db.SetConnMaxLifetime(connMaxLifetime)

    schema := `
    CREATE TABLE IF NOT EXISTS token_balances (
        client_id TEXT PRIMARY KEY,
        balance INTEGER DEFAULT 0,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );
    CREATE TABLE IF NOT EXISTS memes (
        id TEXT PRIMARY KEY,
        url TEXT NOT NULL,
        caption TEXT NOT NULL,
        query TEXT,
        latitude REAL,
        longitude REAL,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );
    `
    _, err = db.Exec(schema)
    if err != nil {
        return nil, fmt.Errorf("failed to create schema: %w", err)
    }

    return db, nil
}
