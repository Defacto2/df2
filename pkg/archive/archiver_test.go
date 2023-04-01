package archive_test

import (
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/Defacto2/df2/pkg/archive"
)

func TestReadr(t *testing.T) {
	t.Parallel()
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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotFiles, _, err := archive.Readr(nil, tt.args.archive, tt.args.filename)
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
	t.Parallel()
	dir, err := os.MkdirTemp(os.TempDir(), "unarchiver")
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
		{"missing src", args{"", fn, dir}, true},
		{"missing fn", args{src, "", dir}, true},
		{"missing dest", args{src, fn, ""}, true},
		{"okay", args{src, fn, dir}, false},
		{"7z", args{z7, zfn, dir}, true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := archive.Unarchiver(tt.args.source, tt.args.destination, tt.args.filename); (err != nil) != tt.wantErr {
				t.Errorf("Unarchiver() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
	if err := os.RemoveAll(dir); err != nil {
		log.Print(err)
	}
}
