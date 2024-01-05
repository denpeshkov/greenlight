package http

import (
	"net/http"
	"strconv"
	"time"

	greenlight "github.com/denpeshkov/greenlight/internal"
)

func (s *Server) registerMovieHandlers() {
	s.router.HandleFunc("GET /v1/movies/{id}", s.handleMovieGet)
	s.router.HandleFunc("POST /v1/movies", s.handleMovieCreate)
}

func (s *Server) handleMovieGet(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 0 {
		s.Error(w, r, http.StatusBadRequest, ErrorResponse{Msg: "Incorrect ID parameter", err: err})
		return
	}

	m, err := s.MovieService.Movie(id)
	if err != nil {
		s.Error(w, r, http.StatusInternalServerError, ErrorResponse{Msg: "Error processing request", err: err})
		return
	}
	if err := s.sendResponse(w, r, http.StatusOK, m); err != nil {
		s.Error(w, r, http.StatusInternalServerError, ErrorResponse{Msg: "Error processing request", err: err})
		return
	}
}

func (s *Server) handleMovieCreate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title   string    `json:"title"`
		Year    time.Time `json:"year"`
		Runtime int       `json:"runtime"`
		Genres  []string  `json:"genres"`
	}
	if err := s.readRequest(w, r, &req); err != nil {
		s.Error(w, r, http.StatusInternalServerError, ErrorResponse{Msg: "Error processing request", err: err})
		return
	}

	m := &greenlight.Movie{
		Title:   req.Title,
		Year:    req.Year,
		Runtime: req.Runtime,
		Genres:  req.Genres,
	}
	if err := m.Valid(); err != nil {
		s.Error(w, r, http.StatusBadRequest, ErrorResponse{Msg: "Validation failure", err: err})
		return
	}
	if err := s.MovieService.CreateMovie(m); err != nil {
		s.Error(w, r, http.StatusInternalServerError, ErrorResponse{Msg: "Error processing request", err: err})
		return
	}
}
