package archive_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Defacto2/df2/pkg/archive"
	"github.com/Defacto2/df2/pkg/archive/internal/content"
	"github.com/Defacto2/df2/pkg/archive/internal/demozoo"
	"github.com/gabriel-vasile/mimetype"
	"github.com/spf13/viper"
)

func testDir(name string) string {
	dir, _ := os.Getwd()
	return filepath.Join(dir, "..", "..", "tests", name)
}

func TestMIME(t *testing.T) {
	type fields struct {
		name       string
		ext        string
		path       string
		mime       *mimetype.MIME
		size       int64
		executable bool
		textfile   bool
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"text", fields{}, true},
		{"text", fields{ext: ".txt", path: testDir("text/test.txt")}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &content.File{
				Name:       tt.fields.name,
				Ext:        tt.fields.ext,
				Path:       tt.fields.path,
				Mime:       tt.fields.mime,
				Size:       tt.fields.size,
				Executable: tt.fields.executable,
				Textfile:   tt.fields.textfile,
			}
			if err := c.MIME(); (err != nil) != tt.wantErr {
				t.Errorf("MIME() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRead(t *testing.T) {
	type args struct {
		archive  string
		filename string
	}
	tests := []struct {
		name      string
		args      args
		wantFiles []string
		wantFile  string
		wantErr   bool
	}{
		{"empty", args{"", ""}, nil, "", true},
		{
			"invalid rar ext",
			args{testDir("demozoo/test.invalid.ext.rar"), "test.invalid.ext.rar"},
			[]string{"test.png", "test.txt"},
			"test.invalid.ext.zip", false,
		},
		{
			"zip",
			args{testDir("demozoo/test.zip"), "test.zip"},
			[]string{"test.png", "test.txt"},
			"test.zip", false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFiles, gotFile, err := archive.Read(tt.args.archive, tt.args.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotFile != tt.wantFile {
				t.Errorf("Read() = %v, want %v", gotFile, tt.wantFile)
			}
			if !reflect.DeepEqual(gotFiles, tt.wantFiles) {
				t.Errorf("Read() = %v, want %v", gotFiles, tt.wantFiles)
			}
		})
	}
}

func TestRestore(t *testing.T) {
	tmp := os.TempDir()
	type args struct {
		source      string
		filename    string
		destination string
	}
	gz := filepath.Join(testDir("demozoo"), "test.tar.gz")
	tar := filepath.Join(testDir("demozoo"), "test.tar")
	z7 := filepath.Join(testDir("demozoo"), "test.7z")
	zip := filepath.Join(testDir("demozoo"), "test.zip")
	res := []string{"test.png", "test.txt"}
	tests := []struct {
		name      string
		args      args
		wantFiles []string
		wantErr   bool
	}{
		{"empty", args{}, nil, true},
		{"err", args{"/dev/null/fake.zip", "fake.zip", tmp}, nil, true},
		{"zip", args{zip, "test.zip", tmp}, res, false},
		{"tar", args{tar, "test.tar", tmp}, res, false},
		{"gz", args{gz, "test.tar.gz", tmp}, res, false},
		{"7z (unsupported)", args{z7, "test.7z", tmp}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFiles, err := archive.Restore(tt.args.source, tt.args.filename, tt.args.destination)
			if (err != nil) != tt.wantErr {
				t.Errorf("Restore() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotFiles, tt.wantFiles) {
				t.Errorf("Restore() = %v, want %v", gotFiles, tt.wantFiles)
			}
		})
	}
}

func TestExtract(t *testing.T) {
	const uuid = "6ba7b814-9dad-11d1-80b4-00c04fd430c8"
	type args struct {
		archive  string
		filename string
		uuid     string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"empty", args{}, true},
		{"error", args{testDir("demozoo/test.zip"), "test.zip", "xxxx"}, true},
		{"7z", args{testDir("demozoo/test.7z"), "test.7z", uuid}, true},
		{"tar", args{testDir("demozoo/test.tar"), "test.tar", uuid}, false},
		{"tar.gz", args{testDir("demozoo/test.tar.gz"), "test.tar.gz", uuid}, false},
		{"tar.bz2", args{testDir("demozoo/test.tar.bz2"), "test.tar.bz2", uuid}, false},
		{"sz (snappy)", args{testDir("demozoo/test.tar.sz"), "test.tar.sz", uuid}, false},
		{"xz", args{testDir("demozoo/test.tar.xz"), "test.tar.xz", uuid}, false},
		{"zip", args{testDir("demozoo/test.zip"), "test.zip", uuid}, false},
		{"zip.bz2", args{testDir("demozoo/test.bz2.zip"), "test.bz2.zip", uuid}, false},
	}
	for _, tt := range tests {
		if viper.GetString("directory.root") == "" {
			return
		}
		t.Run(tt.name, func(t *testing.T) {
			if err := archive.Proof(tt.args.archive, tt.args.filename, tt.args.uuid); (err != nil) != tt.wantErr {
				t.Errorf("Proof(%s) error = %v, wantErr %v", tt.args.archive, err, tt.wantErr)
			}
		})
	}
}

func TestNFO(t *testing.T) {
	const fileDiz = demozoo.FileDiz
	var empty []string
	const (
		ff2 = "hi.nfo"
		ff3 = "random.txt"
	)
	type args struct {
		name  string
		files []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"empty", args{"", empty}, ""},
		{"empty zip", args{"hi.zip", empty}, ""},
		{"1 file", args{"hi.zip", []string{fileDiz}}, fileDiz},
		{"2 files", args{"hi.zip", []string{fileDiz, ff2}}, "hi.nfo"},
		{"3 files", args{"hi.zip", []string{fileDiz, ff2, ff3}}, "hi.nfo"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := archive.NFO(tt.args.name, tt.args.files...); got != tt.want {
				t.Errorf("NFO() = %v, want %v", got, tt.want)
			}
		})
	}
}
