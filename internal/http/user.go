package http

import (
	"net/http"

	"github.com/denpeshkov/greenlight/internal/greenlight"
	"github.com/denpeshkov/greenlight/internal/multierr"
)

func (s *Server) registerUserHandlers() {
	s.router.Handle("POST /v1/users", s.handlerFunc(s.handleUserCreate))
}

// handleUserCreate handles requests to create (register) a user.
func (s *Server) handleUserCreate(w http.ResponseWriter, r *http.Request) (err error) {
	defer multierr.Wrap(&err, "http.Server.handleUserCreate")

	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := s.readRequest(w, r, &req); err != nil {
		return err
	}

	u := &greenlight.User{
		Name:  req.Name,
		Email: req.Email,
	}
	pass, err := greenlight.NewPassword(req.Password)
	if err != nil {
		return err
	}
	u.Password = pass

	if err := u.Valid(); err != nil {
		return err
	}
	if err := s.userService.Create(r.Context(), u); err != nil {
		return err
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
		return err
	}
	return nil
}
