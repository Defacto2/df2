package record

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Defacto2/df2/lib/archive"
	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/Defacto2/df2/lib/str"
	"github.com/Defacto2/df2/lib/zipcontent/internal/scan"
	"github.com/gookit/color"
)

var (
	ErrRawBytes = errors.New("sql rawbytes is missing expected table columns")
	ErrStatNil  = errors.New("stats pointer is nil")
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

func New(values []sql.RawBytes, path string) (Record, error) {
	const id, uuid, filename, readme = 0, 1, 4, 6
	if len(values) < readme+1 {
		return Record{}, ErrRawBytes
	}
	return Record{
		ID:   string(values[id]),
		UUID: string(values[uuid]),
		Name: string(values[filename]),
		NFO:  string(values[readme]),
		File: filepath.Join(path, string(values[uuid])),
	}, nil
}

// Iterate through each value.
func (r *Record) Iterate(s *scan.Stats) error {
	if s == nil {
		return ErrStatNil
	}
	var value string
	for i, raw := range *s.Values {
		value = database.Val(raw)
		switch s.Columns[i] {
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
func (r *Record) files(s *scan.Stats) error {
	if s == nil {
		return ErrStatNil
	}
	var err error
	logs.Print(" • ")
	r.Files, err = archive.Read(r.File, r.Name)
	if err != nil {
		s.Missing++
		return fmt.Errorf("file zip content archive read: %w", err)
	}
	logs.Printf("%d items", len(r.Files))
	if err := r.nfo(s); err != nil {
		return err
	}
	if err := r.save(); err != nil {
		logs.Printf(" %s", str.X())
		return fmt.Errorf("file zip content update: %w", err)
	}
	return nil
}

func (r *Record) nfo(s *scan.Stats) error {
	if s == nil {
		return ErrStatNil
	}
	const txt = ".txt"
	r.NFO = archive.FindNFO(r.Name, r.Files...)
	if r.NFO == "" {
		return nil
	}
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
		if _, err3 := archive.FileMove(src, filepath.Join(s.BasePath, r.UUID+txt)); err3 != nil {
			return err3
		}
		logs.Print(", extracted")
	}
	return nil
}

func (r *Record) id(s *scan.Stats) error {
	if s == nil {
		return ErrStatNil
	}
	logs.Printcrf("%s %0*d. %v ",
		color.Question.Sprint("→"),
		len(strconv.Itoa(s.Total)),
		s.Count,
		color.Primary.Sprint(r.ID))
	return nil
}

func (r *Record) save() error {
	const (
		files = "UPDATE files SET file_zip_content=?,updatedat=NOW(),updatedby=? WHERE id=?"
		nfo   = "UPDATE files SET file_zip_content=?,updatedat=NOW(),updatedby=?," +
			"retrotxt_readme=?,retrotxt_no_readme=? WHERE id=?"
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
