package greenlight

import (
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestError_Ops(t *testing.T) {
	tests := []struct {
		giveErr error
		wantOps []string
	}{
		{
			giveErr: &Error{},
			wantOps: nil,
		},
		{
			giveErr: &Error{Op: "op1"},
			wantOps: []string{"op1"},
		},
		{
			giveErr: &Error{
				Op:  "op1",
				Err: &Error{Op: "op2"},
			},
			wantOps: []string{"op1", "op2"},
		},
		{
			giveErr: &Error{
				Op: "op1",
				Err: &Error{
					Op:  "op2",
					Err: &Error{Op: "op3"},
				},
			},
			wantOps: []string{"op1", "op2", "op3"},
		},
		{
			giveErr: &Error{
				Op: "op1",
				Err: &Error{
					Err: &Error{Op: "op3"},
				},
			},
			wantOps: []string{"op1", "op3"},
		},
		{
			giveErr: &Error{
				Err: &Error{
					Op:  "op2",
					Err: &Error{Op: "op3"},
				},
			},
			wantOps: []string{"op2", "op3"},
		},
		{
			giveErr: &Error{
				Err: &Error{
					Op:  "op2",
					Err: &Error{},
				},
			},
			wantOps: []string{"op2"},
		},
		{
			giveErr: &Error{
				Err: &Error{
					Err: &Error{Op: "op3"},
				},
			},
			wantOps: []string{"op3"},
		},
		{
			giveErr: &Error{
				Err: &Error{
					Err: &Error{},
				},
			},
			wantOps: nil,
		},
		{
			giveErr: &Error{
				Op: "op1",
				Err: &Error{
					Err: errors.New("err"),
				},
			},
			wantOps: []string{"op1"},
		},
		{
			giveErr: &Error{
				Op: "op1",
				Err: &Error{
					Op:  "op2",
					Err: errors.New("err"),
				},
			},
			wantOps: []string{"op1", "op2"},
		},
		{
			giveErr: &Error{
				Err: &Error{
					Op:  "op2",
					Err: errors.New("err"),
				},
			},
			wantOps: []string{"op2"},
		},
		{
			giveErr: &Error{
				Err: &Error{
					Err: errors.New("err"),
				},
			},
			wantOps: nil,
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			if gotOps := ErrorTrace(tt.giveErr); !cmp.Equal(gotOps, tt.wantOps) {
				t.Errorf("Ops() = %v, want %v", gotOps, tt.wantOps)
			}
		})
	}
}
