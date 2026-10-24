package user

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Dimonchester/ttk-room-map/backend/internal/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound     = errors.New("user not found")
	ErrEmailTaken   = errors.New("email already taken")
	ErrInvalidRole  = errors.New("invalid user role")
	ErrInvalidInput = errors.New("email and password hash are required")
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, in CreateUserInput) (*User, error) {
	email := strings.ToLower(strings.TrimSpace(in.Email))
	if email == "" || in.PasswordHash == "" {
		return nil, ErrInvalidInput
	}

	role := in.Role
	if role == "" {
		role = RoleOrganizer
	}
	if !role.IsValid() {
		return nil, ErrInvalidRole
	}

	const query = `
		INSERT INTO users (email, password_hash, role)
		VALUES ($1, $2, $3)
		RETURNING id::text, email, password_hash, role, created_at, updated_at
	`

	var u User
	err := r.pool.QueryRow(ctx, query, email, in.PasswordHash, role).Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.Role,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			logger.Warn(logger.CatDB, "попытка создания дубликата пользователя email=%s (код PG 23505)", email)
			return nil, ErrEmailTaken
		}
		logger.Error(logger.CatDB, "ошибка SQL при создании пользователя email=%s: %v", email, err)
		return nil, fmt.Errorf("insert user: %w", err)
	}

	logger.Info(logger.CatDB, "создан пользователь в БД id=%s email=%s role=%s", u.ID, u.Email, u.Role)
	return &u, nil
}

func (r *Repository) GetByEmail(ctx context.Context, email string) (*User, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	const query = `
		SELECT id::text, email, password_hash, role, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var u User
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.Role,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		logger.Error(logger.CatDB, "ошибка SQL при поиске пользователя по email=%s: %v", email, err)
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	return &u, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*User, error) {
	const query = `
		SELECT id::text, email, password_hash, role, created_at, updated_at
		FROM users
		WHERE id = $1::uuid
	`

	var u User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.Role,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		logger.Error(logger.CatDB, "ошибка SQL при поиске пользователя по id=%s: %v", id, err)
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return &u, nil
}
