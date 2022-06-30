package record

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Defacto2/df2/pkg/archive"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/logs"
	"github.com/Defacto2/df2/pkg/str"
	"github.com/Defacto2/df2/pkg/zipcontent/internal/scan"
	"github.com/gookit/color"
)

const (
	errPrefix = "zipcontent record error,"
)

var (
	ErrCols     = errors.New("the number of values is not the same as the number of columns")
	ErrID       = errors.New("record does not contain a valid value for the id column")
	ErrUUID     = errors.New("record does not contain a valid value for the uuid column")
	ErrRawBytes = errors.New("sql rawbytes is missing expected table columns")
	ErrStatNil  = errors.New("stats pointer is nil or the stats.values field is missing")
)

// Record of a file item.
type Record struct {
	ID    string   // MySQL auto increment Id.
	UUID  string   // Unique Id.
	File  string   // Absolute path to file.
	Name  string   // Filename.
	Files []string // A list of files contained in the archive.
	NFO   string   // NFO or textfile to display on the site.
}

// New returns a Record generated from the sql rawbyte values.
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

// Iterate through each sql rawbyte value.
func (r *Record) Iterate(s *scan.Stats) error {
	if s == nil || s.Values == nil {
		return ErrStatNil
	}
	if len(s.Columns) != len(*s.Values) {
		return fmt.Errorf("%w, columns: %d, values: %d",
			ErrCols, len(s.Columns), len(*s.Values))
	}
	var value string
	for i, raw := range *s.Values {
		value = database.Val(raw)
		switch s.Columns[i] {
		case "id":
			if err := r.id(s); err != nil {
				return err
			}
		case "createdat":
			database.DateTime(raw)
		case "filename":
			logs.Printf("%v", value)
			if err := r.Read(s); err != nil {
				return err
			}
		default:
		}
	}
	return nil
}

// Read and save the archive content to the database.
func (r *Record) Read(s *scan.Stats) error {
	if s == nil {
		return ErrStatNil
	}
	if r.UUID == "" {
		return fmt.Errorf("%w, quoted uuid: %q", ErrUUID, r.NFO)
	}
	var err error
	logs.Print(" • ")
	r.Files, r.Name, err = archive.Read(r.File, r.Name)
	if err != nil {
		s.Missing++
		return fmt.Errorf("%s archive read: %w", errPrefix, err)
	}
	logs.Printf("%d items", len(r.Files))
	if err := r.Nfo(s); err != nil {
		// instead of returning the error, print it.
		// otherwise the results of archive.Read will never be saved
		fmt.Printf(" %s", err)
	}
	updates, err := r.Save()
	if err != nil {
		logs.Printf(" %s", str.X())
		return err
	}
	if updates == 0 {
		logs.Printf(" %s", str.X())
		return nil
	}
	logs.Printf(" %s", str.Y())
	return nil
}

// Nfo finds an appropriate textfile and saves it to the database.
func (r *Record) Nfo(s *scan.Stats) error {
	if s == nil {
		return ErrStatNil
	}
	const txt = ".txt"
	r.NFO = archive.NFO(r.Name, r.Files...)
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
		if _, err3 := archive.Move(src, filepath.Join(s.BasePath, r.UUID+txt)); err3 != nil {
			return err3
		}
		logs.Print(", extracted")
	}
	return nil
}

// Id prints the record ID to stdout.
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

// Save updates the record in the database.
func (r *Record) Save() (int64, error) {
	if r.ID == "" {
		return 0, ErrID
	}
	const (
		files = "UPDATE files SET filename=?,file_zip_content=?,updatedat=NOW(),updatedby=? WHERE id=?"
		nfo   = "UPDATE files SET filename=?,file_zip_content=?,updatedat=NOW(),updatedby=?," +
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
		return 0, fmt.Errorf("%s db prepare: %w", errPrefix, err)
	}
	defer update.Close()
	content := strings.Join(r.Files, "\n")
	if r.NFO == "" {
		a, err := update.Exec(r.Name, content, database.UpdateID, r.ID)
		if err != nil {
			return 0, fmt.Errorf("%s db exec: %w", errPrefix, err)
		}
		return a.RowsAffected()
	}
	a, err := update.Exec(r.Name, content, database.UpdateID, r.NFO, 0, r.ID)
	if err != nil {
		return 0, fmt.Errorf("%s db exec: %w", errPrefix, err)
	}
	return a.RowsAffected()
}
