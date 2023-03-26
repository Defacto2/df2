package arg_test

import (
	"io"
	"testing"

	"github.com/Defacto2/df2/cmd/internal/arg"
	"github.com/stretchr/testify/assert"
)

func Test_options(t *testing.T) {
	t.Parallel()
	type args struct {
		a []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"empty", args{}, "\noptions: MISSING"},
		{"targets", args{arg.Targets()}, "\noptions: all, download, emulation, image"},
		{"test", args{[]string{"test"}}, "\noptions: test"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := arg.CleanOpts(tt.args.a...); got != tt.want {
				t.Errorf("CleanOpts() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestInvalid(t *testing.T) {
	t.Parallel()
	args := []string{"abc", "def"}
	err := arg.Invalid(nil, "", []string{}...)
	assert.NotNil(t, err)
	err = arg.Invalid(io.Discard, "", []string{}...)
	assert.NotNil(t, err)
	err = arg.Invalid(io.Discard, "", args...)
	assert.NotNil(t, err)
	err = arg.Invalid(io.Discard, "abc", args...)
	assert.Nil(t, err)
}
