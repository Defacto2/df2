package tf_test

import (
	"os"
	"path"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Defacto2/df2/lib/directories"
	"github.com/Defacto2/df2/lib/text/internal/tf"
	"github.com/gookit/color"
)

func TestExist(t *testing.T) {
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
			var x tf.TextFile
			x.UUID = tt.fields.UUID
			got, err := x.Exist(&dir)
			if got != tt.want {
				t.Errorf("Exist(%s) = %v, want %v", &dir, got, tt.want)
			}
			if err != nil {
				t.Errorf("Exist(%s) err = %v", &dir, err)
			}
		})
	}
}

func TestArchive(t *testing.T) {
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
			var x tf.TextFile
			x.Name = tt.fields.Filename
			if got := x.Archive(); got != tt.want {
				t.Errorf("Archive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestString(t *testing.T) {
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
			var img tf.TextFile
			img.ID = tt.fields.ID
			img.UUID = tt.fields.UUID
			img.Name = tt.fields.Filename
			img.Ext = tt.fields.FileExt
			img.Size = tt.fields.Filesize
			if got := img.String(); got != tt.want {
				t.Errorf("img.String() = %q, want %q", got, tt.want)
			}
		})
	}
}
