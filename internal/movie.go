package greenlight

import (
	"errors"
	"time"
)

// Movie represents a movie.
type Movie struct {
	Id          int       `json:"id"`
	Title       string    `json:"title"`
	ReleaseDate time.Time `json:"release_date,omitempty"`
	Runtime     int       `json:"runtime,omitempty"`
	Genres      []string  `json:"genres,omitempty"`
}

// Valid returns an error if the validation fails, otherwise nil.
func (m *Movie) Valid() error {
	if m.Id < 0 {
		return errors.New("incorrect ID")
	}
	if m.ReleaseDate.After(time.Now()) {
		return errors.New("incorrect release date")
	}
	if m.Runtime <= 0 {
		return errors.New("incorrect runtime")
	}
	return nil
}

// MovieService is a service for managing movies.
type MovieService interface {
	Movie(id int) (*Movie, error)
	CreateMovie(m *Movie) error
}
