package http

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/denpeshkov/greenlight/internal/greenlight"
)

func (s *Server) registerAuthHandlers() {
	s.router.HandleFunc("POST /v1/auth/token", s.handleCreateToken)
}

// handleCreateToken handles requests to create an authentication token.
func (s *Server) handleCreateToken(w http.ResponseWriter, r *http.Request) {
	op := "http.Server.handleCreateToken"

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := s.readRequest(w, r, &req); err != nil {
		s.Error(w, r, fmt.Errorf("%s: %w", op, err))
		return
	}

	if err := greenlight.PasswordValid(req.Password); err != nil {
		s.Error(w, r, fmt.Errorf("%s: %w", op, err))
		return
	}

	u, err := s.userService.Get(r.Context(), req.Email)
	if err != nil {
		switch {
		case errors.Is(err, greenlight.ErrNotFound):
			s.Error(w, r, greenlight.NewUnauthorizedError("Invalid credentials."))
		default:
			s.Error(w, r, fmt.Errorf("%s: %w", op, err))
		}
		return
	}

	if match, err := u.Password.Matches(req.Password); err != nil {
		s.Error(w, r, fmt.Errorf("%s: %w", op, err))
		return
	} else if !match {
		s.Error(w, r, greenlight.NewUnauthorizedError("Invalid credentials."))
		return
	}

	token, err := s.authService.CreateToken(r.Context(), u.ID)
	if err != nil {
		s.Error(w, r, fmt.Errorf("%s: %w", op, err))
		return
	}

	resp := struct {
		Token string `json:"token"`
	}{
		Token: token,
	}

	if err := s.sendResponse(w, r, http.StatusCreated, resp, nil); err != nil {
		s.Error(w, r, fmt.Errorf("%s: %w", op, err))
		return
	}
}
