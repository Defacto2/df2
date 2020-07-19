package text

import (
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/gookit/color"
)

func TestImage_exists(t *testing.T) {
	type fields struct {
		UUID string
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{"empty", fields{""}, false},
		{"test", fields{"test"}, true},
		{"missingdir", fields{"test"}, false},
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	if runtime.GOOS == "windows" {
		dir.Img000 = path.Join(wd, "..\\..\\tests\\text\\")
		dir.Img150 = path.Join(wd, "..\\..\\tests\\text\\")
		dir.Img400 = path.Join(wd, "..\\..\\tests\\text\\")
	} else {
		dir.Img000 = path.Join(wd, "../../tests/text/")
		dir.Img150 = path.Join(wd, "../../tests/text/")
		dir.Img400 = path.Join(wd, "../../tests/text/")
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "missingdir" {
				dir.Img400 = ""
			}
			x := Image{
				UUID: tt.fields.UUID,
			}
			if got := x.exist(); got != tt.want {
				t.Errorf("Image.exist() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestImage_valid(t *testing.T) {
	type fields struct {
		Filename string
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{"empty", fields{""}, false},
		{"no ext", fields{"hello"}, false},
		{"doc", fields{"hello.doc"}, true},
		{"two exts", fields{"hello.world.doc"}, true},
		{"zip", fields{"hello.zip"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x := Image{
				Filename: tt.fields.Filename,
			}
			if got := x.valid(); got != tt.want {
				t.Errorf("Image.valid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestImage_String(t *testing.T) {
	type fields struct {
		ID       uint
		UUID     string
		Filename string
		FileExt  string
		Filesize int
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"empty", fields{}, "(0)  0 B "},
		{"test", fields{54, "xxx", "somefile", "exe", 346000}, "(54) somefile 346 kB "},
	}
	color.Enable = false
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img := Image{
				ID:       tt.fields.ID,
				UUID:     tt.fields.UUID,
				Filename: tt.fields.Filename,
				FileExt:  tt.fields.FileExt,
				Filesize: tt.fields.Filesize,
			}
			if got := img.String(); got != tt.want {
				t.Errorf("Image.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

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
			if err := Fix(tt.args.simulate); (err != nil) != tt.wantErr {
				t.Errorf("Fix() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
