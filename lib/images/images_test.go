package images

import (
	_ "image/gif"
	_ "image/jpeg"
	"testing"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

func TestNewExt(t *testing.T) {
	type args struct {
		name      string
		extension string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"", args{"hello", ".txt"}, "hello.txt"},
		{"", args{"hello.jpg", ".png"}, "hello.png"},
		{"", args{"hello", ""}, "hello"},
		{"", args{"", ".ssh"}, ".ssh"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewExt(tt.args.name, tt.args.extension); got != tt.want {
				t.Errorf("NewExt() = %v, want %v", got, tt.want)
			}
		})
	}
}
