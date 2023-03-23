package images_test

import (
	"fmt"
	_ "image/gif"
	_ "image/jpeg"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/Defacto2/df2/pkg/configger"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/images"
	"github.com/stretchr/testify/assert"
	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

const imgs, g, j, p, w = "images", "gif", "jpg", "png", "webp"

func testDir(name string) string {
	dir, _ := os.Getwd()
	return filepath.Join(dir, "..", "..", "testdata", name)
}

func testImg(ext string) string {
	return filepath.Join(testDir(imgs), "test."+ext)
}

func testDest(ext string) string {
	return filepath.Join(testDir(imgs), "test-clone."+ext)
}

func testSqr() string {
	return filepath.Join(testDir(imgs), "test-thumb.png")
}

func testTxt() string {
	return filepath.Join(testDir("text"), "test.txt")
}

func TestReplaceExt(t *testing.T) {
	s := images.ReplaceExt(".txt", "document")
	assert.Equal(t, "document.txt", s)
	s = images.ReplaceExt(".png", "image.jpg")
	assert.Equal(t, "image.png", s)
	s = images.ReplaceExt("", "image")
	assert.Equal(t, "image", s)
	s = images.ReplaceExt(".png", "")
	assert.Equal(t, ".png", s)
}

func TestDuplicate(t *testing.T) {
	type args struct {
		filename string
		prefix   string
	}
	path := testImg(p)
	want := filepath.Join(testDir(imgs), "test-duplicate.png")
	tests := []struct {
		name     string
		args     args
		wantName string
		wantErr  bool
	}{
		{"empty", args{}, "", true},
		{"error", args{"hello.jpg", "my"}, "", true},
		{"ok", args{path, "-duplicate"}, want, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, err := images.Duplicate(tt.args.filename, tt.args.prefix)
			if (err != nil) != tt.wantErr {
				t.Errorf("Duplicate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotName != tt.wantName {
				t.Errorf("Duplicate() = %v, want %v", gotName, tt.wantName)
			}
		})
	}
}

func TestInfo(t *testing.T) {
	w, x, f, err := images.Info("")
	assert.NotNil(t, err)
	assert.Equal(t, 0, w)
	assert.Equal(t, 0, x)
	assert.Equal(t, "", f)
	w, x, f, err = images.Info(testTxt())
	assert.NotNil(t, err)
	assert.Equal(t, 0, w)
	assert.Equal(t, 0, x)
	assert.Equal(t, "", f)
	w, x, f, err = images.Info(testImg(p))
	assert.Nil(t, err)
	assert.Equal(t, 1280, w)
	assert.Equal(t, 32, x)
	assert.Equal(t, p, f)
}

func TestGenerate(t *testing.T) {
	const (
		gif = ".gif"
		jpg = ".jpg"
		png = ".png"
		wbm = ".wbm"
		ts  = "test"
		tg  = "testgen"
	)
	dir := testDir(imgs)

	err := images.Generate(nil, "", "", false)
	assert.NotNil(t, err)
	err = images.Generate(io.Discard, "", tg, false)
	assert.NotNil(t, err)
	err = images.Generate(io.Discard, filepath.Join(dir, ts+gif), tg, false)
	assert.Nil(t, err)
	err = images.Generate(io.Discard, filepath.Join(dir, ts+wbm), tg, false)
	assert.NotNil(t, err)
}

func TestLibraries(t *testing.T) {
	const (
		gif = ".gif"
		jpg = ".jpg"
		png = ".png"
		wbm = ".wbm"
		ts  = "test"
		tg  = "testgen"
	)
	dir := testDir(imgs)

	err := images.Libraries(nil, "", "", false)
	assert.NotNil(t, err)
	err = images.Libraries(io.Discard, "", tg, false)
	assert.NotNil(t, err)
	err = images.Libraries(io.Discard, filepath.Join(dir, ts+gif), tg, false)
	assert.Nil(t, err)
}

func TestWidth(t *testing.T) {
	tests := []struct {
		name      string
		wantWidth int
		wantErr   bool
	}{
		{"", 0, true},
		{testTxt(), 0, true},
		{testImg(p), 1280, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotWidth, err := images.Width(tt.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("Width() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotWidth != tt.wantWidth {
				t.Errorf("Width() = %v, want %v", gotWidth, tt.wantWidth)
			}
		})
	}
}

func TestToPNG(t *testing.T) {
	type args struct {
		src    string
		dest   string
		height int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"empty", args{"", "", 0}, true},
		{"noDest", args{testImg(p), "", 0}, true},
		{"noSrc", args{"", testDir("text"), 0}, true},
		{"self", args{testImg(p), testDest(p), 0}, false},
		{j, args{testImg(j), testDest(p), 0}, false},
		{g, args{testImg(g), testDest(p), 0}, false},
		{"unsupported format", args{testImg("wbm"), testDest(p), 0}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := images.ToPNG(tt.args.src, tt.args.dest, tt.args.height, 0)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToPNG() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestToThumb(t *testing.T) {
	type args struct {
		src         string
		dest        string
		sizeSquared int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"empty", args{"", "", 0}, true},
		{"zero size", args{testImg(p), testSqr(), 0}, true},
		{p, args{testImg(p), testSqr(), 100}, false},
		{g, args{testImg(g), testSqr(), 100}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str, err := images.ToThumb(tt.args.src, tt.args.dest, tt.args.sizeSquared)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToThumb() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if str != "" {
				// cleanup
				if err := os.Remove(testSqr()); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestToWebxp(t *testing.T) {
	type args struct {
		src  string
		dest string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"empty", args{"", ""}, true},
		{"invalid src", args{"blahblahblah", ""}, true},
		{g, args{testImg(g), testDest(w)}, false},
		{j, args{testImg(j), testDest(w)}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str, err := images.ToWebp(nil, tt.args.src, tt.args.dest, false)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToWebp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if str != "" {
				// cleanup
				if err := os.Remove(testDest(w)); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestToWebp(t *testing.T) {
	type args struct {
		src  string
		dest string
	}
	tests := []struct {
		name      string
		args      args
		wantPrint string
		wantErr   bool
	}{
		{"empty", args{"", ""}, "", true},
		{"invalid src", args{"blahblahblah", ""}, "", true},
		{g, args{testImg(g), testDest(w)}, "»webp", false},
		{j, args{testImg(j), testDest(w)}, "»webp", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPrint, err := images.ToWebp(nil, tt.args.src, tt.args.dest, true)
			if (err != nil) != tt.wantErr {
				fmt.Fprintf(os.Stderr, "%s -> %s\n", tt.args.src, tt.args.dest)
				t.Errorf("ToWebp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotPrint != tt.wantPrint {
				t.Errorf("ToWebp() = %v, want %v", gotPrint, tt.wantPrint)
			}
			if gotPrint != "" {
				// cleanup
				if err := os.Remove(testDest(w)); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestWebPCalc(t *testing.T) {
	const long = 15000
	type args struct {
		width  int
		height int
	}
	tests := []struct {
		name  string
		args  args
		wantW int
		wantH int
	}{
		{"zero", args{0, 0}, 0, 0},
		{"ignore", args{600, 500}, 600, 500},
		{"15000 long", args{5000, long}, 11383, 5000},
		{"super long", args{640, 869356}, 15743, 640},
		{"square", args{long, long}, 8191, 8191},
		{"sm square", args{500, 500}, 500, 500},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotW, gotH := images.WebPCalc(tt.args.width, tt.args.height)
			if gotW != tt.wantW {
				t.Errorf("WebPCalc() gotW = %v, want %v", gotW, tt.wantW)
			}
			if gotH != tt.wantH {
				t.Errorf("WebPCalc() gotH = %v, want %v", gotH, tt.wantH)
			}
		})
	}
}

func TestFix(t *testing.T) {
	err := images.Fix(nil, nil)
	assert.NotNil(t, err)
	db, err := database.Connect(configger.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	err = images.Fix(db, io.Discard)
	assert.Nil(t, err)
}

func TestMove(t *testing.T) {
	src, err := os.CreateTemp(os.TempDir(), "images-move-test")
	assert.Nil(t, err)
	i, err := src.WriteString("hello world\n")
	src.Sync()
	assert.Nil(t, err)
	assert.Equal(t, 12, i)
	src.Close()
	dst := filepath.Join(os.TempDir(), "images-move-test-xyz")
	err = images.Move(dst, src.Name())
	assert.Nil(t, err)
	_, err = os.Stat(dst)
	assert.Nil(t, err)
	defer os.Remove(dst)
}
