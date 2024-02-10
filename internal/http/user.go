package http

import (
	"fmt"
	"net/http"

	"github.com/denpeshkov/greenlight/internal/greenlight"
)

func (s *Server) registerUserHandlers() {
	s.router.HandleFunc("POST /v1/users", s.handleUserCreate)
}

// handleUserCreate handles requests to create (register) a user.
func (s *Server) handleUserCreate(w http.ResponseWriter, r *http.Request) {
	op := "http.Server.handleUserCreate"

	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := s.readRequest(w, r, &req); err != nil {
		s.Error(w, r, fmt.Errorf("%s: %w", op, err))
		return
	}

	u := &greenlight.User{
		Name:  req.Name,
		Email: req.Email,
	}
	pass, err := greenlight.NewPassword(req.Password)
	if err != nil {
		s.Error(w, r, fmt.Errorf("%s: %w", op, err))
		return
	}
	u.Password = pass

	if err := u.Valid(); err != nil {
		s.Error(w, r, err)
		return
	}
	if err := s.userService.Create(r.Context(), u); err != nil {
		s.Error(w, r, fmt.Errorf("%s: %w", op, err))
		return
	}

	resp := struct {
		ID    int64  `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}{
		ID:    u.ID,
		Name:  u.Name,
		Email: u.Email,
	}

	if err := s.sendResponse(w, r, http.StatusCreated, resp, nil); err != nil {
		s.Error(w, r, fmt.Errorf("%s: %w", op, err))
		return
	}
}
