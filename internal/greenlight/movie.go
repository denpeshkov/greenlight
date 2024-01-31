package greenlight

import (
	"context"
	"time"
)

// Movie represents a movie.
type Movie struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	ReleaseDate time.Time `json:"release_date,omitempty"`
	Runtime     int       `json:"runtime,omitempty"`
	Genres      []string  `json:"genres,omitempty"`
	Version     int32     `json:"version"`
}

// Valid returns an error if the validation fails, otherwise nil.
func (m *Movie) Valid() error {
	err := NewInvalidError("Movie is invalid.")

	if m.ID < 0 {
		err.AddViolationMsg("ID", "Must be greater or equal to 0.")
	}

	if m.ReleaseDate.IsZero() {
		err.AddViolationMsg("release_date", "Must be provided.")
	}
	// Ignore the error since a constant string is used.
	t, _ := time.Parse(time.DateOnly, "1800-01-01")
	if m.ReleaseDate.After(time.Now()) || m.ReleaseDate.Before(t) {
		err.AddViolationMsg("release_date", "Must be greater than 1800-01-01 and not in the future.")
	}

	if m.Runtime == 0 {
		err.AddViolationMsg("runtime", "Must be provided.")
	}
	if m.Runtime < 0 {
		err.AddViolationMsg("runtime", "Must be greater than 0.")
	}

	if len(m.Genres) == 0 {
		err.AddViolationMsg("genres", "Must be provided.")
	}

	if len(err.violations) != 0 {
		return err
	}
	return nil
}

// MovieFilter is a filter used to retrieve movies.
type MovieFilter struct {
	// Title is a title of the movie.
	Title string
	// Genres are genres of the movie.
	Genres []string
	// Page is the number of the page to return.
	Page int
	// PageSize is the size of the page.
	PageSize int
	// Sort is a name of the [Movie] field to sort results on. To sort in descending order, prepend '-' to the field name.
	Sort string
}

func (f *MovieFilter) Valid() error {
	err := NewInvalidError("Movie filter parameter(s) is/are invalid.")

	if f.Page < 1 || f.Page > 10_000_000 {
		err.AddViolationMsg("page", "Must be between 1 and 10_000_000.")
	}

	if f.PageSize < 1 || f.PageSize > 100 {
		err.AddViolationMsg("page_size", "Must be between 1 and 100.")
	}

	switch f.Sort {
	case "id", "title", "release_date", "runtime", "genres", "version":
	case "-id", "-title", "-release_date", "-runtime", "-genres", "-version":
	default:
		err.AddViolationMsg("sort", "Parameter is incorrect.")
	}

	if len(err.violations) != 0 {
		return err
	}
	return nil
}

// MovieService is a service for managing movies.
type MovieService interface {
	GetMovie(ctx context.Context, id int64) (*Movie, error)
	GetMovies(cts context.Context, filter MovieFilter) ([]*Movie, error)
	CreateMovie(ctx context.Context, m *Movie) error
	UpdateMovie(ctx context.Context, m *Movie) error
	DeleteMovie(ctx context.Context, id int64) error
}
