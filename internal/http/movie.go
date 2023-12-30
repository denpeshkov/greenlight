package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	greenlight "github.com/denpeshkov/greenlight/internal"
)

func (s *Server) registerMovieHandlers() {
	s.router.HandleFunc("GET /v1/movies/{id}", s.handleMoveGet)
	s.router.HandleFunc("POST /v1/movies", s.handleMovieCreate)
}

func (s *Server) handleMoveGet(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 0 {
		s.Error(w, r, http.StatusBadRequest, ErrorResponse{Msg: "Incorrect ID parameter", err: err})
		return
	}

	if err := json.NewEncoder(w).Encode(greenlight.Movie{
		Id:      id,
		Title:   "Title",
		Year:    time.Now().Local(),
		Runtime: 120,
		Genres:  []string{"Genre 1", "Genre 2"},
	}); err != nil {
		s.Error(w, r, http.StatusInternalServerError, ErrorResponse{Msg: "Error processing request", err: err})
	}
}

func (s *Server) handleMovieCreate(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Create a new movie")
}
