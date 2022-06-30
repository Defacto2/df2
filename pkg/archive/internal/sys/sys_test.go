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
		{"wrong ext arj", args{testDir("demozoo/test.invalid.ext.arj"),
			"test.invalid.ext.arj"}, "", ".zip", true},
		{"wrong ext rar", args{testDir("demozoo/test.invalid.ext.rar"),
			"test.invalid.ext.rar"}, "", ".zip", true},
		{"no support 7z", args{testDir("demozoo/test.7z"), "test.7z"},
			"", "", true},
		{"arj", args{testDir("demozoo/test.arj"), "test.arj"},
			okay, ".arj", false},
		{"zip deflate", args{testDir("demozoo/test.zip"), "test.zip"},
			okay, ".zip", false},
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

func TestArjItem(t *testing.T) {
	tests := []struct {
		s    string
		want bool
	}{
		{"", false},
		{"001) ", false},
		{"001] somefile", false},
		{"x01) somefile", false},
		{"01) somefile", false},
		{"001) somefile", true},
		{"001) some file.txt", true},
		{"050) somefile", true},
		{"999) somefile", true},
	}
	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			if got := sys.ArjItem(tt.s); got != tt.want {
				t.Errorf("ArjItem() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtract(t *testing.T) {
	type args struct {
		src     string
		targets string
		dest    string
	}
	const tgt = "test.png"
	arj := testDir("demozoo/test.arj")
	zip := testDir("demozoo/test.zip")
	tmp, err := os.MkdirTemp(os.TempDir(), "sys-extract")
	if err != nil {
		t.Error("Extract() error = %w", err)
		return
	}
	defer os.RemoveAll(tmp)
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"empty", args{}, true},
		{"arj", args{arj, tgt, tmp}, false},
		{"zip", args{zip, tgt, tmp}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name := filepath.Base(tt.args.src)
			if err := sys.Extract(name, tt.args.src, tt.args.targets, tt.args.dest); (err != nil) != tt.wantErr {
				t.Errorf("Extract() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
