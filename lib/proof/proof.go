package proof

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
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
	Overwrite bool // overwrite existing files
	All       bool // parse all proofs
	HideMiss  bool // ignore missing uuid files
}

type row struct {
	base    string
	count   int
	missing int
}

var (
	proofID string // ID used for proofs, either a UUID or ID string
)

// Query parses a single proof.
func (req Request) Query(id string) error {
	if !database.UUID(id) && !database.ID(id) {
		return fmt.Errorf("invalid id given %q it needs to be an auto-generated MySQL id or an uuid", id)
	}
	proofID = id
	return req.Queries()
}

// Queries parses all new proofs.
// ow will overwrite any existing proof assets such as images.
// all parses every proof not just records waiting for approval.
func (req Request) Queries() error {
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
		missing: 0}
	for rows.Next() {
		err = rows.Scan(scanArgs...)
		logs.Check(err)
		if new := recordNew(values); !new && !req.All {
			continue
		}
		rw.count++
		r := Record{ID: string(values[0]), UUID: string(values[1]), Name: string(values[4])}
		r.File = filepath.Join(dir.UUID, r.UUID)
		// ping file
		if rw.skip(r, req.HideMiss) {
			continue
		}
		// iterate through each value
		var value string
		for i, col := range values {
			value = val(col)
			switch columns[i] {
			case "id":
				fmt.Printf("%s item %04d (%v) ", logs.Y(), rw.count, value) // rw.count has 3 leading zeros
			case "uuid":
				fmt.Printf("%v ", value)
			case "createdat":
				clock(value)
			case "filename":
				fmt.Printf("%v", value)
			case "file_zip_content":
				r.zip(col, req.Overwrite)
			default:
				//fmt.Printf("  %v: %v\n", columns[i], value)
			}
		}
		println()
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
	print(fmt.Sprintf("  %s approved", logs.Y()))
	return nil
}

func clock(value string) {
	t, err := time.Parse("2006-01-02T15:04:05Z", value)
	logs.Check(err)
	if t.UTC().Format("01 2006") != time.Now().Format("01 2006") {
		fmt.Printf("%v ", color.Info.Sprint(t.UTC().Format("2 Jan 2006")))
	} else {
		fmt.Printf("%v ", color.Info.Sprint(t.UTC().Format("2 Jan 15:04")))
	}
}

// fileZipContent reads an archive and saves its content to the database
func (r Record) fileZipContent() bool {
	a, err := archive.Read(r.File)
	if err != nil {
		logs.Log(err)
		return false
	}
	updateZipContent(r.ID, strings.Join(a, "\n"))
	return true
}

// recordNew reports if a proof is set to unapproved
func recordNew(values []sql.RawBytes) bool {
	if values[2] == nil || string(values[2]) != string(values[3]) {
		return false
	}
	return true
}

func (rw row) skip(r Record, hide bool) bool {
	if _, err := os.Stat(r.File); os.IsNotExist(err) {
		rw.missing++
		if !hide {
			fmt.Printf("%s item %v (%v) missing %v\n", logs.X(), rw.count, r.ID, filepath.Join(rw.base, color.Danger.Sprint(r.UUID)))
		}
		return true
	}
	return false
}

func sqlSelect() string {
	s := "SELECT `id`,`uuid`,`deletedat`,`createdat`,`filename`,`file_zip_content`,`updatedat`,`platform`"
	w := "WHERE `section` = 'releaseproof'"
	if proofID != "" {
		switch {
		case database.UUID(proofID):
			w = fmt.Sprintf("%v AND `uuid`=%q", w, proofID)
		case database.ID(proofID):
			w = fmt.Sprintf("%v AND `id`=%q", w, proofID)
		}
	}
	return s + "FROM `files`" + w
}

func (rw row) summary() {
	t := fmt.Sprintf("Total proofs handled: %v", rw.count)
	fmt.Println(strings.Repeat("─", len(t)))
	fmt.Println(t)
	if rw.missing > 0 {
		fmt.Println("UUID files not found:", rw.missing)
	}
}

// updateZipContent sets the file_zip_content column to match content and platform to "image".
func updateZipContent(id string, content string) {
	db := database.Connect()
	defer db.Close()
	update, err := db.Prepare("UPDATE files SET file_zip_content=?,updatedat=NOW(),updatedby=?,platform=? WHERE id=?")
	logs.Check(err)
	_, err = update.Exec(content, database.UpdateID, "image", id)
	logs.Check(err)
	fmt.Printf("%s file_zip_content", logs.Y())
}

func val(col sql.RawBytes) string {
	if col == nil {
		return "NULL"
	}
	return string(col)
}

func (r Record) zip(col sql.RawBytes, overwrite bool) {
	if col == nil || overwrite {
		fmt.Print("\n   • ")
		if u := r.fileZipContent(); !u {
			return
		}
		if err := archive.Extract(r.File, r.UUID); err != nil {
			logs.Log(err)
		} else {
			err = r.approve()
			logs.Log(err)
		}
	}
}
