package text

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

var dir directories.Dir

// Image preview and thumbnail of a text object.
type Image struct {
	ID       uint
	UUID     string
	Filename string
	FileExt  string
	Filesize int
}

func (i Image) String() string {
	return fmt.Sprintf("(%v) %v %v ", color.Primary.Sprint(i.ID), i.Filename,
		color.Info.Sprint(humanize.Bytes(uint64(i.Filesize))))
}

// Fix generates any missing assets from downloads that are text based.
func Fix(simulate bool) error {
	dir = directories.Init(false)
	db := database.Connect()
	defer db.Close()
	rows, err := db.Query(`SELECT id, uuid, filename, filesize FROM files WHERE platform="text"`)
	if err != nil {
		return err
	}
	c := 0
	for rows.Next() {
		var img Image
		err = rows.Scan(&img.ID, &img.UUID, &img.Filename, &img.Filesize)
		if err != nil {
			logs.Check(err)
		}
		// TODO replace with function that handles archives
		if !img.valid() {
			continue
		}
		if !img.exist() {
			c++
			logs.Printf("%d. %v", c, img)
			input := filepath.Join(dir.UUID, img.UUID)
			if _, err := os.Stat(input); os.IsNotExist(err) {
				logs.Printf("%s\n", logs.X())
				continue
			}
			if simulate {
				logs.Printf("%s\n", color.Question.Sprint("?"))
				continue
			}
			Generate(input, img.UUID)
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
func (i Image) exist() bool {
	const ext = ".png"
	dirs := [3]string{dir.Img000, dir.Img150, dir.Img400}
	for _, path := range dirs {
		if path == "" {
			return false
		}
		if s, err := os.Stat(filepath.Join(path, i.UUID+ext)); os.IsNotExist(err) {
			fmt.Println(s)
			return false
		}
	}
	return true
}

func (i Image) valid() bool {
	switch filepath.Ext(i.Filename) {
	case ".txt", ".diz", ".doc", ".nfo":
		return true
	}
	return false
}
