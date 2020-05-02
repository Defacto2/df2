package archive

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

var tmp = os.TempDir()

func Test_extractr(t *testing.T) {
	type args struct {
		archive  string
		filename string
		tempDir  string
	}
	var zip = filepath.Join(tmp, "zip")
	os.RemoveAll(filepath.Join(zip))
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"empty", args{"", "", ""}, true},
		{"zip", args{testDir("demozoo/test.zip"), "test.zip", zip}, false},
		{"7z", args{testDir("demozoo/test.7z"), "test.7z", ""}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := extractr(tt.args.archive, tt.args.filename, tt.args.tempDir); (err != nil) != tt.wantErr {
				t.Errorf("extractr() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

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
