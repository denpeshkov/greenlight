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

func (s *MovieService) GetMovie(id int) (*greenlight.Movie, error) {
	ctx := context.Background()
	query := `SELECT id, title, release_date, runtime, genres FROM movies WHERE id = $1`
	args := []any{id}
	var m greenlight.Movie
	if err := s.db.db.QueryRowContext(ctx, query, args...).Scan(&m.ID, &m.Title, &m.ReleaseDate, &m.Runtime, pq.Array(&m.Genres)); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, fmt.Errorf("no movie with id=%d: %w", id, ErrRecordNotFound)
		default:
			return nil, fmt.Errorf("get movie record by id=%d: %w", id, err)
		}
	}
	return &m, nil
}

func (s *MovieService) UpdateMovie(m *greenlight.Movie) error {
	panic("not implemented") // TODO: Implement
}

func (s *MovieService) DeleteMovie(id int) error {
	panic("not implemented") // TODO: Implement
}

func (s *MovieService) CreateMovie(m *greenlight.Movie) error {
	ctx := context.Background()
	query := `INSERT INTO movies (title, release_date, runtime, genres) VALUES ($1, $2, $3, $4) RETURNING id`
	args := []any{m.Title, m.ReleaseDate, m.Runtime, pq.Array(m.Genres)}
	if err := s.db.db.QueryRowContext(ctx, query, args...).Scan(&m.ID); err != nil {
		return fmt.Errorf("insert movie record: %w", err)
	}
	return nil
}
