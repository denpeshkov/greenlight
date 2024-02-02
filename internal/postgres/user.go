package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/denpeshkov/greenlight/internal/greenlight"
)

// UserService represents a service for managing users backed by PostgreSQL.
type UserService struct {
	db *DB
}

var _ greenlight.UserService = (*UserService)(nil)

// NewUserService returns a new instance of [UserService].
func NewUserService(db *DB) *UserService {
	return &UserService{
		db: db,
	}
}

func (s *UserService) Get(ctx context.Context, email string) (*greenlight.User, error) {
	op := "postgres.UserService.Get"

	ctx, cancel := context.WithTimeout(ctx, s.db.opts.queryTimeout)
	defer cancel()

	tx, err := s.db.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback()

	query := `SELECT id, name, email, password_hash, activated, version FROM users WHERE email = $1`
	args := []any{email}
	var u greenlight.User
	if err := tx.QueryRowContext(ctx, query, args...).Scan(&u.ID, &u.Name, &u.Password, &u.Activated, &u.Version); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, greenlight.NewNotFoundError("User not found.")
		default:
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &u, nil
}

func (s *UserService) Create(ctx context.Context, u *greenlight.User) error {
	op := "postgres.UserService.Create"

	ctx, cancel := context.WithTimeout(ctx, s.db.opts.queryTimeout)
	defer cancel()

	tx, err := s.db.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback()

	query := `INSERT INTO users (name, email, password_hash, activated) VALUES ($1, $2, $3, $4) RETURNING id, version`
	args := []any{u.ID, u.Email, u.Password, u.Activated}
	if err := tx.QueryRowContext(ctx, query, args...).Scan(&u.ID, &u.Version); err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return greenlight.NewConflictError("A user with this email already exists.")
		default:
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *UserService) Update(ctx context.Context, u *greenlight.User) error {
	op := "postgres.UserService.Update"

	ctx, cancel := context.WithTimeout(ctx, s.db.opts.queryTimeout)
	defer cancel()

	tx, err := s.db.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback()

	query := `UPDATE users SET (name, email, password_hash, activated, version) = ($1, $2, $3, $4, version+1) WHERE id = $5 AND version = $6 RETURNING version`
	args := []any{u.Name, u.Email, u.Password, u.Activated, u.ID, u.Version}
	if err := tx.QueryRowContext(ctx, query, args...).Scan(&u.ID, &u.Version); err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return greenlight.NewConflictError("A user with this email already exists.")
		default:
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
