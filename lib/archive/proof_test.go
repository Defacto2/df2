package archive

import (
	"os"
	"path/filepath"
	"testing"
)

func Test_extract(t *testing.T) {
	type args struct {
		archive string
		tempDir string
	}

	var zip = filepath.Join(tmp, "zip")
	os.RemoveAll(filepath.Join(zip))
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"empty", args{"", ""}, true},
		{"zip", args{testDir("demozoo/test.zip"), zip}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := extract(tt.args.archive, tt.args.tempDir); (err != nil) != tt.wantErr {
				t.Errorf("extract() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExtract(t *testing.T) {
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
		//{"zip", args{testDir("demozoo/test.zip"), "test.zip", "6ba7b814-9dad-11d1-80b4-00c04fd430c8"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Extract(tt.args.archive, tt.args.filename, tt.args.uuid); (err != nil) != tt.wantErr {
				t.Errorf("Extract() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
