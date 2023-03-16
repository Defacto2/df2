package record

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/Defacto2/df2/pkg/archive"
	"github.com/Defacto2/df2/pkg/configger"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/logger"
	"github.com/Defacto2/df2/pkg/proof/internal/stat"
	"github.com/Defacto2/df2/pkg/str"
	"github.com/gookit/color"
	"go.uber.org/zap"
)

var (
	ErrNoRec   = errors.New("record cannot be empty")
	ErrNoStat  = errors.New("stat cannot be empty")
	ErrColVals = errors.New("each database value must have a matching table column")
)

// Record of a file item.
type Record struct {
	ID   string // MySQL auto increment Id.
	UUID string // Unique Id.
	File string // Absolute path to file.
	Name string // Filename.
}

// New returns a file record using values from the database.
func New(values []sql.RawBytes, path string) Record {
	const id, uuid, filename = 0, 1, 4
	if len(values) <= filename {
		return Record{}
	}
	return Record{
		ID:   string(values[id]),
		UUID: string(values[uuid]),
		Name: string(values[filename]),
		File: filepath.Join(path, string(values[uuid])),
	}
}

// Approve sets the record to be publically viewable.
func (r Record) Approve(db *sql.DB, w io.Writer) error {
	if reflect.DeepEqual(r, Record{}) {
		return ErrNoRec
	}
	update, err := db.Prepare("UPDATE files SET updatedat=NOW(),updatedby=?,deletedat=NULL,deletedby=NULL WHERE id=?")
	if err != nil {
		return fmt.Errorf("approve prepare: %w", err)
	}
	defer update.Close()
	res, err := update.Exec(database.UpdateID, r.ID)
	if err != nil {
		return fmt.Errorf("approve exec: %w", err)
	}
	if i, err := res.RowsAffected(); err != nil {
		return fmt.Errorf("approve rows affected: %w", err)
	} else if i == 0 {
		fmt.Fprintf(w, " %s", str.X())
		return nil
	}
	fmt.Fprintf(w, " %s", str.Y())
	return nil
}

// Iterate through each stat value.
func (r Record) Iterate(db *sql.DB, w io.Writer, l *zap.SugaredLogger, cfg configger.Config, s stat.Proof) error { //nolint:cyclop
	if reflect.DeepEqual(r, Record{}) {
		return ErrNoRec
	}
	if reflect.DeepEqual(s, stat.Proof{}) {
		return ErrNoStat
	}
	if s.Values == nil {
		return nil
	}
	if len(*s.Values) != len(s.Columns) {
		return ErrColVals
	}
	var value string
	for i, raw := range *s.Values {
		value = database.Val(raw)
		switch s.Columns[i] {
		case "id":
			r.Prefix(w, &s)
		case "createdat":
			database.DateTime(l, raw)
		case "filename":
			fmt.Fprintf(w, "%v", value)
		case "file_zip_content":
			if err := r.Zip(db, w, l, cfg, raw, &s); err != nil {
				return err
			}
		default:
		}
	}
	return nil
}

// UpdateZipContent sets the file_zip_content column to match content and platform to "image".
func UpdateZipContent(db *sql.DB, w io.Writer, id, filename, content string, items int) error {
	const q = "UPDATE files SET filename=?,file_zip_content=?,updatedat=NOW(),updatedby=?,platform=? WHERE id=?"
	update, err := db.Prepare(q)
	if err != nil {
		return fmt.Errorf("updatezip prepare: %w", err)
	}
	defer update.Close()
	if _, err := update.Exec(filename, content, database.UpdateID, "image", id); err != nil {
		return fmt.Errorf("updatezip exec: %w", err)
	}
	fmt.Fprintf(w, "%d items", items)
	return nil
}

// Prefix prints the stat count and record Id to stdout.
func (r Record) Prefix(w io.Writer, s *stat.Proof) {
	logger.Printcrf(w, "%s %0*d. %v ",
		color.Question.Sprint("→"),
		len(strconv.Itoa(s.Total)),
		s.Count,
		color.Primary.Sprint(r.ID))
}

// Zip reads an archive and saves the content to the database.
func (r Record) Zip(db *sql.DB, w io.Writer, l *zap.SugaredLogger, cfg configger.Config, col sql.RawBytes, s *stat.Proof) error {
	if reflect.DeepEqual(r, Record{}) {
		return ErrNoRec
	}
	if reflect.DeepEqual(s, stat.Proof{}) {
		return ErrNoStat
	}
	if col != nil && !s.Overwrite {
		return nil
	}
	fmt.Fprint(w, " • ")
	if u, err := r.fileZipContent(db, w); !u {
		return nil
	} else if err != nil {
		return fmt.Errorf("zip content: %w", err)
	}
	if err := archive.Proof(w, l, cfg, r.File, r.Name, r.UUID); err != nil {
		return fmt.Errorf("zip proof: %w", err)
	} else if err := r.Approve(db, w); err != nil {
		return fmt.Errorf("zip approve: %w", err)
	}
	return nil
}

func (r Record) fileZipContent(db *sql.DB, w io.Writer) (bool, error) {
	if reflect.DeepEqual(r, Record{}) {
		return false, ErrNoRec
	}
	a, fn, err := archive.Read(w, r.File, r.Name)
	if err != nil {
		return false, err
	}
	if err := UpdateZipContent(db, w, r.ID, fn, strings.Join(a, "\n"), len(a)); err != nil {
		return false, err
	}
	return true, nil
}

// Skip checks if the file of the proof exists.
func Skip(w io.Writer, s stat.Proof, r Record, hide bool) (bool, error) {
	if reflect.DeepEqual(r, Record{}) {
		return false, ErrNoRec
	}
	if reflect.DeepEqual(s, stat.Proof{}) {
		return false, ErrNoStat
	}
	if _, err := os.Stat(r.File); os.IsNotExist(err) {
		s.Missing++
		if !hide {
			fmt.Fprintf(w, "%s %0*d. %v is missing %v %s\n",
				color.Question.Sprint("→"),
				len(strconv.Itoa(s.Total)),
				s.Count,
				color.Primary.Sprint(r.ID),
				filepath.Join(s.Base, color.Danger.Sprint(r.UUID)),
				str.X())
		}
		return true, nil
	} else if err != nil {
		return false, fmt.Errorf("file skip stat %q: %w", r.File, err)
	}
	return false, nil
}
