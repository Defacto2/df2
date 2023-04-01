package arc_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Defacto2/df2/pkg/archive/internal/arc"
	"github.com/Defacto2/df2/pkg/archive/internal/sys"
	"github.com/mholt/archiver"
	"github.com/stretchr/testify/assert"
)

func testDir(name string) string {
	dir, _ := os.Getwd()
	return filepath.Join(dir, "..", "..", "..", "..", "testdata", name)
}

func TestWalkr(t *testing.T) {
	t.Parallel()
	err := arc.Walkr("", "", nil)
	assert.NotNil(t, err)
	// test panic
	err = arc.Walkr(testDir("demozoo/test.tar"), "test.tar", nil)
	assert.NotNil(t, err)

	err = arc.Walkr(testDir("demozoo/test.zip"), "test.zip",
		func(f archiver.File) error {
			return nil
		})
	assert.Nil(t, err)
}

func TestConfigure(t *testing.T) {
	t.Parallel()
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
		f       any
		wantErr bool
	}{
		{"empty", nil, true},
		{"err", "", true},
		{"rar", rar, false},
		{"zip", zip, false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := arc.Configure(tt.f); (err != nil) != tt.wantErr {
				t.Errorf("configure() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMagicExt(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		src     string
		want    string
		wantErr bool
	}{
		{"empty", "", "", true},
		{"invalid rar", testDir("demozoo/test.invalid.ext.rar"), ".zip", false},
		{"7zip", testDir("demozoo/test.7z"), ".7z", false},
		{"bz2", testDir("demozoo/test.tar.bz2"), ".tar.bz2", false},
		{"gz", testDir("demozoo/test.tar.gz"), ".tar.gz", false},
		{"tar", testDir("demozoo/test.tar"), ".tar", false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := sys.MagicExt(tt.src)
			if (err != nil) != tt.wantErr {
				t.Errorf("MagicExt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("MagicExt() = %v, want %v", got, tt.want)
			}
		})
	}
}
