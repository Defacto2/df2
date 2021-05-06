package text

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/directories"
	"github.com/Defacto2/df2/lib/images"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/Defacto2/df2/lib/str"
	"github.com/dustin/go-humanize" //nolint:misspell
	"github.com/gookit/color"       //nolint:misspell
)

const (
	fixStmt = `SELECT id, uuid, filename, filesize FROM files WHERE platform="text" OR platform="ansi" ORDER BY id DESC`

	ans  = ".ans"
	asc  = ".asc"
	diz  = ".diz"
	doc  = ".doc"
	nfo  = ".nfo"
	png  = ".png"
	txt  = ".txt"
	webp = ".webp"

	// 7z,arc,ark,arj,cab,gz,lha,lzh,rar,tar,tar.gz,zip
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
	ID   uint   // database id
	UUID string // database unique id
	Name string // file name
	Ext  string // file extension
	Size int    // file size in bytes
}

func (t textfile) String() string {
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
		if err := rows.Scan(&t.ID, &t.UUID, &t.Name, &t.Size); err != nil {
			return fmt.Errorf("fix rows scan: %w", err)
		}
		if !t.valid() {
			// TODO: extract textfiles from archives.
			continue
		}
		// Missing PNG images.
		ok, err := t.exist(&dir)
		if err != nil {
			return fmt.Errorf("fix exist: %w", err)
		}
		if !ok {
			c++
			logs.Printf("%d. %v", c, t)
			name := filepath.Join(dir.UUID, t.UUID)
			if _, err := os.Stat(name); os.IsNotExist(err) {
				logs.Printf("%s\n", str.X())
				continue
			}
			if simulate {
				logs.Printf("%s\n", color.Question.Sprint("?"))
				continue
			}
			if err := generate(name, t.UUID); err != nil {
				logs.Log(fmt.Errorf("fix generate: %w", err))
			}
			logs.Print("\n")
		}
		// Missing WebP images.
		wfp := filepath.Join(dir.Img000, t.UUID+webp)
		_, err = os.Stat(wfp)
		if os.IsNotExist(err) {
			src := filepath.Join(dir.Img000, t.UUID+png)
			logs.Printf("%s", t.UUID+png)
			s, err := images.ToWebp(src, wfp, true)
			if err != nil {
				logs.Log(fmt.Errorf("fix generate: %w", err))
				continue
			}
			logs.Printf(" %s\n", s)
			continue
		} else if err != nil {
			logs.Log(fmt.Errorf("webp stat: %w", err))
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

// check that [UUID].png exists in all three image subdirectories.
func (t textfile) exist(dir *directories.Dir) (bool, error) {
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

func (t textfile) valid() bool {
	switch filepath.Ext(strings.ToLower(t.Name)) {
	case z7, arc, arj, bz2, cab, gz, lha, lzh, rar, tar, tgz, zip:
		return false
	}
	return true
}
