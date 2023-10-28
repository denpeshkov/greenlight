package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	greenlight "github.com/denpeshkov/greenlight/internal"
	"github.com/julienschmidt/httprouter"
)

func (s *Server) registerMovieRoutes() {
	s.router.GET("/v1/movies/:id", s.showMovieHandler)
	s.router.HandlerFunc(http.MethodPost, "/v1/movies", s.createMovieHandler)
}

func (s *Server) showMovieHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id, err := strconv.Atoi(p.ByName("id"))
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
