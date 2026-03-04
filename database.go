package config

import (
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // The underscore import registers the Postgres driver as a side effect
)

// ConnectDB reads connection details from environment variables and
// returns a live *sqlx.DB connection pool.
//
// sqlx is a thin wrapper over the standard database/sql package.
// It adds conveniences like scanning rows into structs (used in repository.go)
// and named query parameters (:field_name instead of $1, $2).
//
// The connection pool is shared across all requests — we do NOT open
// a new connection per request. This is important for performance.
func ConnectDB() *sqlx.DB {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", "postgres"),
		getEnv("DB_NAME", "device_inventory"),
	)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Connection pool settings — important to understand for interviews
	// MaxOpenConns: max simultaneous connections to Postgres
	// MaxIdleConns: connections kept alive even when not in use (avoid reconnect cost)
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	log.Println("Database connected successfully")
	return db
}

// getEnv reads an environment variable, returning a fallback if not set.
func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
