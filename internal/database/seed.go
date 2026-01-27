package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"study-tracker-backend/internal/models"
)

const (
	defaultAdminEmail    = "admin@gmail.com"
	defaultAdminPassword = "admin"
	defaultAdminName     = "Admin"
)

func EnsureDefaultAdmin(ctx context.Context, pool *pgxpool.Pool) error {
	var adminCount int
	if err := pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM users WHERE role = $1 AND deleted_at IS NULL
	`, models.RoleAdmin).Scan(&adminCount); err != nil {
		return fmt.Errorf("count admins: %w", err)
	}

	if adminCount > 0 {
		return nil
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(defaultAdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash default admin password: %w", err)
	}

	tag, err := pool.Exec(ctx, `
		UPDATE users
		SET deleted_at = NULL, role = $2, password_hash = $3, updated_at = NOW()
		WHERE email = $1
	`, defaultAdminEmail, models.RoleAdmin, string(passwordHash))
	if err != nil {
		return fmt.Errorf("restore default admin: %w", err)
	}

	if tag.RowsAffected() > 0 {
		return nil
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO users (email, password_hash, name, role)
		VALUES ($1, $2, $3, $4)
	`, defaultAdminEmail, string(passwordHash), defaultAdminName, models.RoleAdmin)
	if err != nil {
		return fmt.Errorf("create default admin: %w", err)
	}

	return nil
}
