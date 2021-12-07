package text_test

import (
	"testing"

	"github.com/Defacto2/df2/lib/text"
)

func TestFix(t *testing.T) {
	type args struct {
		simulate bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"simulate", args{true}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := text.Fix(tt.args.simulate); (err != nil) != tt.wantErr {
				t.Errorf("Fix() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
