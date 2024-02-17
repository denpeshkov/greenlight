package http

import (
	"errors"
	"net/http"

	"github.com/denpeshkov/greenlight/internal/greenlight"
	"github.com/denpeshkov/greenlight/internal/multierr"
)

func (s *Server) registerAuthHandlers() {
	s.router.Handle("POST /v1/auth/token", s.handlerFunc(s.handleCreateToken))
}

// handleCreateToken handles requests to create an authentication token.
func (s *Server) handleCreateToken(w http.ResponseWriter, r *http.Request) (err error) {
	defer multierr.Wrap(&err, "http.Server.handleCreateToken")

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err = s.readRequest(w, r, &req); err != nil {
		return err
	}
	if err = greenlight.PasswordValid(req.Password); err != nil {
		return err
	}

	u, err := s.userService.Get(r.Context(), req.Email)
	if err != nil {
		switch {
		case errors.Is(err, greenlight.ErrNotFound):
			return greenlight.NewUnauthorizedError("Invalid credentials.")
		default:
			return err
		}
	}

	if match, err := u.Password.Matches(req.Password); err != nil {
		return err
	} else if !match {
		return greenlight.NewUnauthorizedError("Invalid credentials.")
	}

	token, err := s.authService.CreateToken(r.Context(), u.ID)
	if err != nil {
		return err
	}

	resp := struct {
		Token string `json:"token"`
	}{
		Token: token,
	}

	if err := s.sendResponse(w, r, http.StatusCreated, resp, nil); err != nil {
		return err
	}
	return nil
}
