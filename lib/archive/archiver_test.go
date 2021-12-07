package archive

import (
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/mholt/archiver"
)

func TestReadr(t *testing.T) {
	type args struct {
		archive  string
		filename string
	}
	tests := []struct {
		name      string
		args      args
		wantFiles []string
		wantErr   bool
	}{
		{"empty", args{"", ""}, nil, true},
		{"zip", args{testDir("demozoo/test.zip"), "test.zip"}, []string{"test.png", "test.txt"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFiles, err := Readr(tt.args.archive, tt.args.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("Readr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotFiles, tt.wantFiles) {
				t.Errorf("Readr() = %v, want %v", gotFiles, tt.wantFiles)
			}
		})
	}
}

func TestUnarchiver(t *testing.T) {
	tempDir, err := ioutil.TempDir(os.TempDir(), "unarchiver")
	var (
		src = testDir("demozoo/test.zip")
		fn  = "test.zip"
		z7  = testDir("demozoo/test.7z")
		zfn = "test.7z"
	)
	if err != nil {
		t.Error(err)
	}
	type args struct {
		source      string
		filename    string
		destination string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"empty", args{}, true},
		{"missing src", args{"", fn, tempDir}, true},
		{"missing fn", args{src, "", tempDir}, true},
		{"missing dest", args{src, fn, ""}, true},
		{"okay", args{src, fn, tempDir}, false},
		{"7z", args{z7, zfn, tempDir}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Unarchiver(tt.args.source, tt.args.filename, tt.args.destination); (err != nil) != tt.wantErr {
				t.Errorf("Unarchiver() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
	if err := os.RemoveAll(tempDir); err != nil {
		log.Print(err)
	}
}

func Test_configure(t *testing.T) {
	rar, err := archiver.ByExtension(".tar")
	if err != nil {
		t.Error(err)
	}
	zip, err := archiver.ByExtension(".zip")
	if err != nil {
		t.Error(err)
	}
	tests := []struct {
		name    string
		f       interface{}
		wantErr bool
	}{
		{"empty", nil, true},
		{"err", "", true},
		{"rar", rar, false},
		{"zip", zip, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := configure(tt.f); (err != nil) != tt.wantErr {
				t.Errorf("configure() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
