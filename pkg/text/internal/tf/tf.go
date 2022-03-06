package tf

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Defacto2/df2/pkg/archive"
	"github.com/Defacto2/df2/pkg/directories"
	"github.com/Defacto2/df2/pkg/images"
	"github.com/Defacto2/df2/pkg/logs"
	"github.com/Defacto2/df2/pkg/str"
	"github.com/Defacto2/df2/pkg/text/internal/img"
	"github.com/dustin/go-humanize"
	"github.com/gookit/color"
)

var (
	ErrMeNo  = errors.New("no readme chosen")
	ErrMeUnk = errors.New("unknown readme")
	ErrMeNF  = errors.New("readme not found in archive")
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
	ID       uint           // MySQL auto increment Id.
	UUID     string         // Unique Id.
	Name     string         // Filename.
	Ext      string         // File extension.
	Platform string         // Platform classification of the file.
	Size     int            // Size of the file in bytes.
	NoReadme sql.NullBool   // Disable the display of a readme on the site.
	Readme   sql.NullString // Filename of a readme or NFO textfile to display on the site.
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
	dirs := [2]string{dir.Img000, dir.Img400}
	for _, path := range dirs {
		if path == "" {
			return false, nil
		}
		s, err := os.Stat(filepath.Join(path, t.UUID+png))
		if os.IsNotExist(err) {
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
func (t *TextFile) Extract(dir *directories.Dir) error {
	if t.NoReadme.Valid && !t.NoReadme.Bool {
		return ErrMeNo
	}
	if !t.Readme.Valid {
		return ErrMeUnk
	}
	s := strings.Split(t.Readme.String, ",")
	f, err := archive.Read(filepath.Join(dir.UUID, t.UUID), t.Name)
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
	if err1 := archive.Extractor(src, t.Name, s[0], tmp); err1 != nil {
		return err1
	}
	if err = os.Rename(dest, src+txt); err != nil {
		defer os.Remove(dest)
		return fmt.Errorf("textfile extract: %w", err)
	}
	return nil
}

// ExtractedImgs generates PNG and Webp image assets from a textfile extracted from an archive.
func (t *TextFile) ExtractedImgs(dir string) error {
	j := filepath.Join(dir, t.UUID) + txt
	n, err := filepath.Abs(j)
	if err != nil {
		return fmt.Errorf("extractedimgs: %w", err)
	}
	fmt.Println("n", n)
	if _, err := os.Stat(n); os.IsNotExist(err) {
		return fmt.Errorf("extractedimgs: %w", os.ErrNotExist)
	} else if err != nil {
		return fmt.Errorf("extractedimgs: %s: %w", t.UUID, err)
	}
	amiga := bool(t.Platform == amigaTxt)
	if err := img.Generate(n, t.UUID, amiga); err != nil {
		return fmt.Errorf("extractedimgs: %w", err)
	}
	return nil
}

// TextPng generates PNG format image assets from a textfile.
func (t *TextFile) TextPng(c int, dir string) error {
	logs.Printf("%d. %v", c, t)
	name := filepath.Join(dir, t.UUID)
	if _, err := os.Stat(name); os.IsNotExist(err) {
		logs.Printf("%s\n", str.X())
		return nil
	} else if err != nil {
		return fmt.Errorf("txtpng: %w", err)
	}
	amiga := bool(t.Platform == amigaTxt)
	if err := img.Generate(name, t.UUID, amiga); err != nil {
		return fmt.Errorf("txtpng: %w", err)
	}
	logs.Print("\n")
	return nil
}

// WebP finds and generates missing WebP format images.
func (t *TextFile) WebP(c int, imgDir string) (int, error) {
	c++
	name := filepath.Join(imgDir, t.UUID+webp)
	if sw, err := os.Stat(name); err == nil && sw.Size() > 0 {
		c--
		return c, nil
	} else if !os.IsNotExist(err) && err != nil {
		logs.Printf("%s\n", str.X())
		return c, fmt.Errorf("txtwebp stat: %w", err)
	}
	logs.Printf("%d. %v", c, t)
	src := filepath.Join(imgDir, t.UUID+png)
	if st, err := os.Stat(src); os.IsNotExist(err) || st.Size() == 0 {
		logs.Printf("%s (no src png)\n", str.X())
		return c, nil
	}
	s, err := images.ToWebp(src, name, true)
	if err != nil {
		logs.Printf("%s\n", str.X())
		return c, fmt.Errorf("txtwebp: %w", err)
	}
	logs.Printf("%s %s\n", s, str.Y())
	return c, nil
}
