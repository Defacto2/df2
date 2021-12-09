package tf_test

import (
	"database/sql"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Defacto2/df2/lib/directories"
	"github.com/Defacto2/df2/lib/text/internal/tf"
	"github.com/gookit/color"
)

const (
	textDir       = "../../../../tests/demozoo/"
	uuid          = "21cb94d3-ffc1-4055-8398-b7b4ed1e67e8"
	storedName    = "test.zip"
	fileToExtract = "test.txt"
	txt           = ".txt"
)

func TestExist(t *testing.T) {
	if runtime.GOOS == "windows" {
		return
	}
	const uuid = "815783a6-dd34-4ec8-9527-cdbdaaab612d"
	type fields struct {
		UUID string
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{"empty", fields{""}, false},
		{"test", fields{"test"}, false},
		{"missingdir", fields{uuid}, false}, // "missingdir" will blank dir.Img400
		{"okay", fields{uuid}, true},
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// attempts to find "815783a6-dd34-4ec8-9527-cdbdaaab612d.png" in both dirs
			dir := directories.Init(false)
			dir.Img000 = filepath.Clean(path.Join(wd, "../../../../tests/uuid/"))
			dir.Img400 = filepath.Clean(path.Join(wd, "../../../../tests/uuid/"))
			if tt.name == "missingdir" {
				dir.Img400 = ""
			}
			var f tf.TextFile
			f.UUID = tt.fields.UUID
			got, err := f.Exist(&dir)
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
			var f tf.TextFile
			f.Name = tt.fields.Filename
			if got := f.Archive(); got != tt.want {
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

func TestTextFile_Extract(t *testing.T) {
	var dirInput directories.Dir
	dirInput.UUID = textDir
	type fields struct {
		ID       uint
		UUID     string
		Name     string
		Ext      string
		Platform string
		Size     int
		NoReadme sql.NullBool
		Readme   sql.NullString
	}
	tests := []struct {
		name    string
		fields  fields
		dir     *directories.Dir
		wantErr bool
	}{
		{"empty", fields{}, nil, true},
		{"no dir", fields{ID: 1, UUID: uuid}, nil, true},
		{"okay", fields{
			UUID:   uuid,
			Name:   "test.zip",
			Ext:    ".zip",
			Readme: sql.NullString{String: fileToExtract, Valid: true}}, &dirInput, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &tf.TextFile{
				ID:       tt.fields.ID,
				UUID:     tt.fields.UUID,
				Name:     tt.fields.Name,
				Ext:      tt.fields.Ext,
				Platform: tt.fields.Platform,
				Size:     tt.fields.Size,
				NoReadme: tt.fields.NoReadme,
				Readme:   tt.fields.Readme,
			}
			if err := tr.Extract(tt.dir); (err != nil) != tt.wantErr {
				t.Errorf("TextFile.Extract() error = %v, wantErr %v", err, tt.wantErr)
			}
			defer os.Remove(filepath.Join(textDir, uuid+txt))
		})
	}
}

func TestTextFile_ExtractedImgs(t *testing.T) {
	dir := filepath.Join(textDir, "extracted")
	type fields struct {
		ID       uint
		UUID     string
		Name     string
		Ext      string
		Platform string
		Size     int
		NoReadme sql.NullBool
		Readme   sql.NullString
	}
	tests := []struct {
		name    string
		fields  fields
		dir     string
		wantErr bool
	}{
		{"empty", fields{}, "", true},
		{"no dir", fields{ID: 1, UUID: uuid}, "", true},
		{"okay", fields{
			UUID: uuid,
		}, dir, false},
		// {"okay", fields{
		// 	UUID:   uuid,
		// 	Name:   "test.zip",
		// 	Ext:    ".zip",
		// 	Readme: sql.NullString{String: fileToExtract, Valid: true}}, &dirInput, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &tf.TextFile{
				ID:       tt.fields.ID,
				UUID:     tt.fields.UUID,
				Name:     tt.fields.Name,
				Ext:      tt.fields.Ext,
				Platform: tt.fields.Platform,
				Size:     tt.fields.Size,
				NoReadme: tt.fields.NoReadme,
				Readme:   tt.fields.Readme,
			}
			if err := tr.ExtractedImgs(tt.dir); (err != nil) != tt.wantErr {
				t.Errorf("TextFile.ExtractedImgs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
