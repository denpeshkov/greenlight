package greenlight

import (
	"fmt"

	"github.com/denpeshkov/greenlight/internal/multierr"
)

type InternalError struct {
	Msg string
}

func NewInternalError(format string, args ...any) InternalError {
	return InternalError{
		Msg: fmt.Sprintf(format, args...),
	}
}

func (e InternalError) Error() string {
	return e.Msg
}

type NotFoundError struct {
	Msg string
}

func NewNotFoundError(format string, args ...any) NotFoundError {
	return NotFoundError{
		Msg: fmt.Sprintf(format, args...),
	}
}

func (e NotFoundError) Error() string {
	return e.Msg
}

type InvalidError struct {
	Msg        string
	violations map[string]error
}

func NewInvalidError(format string, args ...any) InvalidError {
	return InvalidError{
		Msg:        fmt.Sprintf(format, args...),
		violations: make(map[string]error),
	}
}

func (e InvalidError) Error() string {
	if len(e.violations) == 0 {
		return e.Msg
	}
	return fmt.Sprintf("%s: %v", e.Msg, e.violations)
}

func (e InvalidError) AddViolation(field string, err error) {
	e.violations[field] = multierr.Join(e.violations[field], err)
}

func (e InvalidError) FieldViolation(field string) error {
	return e.violations[field]
}

func (e InvalidError) Violations() map[string]error {
	return e.violations
}

type ConflictError struct {
	Msg string
}

func NewConflictError(format string, args ...any) error {
	return ConflictError{
		Msg: fmt.Sprintf(format, args...),
	}
}

func (e ConflictError) Error() string {
	return e.Msg
}
