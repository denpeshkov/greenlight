package greenlight

import (
	"bytes"
	"errors"
	"fmt"
)

const (
	// ErrInternal indicates an internal error. Should be used for any non-application error.
	ErrInternal string = "internal"
	// ErrNotFound indicates that the entity was not found.
	ErrNotFound string = "not_found"
	// ErrInvalid indicates that the entity is invalid.
	ErrInvalid string = "invalid"
)

// Error represents an application error.
type Error struct {
	// Op represents an operation that produced an error. Ops are used to construct an error trace.
	Op string `json:"operation,omitempty"`
	// Code represents an error code.
	Code string `json:"code,omitempty"`
	// Msg represents an error message.
	Msg string `json:"message,omitempty"`
	// Err represents a wrapped error.
	Err error `json:"error,omitempty"`
}

// Error implements the error interface.
func (e *Error) Error() string {
	var b bytes.Buffer

	if e.Op != "" {
		fmt.Fprintf(&b, "%s: ", e.Op)
	}

	// If wrapping an error, print its Error() message. Otherwise print the error code & message.
	if e.Err != nil {
		b.WriteString(e.Err.Error())
	} else {
		if e.Code != "" {
			fmt.Fprintf(&b, "[%s] ", e.Code)
		}
		b.WriteString(e.Msg)
	}
	return b.String()
}

// ErrorCode returns the code of the root error, if available. Otherwise returns [ErrInternal].
func ErrorCode(err error) string {
	if err == nil {
		return ""
	}
	if e, ok := err.(*Error); ok && e.Code != "" {
		return e.Code
	} else if ok && e.Err != nil {
		return ErrorCode(e.Err)
	}
	return ErrInternal
}

// ErrorMessage returns the message of the root error, if available. Otherwise returns a generic error message.
// Returned message is intended for the end-user.
func ErrorMessage(err error) string {
	if err == nil {
		return ""
	}
	if e, ok := err.(*Error); ok && e.Msg != "" {
		return e.Msg
	} else if ok && e.Err != nil {
		return ErrorMessage(e.Err)
	} else if e := errors.Unwrap(e); e != nil {
		return ErrorMessage(e)
	}
	return "Internal error occurred."
}

// ErrorTrace returns an error trace.
func ErrorTrace(err error) []string {
	var ops []string

	if e, ok := err.(*Error); ok && e.Op != "" {
		ops = append(ops, e.Op)
		ops = append(ops, ErrorTrace(e.Err)...)
	} else if ok && e.Err != nil {
		ops = append(ops, ErrorTrace(e.Err)...)
	} else if e := errors.Unwrap(err); e != nil {
		ops = append(ops, ErrorTrace(e)...)
	}
	return ops
}

// Unwrap returns a wrapped error.
func (e *Error) Unwrap() error {
	return e.Err
}
