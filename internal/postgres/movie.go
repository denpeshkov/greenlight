package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/denpeshkov/greenlight/internal/greenlight"
	"github.com/denpeshkov/greenlight/internal/multierr"
	"github.com/lib/pq"
)

// MovieService represents a service for managing movies backed by PostgreSQL.
type MovieService struct {
	db *DB
}

var _ greenlight.MovieService = (*MovieService)(nil)

// NewMovieService returns a new instance of [MovieService].
func NewMovieService(db *DB) *MovieService {
	return &MovieService{
		db: db,
	}
}

func (s *MovieService) Get(ctx context.Context, id int64) (_ *greenlight.Movie, err error) {
	defer multierr.Wrap(&err, "postgres.MovieService.Get(%d)", id)

	ctx, cancel := context.WithTimeout(ctx, s.db.opts.queryTimeout)
	defer cancel()

	tx, err := s.db.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query := `SELECT id, title, release_date, runtime, genres, version FROM movies WHERE id = $1`
	args := []any{id}
	var m greenlight.Movie
	if err := tx.QueryRowContext(ctx, query, args...).Scan(&m.ID, &m.Title, &m.ReleaseDate, &m.Runtime, pq.Array(&m.Genres), &m.Version); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, greenlight.ErrNotFound
		default:
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &m, nil
}

func (s *MovieService) GetAll(ctx context.Context, filter greenlight.MovieFilter) (_ []*greenlight.Movie, err error) {
	defer multierr.Wrap(&err, "postgres.MovieService.GetAll")

	ctx, cancel := context.WithTimeout(ctx, s.db.opts.queryTimeout)
	defer cancel()

	tx, err := s.db.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

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
	rs, err := tx.QueryContext(ctx, query, filter.Title, pq.Array(filter.Genres), filter.PageSize, (filter.Page-1)*filter.PageSize)
	if err != nil {
		return nil, err
	}
	defer rs.Close()

	var movies []*greenlight.Movie
	for rs.Next() {
		var m greenlight.Movie
		if err := rs.Scan(&m.ID, &m.Title, &m.ReleaseDate, &m.Runtime, pq.Array(&m.Genres), &m.Version); err != nil {
			return nil, err
		}
		movies = append(movies, &m)
	}
	if err := rs.Err(); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return movies, nil
}

func (s *MovieService) Update(ctx context.Context, m *greenlight.Movie) (err error) {
	defer multierr.Wrap(&err, "postgres.MovieService.Update(%d)", m.ID)

	ctx, cancel := context.WithTimeout(ctx, s.db.opts.queryTimeout)
	defer cancel()

	tx, err := s.db.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `UPDATE movies SET (title, release_date, runtime, genres, version) = ($1, $2, $3, $4, version+1) WHERE id = $5 AND version = $6 RETURNING version`
	args := []any{m.Title, m.ReleaseDate, m.Runtime, pq.Array(m.Genres), m.ID, m.Version}
	if err := tx.QueryRowContext(ctx, query, args...).Scan(&m.Version); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return greenlight.NewConflictError("Conflicting change")
		default:
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (s *MovieService) Delete(ctx context.Context, id int64) (err error) {
	defer multierr.Wrap(&err, "postgres.MovieService.Delete(%d)", id)

	ctx, cancel := context.WithTimeout(ctx, s.db.opts.queryTimeout)
	defer cancel()

	tx, err := s.db.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `DELETE FROM movies where id = $1`
	args := []any{id}
	rs, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	n, err := rs.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return greenlight.ErrNotFound
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (s *MovieService) Create(ctx context.Context, m *greenlight.Movie) (err error) {
	defer multierr.Wrap(&err, "postgres.MovieService.Create")

	ctx, cancel := context.WithTimeout(ctx, s.db.opts.queryTimeout)
	defer cancel()

	tx, err := s.db.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `INSERT INTO movies (title, release_date, runtime, genres) VALUES ($1, $2, $3, $4) RETURNING id, version`
	args := []any{m.Title, m.ReleaseDate, m.Runtime, pq.Array(m.Genres)}
	if err := tx.QueryRowContext(ctx, query, args...).Scan(&m.ID, &m.Version); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}
