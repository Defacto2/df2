package images

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/directories"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/dustin/go-humanize"
	"github.com/gookit/color"
)

var (
	dir      directories.Dir
	simulate = false
)

// Img is an image object.
type Img struct {
	ID       uint
	UUID     string
	Filename string
	FileExt  string
	Filesize int
}

func (i Img) String() string {
	return fmt.Sprintf("(%v) %v %v ", color.Primary.Sprint(i.ID), i.Filename, color.Info.Sprint(humanize.Bytes(uint64(i.Filesize))))
}

// Fix generates any missing assets from downloads that images.
func Fix(sim bool) error {
	simulate = sim
	dir = directories.Init(false)
	db := database.Connect()
	defer db.Close()
	rows, err := db.Query(`SELECT id, uuid, filename, filesize FROM files WHERE platform="image"`)
	if err != nil {
		return err
	}
	c := 0
	for rows.Next() {
		var img Img
		err = rows.Scan(&img.ID, &img.UUID, &img.Filename, &img.Filesize)
		if err != nil {
			logs.Check(err)
		}
		if !img.ext() {
			continue
		}
		if !img.check() {
			c++
			logs.Printf("%d. %v", c, img)
			if _, err := os.Stat(filepath.Join(dir.UUID, img.UUID)); os.IsNotExist(err) {
				logs.Printf("%s\n", logs.X())
				continue
			}
			if simulate {
				logs.Printf("%s\n", color.Question.Sprint("?"))
				continue
			}
			Generate(filepath.Join(dir.UUID, img.UUID), img.UUID)
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

func (i Img) ext() bool {
	switch filepath.Ext(i.Filename) {
	case ".gif", ".jpg", ".jpeg", ".png":
		return true
	}
	return false
}

func (i Img) check() bool {
	dirs := [3]string{dir.Img000, dir.Img150, dir.Img400}
	for _, path := range dirs {
		if _, err := os.Stat(filepath.Join(path, i.UUID+".png")); os.IsNotExist(err) {
			return false
		}
	}
	return true
}
