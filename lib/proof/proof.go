// Package proof handles files that have the section tagged as releaseproof.
package proof

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gookit/color" //nolint:misspell

	"github.com/Defacto2/df2/lib/archive"
	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/directories"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/Defacto2/df2/lib/str"
)

// Record of a file item.
type Record struct {
	ID   string // mysql auto increment id
	UUID string // record unique id
	File string // absolute path to file
	Name string // filename
}

// Request proofs.
type Request struct {
	Overwrite   bool   // overwrite existing files
	AllProofs   bool   // parse all proofs
	HideMissing bool   // ignore missing uuid files
	byID        string // id used for proofs, either a uuid or id string
}

type stat struct {
	base      string          // formatted path to file downloads with UUID as filenames
	basePath  string          // path to file downloads with UUID as filenames
	columns   []string        // column names
	count     int             // row index
	missing   int             // missing UUID files count
	overwrite bool            // --overwrite flag value
	start     time.Time       //
	total     int             // total rows
	values    *[]sql.RawBytes // row values
}

// Query parses a single proof.
func (request *Request) Query(id string) error {
	if err := database.CheckID(id); err != nil {
		return fmt.Errorf("request query id %q: %w", id, err)
	}
	request.byID = id
	if err := request.Queries(); err != nil {
		return fmt.Errorf("request queries: %w", err)
	}
	return nil
}

func statInit() stat {
	dir := directories.Init(false)
	return stat{base: logs.Path(dir.UUID), basePath: dir.UUID, start: time.Now()}
}

// Queries parses all new proofs.
// ow will overwrite any existing proof assets such as images.
// all parses every proof not just records waiting for approval.
func (request Request) Queries() error {
	s := statInit()
	db := database.Connect()
	defer db.Close()
	rows, err := db.Query(sqlSelect(request.byID))
	if err != nil {
		return err
	} else if rows.Err() != nil {
		return rows.Err()
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	values := make([]sql.RawBytes, len(columns))
	// more information: https://github.com/go-sql-driver/mysql/wiki/Examples
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	for rows.Next() {
		s.total++
	}
	if s.total < 1 && request.byID != "" {
		logs.Printf("file record id '%s' does not exist or is not a release proof\n", request.byID)
	} else if s.total > 1 {
		logs.Println("Total records", s.total)
	}
	rows, err = db.Query(sqlSelect(request.byID))
	if err != nil {
		return err
	} else if rows.Err() != nil {
		return rows.Err()
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(scanArgs...); err != nil {
			return err
		}
		if request.flagSkip(values) {
			continue
		}
		s.count++
		r := newRec(values, s.basePath)
		if skip, err := s.fileSkip(r, request.HideMissing); skip {
			continue
		} else if err != nil {
			return err
		}
		s.columns = columns
		s.overwrite = request.Overwrite
		s.values = &values
		if err := r.iterate(&s); err != nil {
			return err
		}
	}
	s.summary(request.byID)
	return nil
}

// iterate through each value.
func (r Record) iterate(s *stat) error {
	var value string
	for i, raw := range *s.values {
		value = val(raw)
		switch s.columns[i] {
		case "id":
			r.printID(s)
		case "createdat":
			database.DateTime(raw)
		case "filename":
			logs.Printf("%v", value)
		case "file_zip_content":
			if err := r.zip(raw, s); err != nil {
				return err
			}
		default:
		}
	}
	return nil
}

func newRec(values []sql.RawBytes, path string) Record {
	var r Record
	r.ID = string(values[0])
	r.UUID = string(values[1])
	r.Name = string(values[4])
	r.File = filepath.Join(path, r.UUID)
	return r
}

func (r Record) printID(s *stat) {
	logs.Printcrf("%s %0*d. %v ",
		color.Question.Sprint("→"),
		len(strconv.Itoa(s.total)),
		s.count,
		color.Primary.Sprint(r.ID))
}

// flagSkip uses argument flags to check if a record is to be ignored.
func (request Request) flagSkip(values []sql.RawBytes) (skip bool) {
	if request.byID != "" && request.Overwrite {
		return false
	}
	if n := database.NewProof(values); !n && !request.AllProofs {
		if request.byID != "" {
			logs.Printf("skip file record id '%s' as it is not new", request.byID)
		}
		return true
	}
	return false
}

// approve sets the record to be publically viewable.
func (r Record) approve() error {
	db := database.Connect()
	defer db.Close()
	update, err := db.Prepare("UPDATE files SET updatedat=NOW(),updatedby=?,deletedat=NULL,deletedby=NULL WHERE id=?")
	if err != nil {
		return fmt.Errorf("approve update prepare: %w", err)
	}
	defer update.Close()
	_, err = update.Exec(database.UpdateID, r.ID)
	if err != nil {
		return fmt.Errorf("approve update exec: %w", err)
	}
	logs.Printf(" %s", str.Y())
	return nil
}

// fileZipContent reads an archive and saves its content to the database.
func (r Record) fileZipContent() (ok bool, err error) {
	a, err := archive.Read(r.File, r.Name)
	if err != nil {
		return false, fmt.Errorf("file zip content archive read: %w", err)
	}
	if err := updateZipContent(r.ID, len(a), strings.Join(a, "\n")); err != nil {
		return false, fmt.Errorf("file zip content update: %w", err)
	}
	return true, nil
}

// fileSkip checks if the file of the proof exists.
func (s *stat) fileSkip(r Record, hide bool) (skip bool, err error) {
	if _, err := os.Stat(r.File); os.IsNotExist(err) {
		s.missing++
		if !hide {
			fmt.Printf("%s %0*d. %v is missing %v %s\n",
				color.Question.Sprint("→"),
				len(strconv.Itoa(s.total)),
				s.count,
				color.Primary.Sprint(r.ID),
				filepath.Join(s.base, color.Danger.Sprint(r.UUID)),
				str.X())
		}
		return true, nil
	} else if err != nil {
		return false, fmt.Errorf("file skip stat %q: %w", r.File, err)
	}
	return false, nil
}

func sqlSelect(id string) string {
	s := "SELECT `id`,`uuid`,`deletedat`,`createdat`,`filename`,`file_zip_content`,`updatedat`,`platform`"
	w := " WHERE `section` = 'releaseproof'"
	if id != "" {
		switch {
		case database.IsUUID(id):
			w = fmt.Sprintf("%v AND `uuid`=%q", w, id)
		case database.IsID(id):
			w = fmt.Sprintf("%v AND `id`=%q", w, id)
		}
	}
	return s + " FROM `files`" + w
}

func (s *stat) summary(id string) {
	if id != "" && s.total < 1 {
		return
	}
	total := s.count - s.missing
	if total == 0 {
		fmt.Print("nothing to do")
	}
	elapsed := time.Since(s.start).Seconds()
	t := fmt.Sprintf("Total proofs handled: %v, time elapsed %.1f seconds", total, elapsed)
	logs.Printf("\n%s\n%s\n", strings.Repeat("─", len(t)), t)
}

// updateZipContent sets the file_zip_content column to match content and platform to "image".
func updateZipContent(id string, items int, content string) error {
	db := database.Connect()
	defer db.Close()
	update, err := db.Prepare("UPDATE files SET file_zip_content=?,updatedat=NOW(),updatedby=?,platform=? WHERE id=?")
	if err != nil {
		return fmt.Errorf("update zip content db prepare: %w", err)
	}
	defer update.Close()
	if _, err := update.Exec(content, database.UpdateID, "image", id); err != nil {
		return fmt.Errorf("update zip content update exec: %w", err)
	}
	logs.Printf("%d items", items)
	return db.Close()
}

func val(col sql.RawBytes) string {
	if col == nil {
		return "NULL"
	}
	return string(col)
}

func (r Record) zip(col sql.RawBytes, s *stat) error {
	if col == nil || s.overwrite {
		logs.Print(" • ")
		if u, err := r.fileZipContent(); !u {
			return nil
		} else if err != nil {
			return fmt.Errorf("zip content: %w", err)
		}
		if err := archive.Extract(r.File, r.Name, r.UUID); err != nil {
			return fmt.Errorf("zip extract: %w", err)
		} else if err := r.approve(); err != nil {
			return fmt.Errorf("zip approve: %w", err)
		}
	}
	return nil
}
