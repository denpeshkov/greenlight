package greenlight

import "errors"

var (
	// ErrNotFound indicates that the entity was not found.
	ErrNotFound error = errors.New("entity not found")
)
