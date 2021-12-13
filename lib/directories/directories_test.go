package directories_test

import (
	"testing"

	"github.com/Defacto2/df2/lib/directories"
)

func TestInit(t *testing.T) {
	const (
		createDir = false
		wantUUID  = "/opt/assets/downloads"
	)
	t.Run("flat", func(t *testing.T) {
		if got := directories.Init(createDir); got.UUID != wantUUID {
			t.Errorf("Init() = %v, want %v", got.UUID, wantUUID)
		}
	})
}

func TestArchiveExt(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"empty", args{}, false},
		{"no period", args{"arj"}, false},
		{"okay", args{".arj"}, true},
		{"caps", args{".ARJ"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := directories.ArchiveExt(tt.args.name); got != tt.want {
				t.Errorf("ArchiveExt() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestFiles(t *testing.T) {
	const (
		name     = "myname"
		wantUUID = "/opt/assets/downloads/myname"
	)
	t.Run("flat", func(t *testing.T) {
		if got := directories.Files(name); got.UUID != wantUUID {
			t.Errorf("Init() = %v, want %v", got.UUID, wantUUID)
		}
	})
}

func TestSize(t *testing.T) {
	tests := []struct {
		name      string
		root      string
		wantCount int64
		wantBytes uint64
		wantErr   bool
	}{
		{"empty", "", 0, 0, true},
		{"nul", "/dev/null/no-such-dir", 0, 0, true},
		{"empty", "../../tests/empty", 0, 0, false},
		{"valid", "../../tests/demozoo", 12, 7156, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCount, gotBytes, err := directories.Size(tt.root)
			if (err != nil) != tt.wantErr {
				t.Errorf("Size() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotCount != tt.wantCount {
				t.Errorf("Size() gotCount = %v, want %v", gotCount, tt.wantCount)
			}
			if gotBytes != tt.wantBytes {
				t.Errorf("Size() gotBytes = %v, want %v", gotBytes, tt.wantBytes)
			}
		})
	}
}
