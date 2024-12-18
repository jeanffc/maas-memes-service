package db

import (
    "database/sql"
    "time"

    _ "github.com/mattn/go-sqlite3"
)

const (
    dbPath          = "./maas.db"
    maxOpenConns    = 25
    maxIdleConns    = 25
    connMaxLifetime = 5 * time.Minute
)

func InitDB() (*sql.DB, error) {
    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        return nil, err
    }

    db.SetMaxOpenConns(maxOpenConns)
    db.SetMaxIdleConns(maxIdleConns)
    db.SetConnMaxLifetime(connMaxLifetime)

    createTablesSQL := `
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
    CREATE INDEX IF NOT EXISTS idx_token_balances_client_id ON token_balances(client_id);
    CREATE INDEX IF NOT EXISTS idx_memes_query ON memes(query);
    `

    _, err = db.Exec(createTablesSQL)
    return db, err
}
