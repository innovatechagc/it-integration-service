package repository

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// PostgresDB wraps the database connection
type PostgresDB struct {
	DB *sql.DB
}

// NewPostgresDB creates a new PostgreSQL database connection
func NewPostgresDB(host, port, user, password, dbname, sslmode string) (*PostgresDB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresDB{DB: db}, nil
}

// Close closes the database connection
func (p *PostgresDB) Close() error {
	return p.DB.Close()
}