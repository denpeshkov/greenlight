package greenlight

import (
	"context"
	"errors"
	"net/mail"
	"unicode/utf8"

	"github.com/denpeshkov/greenlight/internal/multierr"
	"golang.org/x/crypto/bcrypt"
)

// User represents a user.
type User struct {
	ID       int64    `json:"id"`
	Name     string   `json:"name"`
	Email    string   `json:"email"`
	Password Password `json:"-"`
	Version  int      `json:"-"`
}

// Valid returns an error if the validation fails, otherwise nil.
func (u *User) Valid() error {
	err := NewInvalidError("User is invalid.")

	if u.ID < 0 {
		err.AddViolationMsg("ID", "Must be greater or equal to 0.")
	}

	if u.Name == "" {
		err.AddViolationMsg("Name", "Must be provided.")
	}
	if utf8.RuneCount([]byte(u.Name)) > 500 {
		err.AddViolationMsg("Name", "Must not be more than 500 characters long.")
	}

	if u.Email == "" {
		err.AddViolationMsg("Email", "Must be provided.")
	}
	if _, e := mail.ParseAddress(u.Email); e != nil {
		err.AddViolationMsg("Email", "Is invalid.")
	}

	if len(u.Password) == 0 {
		err.AddViolationMsg("Password", "Must be provided.")
	}

	return nil
}

// Password represents a hash of the user password.
type Password []byte

// NewPasswords generates a hashed password from the plaintext password.
func NewPassword(plaintext string) (_ Password, err error) {
	defer multierr.Wrap(&err, "greenlight.NewPassword")

	if err := PasswordValid(plaintext); err != nil {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), 12)
	if err != nil {
		return nil, err
	}
	return hash, nil
}

// PasswordValid validates a plaintext password.
func PasswordValid(plaintext string) error {
	err := NewInvalidError("Password is invalid.")
	if plaintext == "" {
		err.AddViolationMsg("Password", "Must be provided.")
	}
	if len(plaintext) < 8 {
		err.AddViolationMsg("Password", "Must be at least 8 bytes long.")
	}
	if len(plaintext) > 72 {
		err.AddViolationMsg("Password", "Must not be more than 72 bytes long.")
	}
	if len(err.Violations()) != 0 {
		return err
	}
	return nil
}

// Matches tests whether the provided plaintext password matches the hashed password.
func (p *Password) Matches(plaintext string) (_ bool, err error) {
	defer multierr.Wrap(&err, "greenlight.password.Matches")

	if err := bcrypt.CompareHashAndPassword(*p, []byte(plaintext)); err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

// UserService is a service for managing users.
type UserService interface {
	Get(ctx context.Context, email string) (*User, error)
	Create(ctx context.Context, u *User) error
	Update(ctx context.Context, u *User) error
}
