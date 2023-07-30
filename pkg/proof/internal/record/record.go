// Package record handles the proof database record.
package record

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/df2/pkg/archive"
	"github.com/Defacto2/df2/pkg/conf"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/directories"
	"github.com/Defacto2/df2/pkg/logger"
	"github.com/Defacto2/df2/pkg/str"
	"github.com/gookit/color"
)

var (
	ErrColVals = errors.New("each database value must have a matching table column")
	ErrID      = errors.New("record id cannot be empty")
	ErrProof   = errors.New("proof structure cannot be empty")
	ErrRecord  = errors.New("record structure cannot be empty")
	ErrPointer = errors.New("pointer value cannot be nil")
)

// Proof data.
type Proof struct {
	Base      string          // Base is the relative path to file downloads which use UUID as filenames.
	BasePath  string          // BasePath to file downloads which use UUID as filenames.
	Columns   []string        // Column names.
	Count     int             // Count row index.
	Missing   int             // Missing UUID files count.
	Overwrite bool            // Overwrite flag (--overwrite) value.
	Total     int             // Total rows.
	Values    *[]sql.RawBytes // Values of the rows.
	start     time.Time       // processing time
}

func Init(cfg conf.Config) (Proof, error) {
	dir, err := directories.Init(cfg, false)
	if err != nil {
		return Proof{}, err
	}
	return Proof{
		Base:     logger.SPrintPath(dir.UUID),
		BasePath: dir.UUID,
		start:    time.Now(),
	}, nil
}

// Summary of the proofs.
func (p *Proof) Summary(id string) string {
	if p == nil {
		return ""
	}
	if id != "" && p.Total < 1 {
		return ""
	}
	w := &strings.Builder{}
	total := p.Count - p.Missing
	str.Total(w, total, "proofs handled")
	str.TimeTaken(w, time.Since(p.start).Seconds())
	return w.String()
}

// Record of a file item.
type Record struct {
	ID   string // ID is a database generated, auto increment identifier.
	UUID string // Universal unique ID.
	File string // File is an absolute path to the hosted file download.
	Name string // Name is the original filename of the download.
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
	if db == nil {
		return database.ErrDB
	}
	if r.ID == "" {
		return ErrID
	}
	if w == nil {
		w = io.Discard
	}
	const stmt = "UPDATE files SET updatedat=NOW(),updatedby=?,deletedat=NULL,deletedby=NULL WHERE id=?"
	prep, err := db.Prepare(stmt)
	if err != nil {
		return fmt.Errorf("approve prepare: %w", err)
	}
	defer prep.Close()
	res, err := prep.Exec(database.UpdateID, r.ID)
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
func (r Record) Iterate(db *sql.DB, w io.Writer, cfg conf.Config, p Proof) error { //nolint:cyclop
	if db == nil {
		return database.ErrDB
	}
	if r == (Record{}) {
		return ErrRecord
	}
	if reflect.DeepEqual(p, Proof{}) {
		return ErrProof
	}
	if p.Values == nil {
		return nil
	}
	if len(*p.Values) != len(p.Columns) {
		return ErrColVals
	}
	for i, raw := range *p.Values {
		s := database.Val(raw)
		switch p.Columns[i] {
		case "id":
			if err := r.Prefix(w, &p); err != nil {
				return err
			}
		case "createdat":
			s, err := database.DateTime(raw)
			if err != nil {
				return err
			}
			fmt.Fprintf(w, "%v", s)
		case "filename":
			fmt.Fprintf(w, "%v", s)
		case "file_zip_content":
			if err := r.Zip(db, w, cfg, p.Overwrite); err != nil {
				return err
			}
		default:
		}
	}
	return nil
}

// UpdateZipContent sets the file_zip_content column to match content and platform to "image".
func UpdateZipContent(db *sql.DB, w io.Writer, id, filename, content string, items int) error {
	if db == nil {
		return database.ErrDB
	}
	if id == "" {
		return ErrID
	}
	if w == nil {
		w = io.Discard
	}
	const stmt = "UPDATE files SET filename=?,file_zip_content=?,updatedat=NOW(),updatedby=?,platform=? WHERE id=?"
	prep, err := db.Prepare(stmt)
	if err != nil {
		return fmt.Errorf("updatezip prepare: %w", err)
	}
	defer prep.Close()
	if _, err := prep.Exec(filename, content, database.UpdateID, "image", id); err != nil {
		return fmt.Errorf("updatezip exec: %w", err)
	}
	fmt.Fprintf(w, "%d items", items)
	return nil
}

// Prefix prints the stat count and record Id to stdout.
func (r Record) Prefix(w io.Writer, s *Proof) error {
	if s == nil {
		return fmt.Errorf("stat proof %w", ErrPointer)
	}
	if w == nil {
		w = io.Discard
	}
	logger.PrintfCR(w, "->%s %0*d. %v ",
		color.Question.Sprint("→"),
		len(strconv.Itoa(s.Total)),
		s.Count,
		color.Primary.Sprint(r.ID))
	return nil
}

// Zip reads an archive and saves the content to the database.
func (r Record) Zip(db *sql.DB, w io.Writer, cfg conf.Config, overwrite bool) error {
	if db == nil {
		return database.ErrDB
	}
	if r == (Record{}) {
		return ErrRecord
	}
	if w == nil {
		w = io.Discard
	}
	if !overwrite {
		return nil
	}
	u, err := r.fileZipContent(db, w)
	if err != nil {
		return fmt.Errorf("zip content: %w", err)
	}
	if !u {
		return nil
	}
	fmt.Fprint(w, " • ")
	proof := archive.Proof{
		Source: r.File,
		Name:   r.Name,
		UUID:   r.UUID,
		Config: cfg,
	}
	if err := proof.Decompress(w); err != nil {
		return fmt.Errorf("zip proof: %w", err)
	}
	if err := r.Approve(db, w); err != nil {
		return fmt.Errorf("zip approve: %w", err)
	}
	return nil
}

func (r Record) fileZipContent(db *sql.DB, w io.Writer) (bool, error) {
	if db == nil {
		return false, database.ErrDB
	}
	if r == (Record{}) {
		return false, ErrRecord
	}
	if w == nil {
		w = io.Discard
	}
	list, name, err := archive.Read(w, r.File, r.Name)
	if err != nil {
		return false, err
	}
	if err := UpdateZipContent(db, w, r.ID, name, strings.Join(list, "\n"), len(list)); err != nil {
		return false, err
	}
	return true, nil
}

// Skip checks if the file of the proof exists.
func Skip(w io.Writer, s Proof, r Record) (bool, error) {
	if reflect.DeepEqual(s, Proof{}) {
		return false, ErrProof
	}
	if r == (Record{}) {
		return false, ErrRecord
	}
	if w == nil {
		w = io.Discard
	}
	if _, err := os.Stat(r.File); errors.Is(err, fs.ErrNotExist) {
		fmt.Fprintf(w, "%s %0*d. %v is missing %v %s\n",
			color.Question.Sprint("→"),
			len(strconv.Itoa(s.Total)),
			s.Count,
			color.Primary.Sprint(r.ID),
			filepath.Join(s.Base, color.Danger.Sprint(r.UUID)),
			str.X())
		return true, nil
	} else if err != nil {
		return false, fmt.Errorf("file skip stat %q: %w", r.File, err)
	}
	return false, nil
}
