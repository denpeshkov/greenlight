package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"

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

func (s *MovieService) GetMovie(ctx context.Context, id int64) (*greenlight.Movie, error) {
	op := "postgres.MovieService.GetMovie"

	ctx, cancel := context.WithTimeout(ctx, s.db.opts.queryTimeout)
	defer cancel()

	query := `SELECT id, title, release_date, runtime, genres, version FROM movies WHERE id = $1`
	args := []any{id}
	var m greenlight.Movie
	if err := s.db.db.QueryRowContext(ctx, query, args...).Scan(&m.ID, &m.Title, &m.ReleaseDate, &m.Runtime, pq.Array(&m.Genres), &m.Version); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, greenlight.NewNotFoundError("Movie with id=%d is not found.", id)
		default:
			return nil, fmt.Errorf("%s: movie with id=%d: %w", op, id, err)
		}
	}

	return &m, nil
}

func (s *MovieService) GetMovies(ctx context.Context, filter greenlight.MovieFilter) ([]*greenlight.Movie, error) {
	op := "postgres.MovieService.GetMovies"

	ctx, cancel := context.WithTimeout(ctx, s.db.opts.queryTimeout)
	defer cancel()

	sortCol, sortDir := filter.Sort, "ASC"
	if v, ok := strings.CutPrefix(sortCol, "-"); ok {
		sortCol = v
		sortDir = "DESC"
	}

	query := fmt.Sprintf(`
		SELECT id, title, release_date, runtime, genres, version 
		FROM movies
		WHERE (LOWER(title) = LOWER($1) OR $1 = '') AND (genres @> $2 OR $2 = '{}')
		ORDER BY %s %s, id ASC
		LIMIT $3 OFFSET $4`, sortCol, sortDir)
	rs, err := s.db.db.QueryContext(ctx, query, filter.Title, pq.Array(filter.Genres), filter.PageSize, (filter.Page-1)*filter.PageSize)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rs.Close()

	var movies []*greenlight.Movie
	for rs.Next() {
		var m greenlight.Movie
		if err := rs.Scan(&m.ID, &m.Title, &m.ReleaseDate, &m.Runtime, pq.Array(&m.Genres), &m.Version); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		movies = append(movies, &m)
	}
	if err := rs.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return movies, nil
}

func (s *MovieService) UpdateMovie(ctx context.Context, m *greenlight.Movie) error {
	op := "postgres.MovieService.UpdateMovie"

	ctx, cancel := context.WithTimeout(ctx, s.db.opts.queryTimeout)
	defer cancel()

	query := `UPDATE movies SET (title, release_date, runtime, genres, version) = ($1, $2, $3, $4, version+1) WHERE id = $5 AND version = $6 RETURNING version`
	args := []any{m.Title, m.ReleaseDate, m.Runtime, pq.Array(m.Genres), m.ID, m.Version}
	if err := s.db.db.QueryRowContext(ctx, query, args...).Scan(&m.Version); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return greenlight.NewConflictError("Conflicting change for the movie with id=%d", m.ID)
		default:
			return fmt.Errorf("%s: movie with id=%d: %w", op, m.ID, err)
		}
	}
	return nil
}

func (s *MovieService) DeleteMovie(ctx context.Context, id int64) error {
	op := "postgres.MovieService.DeleteMovie"

	ctx, cancel := context.WithTimeout(ctx, s.db.opts.queryTimeout)
	defer cancel()

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

func (s *MovieService) CreateMovie(ctx context.Context, m *greenlight.Movie) error {
	op := "postgres.MovieService.CreateMovie"

	ctx, cancel := context.WithTimeout(ctx, s.db.opts.queryTimeout)
	defer cancel()

	query := `INSERT INTO movies (title, release_date, runtime, genres) VALUES ($1, $2, $3, $4) RETURNING id, version`
	args := []any{m.Title, m.ReleaseDate, m.Runtime, pq.Array(m.Genres)}
	if err := s.db.db.QueryRowContext(ctx, query, args...).Scan(&m.ID, &m.Version); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
