package text

import (
	"database/sql"
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
	fixStmt = "SELECT id, uuid, filename, filesize, retrotxt_no_readme, retrotxt_readme " +
		"FROM files WHERE platform=\"text\" OR platform=\"ansi\" ORDER BY id DESC"
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
		if err := rows.Scan(&t.ID, &t.UUID, &t.Name, &t.Size, &t.NoReadme, &t.Readme); err != nil {
			return fmt.Errorf("fix rows scan: %w", err)
		}
		ok, err := t.exist(&dir)
		if err != nil {
			return fmt.Errorf("fix exist: %w", err)
		}
		if !t.valid() {
			// Extract textfiles from archives.
			t.extract(&dir)
			t.generate(ok, dir.UUID)
			continue
		}
		if !ok {
			// Convert raw textfiles to PNGs.
			c++
			if !t.png(c, simulate, dir.UUID) {
				continue
			}
		}
		if !t.webP(dir.Img000) {
			continue
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

// extract a textfile readme from an archive.
func (t *textfile) extract(dir *directories.Dir) {
	if t.NoReadme.Valid && !t.NoReadme.Bool {
		return
	}
	if !t.Readme.Valid {
		return
	}
	s := strings.Split(t.Readme.String, ",")
	f, err := archive.Read(filepath.Join(dir.UUID, t.UUID), t.Name)
	if err != nil {
		return
	}
	found := false
	for _, key := range f {
		if key == s[0] {
			found = true
			break
		}
	}
	if !found {
		return
	}
	tmp, src := os.TempDir(), filepath.Join(dir.UUID, t.UUID)
	dest := filepath.Join(tmp, s[0])
	if err = archive.Extractor(src, t.Name, s[0], tmp); err != nil {
		fmt.Println(err)
		return
	}
	if err = os.Rename(dest, src+txt); err != nil {
		defer os.Remove(dest)
		logs.Log(fmt.Errorf("extract and move: %w", err))
	}
}

// png generates PNG format image assets from a textfile.
func (t *textfile) png(c int, simulate bool, dir string) bool {
	logs.Printf("%d. %v", c, t)
	name := filepath.Join(dir, t.UUID)
	if _, err := os.Stat(name); os.IsNotExist(err) {
		logs.Printf("%s\n", str.X())
		return false
	}
	if simulate {
		logs.Printf("%s\n", color.Question.Sprint("?"))
		return false
	}
	if err := generate(name, t.UUID); err != nil {
		logs.Log(fmt.Errorf("fix png: %w", err))
	}
	logs.Print("\n")
	return true
}

// generate PNG format image assets from an extracted textfile.
func (t *textfile) generate(ok bool, dir string) {
	if !ok {
		n := filepath.Join(dir, t.UUID) + txt
		if err := generate(n, t.UUID); err != nil {
			logs.Log(fmt.Errorf("fix uuid+txt: %w", err))
		}
	}
}

// webP finds and generates missing WebP format images.
func (t *textfile) webP(imgDir string) bool {
	var err error
	wfp := filepath.Join(imgDir, t.UUID+webp)
	_, err = os.Stat(wfp)
	if os.IsNotExist(err) {
		src := filepath.Join(imgDir, t.UUID+png)
		logs.Printf("%s", t.UUID+png)
		var s string
		s, err = images.ToWebp(src, wfp, true)
		if err != nil {
			logs.Log(fmt.Errorf("fix webp: %w", err))
			return false
		}
		logs.Printf(" %s\n", s)
		return false
	} else if err != nil {
		logs.Log(fmt.Errorf("webp stat: %w", err))
		return false
	}
	return true
}

// check that [UUID].png exists in all three image subdirectories.
func (t *textfile) exist(dir *directories.Dir) (bool, error) {
	dirs := [3]string{dir.Img000, dir.Img150, dir.Img400}
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

// valid confirms that the named file is not a known archive.
func (t *textfile) valid() bool {
	switch filepath.Ext(strings.ToLower(t.Name)) {
	case z7, arc, arj, bz2, cab, gz, lha, lzh, rar, tar, tgz, zip:
		return false
	}
	return true
}
