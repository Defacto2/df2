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

var (
	dir      directories.Dir
	simulate = false
)

// Txt is an image object.
type Txt struct {
	ID       uint
	UUID     string
	Filename string
	FileExt  string
	Filesize int
}

func (t Txt) String() string {
	return fmt.Sprintf("(%v) %v %v ", color.Primary.Sprint(t.ID), t.Filename, color.Info.Sprint(humanize.Bytes(uint64(t.Filesize))))
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
		var txt Txt
		err = rows.Scan(&txt.ID, &txt.UUID, &txt.Filename, &txt.Filesize)
		if err != nil {
			logs.Check(err)
		}
		// TODO replace with function that handles archives
		if !txt.ext() {
			continue
		}
		if !txt.check() {
			c++
			logs.Printf("%d. %v", c, txt)
			input := filepath.Join(dir.UUID, txt.UUID)
			if _, err := os.Stat(input); os.IsNotExist(err) {
				logs.Printf("%s\n", logs.X())
				continue
			}
			if simulate {
				logs.Printf("%s\n", color.Question.Sprint("?"))
				continue
			}
			Generate(input, txt.UUID)
			//images.Generate(input, txt.UUID)
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

func (t Txt) ext() bool {
	switch filepath.Ext(t.Filename) {
	case ".txt", ".diz", ".doc", ".nfo":
		return true
	}
	return false
}

// check that [UUID].png exists in all three image subdirectories.
func (t Txt) check() bool {
	dirs := [3]string{dir.Img000, dir.Img150, dir.Img400}
	for _, path := range dirs {
		if _, err := os.Stat(filepath.Join(path, t.UUID+".png")); os.IsNotExist(err) {
			return false
		}
	}
	return true
}
