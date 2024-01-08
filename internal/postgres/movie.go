package postgres

import (
	"errors"
	"log/slog"

	greenlight "github.com/denpeshkov/greenlight/internal"
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

func (s *MovieService) Movie(id int) (*greenlight.Movie, error) {
	return nil, errors.New("not implemented")
}

func (s *MovieService) CreateMovie(m *greenlight.Movie) error {
	return errors.New("not implemented")
}
