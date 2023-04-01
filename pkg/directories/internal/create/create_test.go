package create_test

import (
	"os"
	"testing"

	"github.com/Defacto2/df2/pkg/directories/internal/create"
)

func TestCreate(t *testing.T) {
	t.Parallel()
	tempDir, err := os.MkdirTemp(os.TempDir(), "test-create-dir")
	if err != nil {
		t.Error(err)
	}
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"empty", "", true},
		{"new", tempDir, false},
		{"temp", os.TempDir(), false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := create.MkDir(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("MkDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
	if err := os.RemoveAll(tempDir); err != nil {
		t.Error(err)
	}
}

func TestHolders(t *testing.T) {
	t.Parallel()
	tempDir, err := os.MkdirTemp(os.TempDir(), "test-create-holders")
	if err != nil {
		t.Error(err)
	}
	type args struct {
		dir    string
		size   int
		number uint
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"empty", args{}, false},
		{"temp", args{tempDir, 50, 1}, false},
		{"too big", args{tempDir, 50, 100}, true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := create.Holders(tt.args.dir, tt.args.size, tt.args.number); (err != nil) != tt.wantErr {
				t.Errorf("Holders() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
	if err := os.RemoveAll(tempDir); err != nil {
		t.Error(err)
	}
	if err := os.Remove("00000000-0000-0000-0000-000000000000"); err != nil {
		t.Error(err)
	}
}
