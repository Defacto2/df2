package images

import (
	"testing"

	"github.com/Defacto2/df2/lib/directories"
)

func TestImg_valid(t *testing.T) {
	dir := directories.Init(false)
	tests := []struct {
		name string
		i    imageFile
		want bool
	}{
		{"not exist", imageFile{UUID: "false id"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.i.valid(&dir); got != tt.want {
				t.Errorf("Img.valid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestImg_ext(t *testing.T) {
	tests := []struct {
		name   string
		i      imageFile
		wantOk bool
	}{
		{"empty", imageFile{}, false},
		{"text", imageFile{Name: "some.txt"}, false},
		{"png", imageFile{Name: "some.png"}, true},
		{"jpeg", imageFile{Name: "some other.jpeg"}, true},
		{"jpeg", imageFile{Name: "some.other.jpeg"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotOk := tt.i.ext(); gotOk != tt.wantOk {
				t.Errorf("Img.ext() = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func TestFix(t *testing.T) {
	tests := []struct {
		name     string
		simulate bool
		wantErr  bool
	}{
		{"name", true, false},
		{"name", false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Fix(tt.simulate); (err != nil) != tt.wantErr {
				t.Errorf("Fix() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
