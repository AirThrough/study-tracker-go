package database

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return pool, nil
}

func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	sql, err := os.ReadFile("migrations/001_create_users.sql")
	if err != nil {
		return fmt.Errorf("read migration: %w", err)
	}

	if _, err := pool.Exec(ctx, string(sql)); err != nil {
		return fmt.Errorf("run migration: %w", err)
	}

	return nil
}
