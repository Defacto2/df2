package cmmt

import (
	"archive/zip"
	"database/sql"
	"fmt"
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

func (z *Zipfile) Save(path string) error {
	file := filepath.Join(fmt.Sprint(path), z.UUID)
	name := file + Filename
	// Open a zip archive for reading.
	r, err := zip.OpenReader(file)
	if err != nil {
		return err
	}
	defer r.Close()
	// Parse and save zip comment
	cmmt := r.Comment
	if cmmt == "" {
		return nil
	}
	if strings.TrimSpace(cmmt) == "" {
		return nil
	}
	if strings.Contains(cmmt, SceneOrg) {
		return nil
	}
	fmt.Print(z.Print(&cmmt))
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	defer f.Close()
	if i, err := f.Write([]byte(cmmt)); err != nil {
		return err
	} else if i == 0 {
		return os.Remove(name)
	}
	return nil
}

func (z *Zipfile) Print(cmmt *string) string {
	if z.ID == 0 {
		return ""
	}
	var b strings.Builder
	fmt.Fprintf(&b, "\n%v. - %s", z.ID, z.Name)
	if z.Magic.Valid {
		fmt.Fprintf(&b, " [%s]", z.Magic.String)
	}
	fmt.Fprintln(&b)
	if z.ASCII {
		fmt.Fprint(&b, z.printASCII(cmmt))
	}
	if z.Unicode {
		fmt.Fprint(&b, z.printUnicode(cmmt))
	}
	return b.String()
}

func (z *Zipfile) printASCII(cmmt *string) string {
	return fmt.Sprintf("%s%s\n", *cmmt, resetCmd)
}

func (z *Zipfile) printUnicode(cmmt *string) string {
	b, err := convert.D437(*cmmt)
	if err != nil {
		return fmt.Sprintf("Could not convert to Unicode:\n%s%s\n", *cmmt, resetCmd)
	}
	return fmt.Sprintf("%s%s\n", b, resetCmd)
}
