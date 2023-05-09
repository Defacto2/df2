// Package tf has the functions to extract and convert text files into images.
package tf

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/Defacto2/df2/pkg/archive"
	"github.com/Defacto2/df2/pkg/conf"
	"github.com/Defacto2/df2/pkg/directories"
	"github.com/Defacto2/df2/pkg/images"
	"github.com/Defacto2/df2/pkg/str"
	"github.com/Defacto2/df2/pkg/text/internal/img"
	"github.com/dustin/go-humanize"
	"github.com/gookit/color"
)

var (
	ErrReadmeBlank = errors.New("readme file cannot be blank or invalid")
	ErrReadmeOff   = errors.New("noreadme bool cannot be false")
	ErrMeNF        = errors.New("readme not found in archive")
	ErrPNG         = errors.New("no such png file")
	ErrPointer     = errors.New("pointer value cannot be nil")
	ErrUUID        = errors.New("readme file cannot be blank or invalid")
)

const (
	// Images.
	png  = ".png"
	webp = ".webp"
	// Texts.
	txt = ".txt"
	// Archives.
	z7  = ".7z"
	arc = ".arc"
	arj = ".arj"
	bz2 = ".bz2"
	cab = ".cab"
	gz  = ".gz"
	lha = ".lha"
	lzh = ".lzh"
	rar = ".rar"
	tar = ".tar"
	tgz = ".tar.gz"
	zip = ".zip"

	amigaTxt = "textamiga"
)

// TextFile is a text file object.
type TextFile struct {
	ID       uint           // ID is a database auto increment ID.
	UUID     string         // Universal unique ID.
	Name     string         // Name of the file.
	Ext      string         // Ext is the file extension of the name.
	Platform string         // Platform classification for the file.
	Size     int64          // Size of the file in bytes.
	NoReadme sql.NullBool   // NoReadme will disable the display of a readme file on the webpage.
	Readme   sql.NullString // Readme is the filename of a readme or NFO textfile to display on the webpage.
}

func (t *TextFile) String() string {
	return fmt.Sprintf("(%v) %v %v ", color.Primary.Sprint(t.ID), t.Name,
		color.Info.Sprint(humanize.Bytes(uint64(t.Size))))
}

// Archive confirms that the named file is a known archive.
func (t *TextFile) Archive() bool {
	switch filepath.Ext(strings.ToLower(t.Name)) {
	case z7, arc, arj, bz2, cab, gz, lha, lzh, rar, tar, tgz, zip:
		return true
	}
	return false
}

// Exist checks that [UUID].png exists in both thumbnail subdirectories.
func (t *TextFile) Exist(dir *directories.Dir) (bool, error) {
	if dir == nil {
		return false, fmt.Errorf("dir %w", ErrPointer)
	}
	dirs := [2]string{dir.Img000, dir.Img400}
	for _, path := range dirs {
		if path == "" {
			return false, nil
		}
		s, err := os.Stat(filepath.Join(path, t.UUID+png))
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		} else if err != nil {
			return false, fmt.Errorf("textfile exist: %w", err)
		}
		if s.Size() == 0 {
			return false, nil
		}
	}
	return true, nil
}

// Extract a textfile readme from an archive.
func (t *TextFile) Extract(w io.Writer, dir *directories.Dir) error {
	if dir == nil {
		return fmt.Errorf("dir %w", ErrPointer)
	}
	if t.NoReadme.Valid && !t.NoReadme.Bool {
		return ErrReadmeOff
	}
	if !t.Readme.Valid {
		return ErrReadmeBlank
	}
	if w == nil {
		w = io.Discard
	}
	s := strings.Split(t.Readme.String, ",")
	f, _, err := archive.Read(w, filepath.Join(dir.UUID, t.UUID), t.Name)
	if err != nil {
		return err
	}
	found := false
	for _, key := range f {
		if key == s[0] {
			found = true
			break
		}
	}
	if !found {
		return ErrMeNF
	}
	tmp, src := os.TempDir(), filepath.Join(dir.UUID, t.UUID)
	dest := filepath.Join(tmp, s[0])
	if err1 := archive.Extractor(t.Name, src, s[0], tmp); err1 != nil {
		return err1
	}
	if err = os.Rename(dest, src+txt); err != nil {
		defer os.Remove(dest)
		return fmt.Errorf("textfile extract: %w", err)
	}
	return nil
}

// ExtractedImgs generates PNG and Webp image assets from a textfile extracted from an archive.
func (t *TextFile) ExtractedImgs(w io.Writer, cfg conf.Config, dir string) error {
	if w == nil {
		w = io.Discard
	}
	j := filepath.Join(dir, t.UUID) + txt
	n, err := filepath.Abs(j)
	if err != nil {
		return fmt.Errorf("extractedimgs: %w", err)
	}
	if _, err := os.Stat(n); errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("%w: %s", os.ErrNotExist, n)
	} else if err != nil {
		return fmt.Errorf("extractedimgs: %s: %w", t.UUID, err)
	}
	amiga := bool(t.Platform == amigaTxt)
	if err := img.Make(w, cfg, n, t.UUID, amiga); err != nil {
		return fmt.Errorf("extractedimgs: %w", err)
	}
	return nil
}

// TextPNG generates PNG format image assets from a textfile.
func (t *TextFile) TextPNG(w io.Writer, cfg conf.Config, dir string) error {
	if w == nil {
		w = io.Discard
	}
	name := filepath.Join(dir, t.UUID)
	if _, err := os.Stat(name); errors.Is(err, fs.ErrNotExist) {
		fmt.Fprintf(w, "%s ", str.X())
		return fmt.Errorf("%w: %s", ErrPNG, name)
	} else if err != nil {
		return fmt.Errorf("txtpng: %w", err)
	}
	amiga := bool(t.Platform == amigaTxt)
	if err := img.Make(w, cfg, name, t.UUID, amiga); err != nil {
		return fmt.Errorf("txtpng: %w", err)
	}
	return nil
}

// WebP finds and generates missing WebP format images.
func (t *TextFile) WebP(w io.Writer, c int, imgDir string) (int, error) {
	if t.UUID == "" {
		return c, ErrUUID
	}
	if w == nil {
		w = io.Discard
	}
	c++
	name := filepath.Join(imgDir, t.UUID+webp)
	if sw, err := os.Stat(name); !errors.Is(err, fs.ErrNotExist) && err != nil {
		fmt.Fprintf(w, "%s\n", str.X())
		return c, fmt.Errorf("webp stat %w: %s", err, name)
	} else if err == nil && sw.Size() > 0 {
		// skip any existing webp images
		c--
		return c, nil
	}
	// fmt.Fprintf(w, "\t%d. %v", c, t)
	src := filepath.Join(imgDir, t.UUID+png)
	if st, err := os.Stat(src); errors.Is(err, fs.ErrNotExist) || st.Size() == 0 {
		fmt.Fprintf(w, "%s (no src png)\n", str.X())
		return c, nil
	}
	s, err := images.ToWebp(w, src, name, true)
	if err != nil {
		fmt.Fprintf(w, "%s\n", str.X())
		return c, fmt.Errorf("txtwebp: %w", err)
	}
	fmt.Fprintf(w, "%s %s\n", s, str.Y())
	return c, nil
}
