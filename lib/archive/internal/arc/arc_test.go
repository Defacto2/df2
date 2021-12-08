package arc_test

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/Defacto2/df2/lib/archive/internal/arc"
	"github.com/Defacto2/df2/lib/archive/internal/file"
	"github.com/mholt/archiver"
)

func TestConfigure(t *testing.T) {
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
			if err := arc.Configure(tt.f); (err != nil) != tt.wantErr {
				t.Errorf("configure() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDir(t *testing.T) {
	tempDir, err := ioutil.TempDir(os.TempDir(), "test-dir")
	if err != nil {
		t.Error(err)
	}
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"", true},
		{tempDir, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := file.Dir(tt.name); (err != nil) != tt.wantErr {
				t.Errorf("Dir() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
	if err := os.RemoveAll(tempDir); err != nil {
		log.Print(err)
	}
}
