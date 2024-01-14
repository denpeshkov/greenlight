package http

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/denpeshkov/greenlight/internal/greenlight"
)

func (s *Server) registerMovieHandlers() {
	s.router.HandleFunc("GET /v1/movies/{id}", s.handleMovieGet)
	s.router.HandleFunc("POST /v1/movies", s.handleMovieCreate)
}

// handleMovieGet handles requests to get a specified movie.
func (s *Server) handleMovieGet(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 0 {
		s.Error(w, r, http.StatusBadRequest, ErrorResponse{Msg: "Incorrect ID parameter", err: err})
		return
	}

	m, err := s.MovieService.GetMovie(id)
	if err != nil {
		s.Error(w, r, http.StatusInternalServerError, ErrorResponse{Msg: "Error processing request", err: err})
		return
	}
	if err := s.sendResponse(w, r, http.StatusOK, m, nil); err != nil {
		s.Error(w, r, http.StatusInternalServerError, ErrorResponse{Msg: "Error processing request", err: err})
		return
	}
}

// handleMovieCreate handles requests to create a movie.
func (s *Server) handleMovieCreate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title       string   `json:"title"`
		ReleaseDate Date     `json:"release_date"`
		Runtime     int      `json:"runtime"`
		Genres      []string `json:"genres"`
	}
	if err := s.readRequest(w, r, &req); err != nil {
		s.Error(w, r, http.StatusInternalServerError, ErrorResponse{Msg: "Error processing request", err: err})
		return
	}

	m := &greenlight.Movie{
		Title:       req.Title,
		ReleaseDate: time.Time(req.ReleaseDate),
		Runtime:     req.Runtime,
		Genres:      req.Genres,
	}
	if err := m.Valid(); err != nil {
		s.Error(w, r, http.StatusBadRequest, ErrorResponse{Msg: "Validation failure", err: err})
		return
	}
	if err := s.MovieService.CreateMovie(m); err != nil {
		s.Error(w, r, http.StatusInternalServerError, ErrorResponse{Msg: "Error processing request", err: err})
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", m.ID))
	if err := s.sendResponse(w, r, http.StatusCreated, m, headers); err != nil {
		s.Error(w, r, http.StatusInternalServerError, ErrorResponse{Msg: "Error processing request", err: err})
		return
	}
}

// Date represents a date in the format "YYYY-MM-DD".
type Date time.Time

func (d *Date) UnmarshalJSON(b []byte) error {
	value := strings.Trim(string(b), `"`) //get rid of "
	if value == "" || value == "null" {
		return nil
	}

	t, err := time.Parse(time.DateOnly, value) //parse time
	if err != nil {
		return err
	}
	*d = Date(t) //set result using the pointer
	return nil
}

func (c Date) MarshalJSON() ([]byte, error) {
	return []byte(`"` + time.Time(c).Format(time.DateOnly) + `"`), nil
}
