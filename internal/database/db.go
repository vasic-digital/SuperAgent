package database

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB interface {
	Ping() error
	Exec(query string, args ...any) error
	Query(query string, args ...any) ([]any, error)
	Close()
}

type PostgresDB struct {
	pool *pgxpool.Pool
}

func (p *PostgresDB) Ping() error {
	return p.pool.Ping(context.Background())
}

func (p *PostgresDB) Exec(query string, args ...any) error {
	_, err := p.pool.Exec(context.Background(), query, args...)
	return err
}

func (p *PostgresDB) Query(query string, args ...any) ([]any, error) {
	rows, err := p.pool.Query(context.Background(), query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []any
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, err
		}
		results = append(results, values)
	}
	return results, nil
}

func (p *PostgresDB) Close() {
	p.pool.Close()
}

// HealthCheck performs a health check on the database.
func (p *PostgresDB) HealthCheck() error {
	return p.Ping()
}

// Connect establishes a real PostgreSQL connection via pgx.
func Connect() (DB, error) {
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "superagent")
	dbPassword := getEnv("DB_PASSWORD", "secret")
	dbName := getEnv("DB_NAME", "superagent_db")

	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &PostgresDB{pool: pool}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
