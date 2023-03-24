package record

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Defacto2/df2/pkg/archive"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/logger"
	"github.com/Defacto2/df2/pkg/str"
	"github.com/Defacto2/df2/pkg/zipcontent/internal/scan"
	"github.com/gookit/color"
)

const (
	errPrefix = "zipcontent record error,"
)

var (
	ErrID       = errors.New("record does not contain a valid value for the id column")
	ErrUUID     = errors.New("record does not contain a valid value for the uuid column")
	ErrRawBytes = errors.New("sql rawbytes is missing an expected table columns")
	ErrStatNil  = errors.New("scan stats pointer is nil or the stats.values field is missing")
	ErrValues   = errors.New("the number of values is not the same as the number of columns")
)

// Record object.
type Record struct {
	ID    string   // ID is the database auto increment ID.
	UUID  string   // Universal unique Id.
	File  string   // File is the absolute path to file archive.
	Name  string   // Name of the file archive.
	Files []string // Files contained in the archive.
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
func (r *Record) Iterate(db *sql.DB, w io.Writer, s *scan.Stats) error {
	if db == nil {
		return database.ErrDB
	}
	if s == nil || s.Values == nil {
		return ErrStatNil
	}
	if w == nil {
		w = io.Discard
	}
	if len(s.Columns) != len(*s.Values) {
		return fmt.Errorf("%w, columns: %d, values: %d",
			ErrValues, len(s.Columns), len(*s.Values))
	}
	value := ""
	for i, raw := range *s.Values {
		value = database.Val(raw)
		switch s.Columns[i] {
		case "id":
			if err := r.id(w, s); err != nil {
				return err
			}
		case "createdat":
			s, err := database.DateTime(raw)
			if err == nil {
				fmt.Fprintf(w, "%v", s)
			}
		case "filename":
			fmt.Fprintf(w, "%v", value)
			if err := r.Archive(db, w, s); err != nil {
				return err
			}
		default:
		}
	}
	return nil
}

// Archive reads and saves the archive content to the database.
func (r *Record) Archive(db *sql.DB, w io.Writer, s *scan.Stats) error {
	if db == nil {
		return database.ErrDB
	}
	if s == nil {
		return ErrStatNil
	}
	if r.UUID == "" {
		return fmt.Errorf("%w, quoted uuid: %q", ErrUUID, r.NFO)
	}
	if w == nil {
		w = io.Discard
	}
	fmt.Fprint(w, " • ")
	var err error
	r.Files, r.Name, err = archive.Read(w, r.File, r.Name)
	if err != nil {
		s.Missing++
		return fmt.Errorf("%s archive read: %w", errPrefix, err)
	}
	fmt.Fprintf(w, "%d items", len(r.Files))
	if err := r.Textfile(w, s); err != nil {
		// instead of returning the error, print it.
		// otherwise the results of archive.Read will never be saved
		fmt.Fprintf(w, " %s", err)
	}
	updates, err := r.Save(db)
	if err != nil {
		fmt.Fprintf(w, " %s", str.X())
		return err
	}
	if updates == 0 {
		fmt.Fprintf(w, " %s", str.X())
		return nil
	}
	fmt.Fprintf(w, " %s", str.Y())
	return nil
}

// Textfile finds an appropriate text or NFO file and saves it to the database.
func (r *Record) Textfile(w io.Writer, s *scan.Stats) error {
	if s == nil {
		return ErrStatNil
	}
	if w == nil {
		w = io.Discard
	}
	const txt = ".txt"
	r.NFO = archive.NFO(r.Name, r.Files...)
	if r.NFO == "" {
		return nil
	}
	fmt.Fprintf(w, ", text file: %s", r.NFO)
	if _, err := os.Stat(r.File + txt); errors.Is(err, fs.ErrNotExist) {
		tmp, err1 := os.MkdirTemp(os.TempDir(), "zipcontent")
		if err1 != nil {
			return err1
		}
		defer os.RemoveAll(tmp)
		if err2 := archive.Extractor(r.Name, r.File, r.NFO, tmp); err2 != nil {
			return err2
		}
		src := filepath.Join(tmp, r.NFO)
		if _, err3 := archive.Move(src, filepath.Join(s.BasePath, r.UUID+txt)); err3 != nil {
			return err3
		}
		fmt.Fprint(w, ", extracted")
	}
	return nil
}

// Id prints the record ID to stdout.
func (r *Record) id(w io.Writer, s *scan.Stats) error {
	if s == nil {
		return ErrStatNil
	}
	if w == nil {
		w = io.Discard
	}
	logger.Printcrf(w, "%s %0*d. %v ",
		color.Question.Sprint("→"),
		len(strconv.Itoa(s.Total)),
		s.Count,
		color.Primary.Sprint(r.ID))
	return nil
}

// Save updates the record in the database.
func (r *Record) Save(db *sql.DB) (int64, error) {
	if db == nil {
		return 0, database.ErrDB
	}
	if r.ID == "" {
		return 0, ErrID
	}
	const (
		files = "UPDATE files SET filename=?,file_zip_content=?,updatedat=NOW(),updatedby=? WHERE id=?"
		nfo   = "UPDATE files SET filename=?,file_zip_content=?,updatedat=NOW(),updatedby=?," +
			"retrotxt_readme=?,retrotxt_no_readme=? WHERE id=?"
	)
	var err error
	stmt := &sql.Stmt{}
	if r.NFO == "" {
		stmt, err = db.Prepare(files)
	} else {
		stmt, err = db.Prepare(nfo)
	}
	if err != nil {
		return 0, fmt.Errorf("%s db prepare: %w", errPrefix, err)
	}
	defer stmt.Close()
	content := strings.Join(r.Files, "\n")
	var args []any
	switch r.NFO {
	case "":
		args = []any{r.Name, content, database.UpdateID, r.ID}
	default:
		args = []any{r.Name, content, database.UpdateID, r.NFO, 0, r.ID}
	}
	res, err := stmt.Exec(args...)
	if err != nil {
		return 0, fmt.Errorf("%s db exec: %w", errPrefix, err)
	}
	// RowsAffected & LastInsertId behaves differently depending on the database server.
	// On MariaDB DB, RowsAffected returns a 1 value for an invalid record primary key.
	// But LastInsertId returns an expected 0 value.
	if rows, err := res.RowsAffected(); err != nil {
		return 0, err
	} else if rows > 0 {
		return rows, nil
	}
	if id, err := res.LastInsertId(); err != nil {
		return 0, fmt.Errorf("%s db last insert: %w", errPrefix, err)
	} else if id == 0 {
		return 0, ErrID
	}
	return 1, nil
}
