package http

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/denpeshkov/greenlight/internal/greenlight"
)

func (s *Server) registerMovieHandlers() {
	s.router.Handle("GET /v1/movies/{id}", s.authenticate(http.HandlerFunc(s.handleMovieGet)))
	s.router.Handle("GET /v1/movies", s.authenticate(http.HandlerFunc(s.handleMoviesGet)))
	s.router.Handle("POST /v1/movies", s.authenticate(http.HandlerFunc(s.handleMovieCreate)))
	s.router.Handle("PATCH /v1/movies/{id}", s.authenticate(http.HandlerFunc(s.handleMovieUpdate)))
	s.router.Handle("DELETE /v1/movies/{id}", s.authenticate(http.HandlerFunc(s.handleMovieDelete)))
}

// handleMovieGet handles requests to get a specified movie.
func (s *Server) handleMovieGet(w http.ResponseWriter, r *http.Request) {
	op := "http.Server.handleMovieGet"

	idRaw := r.PathValue("id")
	id, err := strconv.ParseInt(idRaw, 10, 64)
	if err != nil || id < 0 {
		s.Error(w, r, greenlight.NewInvalidError(`Invalid "ID" parameter format: %s`, idRaw))
		return
	}

	m, err := s.movieService.Get(r.Context(), id)
	if err != nil {
		s.Error(w, r, fmt.Errorf("%s: %w", op, err))
		return
	}

	resp := struct {
		ID          int64     `json:"id"`
		Title       string    `json:"title"`
		ReleaseDate time.Time `json:"release_date,omitempty"`
		Runtime     int       `json:"runtime,omitempty"`
		Genres      []string  `json:"genres,omitempty"`
	}{
		ID:          m.ID,
		Title:       m.Title,
		ReleaseDate: m.ReleaseDate,
		Runtime:     m.Runtime,
		Genres:      m.Genres,
	}

	if err := s.sendResponse(w, r, http.StatusOK, resp, nil); err != nil {
		s.Error(w, r, fmt.Errorf("%s: %w", op, err))
		return
	}
}

// handleMoviesGet handles requests to get movies based on provided filter parameters.
func (s *Server) handleMoviesGet(w http.ResponseWriter, r *http.Request) {
	op := "http.Server.handleMoviesGet"

	filter := greenlight.MovieFilter{
		Title:    "",
		Genres:   []string{},
		Page:     1,
		PageSize: 20,
		Sort:     "id",
	}

	vs := r.URL.Query()

	filter.Title = vs.Get("title")
	if vs.Has("genres") {
		filter.Genres = strings.Split(vs.Get("genres"), ",")
	}
	if vs.Has("page") {
		pageRaw := vs.Get("page")
		page, err := strconv.Atoi(pageRaw)
		if err != nil {
			s.Error(w, r, greenlight.NewInvalidError(`Invalid "page" parameter format: %s`, pageRaw))
			return
		}
		filter.Page = page
	}
	if vs.Has("page_size") {
		pageSzRaw := vs.Get("page_size")
		pageSz, err := strconv.Atoi(pageSzRaw)
		if err != nil {
			s.Error(w, r, greenlight.NewInvalidError(`Invalid "page_size" parameter format: %s`, pageSzRaw))
			return
		}
		filter.PageSize = pageSz
	}
	if vs.Has("sort") {
		filter.Sort = vs.Get("sort")
	}

	if err := filter.Valid(); err != nil {
		s.Error(w, r, err)
		return
	}

	movies, err := s.movieService.GetAll(r.Context(), filter)
	if err != nil {
		s.Error(w, r, fmt.Errorf("%s: %w", op, err))
		return
	}

	type respEl struct {
		ID          int64     `json:"id"`
		Title       string    `json:"title"`
		ReleaseDate time.Time `json:"release_date,omitempty"`
		Runtime     int       `json:"runtime,omitempty"`
		Genres      []string  `json:"genres,omitempty"`
	}
	resp := make([]*respEl, len(movies))
	for i, m := range movies {
		resp[i] = &respEl{
			ID:          m.ID,
			Title:       m.Title,
			ReleaseDate: m.ReleaseDate,
			Runtime:     m.Runtime,
			Genres:      m.Genres,
		}
	}

	if err := s.sendResponse(w, r, http.StatusOK, resp, nil); err != nil {
		s.Error(w, r, fmt.Errorf("%s: %w", op, err))
		return
	}
}

// handleMovieCreate handles requests to create a movie.
func (s *Server) handleMovieCreate(w http.ResponseWriter, r *http.Request) {
	op := "http.Server.handleMovieCreate"

	var req struct {
		Title       string   `json:"title"`
		ReleaseDate date     `json:"release_date"`
		Runtime     int      `json:"runtime"`
		Genres      []string `json:"genres"`
	}
	if err := s.readRequest(w, r, &req); err != nil {
		s.Error(w, r, fmt.Errorf("%s: %w", op, err))
		return
	}

	m := &greenlight.Movie{
		Title:       req.Title,
		ReleaseDate: time.Time(req.ReleaseDate),
		Runtime:     req.Runtime,
		Genres:      req.Genres,
	}
	if err := m.Valid(); err != nil {
		s.Error(w, r, err)
		return
	}
	if err := s.movieService.Create(r.Context(), m); err != nil {
		s.Error(w, r, fmt.Errorf("%s: %w", op, err))
		return
	}

	resp := struct {
		ID          int64     `json:"id"`
		Title       string    `json:"title"`
		ReleaseDate time.Time `json:"release_date,omitempty"`
		Runtime     int       `json:"runtime,omitempty"`
		Genres      []string  `json:"genres,omitempty"`
	}{
		ID:          m.ID,
		Title:       m.Title,
		ReleaseDate: m.ReleaseDate,
		Runtime:     m.Runtime,
		Genres:      m.Genres,
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", resp.ID))
	if err := s.sendResponse(w, r, http.StatusCreated, resp, headers); err != nil {
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

	m, err := s.movieService.Get(r.Context(), id)
	if err != nil {
		s.Error(w, r, fmt.Errorf("%s: %w", op, err))
		return
	}

	// use pointers to allow partial updates
	var req struct {
		Title       *string  `json:"title"`
		ReleaseDate *date    `json:"release_date"`
		Runtime     *int     `json:"runtime"`
		Genres      []string `json:"genres"`
	}
	if err := s.readRequest(w, r, &req); err != nil {
		s.Error(w, r, fmt.Errorf("%s: %w", op, err))
		return
	}

	if req.Title != nil {
		m.Title = *req.Title
	}
	if req.ReleaseDate != nil {
		m.ReleaseDate = time.Time(*req.ReleaseDate)
	}
	if req.Runtime != nil {
		m.Runtime = *req.Runtime
	}
	if req.Genres != nil {
		m.Genres = req.Genres
	}

	if err := m.Valid(); err != nil {
		s.Error(w, r, err)
		return
	}
	if err := s.movieService.Update(r.Context(), m); err != nil {
		s.Error(w, r, fmt.Errorf("%s: %w", op, err))
		return
	}

	resp := struct {
		ID          int64     `json:"id"`
		Title       string    `json:"title"`
		ReleaseDate time.Time `json:"release_date,omitempty"`
		Runtime     int       `json:"runtime,omitempty"`
		Genres      []string  `json:"genres,omitempty"`
	}{
		ID:          m.ID,
		Title:       m.Title,
		ReleaseDate: m.ReleaseDate,
		Runtime:     m.Runtime,
		Genres:      m.Genres,
	}

	if err := s.sendResponse(w, r, http.StatusOK, resp, nil); err != nil {
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

	if err := s.movieService.Delete(r.Context(), id); err != nil {
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
