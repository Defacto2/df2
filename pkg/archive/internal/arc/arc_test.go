package arc_test

import (
	"testing"

	"github.com/Defacto2/df2/pkg/archive/internal/arc"
	"github.com/mholt/archiver"
)

func TestConfigure(t *testing.T) {
	rar, err := archiver.ByExtension(".tar")
	if err != nil {
		t.Error(err)
	}
	zip, err := archiver.ByExtension(".zip")
	if err != nil {
		t.Error(err)
	}
	tests := []struct {
		name    string
		f       interface{}
		wantErr bool
	}{
		{"empty", nil, true},
		{"err", "", true},
		{"rar", rar, false},
		{"zip", zip, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := arc.Configure(tt.f); (err != nil) != tt.wantErr {
				t.Errorf("configure() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
