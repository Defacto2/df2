// Package zipcontent scans archives for file and directory content.
package zipcontent

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/df2/lib/archive"
	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/directories"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/Defacto2/df2/lib/str"
	"github.com/gookit/color" // nolint: misspell
)

// Record of a file item.
type Record struct {
	ID    string   // mysql auto increment id
	UUID  string   // record unique id
	File  string   // absolute path to file
	Name  string   // filename
	Files []string // list of files contained in the archive
	NFO   string   // NFO or textfile to display
}

type stat struct {
	start    time.Time       // processing time
	basePath string          // path to file downloads with UUID as filenames
	count    int             // row index
	missing  int             // missing UUID files count
	total    int             // total rows
	columns  []string        // column names
	values   *[]sql.RawBytes // row values
}

func statInit() stat {
	dir := directories.Init(false)
	return stat{basePath: dir.UUID, start: time.Now()}
}

func Fix() error {
	s := statInit()
	db := database.Connect()
	defer db.Close()
	rows, err := db.Query(where())
	if err != nil {
		return err
	}
	if rows.Err() != nil {
		return rows.Err()
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	for rows.Next() {
		s.total++
	}
	rows, err = db.Query(where())
	if err != nil {
		return err
	}
	if rows.Err() != nil {
		return rows.Err()
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(scanArgs...); err != nil {
			return err
		}
		s.count++
		r := record(values, s.basePath)
		s.columns = columns
		s.values = &values
		if err := r.iterate(&s); err != nil {
			fmt.Println()
			log.Printf("%s\n", err)
			continue
		}
		logs.Println()
	}
	s.summary()
	return nil
}

func record(values []sql.RawBytes, path string) Record {
	const id, uuid, filename, readme = 0, 1, 4, 6
	return Record{
		ID:   string(values[id]),
		UUID: string(values[uuid]),
		Name: string(values[filename]),
		NFO:  string(values[readme]),
		File: filepath.Join(path, string(values[uuid])),
	}
}

func where() string {
	const s = "SELECT `id`,`uuid`,`deletedat`,`createdat`,`filename`,`updatedat`,`retrotxt_readme`"
	const w = " WHERE file_zip_content IS NULL AND (`filename` LIKE '%.zip' OR `filename` LIKE '%.rar' OR `filename` LIKE '%.7z')"
	return fmt.Sprintf("%s FROM `files` %s", s, w)
}

// iterate through each value.
func (r *Record) iterate(s *stat) error {
	var value string
	for i, raw := range *s.values {
		value = database.Val(raw)
		switch s.columns[i] {
		case "id":
			r.id(s)
		case "createdat":
			database.DateTime(raw)
		case "filename":
			logs.Printf("%v", value)
			if err := r.files(s); err != nil {
				return err
			}
		default:
		}
	}
	return nil
}

// files reads an archive and saves its content to the database.
func (r *Record) files(s *stat) (err error) {
	const txt = ".txt"
	logs.Print(" • ")
	r.Files, err = archive.Read(r.File, r.Name)
	if err != nil {
		s.missing++
		return fmt.Errorf("file zip content archive read: %w", err)
	}
	logs.Printf("%d items", len(r.Files))
	if r.NFO == "" {
		if r.NFO = archive.FindNFO(r.Name, r.Files...); r.NFO != "" {
			logs.Printf(", text file: %s", r.NFO)
			if _, err := os.Stat(r.File + txt); os.IsNotExist(err) {
				tmp, err1 := os.MkdirTemp(os.TempDir(), "zipcontent")
				if err1 != nil {
					return err1
				}
				defer os.RemoveAll(tmp)
				if err2 := archive.Extractor(r.File, r.Name, r.NFO, tmp); err2 != nil {
					return err2
				}
				src := filepath.Join(tmp, r.NFO)
				if _, err3 := archive.FileMove(src, filepath.Join(s.basePath, r.UUID+txt)); err3 != nil {
					return err3
				}
				logs.Print(", extracted")
			}
		}
	}
	if err := r.save(); err != nil {
		logs.Printf(" %s", str.X())
		return fmt.Errorf("file zip content update: %w", err)
	}
	return nil
}

func (r *Record) id(s *stat) {
	logs.Printcrf("%s %0*d. %v ",
		color.Question.Sprint("→"),
		len(strconv.Itoa(s.total)),
		s.count,
		color.Primary.Sprint(r.ID))
}

func (r *Record) save() error {
	const (
		files = "UPDATE files SET file_zip_content=?,updatedat=NOW(),updatedby=? WHERE id=?"
		nfo   = "UPDATE files SET file_zip_content=?,updatedat=NOW(),updatedby=?,retrotxt_readme=?,retrotxt_no_readme=? WHERE id=?"
	)
	var err error
	var update *sql.Stmt
	db := database.Connect()
	defer db.Close()
	if r.NFO == "" {
		update, err = db.Prepare(files)
	} else {
		update, err = db.Prepare(nfo)
	}
	if err != nil {
		return fmt.Errorf("update zip content db prepare: %w", err)
	}
	defer update.Close()
	content := strings.Join(r.Files, "\n")
	if r.NFO == "" {
		if _, err := update.Exec(content, database.UpdateID, r.ID); err != nil {
			return fmt.Errorf("update zip content update exec: %w", err)
		}
	} else if _, err := update.Exec(content, database.UpdateID, r.NFO, 0, r.ID); err != nil {
		return fmt.Errorf("update zip content update exec: %w", err)
	}
	logs.Printf(" %s", str.Y())
	return db.Close()
}

func (s *stat) summary() {
	total := s.count - s.missing
	if total == 0 {
		fmt.Print("nothing to do")
	}
	elapsed := time.Since(s.start).Seconds()
	t := fmt.Sprintf("Total archives scanned: %v, time elapsed %.1f seconds", total, elapsed)
	logs.Printf("\n%s\n%s\n", strings.Repeat("─", len(t)), t)
}
