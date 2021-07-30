package text

import (
	"os"
	"path"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Defacto2/df2/lib/directories"
	"github.com/gookit/color" //nolint:misspell
)

func TestImage_exists(t *testing.T) {
	if runtime.GOOS == "windows" {
		return
	}
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
	dir := directories.Init(false)
	dir.Img000 = filepath.Clean(path.Join(wd, "../../tests/text/"))
	dir.Img400 = filepath.Clean(path.Join(wd, "../../tests/text/"))
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "missingdir" {
				dir.Img400 = ""
			}
			x := textfile{
				UUID: tt.fields.UUID,
			}
			got, err := x.exist(&dir)
			if got != tt.want {
				t.Errorf("image.exist(%s) = %v, want %v", &dir, got, tt.want)
			}
			if err != nil {
				t.Errorf("image.exists(%s) err = %v", &dir, err)
			}
		})
	}
}

func TestImage_archive(t *testing.T) {
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
		{"doc", fields{"hello.doc"}, false},
		{"two exts", fields{"hello.world.doc"}, false},
		{"bz2", fields{"hello.bz2"}, true},
		{"zip", fields{"hello.zip"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x := textfile{
				Name: tt.fields.Filename,
			}
			if got := x.archive(); got != tt.want {
				t.Errorf("image.valid() = %v, want %v", got, tt.want)
			}
		})
	}
}

//nolint: revive
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
			img := textfile{
				ID:   tt.fields.ID,
				UUID: tt.fields.UUID,
				Name: tt.fields.Filename,
				Ext:  tt.fields.FileExt,
				Size: tt.fields.Filesize,
			}
			if got := img.String(); got != tt.want {
				t.Errorf("image.String() = %q, want %q", got, tt.want)
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
