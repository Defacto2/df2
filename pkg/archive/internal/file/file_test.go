package file_test

import (
	"archive/tar"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/Defacto2/df2/pkg/archive/internal/file"
	"github.com/stretchr/testify/assert"
)

func join(s string) string {
	x := filepath.Join("text", s)
	return testDir(x)
}

func testDir(name string) string {
	dir, _ := os.Getwd()
	return filepath.Join(dir, "..", "..", "..", "..", "tests", name)
}

func TestAdd(t *testing.T) {
	err := file.Add(nil, "")
	assert.NotNil(t, err)
	err = file.Add(nil, testDir("file-does-not-exists"))
	assert.NotNil(t, err)

	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	err = file.Add(tw, testDir("text/test.png"))
	assert.Nil(t, err)
	err = file.Add(tw, testDir("text/test.txt"))
	assert.Nil(t, err)
}

func TestDir(t *testing.T) {
	err := file.Dir(nil, "")
	assert.NotNil(t, err)
	err = file.Dir(io.Discard, testDir(""))
	assert.Nil(t, err)
}

func TestMove(t *testing.T) {
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
		{"one way", args{join("test.txt"), join("test.txt~")}, 12, false},
		{"restore way", args{join("test.txt~"), join("test.txt")}, 12, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotWritten, err := file.Move(tt.args.name, tt.args.dest)
			if (err != nil) != tt.wantErr {
				t.Errorf("Move() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotWritten != tt.wantWritten {
				t.Errorf("Move() = %v, want %v", gotWritten, tt.wantWritten)
			}
		})
	}
}

func TestCopy(t *testing.T) {
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
		{"empty", args{join("test.txt"), join("test.txt")}, 12, false},
		{"empty", args{join("test.txt"), join("test.txt~")}, 12, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotWritten, err := file.Copy(tt.args.name, tt.args.dest)
			if (err != nil) != tt.wantErr {
				t.Errorf("Copy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotWritten != tt.wantWritten {
				t.Errorf("Copy() = %v, want %v", gotWritten, tt.wantWritten)
			}
			if err == nil && tt.args.name != tt.args.dest {
				if _, err := os.Stat(tt.args.dest); !os.IsNotExist(err) {
					if err = os.Remove(tt.args.dest); err != nil {
						t.Errorf("Copy() failed to cleanup copy %v", tt.args.dest)
					}
				}
			}
		})
	}
}
