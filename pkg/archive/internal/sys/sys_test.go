package sys_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Defacto2/df2/pkg/archive/internal/sys"
)

func testDir(name string) string {
	dir, _ := os.Getwd()
	return filepath.Join(dir, "..", "..", "..", "..", "tests", name)
}

func TestReadr(t *testing.T) {
	const okay = "test.png;test.txt;"
	type args struct {
		src      string
		filename string
	}
	tests := []struct {
		name    string
		args    args
		wants   string
		wantExt string
		wantErr bool
	}{
		{"empty", args{}, "", "", true},
		{"invalid rar", args{testDir("demozoo/test.invalid.ext.rar"), "test.invalid.ext.rar"}, "", ".zip", true},
		{"7z", args{testDir("demozoo/test.7z"), "test.7z"}, "", "", true},
		{"zip deflate", args{testDir("demozoo/test.zip"), "test.zip"}, okay, ".zip", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gots, got, err := sys.Readr(tt.args.src, tt.args.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("Readr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.wantExt {
				t.Errorf("Readr() = %v, want %v", got, tt.wantExt)
				return
			}
			x := strings.Join(gots, ";")
			if x != tt.wants {
				t.Errorf("Readr() = %v, want %v", x, tt.wants)
			}
		})
	}
}
