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

	err := greenlight.NewInvalidError("User is invalid.")
	if req.Password == "" {
		err.AddViolationMsg("Password", "Must be provided.")
	}
	if len(req.Password) < 8 {
		err.AddViolationMsg("Password", "Must be at least 8 characters long.")
	}
	if len(req.Password) > 72 {
		err.AddViolationMsg("Password", "Must not be more than 72 bytes long.")
	}
	if len(err.Violations()) != 0 {
		s.Error(w, r, err)
		return
	}

	u := &greenlight.User{
		Name:      req.Name,
		Email:     req.Email,
		Activated: false,
	}
	pass, errPas := greenlight.NewPassword(req.Password)
	if errPas != nil {
		s.Error(w, r, fmt.Errorf("%s: %w", op, err))
	}
	u.Password = pass

	if err := u.Valid(); err != nil {
		s.Error(w, r, err)
		return
	}
	if err := s.UserService.Create(r.Context(), u); err != nil {
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
