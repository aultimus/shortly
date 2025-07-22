package db

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

type PostgresDB struct {
	db *sql.DB
}

func NewPostgresDB(connStr string) (*PostgresDB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres connection: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}
	return &PostgresDB{db: db}, nil
}

func (p *PostgresDB) Create(key string, value *StoredURL) error {
	_, err := p.db.Exec(`INSERT INTO urls (id, original_url, created_at) VALUES ($1, $2, now()) ON CONFLICT (id) DO NOTHING`, key, value.OriginalURL)
	if err != nil {
		return NewErrDB(fmt.Sprintf("postgres insert error: %v", err))
	}
	return nil
}

func (p *PostgresDB) Get(key string) (*StoredURL, error) {
	var originalURL string
	err := p.db.QueryRow(`SELECT original_url FROM urls WHERE id = $1`, key).Scan(&originalURL)
	if err == sql.ErrNoRows {
		return nil, NewErrNotFound(fmt.Sprintf("could not find key %s", key))
	}
	if err != nil {
		return nil, NewErrDB(fmt.Sprintf("postgres select error: %v", err))
	}
	return &StoredURL{OriginalURL: originalURL}, nil
}

func (p *PostgresDB) Close() error {
	return p.db.Close()
}
