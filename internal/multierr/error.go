// Package multierror implements types to combine together and handle multiple errors.
package multierr

import (
	"bytes"
)

// joinError represents an error that wraps the errors.
type joinError []error

// Join returns an error that wraps the given errors.
// If any of the passed errors is a [MultiError] error, it will be flattened along with the other errors.
func Join(errs ...error) error {
	var mErrs []error
	for _, err := range errs {
		if err == nil {
			continue
		}

		if nested, ok := err.(joinError); ok {
			mErrs = append(mErrs, nested...)
		} else {
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

// Error returns a string representation of an error. Errors are formatted as a flat structure separated by a ';' symbol.
func (e joinError) Error() string {
	if e == nil {
		return ""
	}

	var buf bytes.Buffer
	for i, err := range e {
		if i != 0 {
			buf.WriteString("; ")
		}
		buf.WriteString(err.Error())
	}
	return buf.String()
}

// Unwrap returns this and all the wrapped errors.
func (e joinError) Unwrap() []error {
	if len(e) == 0 {
		return nil
	}
	return e
}
