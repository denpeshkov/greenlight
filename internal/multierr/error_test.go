// Based on: https://github.com/uber-go/multierr/blob/master/error_test.go#L123

package multierr

import (
	"errors"
	"fmt"
	"slices"
	"testing"
)

// errors.New always creates distinct errors so we reuse the same values
var (
	errFoo = errors.New("foo")
	errBar = errors.New("bar")
	errBaz = errors.New("baz")
	errQux = errors.New("qux")
)

func TestJoin(t *testing.T) {
	tests := []struct {
		giveErrors []error
		wantError  error
		wantString string
	}{
		{
			giveErrors: nil,
			wantError:  nil,
		},
		{
			giveErrors: []error{},
			wantError:  nil,
		},
		{
			giveErrors: []error{
				errFoo,
				nil,
				newJoinError(
					errBar,
				),
				nil,
			},
			wantError: newJoinError(
				errFoo,
				errBar,
			),
			wantString: `["foo", "bar"]`,
		},
		{
			giveErrors: []error{nil, nil, errFoo, nil},
			wantError:  errFoo,
			wantString: "foo",
		},
		{
			giveErrors: []error{
				errFoo,
				newJoinError(
					errBar,
				),
			},
			wantError: newJoinError(
				errFoo,
				errBar,
			),
			wantString: `["foo", "bar"]`,
		},
		{
			giveErrors: []error{errFoo},
			wantError:  errFoo,
			wantString: "foo",
		},
		{
			giveErrors: []error{
				errFoo,
				errBar,
			},
			wantError: newJoinError(
				errFoo,
				errBar,
			),
			wantString: `["foo", "bar"]`,
		},
		{
			giveErrors: []error{
				errFoo, errBar, errBaz,
			},
			wantError: newJoinError(
				errFoo, errBar, errBaz,
			),
			wantString: `["foo", "bar", "baz"]`,
		},
		{
			giveErrors: []error{
				errFoo,
				newJoinError(
					errBar,
					errBaz,
				),
				errQux,
			},
			wantError: newJoinError(
				errFoo,
				errBar,
				errBaz,
				errQux,
			),
			wantString: `["foo", "bar", "baz", "qux"]`,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			err := Join(tt.giveErrors...)
			if !equal(tt.wantError, err) {
				t.Errorf("want: %#v, got: %#v", tt.wantError, err)
			}
			if tt.wantString != "" && err.Error() != tt.wantString {
				t.Errorf("want string: %s, got: %s", tt.wantString, err.Error())
			}
		})
	}
}

func equal(err1, err2 error) bool {
	if (err1 != nil && err2 == nil) || (err1 == nil && err2 != nil) {
		return false
	}
	jerr1, ok1 := err1.(joinError)
	jerr2, ok2 := err2.(joinError)
	switch {
	case ok1 && ok2:
		return slices.EqualFunc(jerr1, jerr2, errors.Is)
	case ok1 || ok2:
		return false
	default:
		return err1 == err2
	}
}

func newJoinError(errors ...error) error {
	return joinError(errors)
}
