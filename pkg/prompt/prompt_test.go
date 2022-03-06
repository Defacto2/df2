package prompt_test

import (
	"testing"

	"github.com/Defacto2/df2/pkg/prompt"
)

func TestYN(t *testing.T) {
	type args struct {
		query string
		yes   bool
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"y", args{"", false}, false},
		{"n", args{"", true}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := prompt.YN(tt.args.query, tt.args.yes); got != tt.want {
				t.Errorf("YN() = %v, want %v", got, tt.want)
			}
		})
	}
}
