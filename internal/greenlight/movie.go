package greenlight

import (
	"errors"
	"fmt"
	"time"

	"github.com/denpeshkov/greenlight/internal/multierr"
)

// Movie represents a movie.
type Movie struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	ReleaseDate time.Time `json:"release_date,omitempty"`
	Runtime     int       `json:"runtime,omitempty"`
	Genres      []string  `json:"genres,omitempty"`
}

// Valid returns an error if the validation fails, otherwise nil.
func (m *Movie) Valid() error {
	op := "Movie.Valid"

	var err error
	if m.ID < 0 {
		err = multierr.Join(err, errors.New("ID must be greater than 0"))
	}

	if m.ReleaseDate.IsZero() {
		err = multierr.Join(err, errors.New("release_date must be provided"))
	}
	// Ignore the error since a constant string is used.
	t, _ := time.Parse(time.DateOnly, "1800-01-01")
	if m.ReleaseDate.After(time.Now()) || m.ReleaseDate.Before(t) {
		err = multierr.Join(err, errors.New("release_date must be greater than 1800-01-01 and not in the future"))
	}

	if m.Runtime == 0 {
		err = multierr.Join(err, errors.New("runtime must be provided"))
	}
	if m.Runtime < 0 {
		err = multierr.Join(err, errors.New("runtime must be greater than 0"))
	}

	if len(m.Genres) == 0 {
		err = multierr.Join(err, errors.New("genres must be provided"))
	}

	if err != nil {
		msg := fmt.Sprintf("Movie is invalid. %s", err)
		return &Error{Op: op, Code: ErrInvalid, Msg: msg, Err: err}
	}
	return nil
}

// MovieService is a service for managing movies.
type MovieService interface {
	GetMovie(id int64) (*Movie, error)
	CreateMovie(m *Movie) error
	UpdateMovie(m *Movie) error
	DeleteMovie(id int64) error
}
