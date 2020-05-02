package archive

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gabriel-vasile/mimetype"
)

func testDir(name string) string {
	dir, _ := os.Getwd()
	return filepath.Join(dir, "../../tests/", name)
}

func TestFileCopy(t *testing.T) {

	type args struct {
		name string
		dest string
	}
	tests := []struct {
		name        string
		args        args
		wantWritten int64
		wantErr     bool
	}{
		{"empty", args{"", ""}, 0, true},
		{"empty", args{testDir("text/test.txt"), testDir("text/test.txt")}, 12, false},
		{"empty", args{testDir("text/test.txt"), testDir("text/test.txt~")}, 12, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotWritten, err := FileCopy(tt.args.name, tt.args.dest)
			if (err != nil) != tt.wantErr {
				t.Errorf("FileCopy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotWritten != tt.wantWritten {
				t.Errorf("FileCopy() = %v, want %v", gotWritten, tt.wantWritten)
			}
			if err == nil && tt.args.name != tt.args.dest {
				if _, err := os.Stat(tt.args.dest); !os.IsNotExist(err) {
					if err = os.Remove(tt.args.dest); err != nil {
						t.Errorf("FileCopy() failed to cleanup copy %v", tt.args.dest)
					}
				}
			}
		})
	}
}

func Test_content_filemime(t *testing.T) {
	type fields struct {
		name       string
		ext        string
		path       string
		mime       *mimetype.MIME
		size       int64
		executable bool
		textfile   bool
	}
	type args struct {
		f os.FileInfo
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"text", fields{}, args{}, true},
		{"text", fields{ext: ".txt", path: testDir("text/test.txt")}, args{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &content{
				name:       tt.fields.name,
				ext:        tt.fields.ext,
				path:       tt.fields.path,
				mime:       tt.fields.mime,
				size:       tt.fields.size,
				executable: tt.fields.executable,
				textfile:   tt.fields.textfile,
			}
			if err := c.filemime(tt.args.f); (err != nil) != tt.wantErr {
				t.Errorf("content.filemime() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFileMove(t *testing.T) {
	type args struct {
		name string
		dest string
	}
	println(testDir("text/test.txt~"))
	tests := []struct {
		name        string
		args        args
		wantWritten int64
		wantErr     bool
	}{
		{"empty", args{"", ""}, 0, false},
		{"one way", args{testDir("text/test.txt"), testDir("text/test.txt~")}, 12, false},
		{"restore way", args{testDir("text/test.txt~"), testDir("text/test.txt")}, 12, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotWritten, err := FileMove(tt.args.name, tt.args.dest)
			if (err != nil) != tt.wantErr {
				t.Errorf("FileMove() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotWritten != tt.wantWritten {
				t.Errorf("FileMove() = %v, want %v", gotWritten, tt.wantWritten)
			}
		})
	}
}
