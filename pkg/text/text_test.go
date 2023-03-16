package text_test

import (
	"testing"

	"github.com/Defacto2/df2/pkg/text"
)

func TestFix(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"fix", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := text.Fix(nil, nil); (err != nil) != tt.wantErr {
				t.Errorf("Fix() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
