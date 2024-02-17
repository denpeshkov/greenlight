// Package multierror implements types to combine together and handle multiple errors.
package multierr

import (
	"bytes"
	"fmt"
)

// joinError represents an error that wraps the errors.
type joinError []error

// Join returns an error that wraps the given errors.
// If any of the passed errors is a [MultiError] error, it will be flattened along with the other errors.
func Join(errs ...error) error {
	var mErrs []error
	for _, e := range errs {
		switch err := e.(type) {
		case nil:
			continue
		case joinError:
			mErrs = append(mErrs, err...)
		default:
			mErrs = append(mErrs, err)
		}
	}

	// If nil or one error return directly instead of a slice.
	switch len(mErrs) {
	case 0:
		return nil
	case 1:
		return mErrs[0]
	}
	return joinError(mErrs)
}

// Error returns a string representation of an error. Errors are formatted as a flat structure ["", ""].
func (e joinError) Error() string {
	if e == nil {
		return ""
	}

	var b bytes.Buffer
	b.WriteRune('[')
	for i, err := range e {
		if i != 0 {
			b.WriteString(", ")
		}
		b.WriteRune('"')
		b.WriteString(err.Error())
		b.WriteRune('"')
	}
	b.WriteRune(']')
	return b.String()
}

// Unwrap returns this and all the wrapped errors.
func (e joinError) Unwrap() []error {
	if len(e) == 0 {
		return nil
	}
	return e
}

// Wrap adds context to the error and allows unwrapping the result to recover the original error.
func Wrap(err *error, format string, args ...any) {
	if *err != nil {
		*err = fmt.Errorf("%s: %w", fmt.Sprintf(format, args...), *err)
	}
}
