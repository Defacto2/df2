package proof

// os.Exit() = 9x

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/df2/lib/archive"
	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/directories"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/gookit/color"
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
	Overwrite   bool // overwrite existing files
	AllProofs   bool // parse all proofs
	HideMissing bool // ignore missing uuid files
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

var proofID string // ID used for proofs, either a UUID or ID string

// Query parses a single proof.
func (request Request) Query(id string) error {
	if err := database.CheckID(id); err != nil {
		return err
	}
	proofID = id
	return request.Queries()
}

func proofChk(text string) {
	if proofID == "" {
		return
	}
	logs.Println(text)
}

func statInit() stat {
	dir := directories.Init(false)
	return stat{
		base:     logs.Path(dir.UUID),
		basePath: dir.UUID,
		count:    0,
		missing:  0,
		start:    time.Now(),
		total:    0}
}

// Queries parses all new proofs.
// ow will overwrite any existing proof assets such as images.
// all parses every proof not just records waiting for approval.
func (request Request) Queries() error {
	s := statInit()
	db := database.Connect()
	defer db.Close()
	rows, err := db.Query(sqlSelect())
	if err != nil {
		return err
	}
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
	if s.total < 1 {
		proofChk(fmt.Sprintf("file record id '%s' does not exist", proofID))
	} else if s.total > 1 {
		logs.Println("Total records", s.total)
	}
	rows, err = db.Query(sqlSelect())
	for rows.Next() {
		err = rows.Scan(scanArgs...)
		logs.Check(err)
		if request.flagSkip(values) {
			continue
		}
		s.count++
		var r = new(values, s.basePath)
		if s.fileSkip(r, request.HideMissing) {
			continue
		}
		s.columns = columns
		s.overwrite = request.Overwrite
		s.values = &values
		r.iterate(s)
	}
	logs.Check(rows.Err())
	s.summary()
	return nil
}

// iterate through each value.
func (r Record) iterate(s stat) {
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
			r.zip(raw, s)
		default:
		}
	}
}

func new(values []sql.RawBytes, path string) Record {
	var r Record
	r.ID = string(values[0])
	r.UUID = string(values[1])
	r.Name = string(values[4])
	r.File = filepath.Join(path, r.UUID)
	return r
}

func (r Record) printID(s stat) {
	logs.Printfcr("%s %0*d. %v ",
		color.Question.Sprint("→"),
		len(strconv.Itoa(s.total)),
		s.count,
		color.Primary.Sprint(r.ID))
}

// flagSkip uses argument flags to check if a record is to be ignored.
func (request Request) flagSkip(values []sql.RawBytes) bool {
	if proofID != "" && request.Overwrite {
		return false
	} else if new := database.IsNew(values); !new && !request.AllProofs {
		proofChk(fmt.Sprintf("skip file record id '%s' as it is not new", proofID))
		return true
	}
	return false
}

// approve sets the record to be publically viewable.
func (r Record) approve() error {
	db := database.Connect()
	defer db.Close()
	update, err := db.Prepare("UPDATE files SET updatedat=NOW(),updatedby=?,deletedat=NULL,deletedby=NULL WHERE id=?")
	logs.Check(err)
	_, err = update.Exec(database.UpdateID, r.ID)
	logs.Check(err)
	logs.Printf(" %s", logs.Y())
	return nil
}

// fileZipContent reads an archive and saves its content to the database.
func (r Record) fileZipContent() bool {
	a, err := archive.Read(r.File, r.Name)
	if err != nil {
		logs.Log(err)
		return false
	}
	updateZipContent(r.ID, len(a), strings.Join(a, "\n"))
	return true
}

// fileSkip checks if the file of the proof exists.
func (s *stat) fileSkip(r Record, hide bool) bool {
	if _, err := os.Stat(r.File); os.IsNotExist(err) {
		s.missing++
		if !hide {
			fmt.Printf("%s %0*d. %v is missing %v %s\n",
				color.Question.Sprint("→"),
				len(strconv.Itoa(s.total)),
				s.count,
				color.Primary.Sprint(r.ID),
				filepath.Join(s.base, color.Danger.Sprint(r.UUID)),
				logs.X())
		}
		return true
	}
	return false
}

func sqlSelect() string {
	s := "SELECT `id`,`uuid`,`deletedat`,`createdat`,`filename`,`file_zip_content`,`updatedat`,`platform`"
	w := " WHERE `section` = 'releaseproof'"
	if proofID != "" {
		switch {
		case database.IsUUID(proofID):
			w = fmt.Sprintf("%v AND `uuid`=%q", w, proofID)
		case database.IsID(proofID):
			w = fmt.Sprintf("%v AND `id`=%q", w, proofID)
		}
	}
	return s + " FROM `files`" + w
}

func (s stat) summary() {
	if proofID != "" && s.total < 1 {
		return
	}
	total := s.count - s.missing
	if total == 0 {
		fmt.Print("nothing to do")
	}
	elapsed := time.Since(s.start)
	t := fmt.Sprintf("Total proofs handled: %v, time elapsed %s", total, elapsed)
	logs.Printf("\n%s\n%s\n", strings.Repeat("─", len(t)), t)
}

// updateZipContent sets the file_zip_content column to match content and platform to "image".
func updateZipContent(id string, items int, content string) {
	db := database.Connect()
	defer db.Close()
	update, err := db.Prepare("UPDATE files SET file_zip_content=?,updatedat=NOW(),updatedby=?,platform=? WHERE id=?")
	logs.Check(err)
	_, err = update.Exec(content, database.UpdateID, "image", id)
	logs.Check(err)
	logs.Printf("%d items", items)
}

func val(col sql.RawBytes) string {
	if col == nil {
		return "NULL"
	}
	return string(col)
}

func (r Record) zip(col sql.RawBytes, s stat) {
	if col == nil || s.overwrite {
		logs.Print(" • ")
		if u := r.fileZipContent(); !u {
			return
		}
		if err := archive.Extract(r.File, r.Name, r.UUID); err != nil {
			logs.Log(err)
		} else {
			err = r.approve()
			logs.Log(err)
		}
	}
}
