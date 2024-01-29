package http

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/denpeshkov/greenlight/internal/greenlight"
)

func (s *Server) registerMovieHandlers() {
	s.router.HandleFunc("GET /v1/movies/{id}", s.handleMovieGet)
	s.router.HandleFunc("POST /v1/movies", s.handleMovieCreate)
	s.router.HandleFunc("PATCH /v1/movies/{id}", s.handleMovieUpdate)
	s.router.HandleFunc("DELETE /v1/movies/{id}", s.handleMovieDelete)
}

// handleMovieGet handles requests to get a specified movie.
func (s *Server) handleMovieGet(w http.ResponseWriter, r *http.Request) {
	op := "http.Server.handleMovieGet"

	idRaw := r.PathValue("id")
	id, err := strconv.ParseInt(idRaw, 10, 64)
	if err != nil || id < 0 {
		s.Error(w, r, greenlight.NewInvalidError("Invalid ID format: %s", idRaw))
		return
	}

	m, err := s.MovieService.GetMovie(r.Context(), id)
	if err != nil {
		s.Error(w, r, fmt.Errorf("%s: %w", op, err))
		return
	}
	if err := s.sendResponse(w, r, http.StatusOK, m, nil); err != nil {
		s.Error(w, r, fmt.Errorf("%s: %w", op, err))
		return
	}
}

// handleMovieCreate handles requests to create a movie.
func (s *Server) handleMovieCreate(w http.ResponseWriter, r *http.Request) {
	op := "http.Server.handleMovieCreate"

	var input struct {
		Title       string   `json:"title"`
		ReleaseDate date     `json:"release_date"`
		Runtime     int      `json:"runtime"`
		Genres      []string `json:"genres"`
	}
	if err := s.readRequest(w, r, &input); err != nil {
		s.Error(w, r, fmt.Errorf("%s: %w", op, err))
		return
	}

	m := &greenlight.Movie{
		Title:       input.Title,
		ReleaseDate: time.Time(input.ReleaseDate),
		Runtime:     input.Runtime,
		Genres:      input.Genres,
	}
	if err := m.Valid(); err != nil {
		s.Error(w, r, err)
		return
	}
	if err := s.MovieService.CreateMovie(r.Context(), m); err != nil {
		s.Error(w, r, fmt.Errorf("%s: %w", op, err))
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", m.ID))
	if err := s.sendResponse(w, r, http.StatusCreated, m, headers); err != nil {
		s.Error(w, r, fmt.Errorf("%s: %w", op, err))
		return
	}
}

// handleMovieUpdate handles requests to update a specified movie.
func (s *Server) handleMovieUpdate(w http.ResponseWriter, r *http.Request) {
	op := "http.Server.handleMovieUpdate"

	idRaw := r.PathValue("id")
	id, err := strconv.ParseInt(idRaw, 10, 64)
	if err != nil || id < 0 {
		s.Error(w, r, greenlight.NewInvalidError("Invalid ID format: %s", idRaw))
		return
	}

	movie, err := s.MovieService.GetMovie(r.Context(), id)
	if err != nil {
		s.Error(w, r, fmt.Errorf("%s: %w", op, err))
	}

	// use pointers to allow partial updates
	var input struct {
		Title       *string  `json:"title"`
		ReleaseDate *date    `json:"release_date"`
		Runtime     *int     `json:"runtime"`
		Genres      []string `json:"genres"`
	}
	if err := s.readRequest(w, r, &input); err != nil {
		s.Error(w, r, fmt.Errorf("%s: %w", op, err))
		return
	}

	if input.Title != nil {
		movie.Title = *input.Title
	}
	if input.ReleaseDate != nil {
		movie.ReleaseDate = time.Time(*input.ReleaseDate)
	}
	if input.Runtime != nil {
		movie.Runtime = *input.Runtime
	}
	if input.Genres != nil {
		movie.Genres = input.Genres
	}

	if err := movie.Valid(); err != nil {
		s.Error(w, r, err)
		return
	}
	if err := s.MovieService.UpdateMovie(r.Context(), movie); err != nil {
		s.Error(w, r, fmt.Errorf("%s: %w", op, err))
		return
	}
	if err := s.sendResponse(w, r, http.StatusOK, movie, nil); err != nil {
		s.Error(w, r, fmt.Errorf("%s: %w", op, err))
		return
	}
}

// handleMovieDelete handles requests to delete a specified movie.
func (s *Server) handleMovieDelete(w http.ResponseWriter, r *http.Request) {
	op := "http.Server.handleMovieDelete"

	idRaw := r.PathValue("id")
	id, err := strconv.ParseInt(idRaw, 10, 64)
	if err != nil || id < 0 {
		s.Error(w, r, greenlight.NewInvalidError("Invalid ID format: %s", idRaw))
		return
	}

	if err := s.MovieService.DeleteMovie(r.Context(), id); err != nil {
		s.Error(w, r, fmt.Errorf("%s: %w", op, err))
		return
	}
	if err := s.sendResponse(w, r, http.StatusNoContent, nil, nil); err != nil {
		s.Error(w, r, fmt.Errorf("%s: %w", op, err))
		return
	}
}

// date represents a date in the format "YYYY-MM-DD".
type date time.Time

func (d *date) UnmarshalJSON(b []byte) error {
	value := string(bytes.Trim(b, `"`)) // get rid of "
	if value == "" || value == "null" {
		return nil
	}

	t, err := time.Parse(time.DateOnly, value) // parse time
	if err != nil {
		return greenlight.NewInvalidError("Invalid date format: %s", value)
	}
	*d = date(t) // set result using the pointer
	return nil
}

func (c date) MarshalJSON() ([]byte, error) {
	return []byte(`"` + time.Time(c).Format(time.DateOnly) + `"`), nil
}
