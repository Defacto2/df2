package sys_test

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Defacto2/df2/pkg/archive/internal/sys"
)

func testDir(name string) string {
	dir, _ := os.Getwd()
	return filepath.Join(dir, "..", "..", "..", "..", "testdata", name)
}

func TestRename(t *testing.T) {
	type args struct {
		ext      string
		filename string
	}
	tests := []struct {
		args args
		want string
	}{
		{args{}, ""},
		{args{"", "somefile"}, "somefile"},
		{args{"", "some.file"}, "some"},
		{args{"txt", "somefile"}, "somefile.txt"},
		{args{"text", "some.file"}, "some.text"},
		{args{".txt", "some.file"}, "some.txt"},
		{args{".txt", "some.file.text"}, "some.file.txt"},
	}
	for _, tt := range tests {
		t.Run(tt.args.filename, func(t *testing.T) {
			if got := sys.Rename(tt.args.ext, tt.args.filename); got != tt.want {
				t.Errorf("Rename() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestPKZip tests against archives created with various versions
// of the MS-DOS PKZIP that was commonly used on PC BBSes.
// Each major version of PKZIP introduced new but incompatible
// compression methods.
// The test files were created by me but sourced from:
// https://github.com/jvilk/browserfs-zipfs-extras/tree/master/test/fixtures
func TestPKZip(t *testing.T) {
	const okay = "TEST.ANS;TEST.ASC;TEST.BMP;TEST.CAP;TEST.DIZ;TEST.DOC;TEST.EXE;TEST.GIF;" +
		"TEST.JPG;TEST.ME;TEST.NFO;TEST.PCX;TEST.PNG;TEST.TXT;TEST~1.JPE;"
	v080a1 := testDir("pkzip/PKZ80A1.ZIP")
	v080b4 := testDir("pkzip/PKZ80B4.ZIP")
	v110 := testDir("pkzip/PKZ110.ZIP")
	v110i := testDir("pkzip/PKZ110EI.ZIP")
	v110e := testDir("pkzip/PKZ110ES.ZIP")
	v110x := testDir("pkzip/PKZ110EX.ZIP")
	v20 := testDir("pkzip/PKZ204E0.ZIP")
	v2f := testDir("pkzip/PKZ204EF.ZIP")
	v2n := testDir("pkzip/PKZ204EN.ZIP")
	v2s := testDir("pkzip/PKZ204ES.ZIP")
	v2x := testDir("pkzip/PKZ204EX.ZIP")

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
		{"reduce ascii 1 compression", args{v080a1, "PKZ80A1.ZIP"}, okay, ".zip", false},
		{"reduce binary 4 compression", args{v080b4, "PKZ80B4.ZIP"}, okay, ".zip", false},
		{"reduce using no args", args{v110, "PKZ110.ZIP"}, okay, ".zip", false},
		{"implode method", args{v110i, "PKZ110EI.ZIP"}, okay, ".zip", false},
		{"shrink method", args{v110e, "PKZ110ES.ZIP"}, okay, ".zip", false},
		{"maXimal compression", args{v110x, "PKZ110EX.ZIP"}, okay, ".zip", false},
		{"deflate no compression", args{v20, "PKZ204E0.ZIP"}, okay, ".zip", false},
		{"deflate fast", args{v2f, "PKZ204EF.ZIP"}, okay, ".zip", false},
		{"deflate normal", args{v2n, "PKZ204EN.ZIP"}, okay, ".zip", false},
		{"deflate superfast", args{v2s, "PKZ204ES.ZIP"}, okay, ".zip", false},
		{"deflate extra", args{v2x, "PKZ204EX.ZIP"}, okay, ".zip", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gots, got, err := sys.Readr(nil, tt.args.src, tt.args.filename)
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

func TestReadr(t *testing.T) {
	const okay = "test.png;test.txt;"
	const extra = "ext dir/test file.text;test.png;test.txt;"
	const rarextra = "ext dir/test file.text;test.png;test.txt;ext dir;"
	arj := testDir("demozoo/test.arj")
	lha := testDir("demozoo/test.lha")
	rar := testDir("demozoo/test.rar")
	zip := testDir("demozoo/test.zip")
	z7 := testDir("demozoo/test.7z")
	errarj := testDir("demozoo/test.invalid.ext.arj")
	errrar := testDir("demozoo/test.invalid.ext.rar")
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
		{"wrong ext arj", args{errarj, "test.invalid.ext.arj"}, "", ".zip", true},
		{"wrong ext rar", args{errrar, "test.invalid.ext.rar"}, "", ".zip", true},
		{"no support 7z", args{z7, "test.7z"}, "", "", true},
		{"arj", args{arj, "test.arj"}, okay, ".arj", false},
		{"lha", args{lha, "test.lha"}, extra, ".lha", false},
		{"rar", args{rar, "test.rar"}, rarextra, ".rar", false},
		{"zip deflate", args{zip, "test.zip"}, okay, ".zip", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gots, got, err := sys.Readr(nil, tt.args.src, tt.args.filename)
			// special case for macos which doesn't have easy access to arj.
			if runtime.GOOS == "darwin" {
				if errors.Is(err, exec.ErrNotFound) {
					return
				}
			}
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
			if got := sys.ARJItem(tt.s); got != tt.want {
				t.Errorf("ARJItem() = %v, want %v", got, tt.want)
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
	lha := testDir("demozoo/test.lha")
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
		{"lha", args{lha, tgt, tmp}, false},
		{"zip", args{zip, tgt, tmp}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name := filepath.Base(tt.args.src)
			err := sys.Extract(name, tt.args.src, tt.args.targets, tt.args.dest)
			// special case for macos which doesn't have easy access to arj.
			if runtime.GOOS == "darwin" {
				if errors.Is(err, exec.ErrNotFound) {
					return
				}
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("Extract() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMagicLHA(t *testing.T) {
	tests := []struct {
		magic string
		want  bool
	}{
		{"", false},
		{"abc def ghi", false},
		{"lha archive data", true},
		{"lha 2.x? archive data  [lha]", true},
	}
	for _, tt := range tests {
		t.Run(tt.magic, func(t *testing.T) {
			if got := sys.MagicLHA(tt.magic); got != tt.want {
				t.Errorf("MagicLHA() = %v, want %v", got, tt.want)
			}
		})
	}
}
