package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrRefreshTokenNotFound = errors.New("refresh token not found")

type RefreshTokenRepository struct {
	pool *pgxpool.Pool
}

func NewRefreshTokenRepository(pool *pgxpool.Pool) *RefreshTokenRepository {
	return &RefreshTokenRepository{pool: pool}
}

func (r *RefreshTokenRepository) Create(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error {
	const query = `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`

	if _, err := r.pool.Exec(ctx, query, userID, tokenHash, expiresAt); err != nil {
		return fmt.Errorf("create refresh token: %w", err)
	}

	return nil
}

func (r *RefreshTokenRepository) GetValid(ctx context.Context, tokenHash string) (string, error) {
	const query = `
		SELECT user_id
		FROM refresh_tokens
		WHERE token_hash = $1 AND expires_at > NOW()
	`

	var userID string
	err := r.pool.QueryRow(ctx, query, tokenHash).Scan(&userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrRefreshTokenNotFound
		}
		return "", fmt.Errorf("get refresh token: %w", err)
	}

	return userID, nil
}

func (r *RefreshTokenRepository) Delete(ctx context.Context, tokenHash string) error {
	const query = `DELETE FROM refresh_tokens WHERE token_hash = $1`

	if _, err := r.pool.Exec(ctx, query, tokenHash); err != nil {
		return fmt.Errorf("delete refresh token: %w", err)
	}

	return nil
}
