package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	greenlight "github.com/denpeshkov/greenlight/internal"
)

func (s *Server) registerMovieRoutes() {
	s.router.HandleFunc("GET /v1/movies/{id}", s.showMovieHandler)
	s.router.HandleFunc("POST /v1/movies", s.createMovieHandler)
}

func (s *Server) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		s.Error(w, r, http.StatusBadRequest, ErrorResponse{Msg: "Incorrect ID parameter format", err: err})
		return
	}

	if err := json.NewEncoder(w).Encode(greenlight.Movie{
		Id:      id,
		Title:   "Movie title",
		Year:    time.Now().Local(),
		Runtime: 0,
		Genres:  []string{"Genre 1", "Genre 2"},
	}); err != nil {
		s.Error(w, r, http.StatusInternalServerError, ErrorResponse{Msg: "Error processing request", err: err})
	}
}

func (s *Server) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "create a new movie")
}
