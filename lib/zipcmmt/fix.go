package zipcmmt

import (
	"archive/zip"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/directories"
	"github.com/Defacto2/df2/lib/logs"
	rt "github.com/bengarrett/retrotxtgo/lib/convert"
)

const (
	fixStmt  = `SELECT id, uuid, filename, filesize, file_magic_type FROM files WHERE filename LIKE "%.zip"`
	filename = `_zipcomment.txt`
	resetCmd = "\033[0m"
	sceneOrg = `scene.org`
)

// zipfile is a file file object.
type zipfile struct {
	ID        uint           // database id
	UUID      string         // database unique id
	Name      string         // file name
	Ext       string         // file extension
	Size      int            // file size in bytes
	Magic     sql.NullString // file magic type
	ascii     bool
	unicode   bool
	overwrite bool
}

func Fix(ascii, unicode, overwrite bool) error {
	start := time.Now()
	dir, db := directories.Init(false), database.Connect()
	defer db.Close()
	rows, err := db.Query(fixStmt)
	if err != nil {
		return fmt.Errorf("fix db query: %w", err)
	} else if rows.Err() != nil {
		return fmt.Errorf("fix rows: %w", rows.Err())
	}
	defer rows.Close()
	i := 0
	for rows.Next() {
		z := zipfile{
			ascii:     ascii,
			unicode:   unicode,
			overwrite: overwrite,
		}
		if err := rows.Scan(&z.ID, &z.UUID, &z.Name, &z.Size, &z.Magic); err != nil {
			return fmt.Errorf("fix rows scan: %w", err)
		}
		i++
		if ok := z.checkDownload(dir.UUID); !ok {
			continue
		}
		if ok := z.checkCmmtFile(dir.UUID); !ok {
			continue
		}
		if ascii || unicode {
			z.save(dir.UUID)
			continue
		}
		z.save(dir.UUID)
	}
	elapsed := time.Since(start).Seconds()
	if ascii || unicode {
		logs.Println()
	}
	logs.Print(fmt.Sprintf("%d zip archives scanned for comments", i))
	logs.Print(fmt.Sprintf(", time taken %.3f seconds\n", elapsed))
	return nil
}

func (z zipfile) save(path string) bool {
	file := filepath.Join(fmt.Sprint(path), z.UUID)
	name := file + filename
	// Open a zip archive for reading.
	r, err := zip.OpenReader(file)
	if err != nil {
		log.Println(err)
		return false
	}
	defer r.Close()
	// Parse and save zip comment
	if cmmt := r.Comment; cmmt != "" {
		if strings.TrimSpace(cmmt) == "" {
			return false
		}
		if strings.Contains(cmmt, sceneOrg) {
			return false
		}
		z.print(&cmmt)
		f, err := os.Create(name)
		if err != nil {
			log.Panicln(err)
			return false
		}
		defer f.Close()
		if i, err := f.Write([]byte(cmmt)); err != nil {
			log.Panicln(err)
			return false
		} else if i == 0 {
			os.Remove(name)
			return false
		}
	}
	return true
}

func (z zipfile) print(cmmt *string) {
	fmt.Printf("\n%v. - %s", z.ID, z.Name)
	if z.Magic.Valid {
		fmt.Printf(" [%s]", z.Magic.String)
	}
	fmt.Println("")
	if z.ascii {
		z.printAscii(cmmt)
	}
	if z.unicode {
		z.printUnicode(cmmt)
	}
}

func (z zipfile) printAscii(cmmt *string) {
	fmt.Printf("%s%s\n", *cmmt, resetCmd)
}

func (z zipfile) printUnicode(cmmt *string) {
	b, err := rt.D437(*cmmt)
	if err != nil {
		fmt.Printf("Could not convert to Unicode:\n%s%s\n", *cmmt, resetCmd)
	}
	fmt.Printf("%s%s\n", b, resetCmd)
}
