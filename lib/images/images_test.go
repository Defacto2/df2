package images

import (
	_ "image/gif"
	_ "image/jpeg"
	"log"
	"os"
	"path/filepath"
	"testing"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

func testDir(name string) string {
	dir, _ := os.Getwd()
	return filepath.Join(dir, "..", "..", "tests", name)
}

func testImg(ext string) string {
	return filepath.Join(testDir("images"), "test."+ext)
}

func testDest(ext string) string {
	return filepath.Join(testDir("images"), "test-clone."+ext)
}

func testSqr() string {
	return filepath.Join(testDir("images"), "test-thumb.png")
}

func testTxt() string {
	return filepath.Join(testDir("text"), "test.txt")
}

func TestNewExt(t *testing.T) {
	type args struct {
		name      string
		extension string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"", args{"hello", ".txt"}, "hello.txt"},
		{"", args{"hello.jpg", ".png"}, "hello.png"},
		{"", args{"hello", ""}, "hello"},
		{"", args{"", ".ssh"}, ".ssh"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewExt(tt.args.name, tt.args.extension); got != tt.want {
				t.Errorf("NewExt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDuplicate(t *testing.T) {
	type args struct {
		filename string
		prefix   string
	}
	path := testImg("png")
	want := filepath.Join(testDir("images"), "test-duplicate.png")
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
			gotName, err := Duplicate(tt.args.filename, tt.args.prefix)
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
	tests := []struct {
		name       string
		wantHeight int
		wantWidth  int
		wantFormat string
		wantErr    bool
	}{
		{"", 0, 0, "", true},
		{testTxt(), 0, 0, "", true},
		{testImg("png"), 32, 1280, "png", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotWidth, gotHeight, gotFormat, err := Info(tt.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("Info() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotHeight != tt.wantHeight {
				t.Errorf("Info() gotHeight = %v, want %v", gotHeight, tt.wantHeight)
			}
			if gotWidth != tt.wantWidth {
				t.Errorf("Info() gotWidth = %v, want %v", gotWidth, tt.wantWidth)
			}
			if gotFormat != tt.wantFormat {
				t.Errorf("Info() gotFormat = %v, want %v", gotFormat, tt.wantFormat)
			}
		})
	}
}

func TestGenerate(t *testing.T) {
	dir := testDir("images")
	const (
		gif = ".gif"
		jpg = ".jpg"
		png = ".png"
		wbm = ".wbm"
	)
	type args struct {
		src    string
		id     string
		remove bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"empty", args{}, true},
		{"gif", args{filepath.Join(dir, "test"+gif), "testgen", false}, false},
		{"jpg", args{filepath.Join(dir, "test"+jpg), "testgen", false}, false},
		{"png", args{filepath.Join(dir, "test"+png), "testgen", false}, false},
		{"wbm", args{filepath.Join(dir, "test"+wbm), "testgen", false}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Generate(tt.args.src, tt.args.id, tt.args.remove); (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
	for _, ext := range []string{gif, jpg, png, wbm} {
		if err := os.Remove(filepath.Join(dir, "test_400x"+ext)); err != nil {
			log.Print(err)
		}
	}
}

func TestWidth(t *testing.T) {
	tests := []struct {
		name      string
		wantWidth int
		wantErr   bool
	}{
		{"", 0, true},
		{testTxt(), 0, true},
		{testImg("png"), 1280, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotWidth, err := Width(tt.name)
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

func TestToPng(t *testing.T) {
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
		{"noDest", args{testImg("png"), "", 0}, true},
		{"noSrc", args{"", testDir("text"), 0}, true},
		{"self", args{testImg("png"), testDest("png"), 0}, false},
		{"jpg", args{testImg("jpg"), testDest("png"), 0}, false},
		{"gif", args{testImg("gif"), testDest("png"), 0}, false},
		{"unsupported format", args{testImg("wbm"), testDest("png"), 0}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ToPng(tt.args.src, tt.args.dest, tt.args.height, 0)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToPng() error = %v, wantErr %v", err, tt.wantErr)
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
		{"zero size", args{testImg("png"), testSqr(), 0}, true},
		{"png", args{testImg("png"), testSqr(), 100}, false},
		{"gif", args{testImg("gif"), testSqr(), 100}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str, err := ToThumb(tt.args.src, tt.args.dest, tt.args.sizeSquared)
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
		{"gif", args{testImg("gif"), testDest("webp")}, true},
		{"jpg", args{testImg("jpg"), testDest("webp")}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str, err := ToWebp(tt.args.src, tt.args.dest, false)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToWebp() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if str != "" {
				// cleanup
				if err := os.Remove(testDest("webp")); err != nil {
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
		{"gif", args{testImg("gif"), testDest("webp")}, "", true},
		{"jpg", args{testImg("jpg"), testDest("webp")}, "»webp", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPrint, err := ToWebp(tt.args.src, tt.args.dest, true)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToWebp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotPrint != tt.wantPrint {
				t.Errorf("ToWebp() = %v, want %v", gotPrint, tt.wantPrint)
			}
			if gotPrint != "" {
				// cleanup
				if err := os.Remove(testDest("webp")); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestWebPCalc(t *testing.T) {
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
		{"15000 long", args{5000, 15000}, 11383, 5000},
		{"super long", args{640, 869356}, 15743, 640},
		{"square", args{15000, 15000}, 8191, 8191},
		{"sm square", args{500, 500}, 500, 500},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotW, gotH := WebPCalc(tt.args.width, tt.args.height)
			if gotW != tt.wantW {
				t.Errorf("WebPCalc() gotW = %v, want %v", gotW, tt.wantW)
			}
			if gotH != tt.wantH {
				t.Errorf("WebPCalc() gotH = %v, want %v", gotH, tt.wantH)
			}
		})
	}
}
