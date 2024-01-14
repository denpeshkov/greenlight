package postgres

import "errors"

var (
	// ErrRecordNotFound indicates that the record was not found.
	ErrRecordNotFound error = errors.New("record not found")
)
