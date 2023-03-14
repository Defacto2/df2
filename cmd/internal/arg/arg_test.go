package arg_test

import (
	"testing"

	"github.com/Defacto2/df2/cmd/internal/arg"
)

func Test_options(t *testing.T) {
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
	type args struct {
		arg  string
		args []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"empty", args{}, true},
		{"args", args{
			arg:  "abc",
			args: []string{"abc", "def"},
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := arg.Invalid(nil, tt.args.arg, tt.args.args...); (err != nil) != tt.wantErr {
				t.Errorf("Invalid() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
