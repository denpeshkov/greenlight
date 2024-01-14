package greenlight

import (
	"errors"
	"time"
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
	if m.ID < 0 {
		return errors.New("incorrect ID")
	}
	t, err := time.Parse(time.DateOnly, "1800-01-01")
	if err != nil {
		return err
	}
	if m.ReleaseDate.After(time.Now()) || m.ReleaseDate.Before(t) {
		return errors.New("incorrect release date")
	}
	if m.Runtime <= 0 {
		return errors.New("incorrect runtime")
	}
	return nil
}

// MovieService is a service for managing movies.
type MovieService interface {
	GetMovie(id int) (*Movie, error)
	CreateMovie(m *Movie) error
	UpdateMovie(m *Movie) error
	DeleteMovie(id int) error
}
