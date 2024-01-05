package postgres

import (
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
	s.logger.Error("not implemented")
	return &greenlight.Movie{}, nil
}

func (s *MovieService) CreateMovie(m *greenlight.Movie) error {
	s.logger.Error("not implemented")
	return nil
}
