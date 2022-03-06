package file_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Defacto2/df2/pkg/assets/internal/file"
	"github.com/gookit/color"
)

const empty = "empty"

func TestWalkName(t *testing.T) {
	const tmp = "file.temp"
	dir := os.TempDir()
	valid, err := file.WalkName(dir, filepath.Join(dir, tmp))
	if err != nil {
		t.Error(err)
	}
	type args struct {
		basepath string
		path     string
	}
	tests := []struct {
		name     string
		args     args
		wantName string
		wantErr  bool
	}{
		{empty, args{}, "", true},
		{"empty path", args{dir, ""}, "", true},
		{"empty dir", args{"", tmp}, tmp, false},
		{"dir", args{dir, tmp}, "", true},
		{"dir", args{dir, filepath.Join(dir, tmp)}, valid, false},
	}
	color.Enable = false
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, err := file.WalkName(tt.args.basepath, tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("WalkName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotName != tt.wantName {
				t.Errorf("WalkName() = %v, want %v", gotName, tt.wantName)
			}
		})
	}
}
