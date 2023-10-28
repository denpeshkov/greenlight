package greenlight

import "time"

// Movie represents a movie.
type Movie struct {
	Id      int       `json:"id"`
	Title   string    `json:"title"`
	Year    time.Time `json:"year,omitempty"`
	Runtime int       `json:"runtime,omitempty"`
	Genres  []string  `json:"genres,omitempty"`
}
