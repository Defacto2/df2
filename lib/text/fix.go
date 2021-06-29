package text

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Defacto2/df2/lib/archive"
	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/directories"
	"github.com/Defacto2/df2/lib/images"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/Defacto2/df2/lib/str"
	"github.com/dustin/go-humanize" //nolint:misspell
	"github.com/gookit/color"       //nolint:misspell
)

const (
	fixStmt = "SELECT id, uuid, filename, filesize, retrotxt_no_readme, retrotxt_readme, platform " +
		"FROM files WHERE platform=\"text\" OR platform=\"textamiga\" OR platform=\"ansi\" ORDER BY id DESC"
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
)

// textfile is a text file object.
type textfile struct {
	ID       uint           // database id
	UUID     string         // database unique id
	Name     string         // file name
	Ext      string         // file extension
	Platform string         // file platform classification
	Size     int            // file size in bytes
	NoReadme sql.NullBool   // disable the display of a readme
	Readme   sql.NullString // filename of a readme textfile
}

func (t *textfile) String() string {
	return fmt.Sprintf("(%v) %v %v ", color.Primary.Sprint(t.ID), t.Name,
		color.Info.Sprint(humanize.Bytes(uint64(t.Size))))
}

// Fix generates any missing assets from downloads that are text based.
func Fix(simulate bool) error {
	dir, db := directories.Init(false), database.Connect()
	defer db.Close()
	rows, err := db.Query(fixStmt)
	if err != nil {
		return fmt.Errorf("fix db query: %w", err)
	} else if rows.Err() != nil {
		return fmt.Errorf("fix rows: %w", rows.Err())
	}
	defer rows.Close()
	c, i := 0, 0
	for rows.Next() {
		var t textfile
		i++
		if err := rows.Scan(&t.ID, &t.UUID, &t.Name, &t.Size, &t.NoReadme, &t.Readme, &t.Platform); err != nil {
			return fmt.Errorf("fix rows scan: %w", err)
		}
		ok, err := t.exist(&dir)
		if err != nil {
			return fmt.Errorf("fix exist: %w", err)
		}
		// missing images + source is an archive
		if !ok && t.archive() {
			c++
			if err := t.extract(&dir); errors.Is(err, ErrMeUnk) {
				continue
			} else if errors.Is(err, ErrMeNo) {
				continue
			} else if err != nil {
				fmt.Println(t.String(), err)
				continue
			}
			if err := t.extractedImgs(dir.UUID); err != nil {
				fmt.Println(t.String(), err)
			}
			continue
		}
		// missing images + source is a textfile
		if !ok {
			c++
			if !t.textPng(c, dir.UUID) {
				continue
			}
		}
		// missing webp specific images that rely on PNG sources
		c, err = t.webP(c, dir.Img000)
		if err != nil {
			logs.Println(err)
		}
	}
	fmt.Println("scanned", c, "fixes from", i, "text file records")
	if simulate && c > 0 {
		logs.Simulate()
	} else if c == 0 {
		logs.Println("everything is okay, there is nothing to do")
	}
	return nil
}

// archive confirms that the named file is a known archive.
func (t *textfile) archive() bool {
	switch filepath.Ext(strings.ToLower(t.Name)) {
	case z7, arc, arj, bz2, cab, gz, lha, lzh, rar, tar, tgz, zip:
		return true
	}
	return false
}

// check that [UUID].png exists in all three image subdirectories.
func (t *textfile) exist(dir *directories.Dir) (bool, error) {
	dirs := [2]string{dir.Img000, dir.Img400}
	for _, path := range dirs {
		if path == "" {
			return false, nil
		}
		s, err := os.Stat(filepath.Join(path, t.UUID+png))
		if os.IsNotExist(err) {
			return false, nil
		} else if err != nil {
			return false, fmt.Errorf("image exist: %w", err)
		}
		if s.Size() == 0 {
			return false, nil
		}
	}
	return true, nil
}

// extract a textfile readme from an archive.
func (t *textfile) extract(dir *directories.Dir) error {
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
	if err = archive.Extractor(src, t.Name, s[0], tmp); err != nil {
		return err
	}
	if err = os.Rename(dest, src+txt); err != nil {
		defer os.Remove(dest)
		return fmt.Errorf("extract+move: %w", err)
	}
	return nil
}

// extractedImgs generates PNG and Webp image assets from a textfile extracted from an archive.
func (t *textfile) extractedImgs(dir string) error {
	n := filepath.Join(dir, t.UUID) + txt
	if _, err := os.Stat(n); os.IsNotExist(err) {
		return fmt.Errorf("t.extImgs: %w", os.ErrNotExist)
	} else if err != nil {
		return fmt.Errorf("fix extImg: %s: %w", t.UUID, err)
	}
	amiga := bool(t.Platform == "textamiga")
	if err := generate(n, t.UUID, amiga); err != nil {
		return fmt.Errorf("fix extImg: %w", err)
	}
	return nil
}

// png generates PNG format image assets from a textfile.
func (t *textfile) textPng(c int, dir string) bool {
	logs.Printf("%d. %v", c, t)
	name := filepath.Join(dir, t.UUID)
	if _, err := os.Stat(name); os.IsNotExist(err) {
		logs.Printf("%s\n", str.X())
		return false
	} else if err != nil {
		logs.Log(fmt.Errorf("txtpng stat: %w", err))
		return false
	}
	amiga := bool(t.Platform == "textamiga")
	if err := generate(name, t.UUID, amiga); err != nil {
		logs.Log(fmt.Errorf("fix txtpng: %w", err))
		return false
	}
	logs.Print("\n")
	return true
}

// webP finds and generates missing WebP format images.
func (t *textfile) webP(c int, imgDir string) (int, error) {
	c++
	name := filepath.Join(imgDir, t.UUID+webp)
	if sw, err := os.Stat(name); err == nil && sw.Size() > 0 {
		c--
		return c, nil
	} else if !os.IsNotExist(err) && err != nil {
		logs.Printf("%s\n", str.X())
		return c, fmt.Errorf("webp stat: %w", err)
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
		return c, fmt.Errorf("fix webp: %w", err)
	}
	logs.Printf("%s %s\n", s, str.Y())
	return c, nil
}
