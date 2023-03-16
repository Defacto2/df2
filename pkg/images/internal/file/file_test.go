package file_test

import (
	"testing"

	"github.com/Defacto2/df2/pkg/configger"
	"github.com/Defacto2/df2/pkg/directories"
	"github.com/Defacto2/df2/pkg/images/internal/file"
)

func TestIsDir(t *testing.T) {
	dir := directories.Init(configger.Defaults(), false)
	tests := []struct {
		name string
		i    file.Image
		want bool
	}{
		{"not exist", file.Image{UUID: "false id"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.i.IsDir(&dir); got != tt.want {
				t.Errorf("IsDir() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsExt(t *testing.T) {
	tests := []struct {
		name   string
		i      file.Image
		wantOk bool
	}{
		{"empty", file.Image{}, false},
		{"text", file.Image{Name: "some.txt"}, false},
		{"png", file.Image{Name: "some.png"}, true},
		{"jpeg", file.Image{Name: "some other.jpeg"}, true},
		{"jpeg", file.Image{Name: "some.other.jpeg"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotOk := tt.i.IsExt(); gotOk != tt.wantOk {
				t.Errorf("IsExt() = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}
