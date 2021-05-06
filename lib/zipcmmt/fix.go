package zipcmmt

import (
	"archive/zip"
	"fmt"
	"log"
	"path/filepath"

	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/directories"
)

const (
	fixStmt  = `SELECT id, uuid, filename, filesize FROM files WHERE filename LIKE "%.zip" OR file_magic_type LIKE "Zip archive data%"`
	filename = `_zipcomment.txt`
)

// WHERE filename LIKE "%.zip";

// zipfile is a file file object.
type zipfile struct {
	ID   uint   // database id
	UUID string // database unique id
	Name string // file name
	Ext  string // file extension
	Size int    // file size in bytes
	// Magic string // file magic type
}

func Fix(simulate bool) error {
	fmt.Println("zip comments")
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
		var z zipfile
		i++
		if err := rows.Scan(&z.ID, &z.UUID, &z.Name, &z.Size); err != nil {
			return fmt.Errorf("fix rows scan: %w", err)
		}
		if ok := z.checkDownload(dir.UUID); !ok {
			continue
		}
		c++
		go z.open(dir.UUID)
	}
	fmt.Printf("Records and uuid files: %d, %d\n", i, c)
	fmt.Println(dir.UUID)

	return nil
}

func (z zipfile) open(path string) {
	if ok := z.checkCmmt(path); ok {
		return
	}

	file := filepath.Join(fmt.Sprint(path), z.UUID)

	// Open a zip archive for reading.
	r, err := zip.OpenReader(file)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()
	if cmmt := r.Comment; cmmt != "" {
		fmt.Println("Comment:", cmmt)
	}

	// // Iterate through the files in the archive,
	// // printing some of their contents.
	// for _, f := range r.File {
	// 	fmt.Printf("Contents of %s:\n", f.Name)
	// 	rc, err := f.Open()
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	// _, err = io.CopyN(os.Stdout, rc, 68)
	// 	// if err != nil {
	// 	// 	log.Fatal(err)
	// 	// }
	// 	rc.Close()
	// 	fmt.Println()
	// }
}
