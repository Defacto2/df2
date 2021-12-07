package cmmt

import (
	"archive/zip"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/bengarrett/retrotxtgo/lib/convert"
)

const (
	Filename = `_zipcomment.txt`
	SceneOrg = `scene.org`
	resetCmd = "\033[0m"
)

// Zipfile is a file file object.
type Zipfile struct {
	ID        uint           // database id
	UUID      string         // database unique id
	Name      string         // file name
	Ext       string         // file extension
	Size      int            // file size in bytes
	Magic     sql.NullString // file magic type
	ASCII     bool
	Unicode   bool
	Overwrite bool
}

func (z *Zipfile) CheckDownload(path string) (ok bool) {
	file := filepath.Join(fmt.Sprint(path), z.UUID)
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return false
	}
	return true
}

func (z *Zipfile) CheckCmmtFile(path string) (ok bool) {
	if z.Overwrite {
		return true
	}
	file := filepath.Join(fmt.Sprint(path), z.UUID+Filename)
	if _, err := os.Stat(file); err == nil {
		return false
	}
	return true
}

func (z *Zipfile) Save(path string) {
	file := filepath.Join(fmt.Sprint(path), z.UUID)
	name := file + Filename
	// Open a zip archive for reading.
	r, err := zip.OpenReader(file)
	if err != nil {
		log.Println(err)
		return
	}
	defer r.Close()
	// Parse and save zip comment
	cmmt := r.Comment
	if cmmt == "" {
		return
	}
	if strings.TrimSpace(cmmt) == "" {
		return
	}
	if strings.Contains(cmmt, SceneOrg) {
		return
	}
	z.print(&cmmt)
	f, err := os.Create(name)
	if err != nil {
		log.Panicln(err)
		return
	}
	defer f.Close()
	if i, err := f.Write([]byte(cmmt)); err != nil {
		log.Panicln(err)
		return
	} else if i == 0 {
		os.Remove(name)
		return
	}
}

func (z *Zipfile) print(cmmt *string) {
	fmt.Printf("\n%v. - %s", z.ID, z.Name)
	if z.Magic.Valid {
		fmt.Printf(" [%s]", z.Magic.String)
	}
	fmt.Println("")
	if z.ASCII {
		z.printASCII(cmmt)
	}
	if z.Unicode {
		z.printUnicode(cmmt)
	}
}

func (z *Zipfile) printASCII(cmmt *string) {
	fmt.Printf("%s%s\n", *cmmt, resetCmd)
}

func (z *Zipfile) printUnicode(cmmt *string) {
	b, err := convert.D437(*cmmt)
	if err != nil {
		fmt.Printf("Could not convert to Unicode:\n%s%s\n", *cmmt, resetCmd)
	}
	fmt.Printf("%s%s\n", b, resetCmd)
}
