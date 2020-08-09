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
	diz  = ".diz"
	doc  = ".doc"
	nfo  = ".nfo"
	png  = ".png"
	txt  = ".txt"
	webp = ".webp"
)

var xdir directories.Dir

// image preview and thumbnail of a text object.
type image struct {
	ID       uint
	UUID     string
	Filename string
	FileExt  string
	Filesize int
}

func (i image) String() string {
	return fmt.Sprintf("(%v) %v %v ", color.Primary.Sprint(i.ID), i.Filename,
		color.Info.Sprint(humanize.Bytes(uint64(i.Filesize))))
}

// Fix generates any missing assets from downloads that are text based.
func Fix(simulate bool) error {
	dir, db := directories.Init(false), database.Connect()
	defer db.Close()
	rows, err := db.Query(`SELECT id, uuid, filename, filesize FROM files WHERE platform="text"`)
	if err != nil {
		return fmt.Errorf("fix db query: %w", err)
	} else if rows.Err() != nil {
		return fmt.Errorf("fix rows: %w", rows.Err())
	}
	defer rows.Close()
	c := 0
	for rows.Next() {
		var img image
		if err := rows.Scan(&img.ID, &img.UUID, &img.Filename, &img.Filesize); err != nil {
			return fmt.Errorf("fix rows scan: %w", err)
		}
		// TODO replace with function that handles archives
		if !img.valid() {
			continue
		}
		if ok, err := img.exist(&dir); err != nil {
			return fmt.Errorf("fix exist: %w", err)
		} else if !ok {
			c++
			logs.Printf("%d. %v", c, img)
			input := filepath.Join(dir.UUID, img.UUID)
			if _, err := os.Stat(input); os.IsNotExist(err) {
				logs.Printf("%s\n", str.X())
				continue
			}
			if simulate {
				logs.Printf("%s\n", color.Question.Sprint("?"))
				continue
			}
			if err := generate(input, img.UUID); err != nil {
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
func (i image) exist(dir *directories.Dir) (bool, error) {
	dirs := [3]string{dir.Img000, dir.Img150, dir.Img400}
	for _, path := range dirs {
		if path == "" {
			return false, nil
		}
		if _, err := os.Stat(filepath.Join(path, i.UUID+png)); os.IsNotExist(err) {
			return false, nil
		} else if err != nil {
			return false, fmt.Errorf("image exist: %w", err)
		}
	}
	return true, nil
}

func (i image) valid() bool {
	switch filepath.Ext(i.Filename) {
	case diz, doc, nfo, txt:
		return true
	}
	return false
}
