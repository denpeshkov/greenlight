package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/denpeshkov/greenlight/internal/greenlight"
	"github.com/lib/pq"
)

// MovieService represents a service for managing movies backed by PostgreSQL.
type MovieService struct {
	db     *DB
	logger *slog.Logger
}

// NewMovieService returns a new instance of [MovieService].
func NewMovieService(db *DB) *MovieService {
	return &MovieService{
		db:     db,
		logger: newLogger(),
	}
}

func (s *MovieService) GetMovie(id int64) (*greenlight.Movie, error) {
	op := "postgres.MovieService.GetMovie"

	// FIXME get context from client
	ctx := context.Background()
	query := `SELECT id, title, release_date, runtime, genres FROM movies WHERE id = $1`
	args := []any{id}
	var m greenlight.Movie
	if err := s.db.db.QueryRowContext(ctx, query, args...).Scan(&m.ID, &m.Title, &m.ReleaseDate, &m.Runtime, pq.Array(&m.Genres)); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, greenlight.NewNotFoundError("Movie with id=%d is not found.", id)
		default:
			return nil, fmt.Errorf("%s: movie with id=%d: %w", op, id, err)
		}
	}

	return &m, nil
}

func (s *MovieService) UpdateMovie(m *greenlight.Movie) error {
	op := "postgres.MovieService.UpdateMovie"

	// FIXME get context from client
	ctx := context.Background()
	query := `UPDATE movies SET (title, release_date, runtime, genres) = ($1, $2, $3, $4) WHERE id = $5`
	args := []any{m.Title, m.ReleaseDate, m.Runtime, pq.Array(m.Genres), m.ID}
	if _, err := s.db.db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("%s: movie with id=%d: %w", op, m.ID, err)
	}
	return nil
}

func (s *MovieService) DeleteMovie(id int64) error {
	op := "postgres.MovieService.DeleteMovie"

	// FIXME get context from client
	ctx := context.Background()
	query := `DELETE FROM movies where id = $1`
	args := []any{id}
	rs, err := s.db.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%s: movie with id=%d: %w", op, id, err)
	}

	n, err := rs.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: movie with id=%d: %w", op, id, err)
	}
	if n == 0 {
		return greenlight.NewNotFoundError("Movie with id=%d is not found.", id)
	}

	return nil
}

func (s *MovieService) CreateMovie(m *greenlight.Movie) error {
	op := "postgres.MovieService.CreateMovie"

	// FIXME get context from client
	ctx := context.Background()
	query := `INSERT INTO movies (title, release_date, runtime, genres) VALUES ($1, $2, $3, $4) RETURNING id`
	args := []any{m.Title, m.ReleaseDate, m.Runtime, pq.Array(m.Genres)}
	if err := s.db.db.QueryRowContext(ctx, query, args...).Scan(&m.ID); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
