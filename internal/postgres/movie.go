package postgres

import (
	"context"
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
	panic("not implemented") // TODO: Implement
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
