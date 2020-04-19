package proof

// os.Exit() = 9x

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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

type row struct {
	base    string
	count   int
	missing int
	total   int
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
	println(text)
}

// Queries parses all new proofs.
// ow will overwrite any existing proof assets such as images.
// all parses every proof not just records waiting for approval.
func (request Request) Queries() error {
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
	dir := directories.Init(false)
	// fetch the rows
	rw := row{
		base:    logs.Path(dir.UUID),
		count:   0,
		missing: 0,
		total:   0}
	for rows.Next() {
		rw.total++
	}
	if rw.total < 1 {
		proofChk(fmt.Sprintf("file record id '%s' does not exist", proofID))
	} else if !logs.Quiet && rw.total > 1 {
		println("Total records", rw.total)
	}
	rows, err = db.Query(sqlSelect())
	for rows.Next() {
		err = rows.Scan(scanArgs...)
		logs.Check(err)
		if proofID != "" && request.Overwrite {
			// skip IsNew check
		} else if new := database.IsNew(values); !new && !request.AllProofs {
			proofChk(fmt.Sprintf("skip file record id '%s' as it is not new", proofID))
			continue
		}
		rw.count++
		r := Record{ID: string(values[0]), UUID: string(values[1]), Name: string(values[4])}
		r.File = filepath.Join(dir.UUID, r.UUID)
		// ping file
		if rw.skip(r, request.HideMissing) {
			continue
		}
		// iterate through each value
		var value string
		for i, raw := range values {
			value = val(raw)
			switch columns[i] {
			case "id":
				logs.Printfcr("%s %0*d. %v ", color.Question.Sprint("→"), len(strconv.Itoa(rw.total)), rw.count, color.Primary.Sprint(r.ID))
			case "createdat":
				database.DateTime(raw)
			case "filename":
				logs.Printf("%v", value)
			case "file_zip_content":
				r.zip(raw, request.Overwrite)
			default:
			}
		}
	}
	logs.Check(rows.Err())
	rw.summary()
	return nil
}

// approve sets the record to be publically viewable.
func (r Record) approve() error {
	db := database.Connect()
	defer db.Close()
	update, err := db.Prepare("UPDATE files SET updatedat=NOW(),updatedby=?,deletedat=NULL,deletedby=NULL WHERE id=?")
	logs.Check(err)
	_, err = update.Exec(database.UpdateID, r.ID)
	logs.Check(err)
	print(fmt.Sprintf(" %s", logs.Y()))
	return nil
}

// fileZipContent reads an archive and saves its content to the database
func (r Record) fileZipContent() bool {
	a, err := archive.Read(r.File, r.Name)
	if err != nil {
		logs.Log(err)
		return false
	}
	updateZipContent(r.ID, len(a), strings.Join(a, "\n"))
	return true
}

func (rw *row) skip(r Record, hide bool) bool {
	if _, err := os.Stat(r.File); os.IsNotExist(err) {
		rw.missing++
		if !hide {
			fmt.Printf("%s %0*d. %v is missing %v %s\n", color.Question.Sprint("→"), len(strconv.Itoa(rw.total)), rw.count, color.Primary.Sprint(r.ID), filepath.Join(rw.base, color.Danger.Sprint(r.UUID)), logs.X())
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

func (rw row) summary() {
	if proofID != "" && rw.total < 1 {
		return
	}
	t := fmt.Sprintf("Total proofs handled: %v", rw.count-rw.missing)
	logs.Println(strings.Repeat("─", len(t)))
	logs.Println(t)
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

func (r Record) zip(col sql.RawBytes, overwrite bool) {
	if col == nil || overwrite {
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
