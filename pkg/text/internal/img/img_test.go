package img_test

import (
	"bufio"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/Defacto2/df2/pkg/conf"
	"github.com/Defacto2/df2/pkg/images"
	"github.com/Defacto2/df2/pkg/text/internal/img"
	"github.com/stretchr/testify/assert"
)

const (
	uuid = "_0000000-0000-0000-0000-000000000000"
)

var (
	testDir = filepath.Join("..", "..", "..", "..", "testdata")
)

func TestMake(t *testing.T) {
	t.Parallel()
	gif := filepath.Join(testDir, "images", "test.gif")
	txt := filepath.Join(testDir, "text", "test.txt")
	cfg := conf.Defaults()

	err := img.Make(nil, conf.Config{}, "", "", false)
	assert.NotNil(t, err)
	err = img.Make(io.Discard, cfg, "", "", false)
	assert.NotNil(t, err)
	err = img.Make(io.Discard, cfg, gif, uuid, false)
	assert.NotNil(t, err, "named file points to an invalid image")
	err = img.Make(io.Discard, cfg, "a-invalid-file", uuid, false)
	assert.NotNil(t, err, "named file points to a missing file")
	err = img.Make(io.Discard, cfg, txt, uuid, false)
	assert.Nil(t, err, "named text file should convert to a image with MS-DOS fonts")
	err = img.Make(io.Discard, cfg, txt, uuid, true)
	assert.Nil(t, err, "named text file should convert to a image with Amiga fonts")

	rm1 := filepath.Join(cfg.Images, uuid+".png")
	rm2 := filepath.Join(cfg.Images, uuid+".webp")
	defer os.Remove(rm1)
	defer os.Remove(rm2)
}

func TestType(t *testing.T) {
	t.Parallel()
	gif := filepath.Join(testDir, "images", "test.gif")
	iff := filepath.Join(testDir, "images", "test.iff")
	png := filepath.Join(testDir, "images", "test.png")
	wbm := filepath.Join(testDir, "images", "test.wbm")

	zip08 := filepath.Join(testDir, "pkzip", "PKZ80A1.ZIP")
	z7 := filepath.Join(testDir, "demozoo", "test.7z")
	arj := filepath.Join(testDir, "demozoo", "test.arj")
	lha := filepath.Join(testDir, "demozoo", "test.lha")
	rar := filepath.Join(testDir, "demozoo", "test.rar")
	xz := filepath.Join(testDir, "demozoo", "test.xz")
	txt := filepath.Join(testDir, "text", "test.txt")

	err := img.Type(gif)
	assert.NotNil(t, err)
	err = img.Type(png)
	assert.NotNil(t, err)
	err = img.Type(zip08)
	assert.NotNil(t, err)
	err = img.Type(rar)
	assert.NotNil(t, err)
	err = img.Type(xz)
	assert.NotNil(t, err)

	// the type test is not reliable, but still useful for the formats above
	err = img.Type(wbm)
	assert.Nil(t, err)
	err = img.Type(iff)
	assert.Nil(t, err)
	err = img.Type(z7)
	assert.Nil(t, err)
	err = img.Type(arj)
	assert.Nil(t, err)
	err = img.Type(lha)
	assert.Nil(t, err)
	err = img.Type(txt)
	assert.Nil(t, err)
}

func TestResize(t *testing.T) {
	t.Parallel()
	err := img.Resize(nil, "")
	assert.NotNil(t, err)
	// draw in memory a test image that's too big for webp
	tooLong := image.NewRGBA(image.Rect(0, 0, images.WebpMaxSize+100, 100))
	draw.Draw(tooLong, tooLong.Bounds(), &image.Uniform{image.Black}, image.Point{}, draw.Src)
	f, err := os.CreateTemp(os.TempDir(), "test-resize-toolong")
	assert.Nil(t, err)
	defer f.Close()
	// save the drawing as a png image file
	png.Encode(f, tooLong)
	name := f.Name()
	defer os.Remove(name)
	// obtain the filesize of the png image file
	st, err := f.Stat()
	assert.Nil(t, err)
	size := st.Size()
	// run Resize() to reduce the dimensions of png image file
	err = img.Resize(io.Discard, name)
	assert.Nil(t, err)
	// read and compare the filesize of the resized png image file
	st, err = os.Stat(name)
	assert.Nil(t, err)
	assert.Less(t, st.Size(), size, "the resized png should be a small file size than the original drawn png")
}

func TestReduce(t *testing.T) {
	t.Parallel()
	s, err := img.Reduce(nil, "", "")
	assert.NotNil(t, err)
	assert.Equal(t, "", s)
	s, err = img.Reduce(io.Discard, "no-such-file", "")
	assert.NotNil(t, err)
	assert.Equal(t, "", s)
	// create a new file with 600 lines of text
	f, err := os.CreateTemp(os.TempDir(), "test-reduce-textfile")
	assert.Nil(t, err)
	defer f.Close()
	defer os.Remove(f.Name())
	n := 0
	for n < 600 {
		n += 1
		_, err := f.WriteString(fmt.Sprintf("line %d\n", n))
		assert.Nil(t, err)
	}
	f.Sync()
	// Reduce file
	s, err = img.Reduce(io.Discard, f.Name(), uuid)
	assert.Nil(t, err)
	assert.NotEqual(t, "", s)
	// count the lines in the reduced files
	r, err := os.Open(s)
	assert.Nil(t, err)
	defer r.Close()
	defer os.Remove(s)
	scan := bufio.NewScanner(r)
	scan.Split(bufio.ScanLines)
	lines := 0
	for scan.Scan() {
		lines++
	}
	const expected = 500
	assert.Equal(t, expected, lines, "reduce should have created a file with no more than 500 lines of text")
}
