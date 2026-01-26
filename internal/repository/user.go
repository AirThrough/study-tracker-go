package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"study-tracker-backend/internal/models"
)

var ErrNotFound = errors.New("user not found")
var ErrEmailTaken = errors.New("email already in use")

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

type CreateUserInput struct {
	Email        string
	PasswordHash string
	Name         string
}

type UpdateUserInput struct {
	Email *string
	Name  *string
}

func (r *UserRepository) Create(ctx context.Context, input CreateUserInput) (models.User, error) {
	const query = `
		INSERT INTO users (email, password_hash, name)
		VALUES ($1, $2, $3)
		RETURNING id, email, name, deleted_at, created_at, updated_at
	`

	var user models.User
	err := r.pool.QueryRow(ctx, query, input.Email, input.PasswordHash, input.Name).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.DeletedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return models.User{}, ErrEmailTaken
		}
		return models.User{}, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (models.User, error) {
	const query = `
		SELECT id, email, name, deleted_at, created_at, updated_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`

	var user models.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.DeletedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, ErrNotFound
		}
		return models.User{}, fmt.Errorf("get user by id: %w", err)
	}

	return user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (models.User, string, error) {
	const query = `
		SELECT id, email, name, password_hash, deleted_at, created_at, updated_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
	`

	var user models.User
	var passwordHash string
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&passwordHash,
		&user.DeletedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, "", ErrNotFound
		}
		return models.User{}, "", fmt.Errorf("get user by email: %w", err)
	}

	return user, passwordHash, nil
}

func (r *UserRepository) List(ctx context.Context) ([]models.User, error) {
	const query = `
		SELECT id, email, name, deleted_at, created_at, updated_at
		FROM users
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	users := make([]models.User, 0)
	for rows.Next() {
		var user models.User
		if err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Name,
			&user.DeletedAt,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate users: %w", err)
	}

	return users, nil
}

func (r *UserRepository) Update(ctx context.Context, id string, input UpdateUserInput) (models.User, error) {
	const query = `
		UPDATE users
		SET
			email = COALESCE($2, email),
			name = COALESCE($3, name),
			updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, email, name, deleted_at, created_at, updated_at
	`

	var user models.User
	err := r.pool.QueryRow(ctx, query, id, input.Email, input.Name).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.DeletedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, ErrNotFound
		}
		if isUniqueViolation(err) {
			return models.User{}, ErrEmailTaken
		}
		return models.User{}, fmt.Errorf("update user: %w", err)
	}

	return user, nil
}

func (r *UserRepository) SoftDelete(ctx context.Context, id string) error {
	const query = `
		UPDATE users
		SET deleted_at = $2, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.pool.Exec(ctx, query, id, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("soft delete user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
