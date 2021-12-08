package images_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/Defacto2/df2/lib/images"
)

func TestIDMagick(t *testing.T) {
	const (
		gif = `../../tests/images/test.gif GIF 1280x32 1280x32+0+0 8-bit sRGB 2c 661B 0.000u 0:00.000`
		jpg = `../../tests/images/test.jpg JPEG 1280x32 1280x32+0+0 8-bit sRGB 3236B 0.000u 0:00.000`
		png = `../../tests/images/test.png PNG 1280x32 1280x32+0+0 8-bit sRGB 2c 240B 0.000u 0:00.000`
	)
	tests := []struct {
		name    string
		src     string
		want    string
		wantErr bool
	}{
		{"empty", "", "", true},
		{"null", "/dev/null", "", true},
		{"not an image", "imagemagick_test.go", "", true},
		{"gif", "../../tests/images/test.gif", gif, false},
		{"jpg", "../../tests/images/test.jpg", jpg, false},
		{"png", "../../tests/images/test.png", png, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := images.IDMagick(tt.src)
			if (err != nil) != tt.wantErr {
				t.Errorf("IDMagick() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if string(bytes.TrimSpace(got)) != tt.want {
				t.Errorf("IDMagick() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestToMagick(t *testing.T) {
	const gif = "../../tests/images/test.gif"
	tmp := filepath.Join(t.TempDir(), "test.png")
	type args struct {
		src  string
		dest string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"empty", args{"", ""}, true},
		{"no dest", args{gif, ""}, false},
		{"tmp", args{gif, tmp}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// set temp dir
			if err := images.ToMagick(tt.args.src, tt.args.dest); (err != nil) != tt.wantErr {
				t.Errorf("ToMagick() error = %v, wantErr %v", err, tt.wantErr)
			}
			defer os.Remove(tmp)
		})
	}
}
