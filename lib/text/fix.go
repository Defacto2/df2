package text

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dustin/go-humanize"
	"github.com/gookit/color"

	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/directories"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/Defacto2/df2/lib/str"
)

const (
	fixStmt = `SELECT id, uuid, filename, filesize FROM files ORDER BY id DESC WHERE platform = 'text'`

	diz  = ".diz"
	doc  = ".doc"
	nfo  = ".nfo"
	png  = ".png"
	txt  = ".txt"
	webp = ".webp"
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
	c := 0
	for rows.Next() {
		var t textfile
		if err := rows.Scan(&t.ID, &t.UUID, &t.Name, &t.Size); err != nil {
			return fmt.Errorf("fix rows scan: %w", err)
		}
		// TODO replace with function that handles archives
		if !t.valid() {
			continue
		}
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
	}
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
		if _, err := os.Stat(filepath.Join(path, t.UUID+png)); os.IsNotExist(err) {
			return false, nil
		} else if err != nil {
			return false, fmt.Errorf("image exist: %w", err)
		}
	}
	return true, nil
}

func (t textfile) valid() bool {
	switch filepath.Ext(t.Name) {
	case diz, doc, nfo, txt:
		return true
	}
	return false
}
